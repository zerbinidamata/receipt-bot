# Technical Requirements Specification

## Overview
This document specifies the technical requirements for Receipt-Bot's next phase features:
1. Auto-Categorization
2. Ingredient Matching
3. Export Integration (Notion & Obsidian)

---

## Feature 1: Auto-Categorization

### Domain Model Requirements

#### Category Type
```go
// Location: internal/domain/recipe/category.go
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

func (c Category) IsValid() bool
func ParseCategory(s string) (Category, error)
func AllCategories() []Category
```

#### Dietary Tags
```go
// Location: internal/domain/recipe/dietary.go
type DietaryTag string

const (
    TagVegetarian  DietaryTag = "vegetarian"
    TagVegan       DietaryTag = "vegan"
    TagGlutenFree  DietaryTag = "gluten-free"
    TagDairyFree   DietaryTag = "dairy-free"
    TagLowCarb     DietaryTag = "low-carb"
    TagQuick       DietaryTag = "quick"      // <30 min total time
    TagOnePot      DietaryTag = "one-pot"
    TagKidFriendly DietaryTag = "kid-friendly"
)

func (t DietaryTag) IsValid() bool
func ParseDietaryTag(s string) (DietaryTag, error)
```

#### Recipe Entity Updates
Add to existing Recipe struct:
```go
type Recipe struct {
    // ... existing fields ...
    Category    Category
    Cuisine     string        // e.g., "Italian", "Mexican", "Asian"
    DietaryTags []DietaryTag
    Tags        []string      // free-form tags
}
```

### LLM Prompt Requirements

Update extraction prompt to return:
```json
{
  "title": "string",
  "category": "Pasta & Noodles",
  "cuisine": "Italian",
  "dietaryTags": ["vegetarian"],
  "tags": ["comfort-food", "bake"],
  "ingredients": [...],
  "instructions": [...],
  "prepTime": 15,
  "cookTime": 30,
  "servings": 4
}
```

Category selection rules for LLM:
- Pasta/noodles as main component → "Pasta & Noodles"
- Rice/grains as main component → "Rice & Grains"
- Primarily liquid/broth-based → "Soups & Stews"
- Raw vegetables as main focus → "Salads"
- Meat/poultry as protein focus → "Meat & Poultry"
- Seafood as protein focus → "Seafood"
- No meat/fish → "Vegetarian"
- Sweet/dessert items → "Desserts & Sweets"
- Morning/brunch dishes → "Breakfast"
- Small plates/starters → "Appetizers & Snacks"
- Drinks → "Beverages"
- Condiments/dressings → "Sauces & Condiments"
- Breads/baked goods (not sweet) → "Bread & Baking"

Dietary tag detection rules:
- No meat, poultry, fish → "vegetarian"
- No animal products (no meat, dairy, eggs, honey) → "vegan"
- No wheat, barley, rye → "gluten-free"
- No milk, cheese, butter, cream → "dairy-free"
- Minimal carbs, no pasta/rice/bread → "low-carb"
- Total time (prep + cook) < 30 min → "quick"

### Database Schema Updates

Firestore `recipes` collection:
```
{
  // ... existing fields ...
  category: string,        // NEW
  cuisine: string,         // NEW
  dietaryTags: []string,   // NEW
  tags: []string           // NEW
}
```

### Query Requirements

New query methods:
```go
// ListRecipesByCategory returns recipes matching the given category
func (q *ListRecipesQuery) ByCategory(userID shared.ID, category recipe.Category) ([]*dto.RecipeDTO, error)

// ListRecipesByTag returns recipes matching the given dietary tag
func (q *ListRecipesQuery) ByTag(userID shared.ID, tag recipe.DietaryTag) ([]*dto.RecipeDTO, error)

// GetCategoryCounts returns count of recipes per category for a user
func (q *ListRecipesQuery) GetCategoryCounts(userID shared.ID) (map[recipe.Category]int, error)
```

