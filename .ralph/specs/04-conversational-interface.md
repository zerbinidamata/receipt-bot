# Specification: Conversational Interface Feature

## Overview
Make the bot conversational by understanding natural language queries instead of requiring explicit commands. Users can type naturally (e.g., "Seafood recipes" or "Salmon recipe") and the bot interprets their intent.

## Requirements

### Intent Detection

#### Supported Intents
```go
type IntentType string

const (
    IntentListRecipes      IntentType = "LIST_RECIPES"      // "show recipes", "my recipes"
    IntentFilterCategory   IntentType = "FILTER_CATEGORY"   // "seafood recipes", "pasta dishes"
    IntentFilterIngredient IntentType = "FILTER_INGREDIENT" // "salmon recipe" (specific ingredient)
    IntentMatchIngredients IntentType = "MATCH_INGREDIENTS" // "I have chicken and rice"
    IntentShowCategories   IntentType = "SHOW_CATEGORIES"   // "categories", "what types"
    IntentManagePantry     IntentType = "MANAGE_PANTRY"     // "add eggs to pantry"
    IntentHelp             IntentType = "HELP"              // "help", "how does this work"
    IntentGreeting         IntentType = "GREETING"          // "hi", "hello"
    IntentUnknown          IntentType = "UNKNOWN"           // fallback
)
```

### Port Interface
Create `internal/ports/intent.go`:
```go
type IntentDetector interface {
    DetectIntent(ctx context.Context, text string) (*Intent, error)
}

type Intent struct {
    Type        IntentType
    Category    *recipe.Category  // For FILTER_CATEGORY
    Ingredients []string          // For MATCH_INGREDIENTS
    SearchTerm  string            // For FILTER_INGREDIENT (e.g., "salmon")
    Confidence  float64
}
```

### LLM-Based Intent Detection
Create `internal/adapters/llm/intent.go`:

```go
const intentPrompt = `
Analyze the user message and determine their intent for a recipe bot.

Available categories: Pasta & Noodles, Rice & Grains, Soups & Stews, Salads,
Meat & Poultry, Seafood, Vegetarian, Desserts & Sweets, Breakfast,
Appetizers & Snacks, Beverages, Sauces & Condiments, Bread & Baking

Return JSON:
{
  "intent": "LIST_RECIPES|FILTER_CATEGORY|FILTER_INGREDIENT|MATCH_INGREDIENTS|SHOW_CATEGORIES|MANAGE_PANTRY|HELP|GREETING|UNKNOWN",
  "category": "category name or null",
  "ingredients": ["list", "of", "ingredients"] or [],
  "searchTerm": "specific ingredient to search for or null",
  "confidence": 0.0-1.0
}

Examples:
- "Seafood recipes" → {"intent": "FILTER_CATEGORY", "category": "Seafood", "confidence": 0.95}
- "Salmon recipe" → {"intent": "FILTER_INGREDIENT", "category": "Seafood", "searchTerm": "salmon", "confidence": 0.9}
- "I have chicken and pasta" → {"intent": "MATCH_INGREDIENTS", "ingredients": ["chicken", "pasta"], "confidence": 0.95}
- "Quick breakfast" → {"intent": "FILTER_CATEGORY", "category": "Breakfast", "searchTerm": "quick", "confidence": 0.85}

User message: %s
`
```

### Handler Updates
Update `internal/adapters/telegram/handlers.go`:

