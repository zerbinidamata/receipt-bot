package llm

import (
	"fmt"
	"strings"

	"receipt-bot/internal/ports"
)

// LLMConfig holds configuration for LLM providers
type LLMConfig struct {
	Provider string // "gemini", "openai", "anthropic"
	APIKey   string
	Model    string
}

// NewLLMAdapter creates an appropriate LLM adapter based on configuration
func NewLLMAdapter(config LLMConfig) (ports.LLMPort, error) {
	provider := strings.ToLower(config.Provider)

	switch provider {
	case "gemini":
		return NewGeminiAdapter(config.APIKey, config.Model)

	case "openai":
		return NewOpenAIAdapter(config.APIKey, config.Model)

	// Future: Add Anthropic support
	// case "anthropic":
	//     return NewAnthropicAdapter(config.APIKey, config.Model)

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: gemini, openai)", provider)
	}
}
