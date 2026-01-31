package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/ports"
)

// IntentPrompt is the system prompt for intent detection
const IntentPrompt = `You are an intent detection assistant for a recipe bot. Analyze the user message and determine their intent.

IMPORTANT: The user may write in English OR Portuguese (Brazilian). You MUST understand both languages.

The bot supports these intents:
- LIST_RECIPES: User wants to see their recipes
  EN: "show recipes", "my recipes", "recipe list", "what recipes do I have"
  PT: "mostrar receitas", "minhas receitas", "lista de receitas", "quais receitas eu tenho"
- FILTER_CATEGORY: User wants to filter recipes by category ONLY
  EN: "seafood recipes", "pasta dishes", "breakfast ideas", "show me desserts"
  PT: "receitas de frutos do mar", "pratos de massa", "ideias de café da manhã", "mostrar sobremesas"
- FILTER_INGREDIENT: User wants to find recipes containing a specific ingredient
  EN: "salmon recipe", "chicken dishes", "recipes with beef"
  PT: "receita de salmão", "pratos com frango", "receitas com carne"
- MATCH_INGREDIENTS: User lists ingredients they have and wants matching recipes
  EN: "I have chicken, pasta, and garlic", "what can I make with rice and beans"
  PT: "tenho frango, macarrão e alho", "o que posso fazer com arroz e feijão"
- SHOW_CATEGORIES: User wants to see available categories
  EN: "categories", "what types do I have", "show categories"
  PT: "categorias", "quais tipos eu tenho", "mostrar categorias"
- MANAGE_PANTRY: User wants to manage their pantry
  EN: "add chicken to pantry", "my pantry", "remove eggs from pantry", "clear my pantry"
  PT: "adicionar frango à despensa", "minha despensa", "remover ovos da despensa", "limpar minha despensa"
- HELP: User needs help
  EN: "help", "how does this work", "what can you do"
  PT: "ajuda", "como funciona", "o que você pode fazer"
- GREETING: User is greeting
  EN: "hi", "hello", "hey", "good morning"
  PT: "oi", "olá", "e aí", "bom dia"
- SHOW_MORE: User wants to see more results from previous query
  EN: "show more", "next", "more recipes", "continue"
  PT: "mostrar mais", "próximo", "mais receitas", "continuar"
- SHOW_DETAILS: User wants to see details of a specific recipe from results
  EN: "show me #3", "details on the first one", "tell me about number 2"
  PT: "mostrar #3", "detalhes do primeiro", "falar sobre o número 2"
- REPEAT_LAST: User wants to repeat the last action
  EN: "show again", "repeat", "one more time"
  PT: "mostrar de novo", "repetir", "mais uma vez"
- COMPOUND_QUERY: User combines a category with dietary/tag filters
  EN: "quick pasta recipes", "vegan breakfast", "easy seafood"
  PT: "receitas rápidas de massa", "café da manhã vegano", "frutos do mar fácil"
- UNKNOWN: Cannot determine intent

Available recipe categories (use English names in response):
Pasta & Noodles, Rice & Grains, Soups & Stews, Salads, Meat & Poultry, Seafood, Vegetarian, Desserts & Sweets, Breakfast, Appetizers & Snacks, Beverages, Sauces & Condiments, Bread & Baking

Portuguese category mappings:
- massas/macarrão/pasta -> Pasta & Noodles
- arroz/grãos -> Rice & Grains
- sopas/ensopados/caldos -> Soups & Stews
- saladas -> Salads
- carnes/aves/frango -> Meat & Poultry
- frutos do mar/peixe/camarão -> Seafood
- vegetariano -> Vegetarian
- sobremesas/doces -> Desserts & Sweets
- café da manhã -> Breakfast
- aperitivos/lanches/petiscos -> Appetizers & Snacks
- bebidas -> Beverages
- molhos/condimentos -> Sauces & Condiments
- pães/assados -> Bread & Baking

Available dietary/modifier tags (use English names in response):
vegetarian, vegan, gluten-free, dairy-free, low-carb, quick, one-pot, kid-friendly

Portuguese tag mappings:
- vegetariano -> vegetarian
- vegano -> vegan
- sem glúten -> gluten-free
- sem lactose/sem leite -> dairy-free
- low-carb/baixo carboidrato -> low-carb
- rápido/fácil -> quick
- panela única -> one-pot
- para crianças -> kid-friendly

Response format - return ONLY valid JSON:
{
  "intent": "INTENT_TYPE",
  "category": "category name in English or null",
  "dietaryTags": ["tag1", "tag2"] or [],
  "ingredients": ["list", "of", "ingredients"] or [],
  "searchTerm": "specific ingredient to filter by or null",
  "pantryAction": "SHOW|ADD|REMOVE|CLEAR or null",
  "pantryItems": ["items", "to", "add/remove"] or [],
  "recipeNumber": number or null,
  "confidence": 0.0-1.0
}

Rules:
- ALWAYS return category names in ENGLISH regardless of input language
- For FILTER_CATEGORY: Set "category" to the closest matching category from the list (NO dietary tags)
- For COMPOUND_QUERY: Set BOTH "category" AND "dietaryTags" when user combines them
- For FILTER_INGREDIENT: Set "searchTerm" to the specific ingredient (in the language user provided)
- For MATCH_INGREDIENTS: Extract all ingredients mentioned into "ingredients" array
- For MANAGE_PANTRY: Set "pantryAction" and "pantryItems" if adding/removing
- For SHOW_DETAILS: Set "recipeNumber" to the 1-based index
- Confidence should be 0.9+ for clear intents, 0.7-0.9 for likely matches, below 0.7 for uncertain
- If a message mentions a specific food item but doesn't say "I have"/"tenho", treat it as FILTER_INGREDIENT`

