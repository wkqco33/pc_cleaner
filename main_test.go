package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/wkqco33/pc_cleaner/internal/ai"
	"github.com/wkqco33/pc_cleaner/internal/cleaner"
	"github.com/wkqco33/pc_cleaner/internal/scanner"
)

// --- filterItems ---

func TestFilterItems_EmptySkipReturnsAll(t *testing.T) {
	items := []scanner.CacheItem{
		{Name: "Gradle 캐시"},
		{Name: "pip 캐시"},
	}
	got := filterItems(items, "")
	if len(got) != 2 {
		t.Errorf("skip이 비어있으면 모든 항목을 유지해야 합니다: %d개", len(got))
	}
}

func TestFilterItems_RemovesMatchingKeywords(t *testing.T) {
	items := []scanner.CacheItem{
		{Name: "Gradle 캐시"},
		{Name: "pip 캐시"},
		{Name: "npm 캐시"},
		{Name: "Maven 로컬 저장소"},
	}
	got := filterItems(items, "pip,npm")

	want := []string{"Gradle 캐시", "Maven 로컬 저장소"}
	if len(got) != len(want) {
		t.Fatalf("결과 개수 불일치: 기대 %d, 실제 %d", len(want), len(got))
	}
	for i, name := range want {
		if got[i].Name != name {
			t.Errorf("결과[%d] 이름 불일치: 기대 %q, 실제 %q", i, name, got[i].Name)
		}
	}
}

func TestFilterItems_IsCaseInsensitive(t *testing.T) {
	items := []scanner.CacheItem{{Name: "Pip 캐시"}}
	got := filterItems(items, "PIP")
	if len(got) != 0 {
		t.Errorf("대소문자 구분 없이 건너뛰어야 합니다: %d개 남음", len(got))
	}
}

func TestFilterItems_TrimsWhitespace(t *testing.T) {
	items := []scanner.CacheItem{{Name: "gradle 캐시"}, {Name: "docker 캐시"}}
	got := filterItems(items, " gradle , docker ")
	if len(got) != 0 {
		t.Errorf("공백을 잘라낸 키워드로 매칭해야 합니다: %d개 남음", len(got))
	}
}

func TestFilterItems_SubstringMatch(t *testing.T) {
	// 키워드가 이름의 부분 문자열이어도 매칭되어야 함 (예: "gradle")
	items := []scanner.CacheItem{{Name: "Gradle 캐시"}}
	got := filterItems(items, "gradle")
	if len(got) != 0 {
		t.Errorf("부분 문자열 매칭이 동작해야 합니다: %d개 남음", len(got))
	}
}

// --- filterCleanable ---

func TestFilterCleanable(t *testing.T) {
	results := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "정리 가능"}, Exists: true, Error: nil, NeedsAdmin: false},
		{Item: scanner.CacheItem{Name: "없는 경로"}, Exists: false},
		{Item: scanner.CacheItem{Name: "에러"}, Exists: true, Error: &stubErr{}},
		{Item: scanner.CacheItem{Name: "권한 필요"}, Exists: true, NeedsAdmin: true},
	}

	got := filterCleanable(results)
	if len(got) != 1 {
		t.Fatalf("정리 가능 항목은 1개여야 합니다: 실제 %d", len(got))
	}
	if got[0].Item.Name != "정리 가능" {
		t.Errorf("Exists=true, Error=nil, NeedsAdmin=false 인 항목만 남아야 합니다: %q", got[0].Item.Name)
	}
}

func TestFilterCleanable_Empty(t *testing.T) {
	if got := filterCleanable(nil); len(got) != 0 {
		t.Errorf("빈 입력이면 빈 결과여야 합니다: %d", len(got))
	}
}

// --- totalScanSize ---

func TestTotalScanSize(t *testing.T) {
	results := []scanner.ScanResult{
		{Size: 100},
		{Size: 0},
		{Size: -1}, // 명령형 항목
		{Size: 50},
	}
	if got := totalScanSize(results); got != 150 {
		t.Errorf("양수 크기만 합산해야 합니다: 기대 150, 실제 %d", got)
	}
}

