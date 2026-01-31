package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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

// NewIntentDetector creates an intent detector based on configuration
// Currently only supports Gemini provider
func NewIntentDetector(config LLMConfig) (ports.IntentDetector, error) {
	provider := strings.ToLower(config.Provider)

	switch provider {
	case "gemini":
		// Normalize model name
		model := config.Model
		if model == "" {
			model = "gemini-pro"
		}
		model = normalizeModelName(model)
		if model == "gemini-1.5-flash" {
			model = "gemini-1.5-flash-latest"
		}

		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client for intent detection: %w", err)
		}

		return NewIntentDetectorAdapter(client, model), nil

	case "openai":
		// TODO: Implement OpenAI intent detector
		return nil, fmt.Errorf("OpenAI intent detector not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported LLM provider for intent detection: %s", provider)
	}
}
