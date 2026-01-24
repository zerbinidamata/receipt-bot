# Product Requirements Document (PRD)
## Receipt-Bot: Next Phase Features

**Version:** 1.0
**Date:** January 2026
**Status:** Draft

---

## Executive Summary

This PRD outlines the next phase of development for Receipt-Bot, focusing on three major feature areas:
1. **Export Integration** - Export recipes to Notion and Obsidian
2. **Auto-Categorization** - Intelligent categorization by ingredients and dish type
3. **Ingredient Matching** - Find recipes based on available ingredients

---

## Current State

Receipt-Bot is a Telegram bot that:
- Extracts recipes from TikTok, YouTube, Instagram, and web pages
- Uses AI (Gemini/OpenAI) to structure recipe data
- Stores recipes in Firebase Firestore
- Provides recipe listing and retrieval via Telegram commands

**Tech Stack:** Go (bot/orchestration) + Python (scraping/transcription) + Firebase + gRPC

---

## Feature 1: Export to Notion & Obsidian

### Overview
Allow users to export their saved recipes to external knowledge management tools.

### User Stories

| ID | Story | Priority |
|----|-------|----------|
| E1 | As a user, I want to export a single recipe to Notion so I can organize it with my other notes | High |
| E2 | As a user, I want to export a single recipe to Obsidian so I can store it in my local vault | High |
| E3 | As a user, I want to bulk export all my recipes to Notion/Obsidian | Medium |
| E4 | As a user, I want to configure my Notion/Obsidian connection once and reuse it | High |

### Functional Requirements

#### Notion Integration
- **Authentication**: OAuth 2.0 flow with Notion API
- **Database Creation**: Auto-create a "Recipes" database if not exists
- **Page Structure**:
  ```
  Title: Recipe Name
  Properties:
    - Category (select)
    - Prep Time (number)
    - Cook Time (number)
    - Servings (number)
    - Source URL (url)
    - Source Platform (select: TikTok, YouTube, Instagram, Web)
    - Created At (date)
  Content:
    - Ingredients section (bulleted list)
    - Instructions section (numbered list)
    - Original source link
  ```
- **Duplicate Handling**: Check by source URL before creating new page

#### Obsidian Integration
- **Export Format**: Markdown files with YAML frontmatter
- **File Structure**:
  ```markdown
  ---
  title: Recipe Name
  category: Pasta
  prep_time: 15
  cook_time: 30
  servings: 4
  source: https://...
  platform: TikTok
  created: 2026-01-23
  tags: [recipe, pasta, italian]
  ---

  # Recipe Name

  ## Ingredients
  - 200g pasta
  - ...

  ## Instructions
  1. Step one
  2. ...

  ## Source
  [Original Recipe](https://...)
  ```
- **Delivery Method**:
  - Option A: Generate downloadable .md file via Telegram
  - Option B: Direct sync via Obsidian plugin/local folder (advanced)

### New Telegram Commands
```
/export notion <recipe_number>     - Export single recipe to Notion
/export obsidian <recipe_number>   - Export single recipe as Markdown
/export notion all                 - Export all recipes to Notion
/export obsidian all               - Export all recipes as ZIP
/connect notion                    - Set up Notion integration
/disconnect notion                 - Remove Notion integration
```

### Technical Implementation

#### New Domain Concepts
```go
// internal/domain/export/
type ExportFormat string
const (
    ExportFormatNotion   ExportFormat = "notion"
    ExportFormatObsidian ExportFormat = "obsidian"
)

type ExportRequest struct {
    UserID    shared.ID
    RecipeIDs []shared.ID  // empty = all recipes
    Format    ExportFormat
}

type ExportResult struct {
    Success   bool
    ExportURL string  // Notion page URL or file download URL
    Errors    []error
}
```

#### New Ports
```go
// internal/ports/exporter.go
type NotionExporter interface {
    Connect(userID shared.ID, authCode string) error
    ExportRecipe(ctx context.Context, recipe *recipe.Recipe) (string, error)
    ExportBulk(ctx context.Context, recipes []*recipe.Recipe) (string, error)
}

type ObsidianExporter interface {
    ExportAsMarkdown(recipe *recipe.Recipe) ([]byte, error)
    ExportBulkAsZip(recipes []*recipe.Recipe) ([]byte, error)
}
```