func TestTotalScanSize_Empty(t *testing.T) {
	if got := totalScanSize(nil); got != 0 {
		t.Errorf("빈 입력이면 0이어야 합니다: %d", got)
	}
}

// --- scanRowStatus ---

func TestScanRowStatus_NotExist(t *testing.T) {
	r := scanner.ScanResult{Item: scanner.CacheItem{Name: "x"}, Exists: false}
	if got := scanRowStatus(r); got != "없음" {
		t.Errorf("존재하지 않는 항목은 '없음'이어야 합니다: %q", got)
	}
}

func TestScanRowStatus_Error(t *testing.T) {
	r := scanner.ScanResult{Item: scanner.CacheItem{Name: "x"}, Exists: true, Error: &stubErr{}}
	if got := scanRowStatus(r); got != "오류" {
		t.Errorf("에러 항목은 '오류'여야 합니다: %q", got)
	}
}

func TestScanRowStatus_NeedsAdmin(t *testing.T) {
	r := scanner.ScanResult{Item: scanner.CacheItem{Name: "x"}, Exists: true, NeedsAdmin: true}
	if got := scanRowStatus(r); got != "권한 필요 (관리자)" {
		t.Errorf("권한 필요 항목의 상태 라벨이 올바르지 않습니다: %q", got)
	}
}

func TestScanRowStatus_Command(t *testing.T) {
	r := scanner.ScanResult{
		Item:   scanner.CacheItem{Name: "x", Type: scanner.TypeCommand, Command: []string{"whoami"}},
		Exists: true,
		Size:   -1,
	}
	if got := scanRowStatus(r); got != "명령 실행" {
		t.Errorf("명령형 항목은 '명령 실행'이어야 합니다: %q", got)
	}
}

func TestScanRowStatus_ZeroSize(t *testing.T) {
	r := scanner.ScanResult{Item: scanner.CacheItem{Name: "x"}, Exists: true, Size: 0}
	if got := scanRowStatus(r); got != "0 B" {
		t.Errorf("크기 0 은 '0 B'여야 합니다: %q", got)
	}
}

func TestScanRowStatus_FormattedSize(t *testing.T) {
	r := scanner.ScanResult{Item: scanner.CacheItem{Name: "x"}, Exists: true, Size: 1024}
	if got := scanRowStatus(r); got != "1.0 KB" {
		t.Errorf("크기는 포맷되어야 합니다(1.0 KB): %q", got)
	}
}

// --- groupScanResults ---

func TestGroupScanResults_Empty(t *testing.T) {
	if got := groupScanResults(nil); len(got) != 0 {
		t.Errorf("빈 입력이면 빈 그룹이어야 합니다: %d", len(got))
	}
}

func TestGroupScanResults_GroupsByCategory(t *testing.T) {
	results := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "a", Category: "cat1"}},
		{Item: scanner.CacheItem{Name: "b", Category: "cat2"}},
		{Item: scanner.CacheItem{Name: "c", Category: "cat1"}},
	}
	groups := groupScanResults(results)
	if len(groups) != 2 {
		t.Fatalf("카테고리는 2개여야 합니다: %d", len(groups))
	}
	// 같은 카테고리 항목이 한 그룹에 모여야 함
	if len(groups[0].items) != 2 || len(groups[1].items) != 1 {
		t.Errorf("그룹별 항목 수 불일치: %+v", groups)
	}
	if groups[0].category != "cat1" {
		t.Errorf("cat1이 첫 그룹이어야 합니다: %q", groups[0].category)
	}
}

func TestGroupScanResults_SortedByCategory(t *testing.T) {
	// sort.Strings는 UTF-8 바이트 순서로 정렬하므로 ASCII 카테고리로 검증한다.
	results := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "B", Category: "beta"}},
		{Item: scanner.CacheItem{Name: "A", Category: "alpha"}},
	}
	groups := groupScanResults(results)
	if groups[0].category != "alpha" {
		t.Errorf("카테고리가 사전순(alpha < beta)이어야 합니다: %q", groups[0].category)
	}
}

