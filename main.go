package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/wkqco33/pc_cleaner/internal/ai"
	"github.com/wkqco33/pc_cleaner/internal/cleaner"
	"github.com/wkqco33/pc_cleaner/internal/scanner"
	"github.com/wkqco33/pc_cleaner/internal/ui"
	"github.com/wkqco33/wcli"
	"github.com/wkqco33/wcli/rich"
)

var version = "0.1.0"

type aiAnalyzer interface {
	Analyze(ctx context.Context, results []scanner.ScanResult, prompt string) (*ai.AnalysisResult, error)
}

// app holds the I/O sinks for the cleaner CLI so it can be exercised with
// injected writers in tests (mirrors rich/wcli's writer injection).
type app struct {
	out        io.Writer
	in         io.Reader
	aiAnalyzer aiAnalyzer
}

func newApp(out io.Writer, in io.Reader) *app {
	return &app{out: out, in: in}
}

func main() {
	a := newApp(os.Stdout, os.Stdin)
	if err := a.rootCommand().Execute(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

// rootCommand builds the wcli command tree. Version auto-registers --version,
// and NewCompletionCommand adds a shell-completion subcommand.
func (a *app) rootCommand() *wcli.Command {
	var dryRun bool
	var skipList string

	cmd := &wcli.Command{
		Use:     "pcc",
		Short:   "불필요한 파일 및 캐시를 정리해 디스크 공간을 확보하는 CLI 도구",
		Long:    "macOS / Windows / Linux에서 캐시 및 임시 파일을 정리해 디스크 공간을 확보합니다.",
		Version: version,
		Run: func(ctx *wcli.Context) error {
			return a.clean(dryRun, skipList)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", "", false, "실제 삭제 없이 분석만 실행")
	cmd.Flags().StringVar(&skipList, "skip", "", "", "건너뛸 항목 (쉼표 구분, 예: gradle,pip,docker)")

	cmd.AddCommand(a.aiCommand())
	cmd.AddCommand(wcli.NewCompletionCommand(cmd))
	return cmd
}

// clean runs the full scan → review → confirm → clean pipeline.
func (a *app) clean(dryRun bool, skipList string) error {
	rich.Fprintln(a.out, "[bold][cyan]PC Cleaner v%s — %s[/cyan][/bold]", version, runtime.GOOS)

	items := scanner.GetItems()
	items = filterItems(items, skipList)

	rich.Fprintln(a.out, "  [cyan]ℹ[/cyan] 총 %d개 항목 스캔 중...", len(items))
	fmt.Fprintln(a.out)

	results := scanner.Scan(items)
	a.printScanTable(results)

	cleanable := filterCleanable(results)
	if len(cleanable) == 0 {
		a.printNoCleanable()
		return nil
	}

	totalSize := totalScanSize(cleanable)
	diskInfo, diskErr := scanner.GetDiskUsage("")

	if dryRun {
		if diskErr == nil {
			fmt.Fprintf(a.out, "  디스크 공간: 전체 %s | 사용 가능 %s (정리 후 예상: %s)\n",
				ui.FormatBytes(int64(diskInfo.Total)),
				ui.FormatBytes(int64(diskInfo.Available)),
				ui.FormatBytes(int64(diskInfo.Available+uint64(totalSize))),
			)
		}
		rich.Fprintln(a.out, "  [yellow]⚠ [DRY-RUN] 정리 가능 용량: [bold]%s[/bold][/yellow]",
			ui.FormatBytes(totalSize))
		cleaner.Clean(cleanable, true)
		return nil
	}

	if diskErr == nil {
		fmt.Fprintf(a.out, "  디스크 공간: 전체 %s | 사용 가능 %s (정리 후 예상: %s)\n",
			ui.FormatBytes(int64(diskInfo.Total)),
			ui.FormatBytes(int64(diskInfo.Available)),
			ui.FormatBytes(int64(diskInfo.Available+uint64(totalSize))),
		)
	}
	rich.Fprintln(a.out, "  [bold]정리 가능 용량: %s[/bold]", ui.FormatBytes(totalSize))

	ok, err := rich.FConfirm(a.out, a.in, "삭제를 진행하시겠습니까?", false)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(a.out)
		rich.Fprintln(a.out, "  [yellow]⚠ 취소되었습니다.[/yellow]")
		return nil
	}

	fmt.Fprintln(a.out)
	cleaned := cleaner.Clean(cleanable, false)
	a.printReport(cleaned)
	return nil
}

func (a *app) printNoCleanable() {
	if diskInfo, err := scanner.GetDiskUsage(""); err == nil {
		fmt.Fprintf(a.out, "  디스크 공간: 전체 %s | 사용 가능 %s\n\n",
			ui.FormatBytes(int64(diskInfo.Total)),
			ui.FormatBytes(int64(diskInfo.Available)),
		)
	}
	rich.Fprintln(a.out, "  [green]✓[/green] 정리할 항목이 없습니다.")
}

// filterItems removes items whose name contains any of the skip keywords.
func filterItems(items []scanner.CacheItem, skipStr string) []scanner.CacheItem {
	if skipStr == "" {
		return items
	}
	skips := strings.Split(strings.ToLower(skipStr), ",")
	var filtered []scanner.CacheItem
	for _, item := range items {
		name := strings.ToLower(item.Name)
		matched := false
		for _, s := range skips {
			if strings.Contains(name, strings.TrimSpace(s)) {
				matched = true
				break
			}
		}
		if !matched {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// filterCleanable returns only results that exist, have no errors, and are not
// blocked by missing privileges.
func filterCleanable(results []scanner.ScanResult) []scanner.ScanResult {
	var out []scanner.ScanResult
	for _, r := range results {
		if r.Exists && r.Error == nil && !r.NeedsAdmin {
			out = append(out, r)
		}
	}
	return out
}

func totalScanSize(results []scanner.ScanResult) int64 {
	var total int64
	for _, r := range results {
		if r.Size > 0 {
			total += r.Size
		}
	}
	return total
}

// scanGroup is a category-bucketed group of scan results (deterministic order).
type scanGroup struct {
	category string
	items    []scanner.ScanResult
}

// groupScanResults buckets results by category, returning groups sorted by
// category name (byte order, consistent with Go string sorting).
func groupScanResults(results []scanner.ScanResult) []scanGroup {
	type acc struct {
		category string
		items    []scanner.ScanResult
	}
	order := []string{}
	byCat := map[string]*acc{}
	for _, r := range results {
		cat := r.Item.Category
		if _, ok := byCat[cat]; !ok {
			byCat[cat] = &acc{category: cat}
			order = append(order, cat)
		}
		byCat[cat].items = append(byCat[cat].items, r)
	}
	sort.Strings(order)

	groups := make([]scanGroup, 0, len(order))
	for _, cat := range order {
		groups = append(groups, scanGroup{category: cat, items: byCat[cat].items})
	}
	return groups
}

// scanRowStatus returns the human-readable status label for one scan result.
func scanRowStatus(r scanner.ScanResult) string {
	switch {
	case !r.Exists:
		return "없음"
	case r.Error != nil:
		return "오류"
	case r.NeedsAdmin:
		return "권한 필요 (관리자)"
	case r.Item.Type == scanner.TypeCommand:
		return "명령 실행"
	case r.Size == 0:
		return "0 B"
	default:
		return ui.FormatBytes(r.Size)
	}
}

// printScanTable renders the grouped scan results as rich tables.
func (a *app) printScanTable(results []scanner.ScanResult) {
	for _, g := range groupScanResults(results) {
		rich.Fprintln(a.out, "  [bold][cyan]%s[/cyan][/bold]", g.category)
		t := rich.NewTable("항목", "상태")
		for _, r := range g.items {
			t.AddRow(r.Item.Name, scanRowStatus(r))
		}
		t.Render(a.out)
		fmt.Fprintln(a.out)
	}
}

// printReport renders a cleanup summary into a.out.
func (a *app) printReport(results []cleaner.Result) {
	fmt.Fprintln(a.out)
	rich.Fprintln(a.out, "[bold][cyan]정리 완료[/cyan][/bold]")

	var totalFreed int64
	success, failed := 0, 0
	for _, r := range results {
		if r.Success {
			success++
			if r.Freed > 0 {
				totalFreed += r.Freed
			}
		} else {
			failed++
		}
	}

	rich.Fprintln(a.out, "  [green]성공: %d개[/green]  [red]실패: %d개[/red]", success, failed)
	rich.Fprintln(a.out, "  [bold][green]확보된 용량: %s[/green][/bold]", ui.FormatBytes(totalFreed))
	fmt.Fprintln(a.out)
}

// aiCommand builds the "pcc ai" subcommand.
func (a *app) aiCommand() *wcli.Command {
	var (
		dryRun   bool
		autoYes  bool
		skipList string
		model    string
		endpoint string
		provider string
	)

	cmd := &wcli.Command{
		Use:   "ai [지시사항]",
		Short: "LLM 에이전트가 캐시를 분석하고 정리할 항목을 추천합니다",
		Long:  "LLM(기본: Ollama)을 활용하여 캐시 항목들의 위험도를 평가하고 최적의 정리 계획을 제시합니다.",
		Run: func(ctx *wcli.Context) error {
			userPrompt := strings.Join(ctx.Args, " ")
			return a.cleanAI(dryRun, autoYes, skipList, provider, endpoint, model, userPrompt)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", "", false, "실제 삭제 없이 분석 및 추천만 실행")
	cmd.Flags().BoolVar(&autoYes, "yes", "y", false, "확인 질문 없이 AI 추천 항목 즉시 정리")
	cmd.Flags().StringVar(&skipList, "skip", "", "", "건너뛸 항목 (쉼표 구분)")
	cmd.Flags().StringVar(&model, "model", "", getEnvOrDefault("PCC_AI_MODEL", ai.DefaultModel), "사용할 AI 모델")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", getEnvOrDefault("PCC_AI_ENDPOINT", ai.DefaultBaseURL), "LLM API 엔드포인트 URL")
	cmd.Flags().StringVar(&provider, "provider", "", getEnvOrDefault("PCC_AI_PROVIDER", string(ai.DefaultProvider)), "LLM 프로바이더 (ollama, openai)")

	return cmd
}

// cleanAI executes the AI-assisted scan → analyze → confirm → clean pipeline.
func (a *app) cleanAI(dryRun, autoYes bool, skipList, provider, endpoint, model, prompt string) error {
	rich.Fprintln(a.out, "[bold][cyan]PC Cleaner AI — AI 디스크 분석 및 스마트 정리[/cyan][/bold]")
	if prompt != "" {
		rich.Fprintln(a.out, "  [dim]지시사항: %s[/dim]", prompt)
	}

	items := scanner.GetItems()
	items = filterItems(items, skipList)

	rich.Fprintln(a.out, "  [cyan]ℹ[/cyan] 총 %d개 항목 스캔 중...", len(items))
	scanResults := scanner.Scan(items)

	cleanable := filterCleanable(scanResults)
	if len(cleanable) == 0 {
		a.printNoCleanable()
		return nil
	}

	analyzer := a.aiAnalyzer
	if analyzer == nil {
		cfg := ai.Config{
			Provider: ai.Provider(provider),
			BaseURL:  endpoint,
			Model:    model,
			APIKey:   os.Getenv("OPENAI_API_KEY"),
		}
		client, effectiveModel, err := ai.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("AI 클라이언트 초기화 실패: %w", err)
		}
		analyzer = ai.NewAnalyzer(client, effectiveModel)
	}

	rich.Fprintln(a.out, "  [cyan]ℹ[/cyan] AI 모델(%s)로 캐시 분석 중...", model)
	analysis, err := analyzer.Analyze(context.Background(), cleanable, prompt)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") && provider == string(ai.ProviderOllama) {
			rich.Fprintln(a.out, "  [red]✗ Ollama 서버에 연결할 수 없습니다 (%s).[/red]", endpoint)
			rich.Fprintln(a.out, "  [yellow]ℹ 'ollama serve'가 실행 중인지 확인하거나, --endpoint 옵션을 확인하십시오.[/yellow]")
			return err
		}
		return fmt.Errorf("AI 분석 실패: %w", err)
	}

	fmt.Fprintln(a.out)
	rich.Fprintln(a.out, "  [bold][cyan]AI 권장 사항:[/cyan][/bold] %s", analysis.Summary)
	fmt.Fprintln(a.out)

	a.printAIRecommendationTable(scanResults, analysis)

	recommendedToClean := ai.FilterRecommended(cleanable, analysis)
	if len(recommendedToClean) == 0 {
		rich.Fprintln(a.out, "  [green]✓[/green] AI가 정리를 권장한 항목이 없거나 정리할 항목이 없습니다.")
		return nil
	}

	totalSize := totalScanSize(recommendedToClean)
	diskInfo, diskErr := scanner.GetDiskUsage("")

	if dryRun {
		if diskErr == nil {
			fmt.Fprintf(a.out, "  디스크 공간: 전체 %s | 사용 가능 %s (정리 후 예상: %s)\n",
				ui.FormatBytes(int64(diskInfo.Total)),
				ui.FormatBytes(int64(diskInfo.Available)),
				ui.FormatBytes(int64(diskInfo.Available+uint64(totalSize))),
			)
		}
		rich.Fprintln(a.out, "  [yellow]⚠ [DRY-RUN] 정리 권장 용량: [bold]%s[/bold][/yellow]",
			ui.FormatBytes(totalSize))
		cleaner.Clean(recommendedToClean, true)
		return nil
	}

	if diskErr == nil {
		fmt.Fprintf(a.out, "  디스크 공간: 전체 %s | 사용 가능 %s (정리 후 예상: %s)\n",
			ui.FormatBytes(int64(diskInfo.Total)),
			ui.FormatBytes(int64(diskInfo.Available)),
			ui.FormatBytes(int64(diskInfo.Available+uint64(totalSize))),
		)
	}
	rich.Fprintln(a.out, "  [bold]정리 권장 용량: %s[/bold]", ui.FormatBytes(totalSize))

	if !autoYes {
		ok, err := rich.FConfirm(a.out, a.in, "AI가 추천한 항목들의 삭제를 진행하시겠습니까?", false)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.out)
			rich.Fprintln(a.out, "  [yellow]⚠ 취소되었습니다.[/yellow]")
			return nil
		}
	}

	fmt.Fprintln(a.out)
	cleaned := cleaner.Clean(recommendedToClean, false)
	a.printReport(cleaned)
	return nil
}

// printAIRecommendationTable renders the AI analysis and recommendations as a rich table.
func (a *app) printAIRecommendationTable(results []scanner.ScanResult, analysis *ai.AnalysisResult) {
	recMap := make(map[string]ai.ItemRecommendation)
	for _, rec := range analysis.Recommendations {
		recMap[rec.Name] = rec
	}

	rich.Fprintln(a.out, "  [bold][cyan]AI 분석 리포트[/cyan][/bold]")
	t := rich.NewTable("항목", "크기", "정리 추천", "위험도", "사유")
	for _, r := range results {
		if !r.Exists || r.Error != nil || r.NeedsAdmin {
			continue
		}
		rec, exists := recMap[r.Item.Name]
		cleanStr := "[dim]보존[/dim]"
		if exists && rec.Clean {
			cleanStr = "[green]정리 권장[/green]"
		}
		riskStr := rec.RiskLevel
		if riskStr == "" {
			riskStr = "-"
		}
		reasonStr := rec.Reason
		if reasonStr == "" {
			reasonStr = "-"
		}
		sizeStr := scanRowStatus(r)

		t.AddRow(r.Item.Name, sizeStr, cleanStr, riskStr, reasonStr)
	}
	t.Render(a.out)
	fmt.Fprintln(a.out)
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
