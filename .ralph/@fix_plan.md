# Receipt-Bot Fix Plan

## Overview
Implementing three new features for the receipt-bot Telegram bot:
1. Auto-categorization of recipes
2. Ingredient-based recipe matching
3. Export to Notion & Obsidian

See PRD.md and .ralph/specs/ for detailed specifications.

## Phase 1: Auto-Categorization (Foundation)

### High Priority
- [ ] Create category.go with Category enum and validation
- [ ] Create dietary_tags.go with DietaryTag enum
- [ ] Update Recipe entity with Category, Cuisine, DietaryTags, Tags fields
- [ ] Update LLM prompts to include categorization in extraction
- [ ] Update Firestore schema and recipe_repository.go
- [ ] Add category queries to application layer
- [ ] Update Telegram formatter to display category
- [ ] Implement /recipes <category> filter command
- [ ] Implement /categories command

### Medium Priority
- [ ] Create backfill command for existing recipes
- [ ] Add unit tests for category validation
- [ ] Add integration tests for category queries

## Phase 2: Ingredient Matching

### High Priority
- [ ] Create ingredient normalizer (rule-based)
- [ ] Implement matching algorithm with scoring
- [ ] Add pantry storage to user model
- [ ] Create match_ingredients command
- [ ] Implement /match Telegram command
- [ ] Implement /pantry commands

### Medium Priority
- [ ] Add ingredient substitution groups
- [ ] Cache normalized ingredients in recipe documents
- [ ] Unit tests for normalizer
- [ ] Unit tests for matcher

## Phase 3: Export Integration

### High Priority (Obsidian - simpler)
- [ ] Create ObsidianExporter port interface
- [ ] Implement markdown generator with YAML frontmatter
- [ ] Add file sending capability to Telegram adapter
- [ ] Implement /export obsidian command

### High Priority (Notion)
- [ ] Set up Notion OAuth application
- [ ] Create NotionExporter port interface
- [ ] Implement Notion API client
- [ ] Implement OAuth flow
- [ ] Implement /connect notion command
- [ ] Implement /export notion command

### Medium Priority
- [ ] Bulk export (ZIP for Obsidian, batch for Notion)
- [ ] Rate limiting and retry logic
- [ ] Token encryption at rest

## Completed
- [x] Project initialization
- [x] PRD created with feature specifications
- [x] Ralph specs created for all three features

## Notes
- Start with Phase 1 (Auto-Categorization) as it's the foundation
- Phase 2 depends on normalized ingredients (benefits from Phase 1)
- Phase 3 can be developed in parallel with Phase 2
- Run `go test ./...` after each implementation
- Update this file after completing each task
