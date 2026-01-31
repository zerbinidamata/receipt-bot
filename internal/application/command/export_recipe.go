package command

import (
	"context"
	"fmt"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/ports"
)

// ExportFormat represents the export format type
type ExportFormat string

const (
	ExportFormatObsidian ExportFormat = "obsidian"
	ExportFormatNotion   ExportFormat = "notion"
)

// ExportRecipeInput contains input for exporting recipes
type ExportRecipeInput struct {
	UserID   shared.ID
	RecipeID *shared.ID   // If nil, export all recipes
	Format   ExportFormat
}

// ExportRecipeCommand handles recipe export operations
type ExportRecipeCommand struct {
	recipeRepo       recipe.Repository
	obsidianExporter ports.ObsidianExporter
	notionExporter   ports.NotionExporter
}

// NewExportRecipeCommand creates a new export recipe command
func NewExportRecipeCommand(
	recipeRepo recipe.Repository,
	obsidianExporter ports.ObsidianExporter,
	notionExporter ports.NotionExporter,
) *ExportRecipeCommand {
	return &ExportRecipeCommand{
		recipeRepo:       recipeRepo,
		obsidianExporter: obsidianExporter,
		notionExporter:   notionExporter,
	}
}

// Execute exports recipes based on the input parameters
func (c *ExportRecipeCommand) Execute(ctx context.Context, input ExportRecipeInput) (*ports.ExportResult, error) {
	switch input.Format {
	case ExportFormatObsidian:
		return c.exportToObsidian(ctx, input)
	case ExportFormatNotion:
		return c.exportToNotion(ctx, input)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", input.Format)
	}
}

// exportToObsidian handles Obsidian export
func (c *ExportRecipeCommand) exportToObsidian(ctx context.Context, input ExportRecipeInput) (*ports.ExportResult, error) {
	if c.obsidianExporter == nil {
		return nil, fmt.Errorf("obsidian exporter not configured")
	}

	// Export single recipe
	if input.RecipeID != nil {
		rec, err := c.recipeRepo.FindByID(ctx, recipe.RecipeID(*input.RecipeID))
		if err != nil {
			return nil, fmt.Errorf("recipe not found: %w", err)
		}

		// Verify ownership
		if rec.UserID() != recipe.UserID(input.UserID) {
			return nil, fmt.Errorf("unauthorized: recipe belongs to another user")
		}

		return c.obsidianExporter.ExportRecipe(rec)
	}

	// Export all recipes for user
	recipes, err := c.recipeRepo.FindByUserID(ctx, recipe.UserID(input.UserID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}

	if len(recipes) == 0 {
		return &ports.ExportResult{
			Success: false,
			Format:  "obsidian",
			Message: "No recipes to export",
		}, nil
	}

	return c.obsidianExporter.ExportRecipes(recipes)
}

// exportToNotion handles Notion export
func (c *ExportRecipeCommand) exportToNotion(ctx context.Context, input ExportRecipeInput) (*ports.ExportResult, error) {
	if c.notionExporter == nil {
		return nil, fmt.Errorf("notion exporter not configured")
	}

	// Check if user is connected to Notion
	connected, err := c.notionExporter.IsConnected(ctx, input.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check Notion connection: %w", err)
	}

	if !connected {
		return &ports.ExportResult{
			Success: false,
			Format:  "notion",
			Message: "Not connected to Notion. Use /connect notion to authorize.",
		}, nil
	}

	// Export single recipe
	if input.RecipeID != nil {
		rec, err := c.recipeRepo.FindByID(ctx, recipe.RecipeID(*input.RecipeID))
		if err != nil {
			return nil, fmt.Errorf("recipe not found: %w", err)
		}

		// Verify ownership
		if rec.UserID() != recipe.UserID(input.UserID) {
			return nil, fmt.Errorf("unauthorized: recipe belongs to another user")
		}

		return c.notionExporter.ExportRecipe(ctx, input.UserID.String(), rec)
	}

	// Export all recipes for user
	recipes, err := c.recipeRepo.FindByUserID(ctx, recipe.UserID(input.UserID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}

	if len(recipes) == 0 {
		return &ports.ExportResult{
			Success: false,
			Format:  "notion",
			Message: "No recipes to export",
		}, nil
	}

	return c.notionExporter.ExportRecipes(ctx, input.UserID.String(), recipes)
}

// HasObsidianExporter returns true if Obsidian export is available
func (c *ExportRecipeCommand) HasObsidianExporter() bool {
	return c.obsidianExporter != nil
}

// HasNotionExporter returns true if Notion export is available
func (c *ExportRecipeCommand) HasNotionExporter() bool {
	return c.notionExporter != nil
}
