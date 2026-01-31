# Specification: Auto-Categorization Feature

## Overview
Automatically categorize recipes by dish type and ingredient composition using AI analysis during the recipe extraction process.

## Requirements

### Domain Model Updates

#### Category Enum
Add to `internal/domain/recipe/category.go`:
```go
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
```

#### Dietary Tags
```go
type DietaryTag string

const (
    TagVegetarian  DietaryTag = "vegetarian"
    TagVegan       DietaryTag = "vegan"
    TagGlutenFree  DietaryTag = "gluten-free"
    TagDairyFree   DietaryTag = "dairy-free"
    TagLowCarb     DietaryTag = "low-carb"
    TagQuick       DietaryTag = "quick"
)
```

#### Recipe Entity Updates
Add fields to `internal/domain/recipe/entity.go`:
```go
type Recipe struct {
    // ... existing fields
    Category    Category
    Cuisine     string
    DietaryTags []DietaryTag
    Tags        []string
}
```

### LLM Prompt Updates
Modify `internal/adapters/llm/prompts.go` to include categorization in the extraction prompt:
- Request category assignment from predefined list
- Request cuisine detection
- Request dietary tag identification
- Return additional JSON fields

### Firestore Schema Updates
Add fields to recipes collection:
- `category: string`
- `cuisine: string`
- `dietaryTags: []string`
- `tags: []string`

### Repository Updates
Update `internal/adapters/firebase/recipe_repository.go`:
- Store new fields
- Add query methods: `FindByCategory`, `FindByTag`, `FindByCuisine`

### Query Layer Updates
Add to `internal/application/query/`:
- `ListRecipesByCategory(userID, category)`
- `ListRecipesByTag(userID, tag)`
- `GetCategoryCounts(userID)` - returns map of category to count

### Telegram Commands
New commands in `internal/adapters/telegram/handlers.go`:
- `/recipes` - Now groups by category
- `/recipes <category>` - Filter by category (e.g., `/recipes pasta`)
- `/recipes --tag <tag>` - Filter by dietary tag
- `/categories` - List all categories with recipe counts

### Backfill Strategy
Create migration command to re-process existing recipes:
- Fetch all recipes without category
- Send to LLM for categorization only
- Update Firestore documents

## Acceptance Criteria
- [ ] New recipes are automatically categorized during extraction
- [ ] Category is displayed in recipe output
- [ ] `/recipes pasta` returns only pasta recipes
- [ ] `/categories` shows category counts
- [ ] Existing recipes can be backfilled with categories
- [ ] Unit tests for category validation
- [ ] Integration tests for category queries

## Technical Notes
- Category defaults to "Other" if LLM fails to categorize
- Dietary tags are computed from ingredients (e.g., no meat = vegetarian candidate)
- Quick tag assigned if total time < 30 minutes