```go
func (h *Handler) handleTextMessage(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
    text := strings.TrimSpace(message.Text)

    // Check if it's a URL first (existing behavior)
    if isURL(text) {
        h.handleRecipeLink(ctx, chatID, userID, text)
        return
    }

    // Detect intent from natural language
    intent, err := h.intentDetector.DetectIntent(ctx, text)
    if err != nil || intent.Type == IntentUnknown || intent.Confidence < 0.5 {
        // Fall back to asking for URL
        h.promptForURL(ctx, chatID)
        return
    }

    // Route based on intent
    h.handleIntent(ctx, chatID, userID, intent)
}

func (h *Handler) handleIntent(ctx context.Context, chatID int64, userID shared.ID, intent *Intent) {
    switch intent.Type {
    case IntentListRecipes:
        h.handleListRecipesConversational(ctx, chatID, userID, nil, "")

    case IntentFilterCategory:
        h.handleListRecipesConversational(ctx, chatID, userID, intent.Category, intent.SearchTerm)

    case IntentFilterIngredient:
        h.handleFilterByIngredient(ctx, chatID, userID, intent.Category, intent.SearchTerm)

    case IntentMatchIngredients:
        h.handleMatchConversational(ctx, chatID, userID, intent.Ingredients)

    case IntentShowCategories:
        h.handleCategories(ctx, chatID, userID)

    case IntentManagePantry:
        h.handlePantryConversational(ctx, chatID, userID, intent.Ingredients)

    case IntentHelp:
        h.bot.SendMessage(ctx, chatID, FormatHelp())

    case IntentGreeting:
        h.bot.SendMessage(ctx, chatID, FormatGreeting())
    }
}
```

### New Application Query
Add to `internal/application/query/list_recipes.go`:

```go
// SearchRecipes searches recipes by title and ingredients
func (q *ListRecipesQuery) Search(ctx context.Context, userID shared.ID, searchTerm string) ([]*dto.RecipeDTO, error)

// SearchByCategory searches within a category for specific terms
func (q *ListRecipesQuery) SearchByCategory(ctx context.Context, userID shared.ID, category recipe.Category, searchTerm string) ([]*dto.RecipeDTO, error)
```

### Repository Updates
Add to `internal/domain/recipe/repository.go`:

```go
type Repository interface {
    // ... existing methods
    SearchByTitleOrIngredient(ctx context.Context, userID shared.ID, searchTerm string) ([]*Recipe, error)
    SearchInCategory(ctx context.Context, userID shared.ID, category Category, searchTerm string) ([]*Recipe, error)
}
```

## Example Flows

### Flow 1: "Seafood recipes"
1. User sends: "Seafood recipes"
2. IntentDetector returns: `{Type: FILTER_CATEGORY, Category: "Seafood"}`
3. Handler calls `listRecipesQuery.ExecuteByCategory(ctx, userID, "Seafood")`
4. Bot responds with seafood recipe list

### Flow 2: "Salmon recipe"
1. User sends: "Salmon recipe"
2. IntentDetector returns: `{Type: FILTER_INGREDIENT, Category: "Seafood", SearchTerm: "salmon"}`
3. Handler calls `listRecipesQuery.SearchByCategory(ctx, userID, "Seafood", "salmon")`
4. Bot responds with recipes containing salmon

### Flow 3: "I have chicken, rice, and garlic"
1. User sends: "I have chicken, rice, and garlic"
2. IntentDetector returns: `{Type: MATCH_INGREDIENTS, Ingredients: ["chicken", "rice", "garlic"]}`
3. Handler calls `matchIngredientsCommand.Execute(ctx, input)`
4. Bot responds with matching recipes

### Flow 4: "Quick pasta"
1. User sends: "Quick pasta"
2. IntentDetector returns: `{Type: FILTER_CATEGORY, Category: "Pasta & Noodles", SearchTerm: "quick"}`
3. Handler filters pasta recipes that have "quick" tag or <30 min cook time
4. Bot responds with quick pasta recipes

## Acceptance Criteria
- [ ] Natural language queries work without / prefix
- [ ] "Seafood recipes" returns seafood category recipes
- [ ] "Salmon recipe" filters seafood for salmon-related recipes
- [ ] "I have X, Y, Z" triggers ingredient matching
- [ ] Low confidence intents fall back to URL prompt
- [ ] Existing command behavior unchanged
- [ ] URLs still processed normally
- [ ] Unit tests for intent detection
- [ ] Integration tests for conversational flows

## Technical Notes
- Intent detection uses same LLM as recipe extraction (Gemini)
- Cache common intents to reduce LLM calls (e.g., "hi", "help")
- Confidence threshold of 0.5 for executing intents
- Fall back gracefully to URL prompt for unknown intents
- Category fuzzy matching: "fish" → "Seafood", "sweets" → "Desserts & Sweets"
