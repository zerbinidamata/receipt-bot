# Architecture Specification

## Overview

Receipt-Bot follows **Hexagonal Architecture** (Ports & Adapters) with **Domain-Driven Design** principles. This document describes how new features should integrate with the existing architecture.

---

## Current Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        TELEGRAM BOT                              │
│                     (User Interface)                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    APPLICATION LAYER                             │
│              Commands (Write) & Queries (Read)                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │ProcessRecipeLink│  │ListRecipes      │  │FindRecipe       │  │
│  │GetOrCreateUser  │  │ByCategory (NEW) │  │ByTag (NEW)      │  │
│  │MatchIngredients │  │GetCategoryCounts│  │                 │  │
│  │     (NEW)       │  │     (NEW)       │  │                 │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DOMAIN LAYER                                │
│                  (Pure Business Logic)                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │    Recipe    │  │    User      │  │   Matching   │           │
│  │  Aggregate   │  │   Entity     │  │   (NEW)      │           │
│  │              │  │              │  │              │           │
│  │ - Category   │  │ - PantryItems│  │ - Normalizer │           │
│  │ - DietaryTags│  │   (NEW)      │  │ - Matcher    │           │
│  │   (NEW)      │  │              │  │ - MatchResult│           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        PORTS                                     │
│                    (Interfaces)                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐  │
│  │RecipeRepo   │ │LLMPort      │ │ScraperPort  │ │Exporter   │  │
│  │UserRepo     │ │MessengerPort│ │StoragePort  │ │Ports (NEW)│  │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       ADAPTERS                                   │
│                  (Implementations)                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │ Firebase │ │ Telegram │ │  Gemini  │ │  Python  │            │
│  │          │ │          │ │ /OpenAI  │ │  gRPC    │            │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘            │
│  ┌──────────┐ ┌──────────┐                                      │
│  │ Obsidian │ │  Notion  │  ◄── NEW ADAPTERS                    │
│  │  (NEW)   │ │  (NEW)   │                                      │
│  └──────────┘ └──────────┘                                      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure (Updated)

```
internal/
├── domain/
│   ├── shared/
│   │   ├── id.go
│   │   └── timestamp.go
│   ├── recipe/
│   │   ├── entity.go          # Recipe aggregate root
│   │   ├── ingredient.go      # Ingredient value object
│   │   ├── instruction.go     # Instruction value object
│   │   ├── source.go          # Source value object
│   │   ├── category.go        # NEW: Category type
│   │   ├── dietary.go         # NEW: DietaryTag type
│   │   ├── repository.go      # Repository interface
│   │   └── service.go         # Domain service
│   ├── user/
│   │   ├── entity.go          # User entity (add PantryItems)
│   │   └── repository.go
│   ├── matching/              # NEW PACKAGE
│   │   ├── normalizer.go      # Ingredient normalization
│   │   ├── matcher.go         # Matching algorithm
│   │   ├── result.go          # MatchResult type
│   │   └── staples.go         # Pantry staples list
│   └── export/                # NEW PACKAGE
│       └── types.go           # ExportFormat, ExportRequest
│
├── application/
│   ├── command/
│   │   ├── process_recipe_link.go
│   │   ├── get_or_create_user.go
│   │   ├── match_ingredients.go    # NEW
│   │   ├── update_pantry.go        # NEW
│   │   └── export_recipes.go       # NEW
│   ├── query/
│   │   ├── list_recipes.go         # Update: add ByCategory, ByTag
│   │   ├── find_recipe.go
│   │   └── get_category_counts.go  # NEW
│   └── dto/
│       ├── recipe_dto.go           # Update: add category fields
│       └── match_result_dto.go     # NEW
│
├── ports/
│   ├── scraper.go
│   ├── llm.go
│   ├── messenger.go
│   ├── storage.go
│   └── exporter.go            # NEW: NotionExporter, ObsidianExporter
│
└── adapters/
    ├── firebase/
    │   ├── recipe_repository.go   # Update: handle new fields
    │   └── user_repository.go     # Update: handle pantry
    ├── llm/
    │   ├── gemini.go              # Update: parse new fields
    │   ├── openai.go
    │   └── prompts.go             # Update: categorization prompt
    ├── telegram/
    │   ├── bot.go
    │   ├── handlers.go            # Update: new commands
    │   └── formatter.go           # Update: category display
    ├── obsidian/                  # NEW PACKAGE
    │   ├── markdown.go            # Markdown generation
    │   └── exporter.go            # ObsidianExporter impl
    └── notion/                    # NEW PACKAGE
        ├── client.go              # Notion API client
        ├── oauth.go               # OAuth handling
        └── exporter.go            # NotionExporter impl
```

