package notion

import (
	"context"
	"fmt"
	"strings"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/user"
	"receipt-bot/internal/ports"
)

// UserRepository interface for accessing user Notion credentials
type UserRepository interface {
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
	UpdateNotionConnection(ctx context.Context, userID user.UserID, accessToken, workspaceID, databaseID string) error
	ClearNotionConnection(ctx context.Context, userID user.UserID) error
}

// Exporter implements the NotionExporter interface
type Exporter struct {
	client   *Client
	userRepo UserRepository
}

// NewExporter creates a new Notion exporter
func NewExporter(client *Client, userRepo UserRepository) *Exporter {
	return &Exporter{
		client:   client,
		userRepo: userRepo,
	}
}

// GetAuthURL returns the OAuth authorization URL for a user
func (e *Exporter) GetAuthURL(userID string, state string) string {
	return e.client.GetAuthURL(state)
}

// HandleCallback processes the OAuth callback and stores tokens
func (e *Exporter) HandleCallback(ctx context.Context, userID string, code string) error {
	// Exchange code for token
	tokenResp, err := e.client.ExchangeCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Search for existing recipe databases or use the first available one
	databases, err := e.client.SearchDatabases(ctx, tokenResp.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to search databases: %w", err)
	}

	var databaseID string
	for _, db := range databases {
		// Look for a database that might be for recipes
		if len(db.Title) > 0 {
			title := strings.ToLower(db.Title[0].PlainText)
			if strings.Contains(title, "recipe") || strings.Contains(title, "receita") {
				databaseID = db.ID
				break
			}
		}
	}

	// If no recipe database found, use the first one (user can change later)
	if databaseID == "" && len(databases) > 0 {
		databaseID = databases[0].ID
	}

	// Store credentials
	err = e.userRepo.UpdateNotionConnection(
		ctx,
		user.UserID(userID),
		tokenResp.AccessToken,
		tokenResp.WorkspaceID,
		databaseID,
	)
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	return nil
}

// ExportRecipe exports a single recipe to the user's Notion database
func (e *Exporter) ExportRecipe(ctx context.Context, userID string, rec *recipe.Recipe) (*ports.ExportResult, error) {
	usr, err := e.userRepo.FindByID(ctx, user.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !usr.HasNotionConnection() {
		return &ports.ExportResult{
			Success: false,
			Format:  "notion",
			Message: "Not connected to Notion. Use /connect notion to authorize.",
		}, nil
	}

	if usr.NotionDatabaseID() == "" {
		return &ports.ExportResult{
			Success: false,
			Format:  "notion",
			Message: "No Notion database selected. Please reconnect with /connect notion.",
		}, nil
	}

	// Build properties for the page
	properties := e.buildProperties(rec)

	// Build content blocks
	children := e.buildContent(rec)

	// Create the page
	page, err := e.client.CreatePage(ctx, usr.NotionAccessToken(), usr.NotionDatabaseID(), properties, children)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return &ports.ExportResult{
		Success: true,
		Format:  "notion",
		URL:     page.URL,
		Message: fmt.Sprintf("Recipe exported: %s", rec.Title()),
	}, nil
}

// ExportRecipes exports multiple recipes to the user's Notion database
func (e *Exporter) ExportRecipes(ctx context.Context, userID string, recipes []*recipe.Recipe) (*ports.ExportResult, error) {
	if len(recipes) == 0 {
		return &ports.ExportResult{
			Success: false,
			Format:  "notion",
			Message: "No recipes to export",
		}, nil
	}

	var exported int
	var lastURL string
	var errors []string

	for _, rec := range recipes {
		result, err := e.ExportRecipe(ctx, userID, rec)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", rec.Title(), err))
			continue
		}
		if result.Success {
			exported++
			lastURL = result.URL
		} else {
			errors = append(errors, fmt.Sprintf("%s: %s", rec.Title(), result.Message))
		}
	}

	message := fmt.Sprintf("Exported %d of %d recipes to Notion", exported, len(recipes))
	if len(errors) > 0 {
		message += fmt.Sprintf(" (%d errors)", len(errors))
	}

	return &ports.ExportResult{
		Success: exported > 0,
		Format:  "notion",
		URL:     lastURL,
		Message: message,
	}, nil
}

// IsConnected checks if the user has a valid Notion connection
func (e *Exporter) IsConnected(ctx context.Context, userID string) (bool, error) {
	usr, err := e.userRepo.FindByID(ctx, user.UserID(userID))
	if err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}

	return usr.HasNotionConnection(), nil
}

