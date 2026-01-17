package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"receipt-bot/internal/ports"
)

// OpenAIAdapter implements the LLMPort using OpenAI
type OpenAIAdapter struct {
	client *openai.Client
	model  string
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(apiKey string, model string) (*OpenAIAdapter, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if model == "" {
		model = "gpt-4o-mini" // Default to cost-effective model
	}

	client := openai.NewClient(apiKey)

	return &OpenAIAdapter{
		client: client,
		model:  model,
	}, nil
}

// ExtractRecipe implements the LLMPort interface
func (a *OpenAIAdapter) ExtractRecipe(ctx context.Context, text string) (*ports.RecipeExtraction, error) {
	// Build messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: BuildUserPrompt(text),
		},
	}

	// Create request with JSON mode
	req := openai.ChatCompletionRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: 0.3,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	// Call OpenAI API
	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Extract response
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	responseText := resp.Choices[0].Message.Content

	// Parse JSON response
	var recipeJSON recipeJSON
	if err := json.Unmarshal([]byte(responseText), &recipeJSON); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response as JSON: %w", err)
	}

	// Convert to domain format
	extraction := convertJSONToExtraction(&recipeJSON)

	return extraction, nil
}
