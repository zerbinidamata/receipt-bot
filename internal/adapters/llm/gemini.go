package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"receipt-bot/internal/ports"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// GeminiAdapter implements the LLMPort using Google Gemini
type GeminiAdapter struct {
	client *genai.Client
	model  string
}

// NewGeminiAdapter creates a new Gemini adapter
func NewGeminiAdapter(apiKey string, model string) (*GeminiAdapter, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	if model == "" {
		model = "gemini-pro" // Default to stable, widely available model
	}

	// Normalize model name - handle common variations
	model = normalizeModelName(model)
	
	// If user specified "gemini-1.5-flash" without -latest, try with -latest suffix for better compatibility
	// Also try gemini-pro as fallback if flash models aren't available
	if model == "gemini-1.5-flash" {
		model = "gemini-1.5-flash-latest"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiAdapter{
		client: client,
		model:  model,
	}, nil
}

// normalizeModelName normalizes Gemini model names to the correct format
func normalizeModelName(model string) string {
	// Map common model name variations to correct format
	modelMap := map[string]string{
		"gemini-1.5-flash":      "gemini-1.5-flash",
		"gemini-1.5-flash-latest": "gemini-1.5-flash-latest",
		"gemini-1.5-pro":        "gemini-1.5-pro",
		"gemini-1.5-pro-latest": "gemini-1.5-pro-latest",
		"gemini-pro":            "gemini-pro",
		"gemini-1.0-pro":        "gemini-1.0-pro",
	}

	if normalized, ok := modelMap[model]; ok {
		return normalized
	}

	// If not in map, return as-is (might be a valid model name we don't know about)
	return model
}

// Close closes the Gemini client
func (a *GeminiAdapter) Close() error {
	return a.client.Close()
}

// ExtractRecipe implements the LLMPort interface
func (a *GeminiAdapter) ExtractRecipe(ctx context.Context, text string) (*ports.RecipeExtraction, error) {
	model := a.client.GenerativeModel(a.model)

	// Configure model for JSON output
	model.SetTemperature(0.3) // Lower temperature for more deterministic output
	model.ResponseMIMEType = "application/json"

	// Build the prompt
	prompt := fmt.Sprintf("%s\n\n%s", SystemPrompt, BuildUserPrompt(text))

	// Add timeout to prevent hanging indefinitely
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Generate content
	resp, err := model.GenerateContent(ctxWithTimeout, genai.Text(prompt))
	if err != nil {
		// Check for timeout
		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("Gemini API call timed out after 60 seconds. The API may be slow or unresponsive. Please try again")
		}

		// Provide helpful error message for model not found errors
		errStr := err.Error()
		if contains(errStr, "not found") || contains(errStr, "not supported") {
			return nil, fmt.Errorf("Gemini API call failed: %w\n\n"+
				"Troubleshooting:\n"+
				"1. Verify the model name is correct. Try: gemini-1.5-flash-latest, gemini-1.5-pro, or gemini-pro\n"+
				"2. Check your API key has access to the requested model\n"+
				"3. Update LLM_MODEL in your .env file to a supported model name", err)
		}
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Get the text response
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	// Log raw response for debugging (first 1000 chars)
	responsePreview := responseText
	if len(responsePreview) > 1000 {
		responsePreview = responsePreview[:1000] + "..."
	}
	fmt.Printf("[DEBUG] Gemini raw response (preview): %s\n", responsePreview)

	// Clean up response - remove markdown code blocks if present
	cleanedResponse := cleanJSONResponse(responseText)

	// Parse JSON response
	var recipeJSON recipeJSON
	if err := json.Unmarshal([]byte(cleanedResponse), &recipeJSON); err != nil {
		fmt.Printf("[DEBUG] Failed to parse JSON. Raw response: %s\n", responseText)
		return nil, fmt.Errorf("failed to parse Gemini response as JSON: %w", err)
	}

	fmt.Printf("[DEBUG] Parsed JSON - Ingredients: %d, Instructions: %d\n", len(recipeJSON.Ingredients), len(recipeJSON.Instructions))

	// Convert to domain format
	extraction := convertJSONToExtraction(&recipeJSON)

	return extraction, nil
}

// cleanJSONResponse removes markdown code blocks and extra whitespace from JSON response
func cleanJSONResponse(response string) string {
	// Remove markdown code blocks (```json ... ``` or ``` ... ```)
	codeBlockRegex := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")
	matches := codeBlockRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		response = matches[1]
	}

	// Trim whitespace
	response = strings.TrimSpace(response)

	// Find JSON object boundaries if there's extra text
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		response = response[startIdx : endIdx+1]
	}

	return response
}

// recipeJSON represents the JSON structure from the LLM
type recipeJSON struct {
	Title           string            `json:"title"`
	Category        string            `json:"category"`
	Cuisine         string            `json:"cuisine"`
	DietaryTags     []string          `json:"dietary_tags"`
	Tags            []string          `json:"tags"`
	Ingredients     []ingredientJSON  `json:"ingredients"`
	Instructions    []instructionJSON `json:"instructions"`
	PrepTimeMinutes *int              `json:"prep_time_minutes"`
	CookTimeMinutes *int              `json:"cook_time_minutes"`
	Servings        *int              `json:"servings"`

	// Multilingual support
	SourceLanguage         string            `json:"source_language"`
	TranslatedTitle        *string           `json:"translated_title"`
	TranslatedIngredients  []ingredientJSON  `json:"translated_ingredients"`
	TranslatedInstructions []instructionJSON `json:"translated_instructions"`
}

