# Specification: Ingredient-Based Recipe Matching

## Overview
Allow users to send a list of ingredients they have available, and the bot suggests recipes that can be made with those ingredients.

## Requirements

### Ingredient Normalizer
Create `internal/domain/matching/normalizer.go`:
```go
type IngredientNormalizer interface {
    // Normalize extracts the base ingredient from a full ingredient string
    // "2 cups all-purpose flour" -> "flour"
    // "fresh basil leaves, chopped" -> "basil"
    Normalize(raw string) string

    // AreSimilar checks if two ingredients can substitute each other
    // "butter" and "margarine" -> true
    // "parmesan" and "romano" -> true
    AreSimilar(a, b string) bool
}
```

### Matching Algorithm
Create `internal/domain/matching/matcher.go`:
```go
type MatchLevel int

const (
    MatchLevelPerfect MatchLevel = iota  // 100% match
    MatchLevelHigh                        // 80%+ match
    MatchLevelMedium                      // 60-80% match
    MatchLevelLow                         // <60% (not shown)
)

type MatchResult struct {
    Recipe          *recipe.Recipe
    MatchPercentage float64
    MatchedItems    []string
    MissingItems    []string
    MatchLevel      MatchLevel
}

type IngredientMatcher struct {
    normalizer  IngredientNormalizer
    pantryItems map[string]bool  // common items to ignore in matching
}

func (m *IngredientMatcher) Match(
    userIngredients []string,
    recipes []*recipe.Recipe,
    options MatchOptions,
) []MatchResult
```

### Common Pantry Staples
Exclude from matching calculations:
```
salt, pepper, black pepper, oil, olive oil, vegetable oil,
water, sugar, flour, butter, garlic powder, onion powder
```

### Ingredient Substitution Groups
```go
var substitutionGroups = [][]string{
    {"milk", "cream", "half-and-half"},
    {"parmesan", "romano", "pecorino"},
    {"butter", "margarine"},
    {"chicken breast", "chicken thigh"},
    {"sugar", "honey", "maple syrup"},
    {"spaghetti", "linguine", "fettuccine", "penne"},
}
```

### User Pantry Storage
Add to user model or create separate collection:
```go
type UserPantry struct {
    UserID     shared.ID
    Items      []string  // normalized ingredient names
    UpdatedAt  time.Time
}
```

Update Firestore users collection:
- `pantryItems: []string`
- `pantryUpdatedAt: timestamp`

### Application Commands
Create `internal/application/command/match_ingredients.go`:
```go
type MatchIngredientsCommand struct {
    UserID         shared.ID
    Ingredients    []string
    CategoryFilter *recipe.Category
    StrictMatch    bool  // only perfect matches
}

type MatchIngredientsResult struct {
    PerfectMatches []dto.MatchResultDTO
    HighMatches    []dto.MatchResultDTO
    MediumMatches  []dto.MatchResultDTO
}
```

### Telegram Commands
New commands:
- `/match chicken, pasta, garlic, tomatoes` - Find matching recipes
- `/match --strict chicken, pasta` - Only perfect matches
- `/match --category pasta chicken, cream` - Filter by category
- `/pantry add butter, eggs, milk` - Save common ingredients
- `/pantry` - Show saved pantry items
- `/pantry clear` - Clear pantry

### Output Format
```
🍳 Here's what you can make:

✅ Perfect Matches (3 recipes):
1. Chicken Pasta Primavera
2. One-Pot Chicken Alfredo
3. Tomato Chicken Penne

🔸 Almost There - Missing 1-2 items (5 recipes):
4. Chicken Parmesan (missing: breadcrumbs, egg)
5. Tuscan Chicken Pasta (missing: cream, spinach)

Reply with a number to see the full recipe!
```

## Acceptance Criteria
- [ ] `/match` command parses comma-separated ingredients
- [ ] Matching algorithm returns recipes sorted by match percentage
- [ ] Perfect matches (100%) displayed first
- [ ] Missing ingredients shown for partial matches
- [ ] Pantry items stored and retrieved per user
- [ ] Common staples excluded from matching
- [ ] Similar ingredients treated as matches
- [ ] Unit tests for normalizer
- [ ] Unit tests for matcher
- [ ] Integration tests for match command

## Technical Notes
- Ingredient normalization should handle:
  - Quantities and units ("2 cups", "1 lb")
  - Adjectives ("fresh", "chopped", "minced")
  - Plurals ("tomatoes" -> "tomato")
- Consider using LLM for complex ingredient parsing if rule-based fails
- Cache normalized ingredients in recipe documents to speed up matching
