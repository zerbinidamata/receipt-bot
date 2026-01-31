package command

import (
	"context"
	"fmt"
	"log"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/ports"
)

// BackfillCategoriesCommand handles re-categorization of existing recipes
type BackfillCategoriesCommand struct {
	recipeRepo recipe.Repository
	llmPort    ports.LLMPort
}

// NewBackfillCategoriesCommand creates a new backfill categories command
func NewBackfillCategoriesCommand(
	recipeRepo recipe.Repository,
	llmPort ports.LLMPort,
) *BackfillCategoriesCommand {
	return &BackfillCategoriesCommand{
		recipeRepo: recipeRepo,
		llmPort:    llmPort,
	}
}

// BackfillResult contains the result of a backfill operation
type BackfillResult struct {
	TotalProcessed int
	Updated        int
	Skipped        int
	Errors         int
	Details        []BackfillDetail
}

// BackfillDetail contains details about a single recipe backfill
type BackfillDetail struct {
	RecipeID    string
	Title       string
	OldCategory recipe.Category
	NewCategory recipe.Category
	Error       error
}

// Execute runs the backfill operation for a specific user
func (c *BackfillCategoriesCommand) Execute(ctx context.Context, userID recipe.UserID) (*BackfillResult, error) {
	// Fetch all recipes for the user
	recipes, err := c.recipeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}

	result := &BackfillResult{
		Details: make([]BackfillDetail, 0, len(recipes)),
	}

	for _, rec := range recipes {
		result.TotalProcessed++
		detail := BackfillDetail{
			RecipeID:    rec.ID().String(),
			Title:       rec.Title(),
			OldCategory: rec.Category(),
		}

		// Skip recipes that already have a non-Other category
		if rec.Category() != recipe.CategoryOther {
			detail.NewCategory = rec.Category()
			result.Skipped++
			result.Details = append(result.Details, detail)
			continue
		}

		// Build text for categorization from existing recipe data
		text := buildRecipeText(rec)

		// Use LLM to extract category
		extraction, err := c.llmPort.ExtractRecipe(ctx, text)
		if err != nil {
			detail.Error = err
			result.Errors++
			result.Details = append(result.Details, detail)
			log.Printf("Failed to categorize recipe %s: %v", rec.ID().String(), err)
			continue
		}

		// Parse the new category
		newCategory := recipe.CategoryFromLLM(extraction.Category)
		detail.NewCategory = newCategory

		// Update the recipe if category changed
		if newCategory != rec.Category() {
			rec.SetCategory(newCategory)

			// Also update cuisine and dietary tags if available
			if extraction.Cuisine != "" && rec.Cuisine() == "" {
				rec.SetCuisine(extraction.Cuisine)
			}

			if len(extraction.DietaryTags) > 0 && len(rec.DietaryTags()) == 0 {
				dietaryTags := recipe.ParseDietaryTags(extraction.DietaryTags)
				rec.SetDietaryTags(dietaryTags)
			}

			if len(extraction.Tags) > 0 && len(rec.Tags()) == 0 {
				rec.SetTags(extraction.Tags)
			}

			if err := c.recipeRepo.Update(ctx, rec); err != nil {
				detail.Error = err
				result.Errors++
				log.Printf("Failed to update recipe %s: %v", rec.ID().String(), err)
			} else {
				result.Updated++
			}
		} else {
			result.Skipped++
		}

		result.Details = append(result.Details, detail)
	}

	return result, nil
}

// ExecuteAll runs the backfill operation for all recipes in the system
// This should be used carefully as it processes all recipes
func (c *BackfillCategoriesCommand) ExecuteAll(ctx context.Context, userIDs []recipe.UserID) (*BackfillResult, error) {
	totalResult := &BackfillResult{
		Details: make([]BackfillDetail, 0),
	}

	for _, userID := range userIDs {
		result, err := c.Execute(ctx, userID)
		if err != nil {
			log.Printf("Failed to process user %s: %v", userID.String(), err)
			continue
		}

		totalResult.TotalProcessed += result.TotalProcessed
		totalResult.Updated += result.Updated
		totalResult.Skipped += result.Skipped
		totalResult.Errors += result.Errors
		totalResult.Details = append(totalResult.Details, result.Details...)
	}

	return totalResult, nil
}

// buildRecipeText builds a text representation of a recipe for categorization
func buildRecipeText(rec *recipe.Recipe) string {
	text := fmt.Sprintf("Recipe: %s\n\n", rec.Title())

	text += "Ingredients:\n"
	for _, ing := range rec.Ingredients() {
		text += fmt.Sprintf("- %s %s %s", ing.Quantity(), ing.Unit(), ing.Name())
		if ing.Notes() != "" {
			text += fmt.Sprintf(" (%s)", ing.Notes())
		}
		text += "\n"
	}

	text += "\nInstructions:\n"
	for _, inst := range rec.Instructions() {
		text += fmt.Sprintf("%d. %s\n", inst.StepNumber(), inst.Text())
	}

	return text
}
