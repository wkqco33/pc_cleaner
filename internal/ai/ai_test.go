package ai

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/pc_cleaner/internal/scanner"
)

// fakeLLMClient implements llm.Client for testing without real network calls.
type fakeLLMClient struct {
	completeFunc func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
}

func (f *fakeLLMClient) Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if f.completeFunc != nil {
		return f.completeFunc(ctx, req)
	}
	return &llm.ChatResponse{}, nil
}

func (f *fakeLLMClient) Stream(ctx context.Context, req llm.ChatRequest) (llm.Stream, error) {
	return nil, errors.New("stream not implemented in fake")
}

func (f *fakeLLMClient) CreateEmbeddings(ctx context.Context, req llm.EmbeddingRequest) (*llm.EmbeddingResponse, error) {
	return nil, errors.New("embeddings not implemented in fake")
}

func (f *fakeLLMClient) TokenCounter(model string) any {
	return nil
}

func TestNewClient_Defaults(t *testing.T) {
	cfg := Config{}
	client, model, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient 실패: %v", err)
	}
	if client == nil {
		t.Fatal("client가 nil이면 안 됩니다")
	}
	if model != DefaultModel {
		t.Errorf("기본 모델 불일치: 기대 %q, 실제 %q", DefaultModel, model)
	}
}

func TestNewClient_Custom(t *testing.T) {
	cfg := Config{
		Provider: ProviderOpenAI,
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   "test-key",
		Model:    "gpt-4o-mini",
	}
	client, model, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient 실패: %v", err)
	}
	if client == nil {
		t.Fatal("client가 nil이면 안 됩니다")
	}
	if model != "gpt-4o-mini" {
		t.Errorf("모델 불일치: 기대 'gpt-4o-mini', 실제 %q", model)
	}
}

func TestAnalyze_SuccessfulRecommendation(t *testing.T) {
	mockResponse := struct {
		Summary         string               `json:"summary"`
		Recommendations []ItemRecommendation `json:"recommendations"`
	}{
		Summary: "임시 파일과 브라우저 캐시 정리를 권장합니다.",
		Recommendations: []ItemRecommendation{
			{
				Name:      "Xcode 캐시",
				Clean:     true,
				Reason:    "최근 빌드 잔여물이 많아 정리 시 큰 공간 확보 가능",
				RiskLevel: "안전",
			},
			{
				Name:      "Gradle 캐시",
				Clean:     false,
				Reason:    "현재 진행 중인 프로젝트의 재빌드 속도 저하 방지",
				RiskLevel: "주의",
			},
		},
	}
	respBytes, _ := json.Marshal(mockResponse)

	fake := &fakeLLMClient{
		completeFunc: func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
			return &llm.ChatResponse{
				Choices: []llm.Choice{
					{
						Message: llm.Message{
							Role:    llm.RoleAssistant,
							Content: string(respBytes),
						},
					},
				},
			}, nil
		},
	}

	analyzer := NewAnalyzer(fake, "test-model")

	scanResults := []scanner.ScanResult{
		{
			Item:   scanner.CacheItem{Name: "Xcode 캐시", Category: "개발"},
			Size:   1024 * 1024 * 500, // 500MB
			Exists: true,
		},
		{
			Item:   scanner.CacheItem{Name: "Gradle 캐시", Category: "개발"},
			Size:   1024 * 1024 * 300, // 300MB
			Exists: true,
		},
	}

	res, err := analyzer.Analyze(context.Background(), scanResults, "안전한 캐시만 정리해줘")
	if err != nil {
		t.Fatalf("Analyze 실패: %v", err)
	}

	if res.Summary != mockResponse.Summary {
		t.Errorf("요약 불일치: 기대 %q, 실제 %q", mockResponse.Summary, res.Summary)
	}
	if len(res.Recommendations) != 2 {
		t.Fatalf("추천 항목 수 불일치: 기대 2, 실제 %d", len(res.Recommendations))
	}
	if !res.Recommendations[0].Clean || res.Recommendations[1].Clean {
		t.Errorf("추천 정리 플래그 불일치: %+v", res.Recommendations)
	}
}

