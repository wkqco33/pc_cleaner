package ai

import (
	"fmt"
	"strings"

	"github.com/wkqco33/pc_cleaner/internal/scanner"
	"github.com/wkqco33/pc_cleaner/internal/ui"
)

const systemPrompt = `당신은 PC 디스크 공간 및 캐시 파일 정리를 전문으로 하는 시스템 관리자 AI 에이전트입니다.
현재 사용자의 PC에서 발견된 캐시 및 임시 파일 목록을 분석하여, 어떤 항목을 정리(삭제)해야 안전하고 효과적인지 판단해 주십시오.

[지침]
1. 반드시 주어진 캐시 항목 목록에 존재하는 항목들만 분석하십시오. 목록에 없는 새로운 항목을 임의로 생성하지 마십시오.
2. 개발 환경 캐시(Gradle, npm, pip, go-build 등)는 삭제 시 다음 빌드가 느려질 수 있으므로, 사용자가 명시적으로 요구하지 않는 한 보존(clean: false)을 권장하거나 '주의' 등급을 부여하십시오.
3. 브라우저 캐시, 썸네일, OS 임시 파일 등은 일반적으로 삭제해도 안전(clean: true, risk_level: "안전")합니다.
4. 사용자의 특별한 요청(프롬프트)이 있다면 이를 최우선으로 반영하십시오.
5. 반드시 아래 JSON 스키마 형식으로만 응답해야 합니다. 다른 설명이나 텍스트를 JSON 앞뒤에 붙이지 마십시오.

{
  "summary": "전체 분석 소견 요약",
  "recommendations": [
    {
      "name": "캐시 항목 이름 (입력된 목록과 정확히 일치해야 함)",
      "clean": true,
      "reason": "정리 권장 또는 보존 권장 사유",
      "risk_level": "안전" 또는 "주의" 또는 "위험"
    }
  ]
}`

func buildUserPrompt(results []scanner.ScanResult, userInstruction string) string {
	var sb strings.Builder
	if userInstruction != "" {
		sb.WriteString(fmt.Sprintf("[사용자 특별 요청]: %s\n\n", userInstruction))
	}

	sb.WriteString("[현재 발견된 캐시 항목 목록]\n")
	for _, r := range results {
		if !r.Exists || r.Error != nil || r.NeedsAdmin {
			continue
		}
		sizeStr := ui.FormatBytes(r.Size)
		if r.Item.Type == scanner.TypeCommand {
			sizeStr = "명령 실행형 항목"
		}
		sb.WriteString(fmt.Sprintf("- 이름: %s | 카테고리: %s | 크기: %s\n", r.Item.Name, r.Item.Category, sizeStr))
	}
	sb.WriteString("\n위 항목들을 분석하여 JSON 형식으로 결과를 출력해 주십시오.")
	return sb.String()
}