// Disconnect removes the Notion connection for a user
func (e *Exporter) Disconnect(ctx context.Context, userID string) error {
	return e.userRepo.ClearNotionConnection(ctx, user.UserID(userID))
}

// buildProperties builds Notion page properties from a recipe
func (e *Exporter) buildProperties(rec *recipe.Recipe) map[string]interface{} {
	props := map[string]interface{}{
		"Name": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": rec.Title(),
					},
				},
			},
		},
	}

	// Category
	if rec.Category() != "" {
		props["Category"] = map[string]interface{}{
			"select": map[string]string{
				"name": string(rec.Category()),
			},
		}
	}

	// Cuisine
	if rec.Cuisine() != "" {
		props["Cuisine"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": rec.Cuisine(),
					},
				},
			},
		}
	}

	// Prep Time
	if rec.PrepTime() != nil {
		props["Prep Time"] = map[string]interface{}{
			"number": int(rec.PrepTime().Minutes()),
		}
	}

	// Cook Time
	if rec.CookTime() != nil {
		props["Cook Time"] = map[string]interface{}{
			"number": int(rec.CookTime().Minutes()),
		}
	}

	// Servings
	if rec.Servings() != nil {
		props["Servings"] = map[string]interface{}{
			"number": *rec.Servings(),
		}
	}

	// Source URL
	if rec.Source().URL() != "" {
		props["Source URL"] = map[string]interface{}{
			"url": rec.Source().URL(),
		}
	}

	// Tags (dietary tags)
	if len(rec.DietaryTags()) > 0 {
		var tags []map[string]string
		for _, tag := range rec.DietaryTags() {
			tags = append(tags, map[string]string{
				"name": string(tag),
			})
		}
		props["Tags"] = map[string]interface{}{
			"multi_select": tags,
		}
	}

	return props
}

// buildContent builds Notion block content from a recipe
func (e *Exporter) buildContent(rec *recipe.Recipe) []interface{} {
	var blocks []interface{}

	// Ingredients heading
	blocks = append(blocks, map[string]interface{}{
		"object": "block",
		"type":   "heading_2",
		"heading_2": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]string{
						"content": "Ingredients",
					},
				},
			},
		},
	})

	// Ingredients as bulleted list
	for _, ing := range rec.Ingredients() {
		text := formatIngredient(ing)
		blocks = append(blocks, map[string]interface{}{
			"object": "block",
			"type":   "bulleted_list_item",
			"bulleted_list_item": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": text,
						},
					},
				},
			},
		})
	}

	// Instructions heading
	blocks = append(blocks, map[string]interface{}{
		"object": "block",
		"type":   "heading_2",
		"heading_2": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]string{
						"content": "Instructions",
					},
				},
			},
		},
	})

	// Instructions as numbered list
	for i, inst := range rec.Instructions() {
		stepNum := inst.StepNumber()
		if stepNum == 0 {
			stepNum = i + 1
		}
		text := inst.Text()
		blocks = append(blocks, map[string]interface{}{
			"object": "block",
			"type":   "numbered_list_item",
			"numbered_list_item": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": text,
						},
					},
				},
			},
		})
	}

	// Source section
	if rec.Source().URL() != "" {
		blocks = append(blocks, map[string]interface{}{
			"object": "block",
			"type":   "divider",
			"divider": map[string]interface{}{},
		})

		sourceText := "Original Recipe"
		if rec.Source().Author() != "" {
			sourceText += " by " + rec.Source().Author()
		}
		if rec.Source().Platform() != "" {
			sourceText += " on " + string(rec.Source().Platform())
		}

		blocks = append(blocks, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]interface{}{
							"content": sourceText,
							"link": map[string]string{
								"url": rec.Source().URL(),
							},
						},
					},
				},
			},
		})
	}

	return blocks
}

// formatIngredient formats an ingredient for display
func formatIngredient(ing recipe.Ingredient) string {
	var parts []string

	if ing.Quantity() != "" {
		parts = append(parts, ing.Quantity())
	}
	if ing.Unit() != "" {
		parts = append(parts, ing.Unit())
	}
	parts = append(parts, ing.Name())

	result := strings.Join(parts, " ")

	if ing.Notes() != "" {
		result += fmt.Sprintf(" (%s)", ing.Notes())
	}

	return result
}