// --- app rendering ---

func TestPrintScanTable_WritesCategoriesAndRows(t *testing.T) {
	var buf bytes.Buffer
	a := newApp(&buf, nil)
	results := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "B 항목", Category: "beta"}, Exists: true, Size: 512},
		{Item: scanner.CacheItem{Name: "A 항목", Category: "alpha"}, Exists: false},
	}
	a.printScanTable(results)
	out := buf.String()

	for _, want := range []string{"alpha", "beta", "A 항목", "B 항목", "없음", "512 B"} {
		if !strings.Contains(out, want) {
			t.Errorf("스캔 테이블에 %q 가 없어야 합니다(실제 있음): %q", want, out)
		}
	}
	// 카테고리 정렬 순서
	if strings.Index(out, "alpha") > strings.Index(out, "beta") {
		t.Errorf("카테고리가 사전순으로 출력되어야 합니다: %q", out)
	}
}

func TestPrintReport_CountsAndFreed(t *testing.T) {
	var buf bytes.Buffer
	a := newApp(&buf, nil)
	results := []cleaner.Result{
		{Success: true, Freed: 100},
		{Success: true, Freed: 50},
		{Success: false},
	}
	a.printReport(results)
	out := buf.String()

	if !strings.Contains(out, "성공") || !strings.Contains(out, "실패") {
		t.Errorf("성공/실패 통계를 출력해야 합니다: %q", out)
	}
	if !strings.Contains(out, "150 B") {
		t.Errorf("확보 용량 합계(150 B)를 출력해야 합니다: %q", out)
	}
}

// --- AI command tests ---

type stubAIAnalyzer struct {
	result     *ai.AnalysisResult
	err        error
	lastPrompt string
}

func (s *stubAIAnalyzer) Analyze(ctx context.Context, results []scanner.ScanResult, prompt string) (*ai.AnalysisResult, error) {
	s.lastPrompt = prompt
	return s.result, s.err
}

func TestAICommand_DryRun(t *testing.T) {
	var buf bytes.Buffer
	a := newApp(&buf, nil)
	stub := &stubAIAnalyzer{
		result: &ai.AnalysisResult{
			Summary: "AI 권장 사항: 불필요한 임시 파일 정리",
			Recommendations: []ai.ItemRecommendation{
				{
					Name:      "Xcode 캐시",
					Clean:     true,
					Reason:    "공간 확보 최적",
					RiskLevel: "안전",
				},
			},
		},
	}
	a.aiAnalyzer = stub

	// Execute with --dry-run
	err := a.rootCommand().Execute([]string{"ai", "--dry-run", "개발 캐시 제외"})
	if err != nil {
		t.Fatalf("명령어 실행 실패: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "AI 디스크 분석") {
		t.Errorf("AI 디스크 분석 헤더가 출력되어야 합니다: %q", out)
	}
	if !strings.Contains(out, "AI 권장 사항") {
		t.Errorf("AI 요약이 출력되어야 합니다: %q", out)
	}
	if stub.lastPrompt != "개발 캐시 제외" {
		t.Errorf("사용자 프롬프트가 전달되어야 합니다: %q", stub.lastPrompt)
	}
}

func TestAICommand_NoCleanable(t *testing.T) {
	var buf bytes.Buffer
	a := newApp(&buf, nil)
	stub := &stubAIAnalyzer{
		result: &ai.AnalysisResult{
			Summary: "정리할 항목이 없습니다.",
			Recommendations: []ai.ItemRecommendation{
				{Name: "Xcode 캐시", Clean: false, Reason: "유지 필요"},
			},
		},
	}
	a.aiAnalyzer = stub

	err := a.rootCommand().Execute([]string{"ai", "--dry-run"})
	if err != nil {
		t.Fatalf("명령어 실행 실패: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "정리할 항목이 없습니다") {
		t.Errorf("정리 항목 없음 메시지가 출력되어야 합니다: %q", out)
	}
}

// --- helpers ---

type stubErr struct{}

func (s *stubErr) Error() string { return "boom" }
