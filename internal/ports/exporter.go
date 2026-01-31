package ports

import (
	"context"

	"receipt-bot/internal/domain/recipe"
)

// ExportResult contains the result of an export operation
type ExportResult struct {
	Success  bool
	Format   string
	Filename string // For file-based exports (Obsidian)
	Data     []byte // File content for downloads
	URL      string // For API-based exports (Notion page URL)
	Message  string // User-facing message
}

// ObsidianExporter defines the interface for exporting recipes to Obsidian format
type ObsidianExporter interface {
	// ExportRecipe exports a single recipe as Obsidian-compatible markdown
	ExportRecipe(recipe *recipe.Recipe) (*ExportResult, error)

	// ExportRecipes exports multiple recipes as a ZIP file
	ExportRecipes(recipes []*recipe.Recipe) (*ExportResult, error)
}

// NotionExporter defines the interface for exporting recipes to Notion
type NotionExporter interface {
	// GetAuthURL returns the OAuth authorization URL for a user
	GetAuthURL(userID string, state string) string

	// HandleCallback processes the OAuth callback and stores tokens
	HandleCallback(ctx context.Context, userID string, code string) error

	// ExportRecipe exports a single recipe to the user's Notion database
	ExportRecipe(ctx context.Context, userID string, recipe *recipe.Recipe) (*ExportResult, error)

	// ExportRecipes exports multiple recipes to the user's Notion database
	ExportRecipes(ctx context.Context, userID string, recipes []*recipe.Recipe) (*ExportResult, error)

	// IsConnected checks if the user has a valid Notion connection
	IsConnected(ctx context.Context, userID string) (bool, error)

	// Disconnect removes the Notion connection for a user
	Disconnect(ctx context.Context, userID string) error
}