### Telegram Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/recipes` | List all recipes (now shows category) | `/recipes` |
| `/recipes <category>` | Filter by category | `/recipes pasta` |
| `/recipes --tag <tag>` | Filter by dietary tag | `/recipes --tag vegetarian` |
| `/categories` | Show categories with counts | `/categories` |

Category shorthand mapping:
- `pasta` → "Pasta & Noodles"
- `rice` → "Rice & Grains"
- `soup`, `soups` → "Soups & Stews"
- `salad`, `salads` → "Salads"
- `meat` → "Meat & Poultry"
- `seafood`, `fish` → "Seafood"
- `vegetarian`, `veggie` → "Vegetarian"
- `dessert`, `desserts`, `sweet` → "Desserts & Sweets"
- `breakfast` → "Breakfast"
- `appetizer`, `snack` → "Appetizers & Snacks"
- `drinks`, `beverage` → "Beverages"
- `sauce`, `sauces` → "Sauces & Condiments"
- `bread`, `baking` → "Bread & Baking"

---

## Feature 2: Ingredient Matching

### Domain Model Requirements

#### Ingredient Normalizer
```go
// Location: internal/domain/matching/normalizer.go
type IngredientNormalizer interface {
    // Normalize converts raw ingredient text to normalized form
    // "2 cups all-purpose flour" → "flour"
    // "fresh basil leaves, chopped" → "basil"
    Normalize(raw string) string

    // NormalizeList normalizes a list of user-provided ingredients
    NormalizeList(ingredients []string) []string

    // AreSimilar checks if two normalized ingredients are substitutable
    // "butter" and "margarine" → true
    AreSimilar(a, b string) bool
}
```

Normalization rules:
1. Remove quantities: "2 cups", "1/2 lb", "200g"
2. Remove units: "cups", "tablespoons", "oz", "grams"
3. Remove preparations: "chopped", "diced", "minced", "fresh", "dried"
4. Remove descriptors: "large", "small", "organic", "boneless"
5. Singularize: "tomatoes" → "tomato"
6. Lowercase everything
7. Handle common variations: "all-purpose flour" → "flour"

#### Matcher
```go
// Location: internal/domain/matching/matcher.go
type MatchResult struct {
    Recipe          *recipe.Recipe
    MatchPercentage float64      // 0.0 to 1.0
    MatchedItems    []string     // ingredients user has
    MissingItems    []string     // ingredients user needs
    MatchLevel      MatchLevel
}

type MatchLevel int
const (
    MatchLevelPerfect MatchLevel = iota  // 100%
    MatchLevelHigh                        // 80-99%
    MatchLevelMedium                      // 60-79%
    MatchLevelLow                         // <60%, not returned
)

type IngredientMatcher struct {
    normalizer    IngredientNormalizer
    pantryStaples map[string]bool  // items to ignore
}

func (m *IngredientMatcher) Match(
    userIngredients []string,
    recipes []*recipe.Recipe,
    options MatchOptions,
) []MatchResult

type MatchOptions struct {
    MinMatchPercentage float64   // default 0.6
    IncludePantryItems bool      // whether to count pantry staples
    CategoryFilter     *Category // optional category filter
    MaxResults         int       // default 20
}
```

#### Pantry Staples
Items excluded from matching calculation by default:
```
salt, pepper, black pepper, white pepper,
oil, olive oil, vegetable oil, cooking oil, canola oil,
water, ice,
sugar, brown sugar, powdered sugar,
flour, all-purpose flour,
butter, margarine,
garlic powder, onion powder, paprika,
baking soda, baking powder, yeast,
vanilla extract, vanilla,
soy sauce, vinegar
```

