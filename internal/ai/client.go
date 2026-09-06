package ai

import (
	"fmt"

	llm "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/LLM_client_go/ollama"
	"github.com/wkqco33/LLM_client_go/openai"
)

// NewClient instantiates an llm.Client based on the configuration.
// It returns the client, the effective model name, and an error if any.
func NewClient(cfg Config) (llm.Client, string, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = DefaultProvider
	}

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	switch provider {
	case ProviderOllama:
		baseURL := cfg.BaseURL
		if baseURL == "" {
			baseURL = DefaultBaseURL
		}
		client := ollama.New(ollama.Config{
			BaseURL: baseURL,
			Timeout: timeout,
		})
		return client, model, nil

	case ProviderOpenAI:
		client := openai.New(openai.Config{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Timeout: timeout,
		})
		return client, model, nil

	default:
		return nil, "", fmt.Errorf("지원되지 않는 AI 프로바이더: %q", provider)
	}
}