#### New Adapters
```
internal/adapters/notion/
  - client.go           # Notion API client
  - oauth.go            # OAuth flow handling
  - exporter.go         # NotionExporter implementation

internal/adapters/obsidian/
  - markdown.go         # Markdown generation
  - exporter.go         # ObsidianExporter implementation
```

#### Database Schema Updates
```
users collection (add fields):
  notionAccessToken: string (encrypted)
  notionDatabaseId: string
  notionConnectedAt: timestamp
```

---

## Feature 2: Auto-Categorization

### Overview
Automatically categorize recipes by dish type and ingredient composition using AI analysis.

### User Stories

| ID | Story | Priority |
|----|-------|----------|
| C1 | As a user, I want my recipes auto-categorized so I can browse by type | High |
| C2 | As a user, I want to filter recipes by category (pasta, soup, dessert) | High |
| C3 | As a user, I want to see ingredient-based tags (vegetarian, gluten-free) | Medium |
| C4 | As a user, I want to manually override categories if the AI got it wrong | Low |

### Category Taxonomy

#### Primary Categories (Dish Type)
```
- Pasta & Noodles
- Rice & Grains
- Soups & Stews
- Salads
- Meat & Poultry
- Seafood
- Vegetarian
- Desserts & Sweets
- Breakfast
- Appetizers & Snacks
- Beverages
- Sauces & Condiments
- Bread & Baking
```

#### Secondary Tags (Dietary/Ingredient-based)
```
- Vegetarian
- Vegan
- Gluten-Free
- Dairy-Free
- Low-Carb
- Quick (<30 min)
- One-Pot
- Kid-Friendly
```

#### Cuisine Tags
```
- Italian, Mexican, Asian, Mediterranean, American, French, Indian, etc.
```

### Categorization Logic

**AI-Based Classification:**
The existing LLM extraction will be enhanced to include categorization:

```json
{
  "title": "Homemade Lasagna",
  "category": "Pasta & Noodles",
  "cuisine": "Italian",
  "tags": ["comfort-food", "bake"],
  "dietary": ["contains-gluten", "contains-dairy"],
  "ingredients": [...],
  "instructions": [...]
}
```

**Ingredient-Based Rules:**
```
- Contains pasta/noodles → Pasta & Noodles
- Contains rice/quinoa/couscous → Rice & Grains
- Main protein is fish/shrimp → Seafood
- No meat/fish + no dairy → Vegan candidate
- Sugar + flour + butter → Desserts candidate
```

### New Telegram Commands
```
/recipes                           - List all (now grouped by category)
/recipes pasta                     - Filter by category
/recipes --tag vegetarian          - Filter by tag
/categories                        - List all categories with counts
/recategorize <recipe_number>      - Trigger re-categorization
```

### Technical Implementation

#### Domain Model Updates
```go
// internal/domain/recipe/category.go
type Category string
const (
    CategoryPasta       Category = "Pasta & Noodles"
    CategoryRice        Category = "Rice & Grains"
    CategorySoups       Category = "Soups & Stews"
    CategorySalads      Category = "Salads"
    CategoryMeat        Category = "Meat & Poultry"
    CategorySeafood     Category = "Seafood"
    CategoryVegetarian  Category = "Vegetarian"
    CategoryDesserts    Category = "Desserts & Sweets"
    CategoryBreakfast   Category = "Breakfast"
    CategoryAppetizers  Category = "Appetizers & Snacks"
    CategoryBeverages   Category = "Beverages"
    CategorySauces      Category = "Sauces & Condiments"
    CategoryBread       Category = "Bread & Baking"
    CategoryOther       Category = "Other"
)

type DietaryTag string
const (
    TagVegetarian  DietaryTag = "vegetarian"
    TagVegan       DietaryTag = "vegan"
    TagGlutenFree  DietaryTag = "gluten-free"
    TagDairyFree   DietaryTag = "dairy-free"
    TagLowCarb     DietaryTag = "low-carb"
    TagQuick       DietaryTag = "quick"
)

// Recipe entity additions
type Recipe struct {
    // ... existing fields
    Category    Category
    Cuisine     string
    DietaryTags []DietaryTag
    Tags        []string
}
```

