# Receipt-Bot Fix Plan

## Overview
Implementing three new features for the receipt-bot Telegram bot:
1. Auto-categorization of recipes
2. Ingredient-based recipe matching
3. Export to Notion & Obsidian

See PRD.md and .ralph/specs/ for detailed specifications.

## Phase 1: Auto-Categorization (Foundation) ✅ COMPLETE

### High Priority
- [x] Create category.go with Category enum and validation
- [x] Create dietary_tags.go with DietaryTag enum
- [x] Update Recipe entity with Category, Cuisine, DietaryTags, Tags fields
- [x] Update LLM prompts to include categorization in extraction
- [x] Update Firestore schema and recipe_repository.go
- [x] Add category queries to application layer
- [x] Update Telegram formatter to display category
- [x] Implement /recipes <category> filter command
- [x] Implement /categories command

### Medium Priority
- [x] Create backfill command for existing recipes
- [x] Add unit tests for category validation
- [x] Add integration tests for category queries

## Phase 2: Ingredient Matching ✅ COMPLETE

### High Priority
- [x] Create ingredient normalizer (rule-based)
- [x] Implement matching algorithm with scoring
- [x] Add pantry storage to user model
- [x] Create match_ingredients command
- [x] Implement /match Telegram command
- [x] Implement /pantry commands

### Medium Priority
- [x] Add ingredient substitution groups (included in normalizer)
- [x] Cache normalized ingredients in recipe documents
- [x] Unit tests for normalizer
- [x] Unit tests for matcher

## Phase 3: Conversational Interface ✅ COMPLETE

### High Priority
- [x] Create intent detection using LLM for natural language queries
- [x] Implement conversational message handler (non-URL, non-command text)
- [x] Support natural queries like "Seafood recipes" → execute /recipes seafood
- [x] Support ingredient-based queries like "Salmon recipe" → filter seafood + salmon
- [x] Support conversational pantry management ("I have chicken and rice")

### Medium Priority
- [x] Add fuzzy matching for category names ("fish" → Seafood) - via LLM intent detection
- [x] Add context-aware responses for follow-up questions ("show more", "details on #3", "repeat")
- [x] Support compound queries ("quick pasta recipes", "vegan breakfast")
- [x] Implement conversation memory for context continuity (ConversationManager)

## Phase 4: PT-BR Multilingual Support ✅ COMPLETE

### High Priority
- [x] Update LLM prompts to detect source language
- [x] Add translation field to Recipe entity (translatedTitle, translatedIngredients, translatedInstructions)
- [x] Store original language + translation (EN↔PT-BR)
- [x] Update Firestore schema for multilingual fields
- [x] Detect user language preference from Telegram settings
- [x] Translate recipe output based on user preference
- [x] Support PT-BR natural language queries

### Medium Priority
- [x] Add /language command to set user preference
- [x] Translate category names and UI strings
- [x] Support ingredient matching in Portuguese (via intent detection)

## Phase 5: Export Integration ✅ COMPLETE

### High Priority (Obsidian - simpler)
- [x] Create ObsidianExporter port interface
- [x] Implement markdown generator with YAML frontmatter
- [x] Add file sending capability to Telegram adapter (SendDocument)
- [x] Implement /export obsidian command
- [x] Bulk export as ZIP for multiple recipes

### High Priority (Notion)
- [x] Create NotionExporter port interface
- [x] Implement Notion API client with OAuth support
- [x] Implement OAuth flow (GetAuthURL, ExchangeCode, HandleCallback)
- [x] Implement /connect notion command
- [x] Implement /disconnect notion command
- [x] Implement /export notion command
- [x] Update User entity with Notion OAuth fields
- [x] Update Firebase user repository with Notion credentials storage

### Medium Priority
- [x] Bulk export (ZIP for Obsidian, batch for Notion)
- [ ] Rate limiting and retry logic (future improvement)
- [ ] Token encryption at rest (future improvement)

## Completed
- [x] Project initialization
- [x] PRD created with feature specifications
- [x] Ralph specs created for all three features

## Notes
- Phase 1 (Auto-Categorization) ✅ COMPLETE - Foundation for all features
- Phase 2 (Ingredient Matching) ✅ COMPLETE - Enables pantry-based matching
- Phase 3 (Conversational Interface) ✅ COMPLETE - Natural language interaction
- Phase 4 (PT-BR Support) ✅ COMPLETE - Multilingual support for Portuguese speakers
- Phase 5 (Export Integration) ✅ COMPLETE - Export to Obsidian (Markdown) and Notion
- Run `go test ./...` after each implementation
- Update this file after completing each task

## Configuration Required for Notion Integration
To enable Notion export, set the following environment variables:
- `NOTION_CLIENT_ID` - Your Notion OAuth app client ID
- `NOTION_CLIENT_SECRET` - Your Notion OAuth app client secret
- `NOTION_REDIRECT_URI` - OAuth redirect URI (e.g., https://yourapp.com/auth/notion/callback)

See https://developers.notion.com/docs/authorization for Notion OAuth setup instructions.
