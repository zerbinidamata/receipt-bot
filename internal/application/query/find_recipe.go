package query

import (
	"context"
	"fmt"

	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/domain/recipe"
)

// FindRecipeQuery handles retrieving a single recipe
type FindRecipeQuery struct {
	recipeRepo recipe.Repository
}

// NewFindRecipeQuery creates a new query
func NewFindRecipeQuery(recipeRepo recipe.Repository) *FindRecipeQuery {
	return &FindRecipeQuery{
		recipeRepo: recipeRepo,
	}
}

// Execute retrieves a recipe by ID
func (q *FindRecipeQuery) Execute(ctx context.Context, recipeID recipe.RecipeID) (*dto.RecipeDTO, error) {
	rec, err := q.recipeRepo.FindByID(ctx, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find recipe: %w", err)
	}

	return convertToDTO(rec), nil
}