type ingredientJSON struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
	Unit     string `json:"unit"`
	Notes    string `json:"notes"`
}

type instructionJSON struct {
	StepNumber      int      `json:"step_number"`
	Text            string   `json:"text"`
	DurationMinutes *float64 `json:"duration_minutes"`
}

// convertJSONToExtraction converts the JSON response to domain format
func convertJSONToExtraction(recipe *recipeJSON) *ports.RecipeExtraction {
	extraction := &ports.RecipeExtraction{
		Title:          recipe.Title,
		Category:       recipe.Category,
		Cuisine:        recipe.Cuisine,
		DietaryTags:    recipe.DietaryTags,
		Tags:           recipe.Tags,
		Ingredients:    make([]ports.IngredientData, len(recipe.Ingredients)),
		Instructions:   make([]ports.InstructionData, len(recipe.Instructions)),
		SourceLanguage: recipe.SourceLanguage,
	}

	// Default source language to English if not specified
	if extraction.SourceLanguage == "" {
		extraction.SourceLanguage = "en"
	}

	// Convert ingredients
	for i, ing := range recipe.Ingredients {
		extraction.Ingredients[i] = ports.IngredientData{
			Name:     ing.Name,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
			Notes:    ing.Notes,
		}
	}

	// Convert instructions
	for i, inst := range recipe.Instructions {
		var duration *time.Duration
		if inst.DurationMinutes != nil && *inst.DurationMinutes > 0 {
			d := time.Duration(*inst.DurationMinutes * float64(time.Minute))
			duration = &d
		}

		extraction.Instructions[i] = ports.InstructionData{
			StepNumber: inst.StepNumber,
			Text:       inst.Text,
			Duration:   duration,
		}
	}

	// Convert times
	if recipe.PrepTimeMinutes != nil && *recipe.PrepTimeMinutes > 0 {
		d := time.Duration(*recipe.PrepTimeMinutes) * time.Minute
		extraction.PrepTime = &d
	}

	if recipe.CookTimeMinutes != nil && *recipe.CookTimeMinutes > 0 {
		d := time.Duration(*recipe.CookTimeMinutes) * time.Minute
		extraction.CookTime = &d
	}

	extraction.Servings = recipe.Servings

	// Convert translated title
	extraction.TranslatedTitle = recipe.TranslatedTitle

	// Convert translated ingredients
	if len(recipe.TranslatedIngredients) > 0 {
		extraction.TranslatedIngredients = make([]ports.IngredientData, len(recipe.TranslatedIngredients))
		for i, ing := range recipe.TranslatedIngredients {
			extraction.TranslatedIngredients[i] = ports.IngredientData{
				Name:     ing.Name,
				Quantity: ing.Quantity,
				Unit:     ing.Unit,
				Notes:    ing.Notes,
			}
		}
	}

	// Convert translated instructions
	if len(recipe.TranslatedInstructions) > 0 {
		extraction.TranslatedInstructions = make([]ports.InstructionData, len(recipe.TranslatedInstructions))
		for i, inst := range recipe.TranslatedInstructions {
			var duration *time.Duration
			if inst.DurationMinutes != nil && *inst.DurationMinutes > 0 {
				d := time.Duration(*inst.DurationMinutes * float64(time.Minute))
				duration = &d
			}

			extraction.TranslatedInstructions[i] = ports.InstructionData{
				StepNumber: inst.StepNumber,
				Text:       inst.Text,
				Duration:   duration,
			}
		}
	}

	return extraction
}

// TranslateRecipe translates a recipe to the target language
func (a *GeminiAdapter) TranslateRecipe(ctx context.Context, recipe *ports.RecipeTranslationInput, targetLang string) (*ports.RecipeTranslationOutput, error) {
	model := a.client.GenerativeModel(a.model)

	// Configure model for JSON output
	model.SetTemperature(0.3)
	model.ResponseMIMEType = "application/json"

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

	// Build prompt
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
- Keep cooking terms natural in the target language`, targetLang, recipe.Title, strings.Join(ingredients, "\n"), strings.Join(instructions, "\n"), targetLang)

	// Add timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Generate content
	resp, err := model.GenerateContent(ctxWithTimeout, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini for translation")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	// Clean up response
	cleanedResponse := cleanJSONResponse(responseText)

	// Parse JSON response
	var translationResp struct {
		Title        string            `json:"title"`
		Ingredients  []ingredientJSON  `json:"ingredients"`
		Instructions []instructionJSON `json:"instructions"`
	}
	if err := json.Unmarshal([]byte(cleanedResponse), &translationResp); err != nil {
		return nil, fmt.Errorf("failed to parse translation response: %w", err)
	}

	// Convert to output format
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
