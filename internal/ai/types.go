package ai

import "time"

// Provider specifies the LLM provider backend.
type Provider string

const (
	ProviderOllama Provider = "ollama"
	ProviderOpenAI Provider = "openai"
)

const (
	DefaultProvider = ProviderOllama
	DefaultModel    = "llama3.2:latest"
	DefaultBaseURL  = "http://localhost:11434/v1"
	DefaultTimeout  = 60 * time.Second
)

// Config configures the AI client connection.
type Config struct {
	Provider Provider
	BaseURL  string
	APIKey   string
	Model    string
	Timeout  time.Duration
}

// ItemRecommendation represents the AI's opinion on a specific cache item.
type ItemRecommendation struct {
	Name      string `json:"name"`
	Clean     bool   `json:"clean"`
	Reason    string `json:"reason"`
	RiskLevel string `json:"risk_level"` // e.g. "안전", "주의", "위험"
}

// AnalysisResult holds the aggregated recommendation report from the AI.
type AnalysisResult struct {
	Summary         string               `json:"summary"`
	Recommendations []ItemRecommendation `json:"recommendations"`
}