#### LLM Prompt Updates
Update `internal/adapters/llm/prompts.go` to request categorization:

```go
const extractionPromptWithCategories = `
Extract the recipe and categorize it.

Categories (pick one):
- Pasta & Noodles
- Rice & Grains
- Soups & Stews
...

Dietary tags (pick all that apply):
- vegetarian (no meat/fish)
- vegan (no animal products)
- gluten-free
- dairy-free
...

Return JSON with these additional fields:
{
  "category": "...",
  "cuisine": "...",
  "dietaryTags": ["...", "..."],
  "tags": ["...", "..."]
}
`
```

#### Database Schema Updates
```
recipes collection (add fields):
  category: string
  cuisine: string
  dietaryTags: []string
  tags: []string
```

#### New Query Methods
```go
// internal/application/query/
func (q *ListRecipesQuery) ByCategory(category recipe.Category) ([]*dto.RecipeDTO, error)
func (q *ListRecipesQuery) ByTag(tag recipe.DietaryTag) ([]*dto.RecipeDTO, error)
func (q *ListRecipesQuery) ByCuisine(cuisine string) ([]*dto.RecipeDTO, error)
```

---

## Feature 3: Ingredient-Based Recipe Matching

### Overview
Users can send a list of ingredients they have, and the bot suggests recipes that can be made with those ingredients.

### User Stories

| ID | Story | Priority |
|----|-------|----------|
| M1 | As a user, I want to list my available ingredients and see what I can cook | High |
| M2 | As a user, I want to see partial matches (recipes where I have most ingredients) | High |
| M3 | As a user, I want to know which ingredients I'm missing for each recipe | Medium |
| M4 | As a user, I want to filter matches by category | Low |

### Matching Algorithm

**Ingredient Normalization:**
Before matching, ingredients must be normalized:
```
"2 cups flour" → "flour"
"fresh basil leaves" → "basil"
"1 lb ground beef" → "ground beef"
"chicken breast, boneless" → "chicken breast"
```

**Matching Levels:**
1. **Perfect Match**: User has all required ingredients
2. **High Match (80%+)**: Missing 1-2 ingredients
3. **Medium Match (60-80%)**: Missing 3-4 ingredients
4. **Low Match (<60%)**: Missing too many ingredients (not shown)

**Smart Matching:**
- Treat similar ingredients as substitutes (butter ↔ margarine, any cheese type)
- Ignore common pantry staples (salt, pepper, oil, water) in matching calculation
- Consider "main ingredients" vs "minor ingredients" weighting

### User Interaction Flow

```
User: "I have chicken, pasta, garlic, tomatoes, onion, cheese"

Bot: "🍳 Here's what you can make:

✅ Perfect Matches (3 recipes):
1. Chicken Pasta Primavera
2. One-Pot Chicken Alfredo
3. Tomato Chicken Penne

🔸 Almost There - Missing 1-2 items (5 recipes):
4. Chicken Parmesan (missing: breadcrumbs, egg)
5. Tuscan Chicken Pasta (missing: cream, spinach)

Reply with a number to see the full recipe!"
```

### New Telegram Commands
```
/match chicken, pasta, garlic, tomatoes    - Find recipes with these ingredients
/match --strict chicken, pasta             - Only perfect matches
/match --category pasta chicken, cream     - Filter by category
/pantry add butter, eggs, milk             - Save common ingredients
/pantry                                    - Show saved pantry items
/pantry clear                              - Clear pantry
```

### Technical Implementation

#### New Domain Concepts
```go
// internal/domain/matching/
type IngredientMatcher struct {
    normalizer IngredientNormalizer
    pantryItems map[string]bool  // common items to ignore
}

type MatchResult struct {
    Recipe          *recipe.Recipe
    MatchPercentage float64
    MatchedItems    []string
    MissingItems    []string
    MatchLevel      MatchLevel
}

type MatchLevel int
const (
    MatchLevelPerfect MatchLevel = iota
    MatchLevelHigh
    MatchLevelMedium
    MatchLevelLow
)

func (m *IngredientMatcher) Match(
    userIngredients []string,
    recipes []*recipe.Recipe,
) []MatchResult
```