func TestAnalyze_FiltersHallucinatedItems(t *testing.T) {
	mockResponse := struct {
		Summary         string               `json:"summary"`
		Recommendations []ItemRecommendation `json:"recommendations"`
	}{
		Summary: "분석 완료",
		Recommendations: []ItemRecommendation{
			{Name: "Xcode 캐시", Clean: true, Reason: "OK"},
			{Name: "전혀없는캐시", Clean: true, Reason: "삭제 권장"},
		},
	}
	respBytes, _ := json.Marshal(mockResponse)

	fake := &fakeLLMClient{
		completeFunc: func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
			return &llm.ChatResponse{
				Choices: []llm.Choice{
					{
						Message: llm.Message{
							Role:    llm.RoleAssistant,
							Content: string(respBytes),
						},
					},
				},
			}, nil
		},
	}

	analyzer := NewAnalyzer(fake, "test-model")
	scanResults := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "Xcode 캐시"}, Exists: true, Size: 100},
	}

	res, err := analyzer.Analyze(context.Background(), scanResults, "")
	if err != nil {
		t.Fatalf("Analyze 실패: %v", err)
	}

	if len(res.Recommendations) != 1 {
		t.Fatalf("환각 항목은 필터링되어야 합니다: 실제 %d개", len(res.Recommendations))
	}
	if res.Recommendations[0].Name != "Xcode 캐시" {
		t.Errorf("스캔 결과에 존재하는 항목만 남아야 합니다: %q", res.Recommendations[0].Name)
	}
}

func TestAnalyze_FallbackOnInvalidJSON(t *testing.T) {
	fake := &fakeLLMClient{
		completeFunc: func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
			return &llm.ChatResponse{
				Choices: []llm.Choice{
					{
						Message: llm.Message{
							Role:    llm.RoleAssistant,
							Content: "이것은 유효하지 않은 JSON 응답입니다.",
						},
					},
				},
			}, nil
		},
	}

	analyzer := NewAnalyzer(fake, "test-model")
	scanResults := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "브라우저 캐시"}, Exists: true, Size: 100},
	}

	res, err := analyzer.Analyze(context.Background(), scanResults, "")
	if err != nil {
		t.Fatalf("JSON 파싱 실패 시 폴백으로 동작해야 하며 에러가 발생하지 않아야 합니다: %v", err)
	}
	if res == nil {
		t.Fatal("결과가 nil이면 안 됩니다")
	}
	if len(res.Recommendations) != 1 {
		t.Errorf("폴백 추천 항목 수 불일치: %d", len(res.Recommendations))
	}
}

func TestAnalyze_LLMError(t *testing.T) {
	fake := &fakeLLMClient{
		completeFunc: func(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
			return nil, errors.New("connection refused")
		},
	}

	analyzer := NewAnalyzer(fake, "test-model")
	scanResults := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "브라우저 캐시"}, Exists: true, Size: 100},
	}

	_, err := analyzer.Analyze(context.Background(), scanResults, "")
	if err == nil {
		t.Fatal("LLM 호출 에러 시 에러를 반환해야 합니다")
	}
}

func TestFilterRecommended(t *testing.T) {
	scanResults := []scanner.ScanResult{
		{Item: scanner.CacheItem{Name: "A"}, Exists: true, Size: 100},
		{Item: scanner.CacheItem{Name: "B"}, Exists: true, Size: 200},
		{Item: scanner.CacheItem{Name: "C"}, Exists: true, Size: 300},
	}

	analysis := &AnalysisResult{
		Recommendations: []ItemRecommendation{
			{Name: "A", Clean: true},
			{Name: "B", Clean: false},
		},
	}

	got := FilterRecommended(scanResults, analysis)
	if len(got) != 1 {
		t.Fatalf("Clean=true인 항목만 추출되어야 합니다: %d개", len(got))
	}
	if got[0].Item.Name != "A" {
		t.Errorf("A 항목이어야 합니다: %s", got[0].Item.Name)
	}
}

func TestNewClient_UnsupportedProvider(t *testing.T) {
	cfg := Config{Provider: "unsupported"}
	_, _, err := NewClient(cfg)
	if err == nil {
		t.Fatal("지원하지 않는 프로바이더에 대해 에러를 반환해야 합니다")
	}
}

func TestAnalyze_EmptyScanResults(t *testing.T) {
	fake := &fakeLLMClient{}
	analyzer := NewAnalyzer(fake, "test-model")
	res, err := analyzer.Analyze(context.Background(), nil, "")
	if err != nil {
		t.Fatalf("빈 스캔 결과일 때 에러 없이 동작해야 합니다: %v", err)
	}
	if len(res.Recommendations) != 0 {
		t.Errorf("추천 항목이 비어있어야 합니다: %d", len(res.Recommendations))
	}
}

func TestExtractJSON_Formats(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`{"a": 1}`, `{"a": 1}`},
		{"```json\n{\"a\": 1}\n```", `{"a": 1}`},
		{"```\n{\"a\": 1}\n```", `{"a": 1}`},
	}
	for _, tc := range cases {
		got := extractJSON(tc.input)
		if got != tc.want {
			t.Errorf("extractJSON(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