#### Substitution Groups
Items that can substitute for each other:
```go
var substitutionGroups = [][]string{
    {"butter", "margarine", "oil"},
    {"milk", "cream", "half-and-half", "oat milk", "almond milk"},
    {"parmesan", "romano", "pecorino", "asiago"},
    {"cheddar", "monterey jack", "colby"},
    {"chicken breast", "chicken thigh", "chicken"},
    {"ground beef", "ground turkey", "ground pork"},
    {"spaghetti", "linguine", "fettuccine", "angel hair"},
    {"penne", "rigatoni", "ziti"},
    {"lemon juice", "lime juice", "citrus juice"},
    {"sugar", "honey", "maple syrup", "agave"},
    {"vegetable broth", "chicken broth", "beef broth", "stock"},
}
```

### User Pantry

Add to User entity:
```go
type User struct {
    // ... existing fields ...
    PantryItems   []string   // normalized ingredient names
    PantryUpdated time.Time
}
```

Firestore `users` collection update:
```
{
  // ... existing fields ...
  pantryItems: []string,
  pantryUpdatedAt: timestamp
}
```

### Application Layer

```go
// Location: internal/application/command/match_ingredients.go
type MatchIngredientsCommand struct {
    UserID          shared.ID
    Ingredients     []string
    CategoryFilter  *recipe.Category
    StrictMatch     bool  // only perfect matches
}

type MatchIngredientsHandler struct {
    recipeRepo  recipe.Repository
    userRepo    user.Repository
    matcher     *matching.IngredientMatcher
}

func (h *MatchIngredientsHandler) Handle(ctx context.Context, cmd MatchIngredientsCommand) (*MatchIngredientsResult, error)

type MatchIngredientsResult struct {
    PerfectMatches []dto.MatchResultDTO
    HighMatches    []dto.MatchResultDTO
    MediumMatches  []dto.MatchResultDTO
    TotalRecipes   int
}
```

### Telegram Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/match <ingredients>` | Find recipes with ingredients | `/match chicken, pasta, garlic` |
| `/match --strict <ingredients>` | Only perfect matches | `/match --strict eggs, cheese` |
| `/match --category <cat> <ingredients>` | Filter by category | `/match --category pasta chicken, cream` |
| `/pantry` | Show saved pantry items | `/pantry` |
| `/pantry add <items>` | Add to pantry | `/pantry add butter, eggs, milk` |
| `/pantry remove <items>` | Remove from pantry | `/pantry remove eggs` |
| `/pantry clear` | Clear all pantry items | `/pantry clear` |

### Match Results Display Format

```
🍳 What You Can Make

✅ Perfect Matches (3):
1. Chicken Pasta Primavera
2. One-Pot Chicken Alfredo
3. Tomato Chicken Penne

🔸 Almost There (5):
4. Chicken Parmesan
   Missing: breadcrumbs, egg
5. Tuscan Chicken Pasta
   Missing: cream, spinach

Reply with number to see recipe!
```

---

## Feature 3: Export Integration

### Obsidian Export

#### Markdown Format
```markdown
---
title: Recipe Title
category: Pasta & Noodles
cuisine: Italian
prep_time: 15
cook_time: 30
servings: 4
source: https://tiktok.com/...
platform: TikTok
author: @username
created: 2026-01-23
tags:
  - recipe
  - pasta
  - italian
  - vegetarian
---

# Recipe Title

## Info
- **Prep Time:** 15 minutes
- **Cook Time:** 30 minutes
- **Servings:** 4
- **Category:** Pasta & Noodles
- **Cuisine:** Italian

## Ingredients
- 200g pasta
- 2 cloves garlic, minced
- 1 cup tomato sauce

## Instructions
1. Boil pasta according to package directions
2. Sauté garlic in olive oil
3. Add tomato sauce and simmer

## Source
[Original Recipe](https://tiktok.com/...) by @username
```

#### Exporter Interface
```go
// Location: internal/ports/exporter.go
type ObsidianExporter interface {
    // ExportAsMarkdown generates markdown for a single recipe
    ExportAsMarkdown(recipe *recipe.Recipe) ([]byte, error)

    // ExportBulkAsZip generates a ZIP file with all recipes
    ExportBulkAsZip(recipes []*recipe.Recipe) ([]byte, error)
}
```