---

## New Components Detail

### Domain Layer Additions

#### 1. Category & Dietary Tags
Location: `internal/domain/recipe/`

```go
// category.go
package recipe

type Category string

// Constants, validation, parsing methods
// Pure domain logic, no external dependencies

// dietary.go
type DietaryTag string

// Constants, validation, parsing methods
```

These are **Value Objects** - immutable, compared by value, self-validating.

#### 2. Matching Package
Location: `internal/domain/matching/`

```go
// normalizer.go
type Normalizer struct {
    // No external dependencies
}

func (n *Normalizer) Normalize(raw string) string
func (n *Normalizer) AreSimilar(a, b string) bool

// matcher.go
type Matcher struct {
    normalizer *Normalizer
    staples    map[string]bool
}

func (m *Matcher) Match(userIngredients []string, recipes []*recipe.Recipe) []MatchResult

// result.go
type MatchResult struct {
    Recipe          *recipe.Recipe
    MatchPercentage float64
    MatchedItems    []string
    MissingItems    []string
    MatchLevel      MatchLevel
}
```

This is a **Domain Service** - stateless, operates on multiple aggregates.

### Application Layer Additions

#### 1. Match Ingredients Command
```go
// match_ingredients.go
type MatchIngredientsCommand struct {
    UserID      shared.ID
    Ingredients []string
    Options     MatchOptions
}

type MatchIngredientsHandler struct {
    recipeRepo recipe.Repository
    userRepo   user.Repository
    matcher    *matching.Matcher
}

func (h *MatchIngredientsHandler) Handle(ctx context.Context, cmd MatchIngredientsCommand) (*MatchResult, error) {
    // 1. Get user's recipes
    // 2. Get user's pantry items (optional)
    // 3. Run matching algorithm
    // 4. Return results
}
```

#### 2. Export Commands
```go
// export_recipes.go
type ExportRecipesCommand struct {
    UserID    shared.ID
    RecipeIDs []shared.ID  // empty = all
    Format    export.Format
}

type ExportRecipesHandler struct {
    recipeRepo       recipe.Repository
    obsidianExporter ports.ObsidianExporter
    notionExporter   ports.NotionExporter
}
```

### Ports (Interfaces)

```go
// exporter.go
package ports

type ObsidianExporter interface {
    ExportAsMarkdown(recipe *recipe.Recipe) ([]byte, error)
    ExportBulkAsZip(recipes []*recipe.Recipe) ([]byte, error)
}

type NotionExporter interface {
    Connect(userID shared.ID, redirectURL string) (authURL string, err error)
    CompleteAuth(userID shared.ID, code string) error
    IsConnected(userID shared.ID) bool
    Disconnect(userID shared.ID) error
    ExportRecipe(ctx context.Context, userID shared.ID, recipe *recipe.Recipe) (pageURL string, err error)
    ExportBulk(ctx context.Context, userID shared.ID, recipes []*recipe.Recipe) (databaseURL string, err error)
}
```

### Adapters

#### Obsidian Adapter
```go
// adapters/obsidian/exporter.go
type ObsidianExporter struct {
    // No external dependencies needed
}

func (e *ObsidianExporter) ExportAsMarkdown(r *recipe.Recipe) ([]byte, error) {
    // Generate YAML frontmatter
    // Generate markdown content
    // Return bytes
}
```

