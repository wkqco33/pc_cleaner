package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/seomini/pc_cleaner/internal/cleaner"
	"github.com/seomini/pc_cleaner/internal/scanner"
	"github.com/seomini/pc_cleaner/internal/ui"
)

var version = "0.1.0"

func main() {
	dryRun := flag.Bool("dry-run", false, "실제 삭제 없이 분석만 실행")
	skipList := flag.String("skip", "", "건너뛸 항목 (쉼표 구분, 예: gradle,pip,docker)")
	showVersion := flag.Bool("version", false, "버전 출력")
	flag.Parse()

	if *showVersion {
		fmt.Printf("pc_cleaner v%s\n", version)
		return
	}

	ui.PrintHeader(fmt.Sprintf("PC Cleaner v%s — %s", version, runtime.GOOS))

	// 스캔 대상 수집
	items := scanner.GetItems()
	items = filterItems(items, *skipList)

	ui.PrintInfo(fmt.Sprintf("총 %d개 항목 스캔 중...", len(items)))
	fmt.Println()

	// 병렬 스캔
	results := scanner.Scan(items)

	// 결과 출력
	printTable(results)

	// 정리 대상만 추출
	cleanable := filterCleanable(results)
	if len(cleanable) == 0 {
		ui.PrintOK("정리할 항목이 없습니다.")
		return
	}

	totalSize := totalScanSize(cleanable)

	fmt.Println()
	if *dryRun {
		ui.PrintWarn(fmt.Sprintf("[DRY-RUN] 정리 가능 용량: %s", ui.Bold(ui.Yellow(ui.FormatBytes(totalSize)))))
		cleaner.Clean(cleanable, true)
		return
	}

	// 사용자 확인
	fmt.Printf("\n  %s\n",
		ui.Bold(fmt.Sprintf("정리 가능 용량: %s", ui.Yellow(ui.FormatBytes(totalSize)))),
	)
	fmt.Printf("  삭제를 진행하시겠습니까? %s ", ui.Gray("[y/N]"))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" {
		fmt.Println()
		ui.PrintWarn("취소되었습니다.")
		return
	}

	fmt.Println()
	ui.PrintHeader("정리 실행 중")
	cleanResults := cleaner.Clean(cleanable, false)

	// 결과 리포트
	printReport(cleanResults)
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

// filterCleanable returns only results that exist and have no errors.
func filterCleanable(results []scanner.ScanResult) []scanner.ScanResult {
	var out []scanner.ScanResult
	for _, r := range results {
		if r.Exists && r.Error == nil {
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

func printTable(results []scanner.ScanResult) {
	// 카테고리별 그룹
	type group struct {
		category string
		items    []scanner.ScanResult
	}

	catMap := map[string]*group{}
	var catOrder []string
	for _, r := range results {
		cat := r.Item.Category
		if _, ok := catMap[cat]; !ok {
			catMap[cat] = &group{category: cat}
			catOrder = append(catOrder, cat)
		}
		catMap[cat].items = append(catMap[cat].items, r)
	}
	sort.Strings(catOrder)

	for _, cat := range catOrder {
		g := catMap[cat]
		fmt.Printf("  %s\n", ui.Bold(ui.Cyan(g.category)))
		for _, r := range g.items {
			printRow(r)
		}
		fmt.Println()
	}
}

func printRow(r scanner.ScanResult) {
	name := fmt.Sprintf("    %-40s", r.Item.Name)

	switch {
	case !r.Exists:
		fmt.Printf("%s %s\n", name, ui.Gray("없음"))
	case r.Error != nil:
		fmt.Printf("%s %s\n", name, ui.Red("오류"))
	case r.Item.Type == scanner.TypeCommand:
		fmt.Printf("%s %s\n", name, ui.Yellow("명령 실행"))
	case r.Size == 0:
		fmt.Printf("%s %s\n", name, ui.Gray("0 B"))
	default:
		fmt.Printf("%s %s\n", name, ui.Yellow(ui.FormatBytes(r.Size)))
	}
}

func printReport(results []cleaner.Result) {
	fmt.Println()
	ui.PrintHeader("정리 완료")

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

	fmt.Printf("  성공: %s  실패: %s\n",
		ui.Green(fmt.Sprintf("%d개", success)),
		ui.Red(fmt.Sprintf("%d개", failed)),
	)
	fmt.Printf("  확보된 용량: %s\n\n", ui.Bold(ui.Green(ui.FormatBytes(totalFreed))))
}