#### File Naming
Single recipe: `{sanitized-title}.md`
- Replace spaces with hyphens
- Remove special characters
- Lowercase
- Example: "Chicken Tikka Masala" → `chicken-tikka-masala.md`

Bulk export: `recipes-export-{timestamp}.zip`
- Example: `recipes-export-2026-01-23.zip`

### Notion Export

#### Authentication
- Use Notion Public OAuth 2.0
- Redirect URI: `https://your-domain.com/notion/callback` (or Telegram deep link)
- Store access token encrypted in Firestore

#### Database Schema
Create "Recipes" database with properties:
| Property | Type | Description |
|----------|------|-------------|
| Name | title | Recipe title |
| Category | select | Recipe category |
| Cuisine | select | Cuisine type |
| Prep Time | number | Minutes |
| Cook Time | number | Minutes |
| Servings | number | Number of servings |
| Source URL | url | Original recipe URL |
| Platform | select | TikTok, YouTube, etc. |
| Tags | multi_select | Dietary and other tags |
| Created | date | When recipe was saved |

#### Page Content Structure
```
# Recipe Title

## Ingredients
• 200g pasta
• 2 cloves garlic

## Instructions
1. Step one
2. Step two

---
Source: [Original Recipe](url) by @author
```

#### Exporter Interface
```go
// Location: internal/ports/exporter.go
type NotionExporter interface {
    // Connect initiates OAuth flow
    Connect(userID shared.ID, redirectURL string) (authURL string, err error)

    // CompleteAuth exchanges code for token
    CompleteAuth(userID shared.ID, code string) error

    // IsConnected checks if user has valid Notion connection
    IsConnected(userID shared.ID) bool

    // Disconnect removes Notion integration
    Disconnect(userID shared.ID) error

    // ExportRecipe exports single recipe to Notion
    ExportRecipe(ctx context.Context, userID shared.ID, recipe *recipe.Recipe) (pageURL string, err error)

    // ExportBulk exports multiple recipes
    ExportBulk(ctx context.Context, userID shared.ID, recipes []*recipe.Recipe) (databaseURL string, err error)
}
```

#### User Schema Updates
Firestore `users` collection:
```
{
  // ... existing fields ...
  notionAccessToken: string,      // encrypted
  notionWorkspaceId: string,
  notionDatabaseId: string,
  notionConnectedAt: timestamp
}
```

### Telegram Commands

| Command | Description |
|---------|-------------|
| `/export obsidian <number>` | Export single recipe as .md file |
| `/export obsidian all` | Export all recipes as .zip |
| `/connect notion` | Start Notion OAuth flow |
| `/export notion <number>` | Export single recipe to Notion |
| `/export notion all` | Export all recipes to Notion |
| `/disconnect notion` | Remove Notion integration |

### Error Handling

| Scenario | User Message |
|----------|--------------|
| Notion not connected | "Please connect Notion first with /connect notion" |
| Notion token expired | "Your Notion connection expired. Please reconnect with /connect notion" |
| Recipe not found | "Recipe #X not found. Use /recipes to see your recipes." |
| Export failed | "Export failed: {reason}. Please try again." |
| Rate limited | "Too many exports. Please wait a few minutes." |

---

## Non-Functional Requirements

### Performance
- Category queries: < 500ms for up to 1000 recipes
- Ingredient matching: < 2s for matching against 100 recipes
- Export generation: < 5s for single recipe, < 30s for bulk (up to 100)

### Security
- Notion tokens encrypted at rest (AES-256)
- No storage of Notion OAuth secrets in code
- Rate limit exports: 10/minute per user

### Reliability
- Retry failed Notion API calls (3 attempts, exponential backoff)
- Graceful degradation if categorization fails (use "Other")
- Validate all user input before processing

### Backwards Compatibility
- Existing recipes without categories display as "Uncategorized"
- Migration script to backfill categories for existing recipes
- All new fields optional in database schema
