package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/pc_cleaner/internal/scanner"
)

// Analyzer coordinates AI-driven cache analysis using an LLM client.
type Analyzer struct {
	client llm.Client
	model  string
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer(client llm.Client, model string) *Analyzer {
	return &Analyzer{
		client: client,
		model:  model,
	}
}

// Analyze sends the scan results to the LLM and produces an AnalysisResult.
func (a *Analyzer) Analyze(ctx context.Context, results []scanner.ScanResult, userInstruction string) (*AnalysisResult, error) {
	// Filter to cleanable results for LLM context
	cleanable := make([]scanner.ScanResult, 0, len(results))
	validNames := make(map[string]scanner.ScanResult)
	for _, r := range results {
		if r.Exists && r.Error == nil && !r.NeedsAdmin {
			cleanable = append(cleanable, r)
			validNames[r.Item.Name] = r
		}
	}

	if len(cleanable) == 0 {
		return &AnalysisResult{
			Summary:         "정리 가능한 캐시 항목이 없습니다.",
			Recommendations: []ItemRecommendation{},
		}, nil
	}

	prompt := buildUserPrompt(cleanable, userInstruction)

	req := llm.ChatRequest{
		Model: a.model,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: prompt},
		},
	}

	resp, err := a.client.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("AI 모델 호출 실패: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI 모델로부터 응답을 받지 못했습니다")
	}

	content := resp.Choices[0].Message.Content
	rawJSON := extractJSON(content)

	var analysis AnalysisResult
	if err := json.Unmarshal([]byte(rawJSON), &analysis); err != nil {
		// If LLM returned invalid JSON, create a safe fallback
		return fallbackAnalysis(cleanable), nil
	}

	// Filter out hallucinated items not present in original scan
	filteredRecs := make([]ItemRecommendation, 0, len(analysis.Recommendations))
	for _, rec := range analysis.Recommendations {
		if _, ok := validNames[rec.Name]; ok {
			filteredRecs = append(filteredRecs, rec)
		}
	}
	analysis.Recommendations = filteredRecs

	return &analysis, nil
}

// FilterRecommended returns only ScanResults that the AI recommended for cleaning.
func FilterRecommended(results []scanner.ScanResult, analysis *AnalysisResult) []scanner.ScanResult {
	if analysis == nil {
		return nil
	}
	toClean := make(map[string]bool)
	for _, rec := range analysis.Recommendations {
		if rec.Clean {
			toClean[rec.Name] = true
		}
	}

	var out []scanner.ScanResult
	for _, r := range results {
		if toClean[r.Item.Name] {
			out = append(out, r)
		}
	}
	return out
}

// extractJSON strips markdown code block fences if present.
func extractJSON(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		if idx := strings.LastIndex(content, "```"); idx != -1 {
			content = content[:idx]
		}
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		if idx := strings.LastIndex(content, "```"); idx != -1 {
			content = content[:idx]
		}
	}
	return strings.TrimSpace(content)
}

// fallbackAnalysis generates a conservative default recommendation when LLM parsing fails.
func fallbackAnalysis(results []scanner.ScanResult) *AnalysisResult {
	recs := make([]ItemRecommendation, 0, len(results))
	for _, r := range results {
		clean := true
		risk := "안전"
		reason := "일반 임시/캐시 파일로 삭제 가능"
		if r.Item.Category == "개발" {
			clean = false
			risk = "주의"
			reason = "개발 도구 캐시이므로 보존 권장"
		}
		recs = append(recs, ItemRecommendation{
			Name:      r.Item.Name,
			Clean:     clean,
			Reason:    reason,
			RiskLevel: risk,
		})
	}
	return &AnalysisResult{
		Summary:         "AI 응답 형식을 파싱할 수 없어 기본 안전 규칙으로 분석을 대체했습니다.",
		Recommendations: recs,
	}
}