// IntentDetectorAdapter implements IntentDetector using LLM
type IntentDetectorAdapter struct {
	client *genai.Client
	model  string
}

// NewIntentDetectorAdapter creates a new intent detector adapter
func NewIntentDetectorAdapter(client *genai.Client, model string) *IntentDetectorAdapter {
	return &IntentDetectorAdapter{
		client: client,
		model:  model,
	}
}

// intentResponse represents the JSON response from the LLM
type intentResponse struct {
	Intent       string   `json:"intent"`
	Category     *string  `json:"category"`
	DietaryTags  []string `json:"dietaryTags"`
	Ingredients  []string `json:"ingredients"`
	SearchTerm   *string  `json:"searchTerm"`
	PantryAction *string  `json:"pantryAction"`
	PantryItems  []string `json:"pantryItems"`
	RecipeNumber *int     `json:"recipeNumber"`
	Confidence   float64  `json:"confidence"`
}

// DetectIntent implements the IntentDetector interface
func (a *IntentDetectorAdapter) DetectIntent(ctx context.Context, text string) (*ports.Intent, error) {
	model := a.client.GenerativeModel(a.model)

	// Configure model for JSON output
	model.SetTemperature(0.2) // Low temperature for deterministic output
	model.ResponseMIMEType = "application/json"

	// Build the prompt
	prompt := fmt.Sprintf("%s\n\nUser message: %s", IntentPrompt, text)

	// Add timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Generate content
	resp, err := model.GenerateContent(ctxWithTimeout, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("intent detection failed: %w", err)
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from LLM for intent detection")
	}

	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	// Clean up response
	cleanedResponse := cleanIntentResponse(responseText)

	// Parse JSON response
	var intentResp intentResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &intentResp); err != nil {
		return nil, fmt.Errorf("failed to parse intent response: %w", err)
	}

	// Convert to domain Intent
	intent := convertToIntent(&intentResp, text)

	return intent, nil
}

// cleanIntentResponse removes markdown code blocks and extra text
func cleanIntentResponse(response string) string {
	// Remove markdown code blocks
	codeBlockRegex := regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")
	matches := codeBlockRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		response = matches[1]
	}

	response = strings.TrimSpace(response)

	// Find JSON object boundaries
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		response = response[startIdx : endIdx+1]
	}

	return response
}

// convertToIntent converts the LLM response to a domain Intent
func convertToIntent(resp *intentResponse, rawText string) *ports.Intent {
	intent := &ports.Intent{
		Type:        parseIntentType(resp.Intent),
		Ingredients: resp.Ingredients,
		Confidence:  resp.Confidence,
		RawResponse: rawText,
	}

	// Handle category
	if resp.Category != nil && *resp.Category != "" {
		cat := recipe.ParseCategory(*resp.Category)
		intent.Category = &cat
	}

	// Handle dietary tags
	if len(resp.DietaryTags) > 0 {
		intent.DietaryTags = recipe.ParseDietaryTags(resp.DietaryTags)
	}

	// Handle search term
	if resp.SearchTerm != nil && *resp.SearchTerm != "" {
		intent.SearchTerm = *resp.SearchTerm
	}

	// Handle pantry action
	if resp.PantryAction != nil && *resp.PantryAction != "" {
		intent.PantryAction = parsePantryAction(*resp.PantryAction)
		intent.PantryItems = resp.PantryItems
	}

	// Handle recipe number for SHOW_DETAILS
	if resp.RecipeNumber != nil && *resp.RecipeNumber > 0 {
		intent.RecipeNumber = *resp.RecipeNumber
	}

	return intent
}

// parseIntentType converts a string to IntentType
func parseIntentType(s string) ports.IntentType {
	switch strings.ToUpper(s) {
	case "LIST_RECIPES":
		return ports.IntentListRecipes
	case "FILTER_CATEGORY":
		return ports.IntentFilterCategory
	case "FILTER_INGREDIENT":
		return ports.IntentFilterIngredient
	case "MATCH_INGREDIENTS":
		return ports.IntentMatchIngredients
	case "SHOW_CATEGORIES":
		return ports.IntentShowCategories
	case "MANAGE_PANTRY":
		return ports.IntentManagePantry
	case "HELP":
		return ports.IntentHelp
	case "GREETING":
		return ports.IntentGreeting
	case "SHOW_MORE":
		return ports.IntentShowMore
	case "SHOW_DETAILS":
		return ports.IntentShowDetails
	case "REPEAT_LAST":
		return ports.IntentRepeatLast
	case "COMPOUND_QUERY":
		return ports.IntentCompoundQuery
	default:
		return ports.IntentUnknown
	}
}

// parsePantryAction converts a string to PantryAction
func parsePantryAction(s string) ports.PantryAction {
	switch strings.ToUpper(s) {
	case "ADD":
		return ports.PantryActionAdd
	case "REMOVE":
		return ports.PantryActionRemove
	case "CLEAR":
		return ports.PantryActionClear
	default:
		return ports.PantryActionShow
	}
}
