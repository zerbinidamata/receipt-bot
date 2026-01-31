package command

import (
	"context"
	"fmt"
	"log"

	"receipt-bot/internal/domain/matching"
	"receipt-bot/internal/domain/recipe"
)

// BackfillNormalizedIngredientsCommand handles caching normalized ingredients for existing recipes
type BackfillNormalizedIngredientsCommand struct {
	recipeRepo recipe.Repository
	normalizer matching.IngredientNormalizer
}

// NewBackfillNormalizedIngredientsCommand creates a new backfill normalized ingredients command
func NewBackfillNormalizedIngredientsCommand(
	recipeRepo recipe.Repository,
) *BackfillNormalizedIngredientsCommand {
	return &BackfillNormalizedIngredientsCommand{
		recipeRepo: recipeRepo,
		normalizer: matching.NewRuleBasedNormalizer(),
	}
}

// BackfillNormalizedResult contains the result of a backfill operation
type BackfillNormalizedResult struct {
	TotalProcessed int
	Updated        int
	Skipped        int
	Errors         int
}

// Execute runs the backfill operation for a specific user
func (c *BackfillNormalizedIngredientsCommand) Execute(ctx context.Context, userID recipe.UserID) (*BackfillNormalizedResult, error) {
	// Fetch all recipes for the user
	recipes, err := c.recipeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}

	result := &BackfillNormalizedResult{}

	for _, rec := range recipes {
		result.TotalProcessed++

		// Skip recipes that already have normalized ingredients
		if rec.HasNormalizedIngredients() {
			result.Skipped++
			continue
		}

		// Normalize all ingredients
		normalizedIngredients := make([]string, 0, len(rec.Ingredients()))
		for _, ing := range rec.Ingredients() {
			normalized := c.normalizer.Normalize(ing.Name())
			if normalized != "" {
				normalizedIngredients = append(normalizedIngredients, normalized)
			}
		}

		// Update the recipe with normalized ingredients
		rec.SetNormalizedIngredients(normalizedIngredients)

		if err := c.recipeRepo.Update(ctx, rec); err != nil {
			result.Errors++
			log.Printf("Failed to update recipe %s: %v", rec.ID().String(), err)
			continue
		}

		result.Updated++
	}

	return result, nil
}

// ExecuteAll runs the backfill operation for all provided user IDs
func (c *BackfillNormalizedIngredientsCommand) ExecuteAll(ctx context.Context, userIDs []recipe.UserID) (*BackfillNormalizedResult, error) {
	totalResult := &BackfillNormalizedResult{}

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
	}

	return totalResult, nil
}
