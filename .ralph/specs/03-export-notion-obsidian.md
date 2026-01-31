# Specification: Export to Notion & Obsidian

## Overview
Allow users to export their saved recipes to external knowledge management tools (Notion and Obsidian).

## Requirements

### Notion Integration

#### OAuth Setup
- Register Notion OAuth application
- Implement OAuth 2.0 flow for user authorization
- Store encrypted access tokens per user

#### Notion Exporter Port
Create `internal/ports/notion_exporter.go`:
```go
type NotionExporter interface {
    // Connect initiates OAuth and stores credentials
    Connect(ctx context.Context, userID shared.ID, authCode string) error

    // IsConnected checks if user has valid Notion connection
    IsConnected(ctx context.Context, userID shared.ID) bool

    // Disconnect removes Notion integration
    Disconnect(ctx context.Context, userID shared.ID) error

    // ExportRecipe exports single recipe, returns Notion page URL
    ExportRecipe(ctx context.Context, userID shared.ID, recipe *recipe.Recipe) (string, error)

    // ExportBulk exports multiple recipes, returns database URL
    ExportBulk(ctx context.Context, userID shared.ID, recipes []*recipe.Recipe) (string, error)
}
```

#### Notion Page Structure
Database properties:
- Title (title)
- Category (select)
- Cuisine (select)
- Prep Time (number, minutes)
- Cook Time (number, minutes)
- Servings (number)
- Source URL (url)
- Platform (select: TikTok, YouTube, Instagram, Web)
- Tags (multi-select)
- Created At (date)

Page content:
```
# {Recipe Title}

## Ingredients
- {ingredient 1}
- {ingredient 2}
...

## Instructions
1. {step 1}
2. {step 2}
...

## Source
[Original Recipe]({source_url})
By: {author}
```

#### Notion Adapter
Create `internal/adapters/notion/`:
- `client.go` - Notion API client wrapper
- `oauth.go` - OAuth flow handling
- `exporter.go` - NotionExporter implementation

### Obsidian Integration

#### Obsidian Exporter Port
Create `internal/ports/obsidian_exporter.go`:
```go
type ObsidianExporter interface {
    // ExportAsMarkdown generates markdown for single recipe
    ExportAsMarkdown(recipe *recipe.Recipe) ([]byte, error)

    // ExportBulkAsZip generates ZIP file with all recipes
    ExportBulkAsZip(recipes []*recipe.Recipe) ([]byte, error)
}
```

#### Markdown Format
```markdown
---
title: {Recipe Title}
category: {category}
cuisine: {cuisine}
prep_time: {prep_time_minutes}
cook_time: {cook_time_minutes}
servings: {servings}
source: {source_url}
platform: {platform}
created: {created_date}
tags: [{tag1}, {tag2}]
---

# {Recipe Title}

## Ingredients
- {quantity} {unit} {ingredient} {notes}
...

## Instructions
1. {step 1}
2. {step 2}
...

## Source
[Original Recipe]({source_url})
By: {author}
```

#### Obsidian Adapter
Create `internal/adapters/obsidian/`:
- `markdown.go` - Markdown generation with YAML frontmatter
- `exporter.go` - ObsidianExporter implementation

### File Delivery via Telegram
- Single recipe: Send as document (.md file)
- Bulk export: Send as document (.zip file)
- Use Telegram's `sendDocument` API

### Database Schema Updates
Add to users collection:
```
notionAccessToken: string (encrypted)
notionRefreshToken: string (encrypted)
notionDatabaseId: string
notionConnectedAt: timestamp
```

### Telegram Commands
```
/export notion <recipe_number>     - Export single recipe to Notion
/export obsidian <recipe_number>   - Export single recipe as Markdown file
/export notion all                 - Export all recipes to Notion
/export obsidian all               - Export all recipes as ZIP
/connect notion                    - Start Notion OAuth flow
/disconnect notion                 - Remove Notion integration
```

### Application Commands
Create `internal/application/command/`:
- `export_to_notion.go`
- `export_to_obsidian.go`
- `connect_notion.go`
- `disconnect_notion.go`

## Acceptance Criteria

### Notion
- [ ] OAuth flow works and stores credentials securely
- [ ] `/connect notion` initiates connection
- [ ] `/export notion 1` exports recipe to Notion
- [ ] `/export notion all` exports all recipes
- [ ] Exported pages have correct properties and content
- [ ] Duplicate detection prevents re-exporting same recipe
- [ ] `/disconnect notion` removes integration

### Obsidian
- [ ] `/export obsidian 1` sends markdown file via Telegram
- [ ] `/export obsidian all` sends ZIP file via Telegram
- [ ] Markdown has valid YAML frontmatter
- [ ] File names are sanitized (no special characters)
- [ ] ZIP contains organized folder structure

### General
- [ ] Error handling for API failures
- [ ] Rate limiting for bulk exports
- [ ] Unit tests for markdown generation
- [ ] Integration tests for Notion API

## Technical Notes
- Notion API rate limits: 3 requests/second
- Implement retry with exponential backoff
- Consider background processing for bulk exports
- File names: sanitize recipe titles for filesystem compatibility
- ZIP structure: `recipes/{category}/{recipe_title}.md`

## Security Considerations
- Encrypt Notion tokens at rest
- Use environment variable for encryption key
- Implement token refresh flow
- Never log access tokens