#### Notion Adapter
```go
// adapters/notion/exporter.go
type NotionExporter struct {
    client     *NotionClient
    userRepo   user.Repository
    encryption *EncryptionService
}

func (e *NotionExporter) ExportRecipe(ctx context.Context, userID shared.ID, r *recipe.Recipe) (string, error) {
    // Get user's Notion token
    // Create page in user's database
    // Return page URL
}
```

---

## Data Flow Examples

### Auto-Categorization Flow
```
1. User sends recipe URL to Telegram
2. Handler calls ProcessRecipeLinkCommand
3. Command fetches content via ScraperPort (Python service)
4. Command extracts recipe via LLMPort (Gemini)
   └── Updated prompt now requests category, tags
5. LLM returns JSON with category, dietaryTags, cuisine
6. Command creates Recipe entity with new fields
7. Command saves via RecipeRepository (Firebase)
8. Handler formats response with category displayed
```

### Ingredient Matching Flow
```
1. User sends "/match chicken, pasta, garlic"
2. Handler parses ingredients from message
3. Handler calls MatchIngredientsCommand
4. Command fetches user's recipes from RecipeRepository
5. Command runs Matcher.Match() (domain logic)
   └── Normalizes user ingredients
   └── Normalizes recipe ingredients
   └── Calculates match percentages
   └── Filters by threshold
6. Command returns MatchResult DTOs
7. Handler formats and sends response
```

### Export to Obsidian Flow
```
1. User sends "/export obsidian 3"
2. Handler parses recipe number
3. Handler calls FindRecipeQuery to get recipe
4. Handler calls ObsidianExporter.ExportAsMarkdown()
5. Exporter generates markdown with frontmatter
6. Handler sends file to user via Telegram
```

### Export to Notion Flow
```
1. User sends "/connect notion"
2. Handler calls NotionExporter.Connect()
3. Exporter returns OAuth URL
4. Handler sends URL to user
5. User authorizes in browser
6. Notion redirects with code
7. Handler calls NotionExporter.CompleteAuth()
8. Exporter exchanges code for token, stores encrypted

Later:
1. User sends "/export notion 3"
2. Handler calls NotionExporter.ExportRecipe()
3. Exporter creates Notion page via API
4. Handler sends success message with page URL
```

---

## Dependency Injection

Update `cmd/bot/main.go` to wire new dependencies:

```go
func main() {
    // Existing...

    // NEW: Create domain services
    normalizer := matching.NewNormalizer()
    matcher := matching.NewMatcher(normalizer)

    // NEW: Create exporters
    obsidianExporter := obsidian.NewExporter()
    notionExporter := notion.NewExporter(notionClient, userRepo, encryptionService)

    // NEW: Create command handlers
    matchHandler := command.NewMatchIngredientsHandler(recipeRepo, userRepo, matcher)
    exportHandler := command.NewExportRecipesHandler(recipeRepo, obsidianExporter, notionExporter)

    // Update Telegram handler with new dependencies
    telegramHandler := telegram.NewHandler(
        // existing...
        matchHandler,
        exportHandler,
    )
}
```

---

## Testing Strategy

### Unit Tests (Domain Layer)
- `category_test.go` - Category validation, parsing
- `dietary_test.go` - DietaryTag validation
- `normalizer_test.go` - Ingredient normalization edge cases
- `matcher_test.go` - Matching algorithm correctness

### Integration Tests (Application Layer)
- `match_ingredients_test.go` - Full matching flow with mocked repos
- `export_recipes_test.go` - Export flow with mocked exporters

### Adapter Tests
- `obsidian_exporter_test.go` - Markdown generation correctness
- `notion_exporter_test.go` - API interaction (use mocks/stubs)

---

## Migration Considerations

### Existing Recipes
- Add default values for new fields:
  - `category`: "Other"
  - `cuisine`: ""
  - `dietaryTags`: []
  - `tags`: []
- Create migration script to re-process existing recipes through LLM for categorization

### Database Indexes
Add Firestore indexes for:
- `recipes` collection: `userId` + `category`
- `recipes` collection: `userId` + `dietaryTags` (array-contains)

### Backwards Compatibility
- All new fields optional in Firestore
- Old recipes display normally (category shows as "Uncategorized")
- No breaking changes to existing commands
