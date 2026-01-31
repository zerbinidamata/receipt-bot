# Ralph Fix Plan

## High Priority

### Phase 1: Auto-Categorization (Foundation)
- [ ] Add Category type and constants to `internal/domain/recipe/category.go`
- [ ] Add DietaryTag type and constants to `internal/domain/recipe/dietary.go`
- [ ] Update Recipe entity with Category, Cuisine, DietaryTags, Tags fields
- [ ] Update LLM extraction prompt in `internal/adapters/llm/prompts.go` to include categorization
- [ ] Update Gemini adapter response parsing to handle new fields
- [ ] Update Firebase recipe repository to store/retrieve new fields
- [ ] Update RecipeDTO to include category fields
- [ ] Add unit tests for new domain types
- [ ] Update Telegram formatter to display category in recipe output

### Phase 1: Category Commands
- [ ] Implement `/recipes <category>` filter command
- [ ] Implement `/recipes --tag <tag>` filter command
- [ ] Implement `/categories` command to list categories with counts
- [ ] Add ListRecipesByCategory query
- [ ] Add ListRecipesByTag query

## Medium Priority

### Phase 2: Ingredient Matching
- [ ] Create `internal/domain/matching/normalizer.go` - ingredient text normalization
- [ ] Create `internal/domain/matching/matcher.go` - matching algorithm
- [ ] Define MatchResult and MatchLevel types
- [ ] Create pantry staples list (items to ignore in matching)
- [ ] Add pantryItems field to User entity
- [ ] Update user repository for pantry storage
- [ ] Create MatchIngredientsCommand in application layer
- [ ] Implement `/match <ingredients>` Telegram command
- [ ] Implement `/pantry add <items>` command
- [ ] Implement `/pantry` and `/pantry clear` commands
- [ ] Add Telegram formatter for match results
- [ ] Add unit tests for normalizer
- [ ] Add unit tests for matcher algorithm

### Phase 3: Obsidian Export
- [ ] Create `internal/adapters/obsidian/markdown.go` - markdown generator
- [ ] Create `internal/adapters/obsidian/exporter.go` - ObsidianExporter implementation
- [ ] Define ExportFormat type in domain
- [ ] Implement `/export obsidian <recipe>` command
- [ ] Implement `/export obsidian all` bulk export as ZIP
- [ ] Add file upload capability to Telegram adapter

## Low Priority

### Phase 3: Notion Export
- [ ] Research Notion API authentication options (internal vs public OAuth)
- [ ] Create `internal/adapters/notion/client.go` - Notion API client
- [ ] Create `internal/adapters/notion/oauth.go` - OAuth flow handling
- [ ] Create `internal/adapters/notion/exporter.go` - NotionExporter implementation
- [ ] Add notionAccessToken, notionDatabaseId to User entity (encrypted)
- [ ] Implement `/connect notion` OAuth flow
- [ ] Implement `/export notion <recipe>` command
- [ ] Implement `/export notion all` bulk export
- [ ] Implement `/disconnect notion` command
- [ ] Add integration tests for Notion adapter

### Backfill & Polish
- [ ] Create migration script to backfill categories for existing recipes
- [ ] Add `/recategorize <recipe>` command for manual re-categorization
- [ ] Implement ingredient substitution groups for smarter matching
- [ ] Add category statistics to `/recipes` command output

## Completed
- [x] Project initialization
- [x] Core recipe extraction pipeline
- [x] TikTok, YouTube, Instagram, Web scraping
- [x] Firebase Firestore integration
- [x] Telegram bot with /start, /help, /recipes, /recipe commands
- [x] LLM integration (Gemini/OpenAI)
- [x] Hexagonal architecture setup
- [x] PRD creation for next phase features