#### Ingredient Normalization
```go
// internal/domain/matching/normalizer.go
type IngredientNormalizer interface {
    Normalize(raw string) string
    AreSimilar(a, b string) bool
}

// Uses stemming, removes quantities/units, handles plurals
// Could use LLM for complex cases
```

#### New Application Commands
```go
// internal/application/command/match_ingredients.go
type MatchIngredientsCommand struct {
    userID       shared.ID
    ingredients  []string
    categoryFilter *recipe.Category
    strictMatch  bool
}

type MatchIngredientsResult struct {
    PerfectMatches []dto.MatchResultDTO
    HighMatches    []dto.MatchResultDTO
    MediumMatches  []dto.MatchResultDTO
}
```

#### User Pantry Storage
```go
// Add to user entity or separate collection
type UserPantry struct {
    UserID     shared.ID
    Items      []string  // normalized ingredient names
    UpdatedAt  time.Time
}
```

#### Database Schema Updates
```
users collection (add fields):
  pantryItems: []string
  pantryUpdatedAt: timestamp
```

---

## Implementation Phases

### Phase 1: Auto-Categorization
**Rationale**: Foundation for other features, low risk, immediate value

1. Update domain model with Category, DietaryTags
2. Modify LLM prompts to include categorization
3. Update Firestore schema and repository
4. Add category to recipe display formatting
5. Implement filter commands (/recipes pasta)
6. Backfill categories for existing recipes

### Phase 2: Ingredient Matching
**Rationale**: High user value, builds on categorization

1. Implement ingredient normalizer
2. Build matching algorithm with scoring
3. Add pantry storage to user model
4. Create match command handler
5. Implement Telegram formatting for match results
6. Add pantry management commands

### Phase 3: Export Integration
**Rationale**: More complex (external APIs), can be developed in parallel

1. Implement Obsidian markdown exporter (simpler)
2. Add file download capability to Telegram handler
3. Set up Notion OAuth flow
4. Implement Notion API client
5. Add Notion exporter with database creation
6. Implement bulk export functionality

---

## Technical Considerations

### Performance
- **Ingredient matching**: Index ingredients in Firestore for efficient queries
- **Bulk export**: Process in background, send notification when ready
- **Caching**: Cache category counts, pantry items

### Security
- **Notion tokens**: Encrypt at rest, use short-lived access tokens
- **Rate limiting**: Limit export frequency to prevent abuse

### Error Handling
- **Notion API failures**: Retry with exponential backoff
- **Categorization failures**: Default to "Other" category
- **Matching edge cases**: Handle empty recipes, missing ingredients gracefully

### Testing Strategy
- Unit tests for normalizer, matcher, categorizer
- Integration tests for Notion/Obsidian exporters
- E2E tests for new Telegram commands

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Categorization accuracy | >90% user agreement |
| Match relevance | >80% of "perfect matches" are actually cookable |
| Export success rate | >95% |
| Feature adoption | >50% of active users try new features |

---

## Open Questions

1. **Notion OAuth**: Should we use Notion's internal integration or public OAuth?
2. **Obsidian sync**: Is file download sufficient or do users want direct vault sync?
3. **Category customization**: Should users be able to create custom categories?
4. **Ingredient synonyms**: Build our own database or use external API?
5. **Partial ingredient matching**: How to handle "chicken" matching "chicken breast"?

---

## Appendix

### A. Category Mapping Examples

| Recipe | Primary Category | Cuisine | Tags |
|--------|-----------------|---------|------|
| Homemade Lasagna | Pasta & Noodles | Italian | comfort-food, bake |
| Chicken Tikka Masala | Meat & Poultry | Indian | curry, spicy |
| Caesar Salad | Salads | American | quick, lunch |
| Chocolate Lava Cake | Desserts & Sweets | French | chocolate, bake |
| Ramen | Pasta & Noodles | Japanese | soup, noodles |

### B. Common Pantry Staples (Excluded from Matching)
```
salt, pepper, black pepper, oil, olive oil, vegetable oil,
water, sugar, flour, butter, garlic powder, onion powder
```

### C. Ingredient Substitution Groups
```
Dairy: milk, cream, half-and-half
Cheese: parmesan, romano, pecorino
Pasta: spaghetti, linguine, fettuccine (shape-interchangeable)
Protein: chicken breast, chicken thigh (same animal)
Sweetener: sugar, honey, maple syrup
```
