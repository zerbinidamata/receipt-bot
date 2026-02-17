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

// TranslateRecipe translates a recipe to the target language
func (a *OpenAIAdapter) TranslateRecipe(ctx context.Context, recipe *ports.RecipeTranslationInput, targetLang string) (*ports.RecipeTranslationOutput, error) {
	// Build ingredients list
	var ingredients []string
	for _, ing := range recipe.Ingredients {
		ingStr := ing.Name
		if ing.Quantity != "" {
			ingStr = ing.Quantity + " " + ing.Unit + " " + ing.Name
		}
		if ing.Notes != "" {
			ingStr += " (" + ing.Notes + ")"
		}
		ingredients = append(ingredients, ingStr)
	}

	// Build instructions list
	var instructions []string
	for _, inst := range recipe.Instructions {
		instructions = append(instructions, inst.Text)
	}

	prompt := fmt.Sprintf(`Translate this recipe to %s. Keep the same structure and format.

Title: %s

Ingredients:
%s

Instructions:
%s

Return ONLY valid JSON in this exact format:
{
  "title": "translated title",
  "ingredients": [
    {"name": "ingredient name", "quantity": "amount", "unit": "unit", "notes": "any notes"}
  ],
  "instructions": [
    {"step_number": 1, "text": "instruction text"}
  ]
}

IMPORTANT:
- Translate ALL text to %s
- Keep quantities and measurements accurate
- Preserve step numbers
- Keep cooking terms natural in the target language`, targetLang, recipe.Title, joinStrings(ingredients, "\n"), joinStrings(instructions, "\n"), targetLang)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: 0.3,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI for translation")
	}

	responseText := resp.Choices[0].Message.Content

	var translationResp struct {
		Title        string            `json:"title"`
		Ingredients  []ingredientJSON  `json:"ingredients"`
		Instructions []instructionJSON `json:"instructions"`
	}
	if err := json.Unmarshal([]byte(responseText), &translationResp); err != nil {
		return nil, fmt.Errorf("failed to parse translation response: %w", err)
	}

	output := &ports.RecipeTranslationOutput{
		Title:        translationResp.Title,
		Ingredients:  make([]ports.IngredientData, len(translationResp.Ingredients)),
		Instructions: make([]ports.InstructionData, len(translationResp.Instructions)),
	}

	for i, ing := range translationResp.Ingredients {
		output.Ingredients[i] = ports.IngredientData{
			Name:     ing.Name,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
			Notes:    ing.Notes,
		}
	}

	for i, inst := range translationResp.Instructions {
		output.Instructions[i] = ports.InstructionData{
			StepNumber: inst.StepNumber,
			Text:       inst.Text,
		}
	}

	return output, nil
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
