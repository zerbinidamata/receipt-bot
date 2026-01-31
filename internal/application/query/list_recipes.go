package query

import (
	"context"
	"fmt"

	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/domain/recipe"
)

// ListRecipesQuery handles retrieving recipes for a user
type ListRecipesQuery struct {
	recipeRepo recipe.Repository
}

// NewListRecipesQuery creates a new query
func NewListRecipesQuery(recipeRepo recipe.Repository) *ListRecipesQuery {
	return &ListRecipesQuery{
		recipeRepo: recipeRepo,
	}
}

// Execute retrieves all recipes for a user
func (q *ListRecipesQuery) Execute(ctx context.Context, userID recipe.UserID) ([]*dto.RecipeDTO, error) {
	recipes, err := q.recipeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	// Convert to DTOs
	dtos := make([]*dto.RecipeDTO, len(recipes))
	for i, rec := range recipes {
		dtos[i] = convertToDTO(rec)
	}

	return dtos, nil
}

// ExecuteByIndex retrieves a specific recipe by its index (1-based) for a user
func (q *ListRecipesQuery) ExecuteByIndex(ctx context.Context, userID recipe.UserID, index int) (*dto.RecipeDTO, error) {
	recipes, err := q.recipeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}

	if index < 1 || index > len(recipes) {
		return nil, fmt.Errorf("recipe #%d not found (you have %d recipes)", index, len(recipes))
	}

	return convertToDTO(recipes[index-1]), nil
}

// ExecuteByCategory retrieves recipes filtered by category
func (q *ListRecipesQuery) ExecuteByCategory(ctx context.Context, userID recipe.UserID, category recipe.Category) ([]*dto.RecipeDTO, error) {
	recipes, err := q.recipeRepo.FindByUserIDAndCategory(ctx, userID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes by category: %w", err)
	}

	dtos := make([]*dto.RecipeDTO, len(recipes))
	for i, rec := range recipes {
		dtos[i] = convertToDTO(rec)
	}

	return dtos, nil
}

// GetCategoryCounts returns the count of recipes per category
func (q *ListRecipesQuery) GetCategoryCounts(ctx context.Context, userID recipe.UserID) (map[string]int, error) {
	counts, err := q.recipeRepo.GetCategoryCounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}

	// Convert to string keys for DTO
	result := make(map[string]int)
	for cat, count := range counts {
		result[string(cat)] = count
	}

	return result, nil
}

// SearchByIngredient searches recipes containing a specific ingredient
func (q *ListRecipesQuery) SearchByIngredient(ctx context.Context, userID recipe.UserID, ingredient string) ([]*dto.RecipeDTO, error) {
	recipes, err := q.recipeRepo.SearchByIngredient(ctx, userID, ingredient)
	if err != nil {
		return nil, fmt.Errorf("failed to search recipes by ingredient: %w", err)
	}

	dtos := make([]*dto.RecipeDTO, len(recipes))
	for i, rec := range recipes {
		dtos[i] = convertToDTO(rec)
	}

	return dtos, nil
}

// ExecuteByFilters retrieves recipes filtered by optional category and dietary tags
func (q *ListRecipesQuery) ExecuteByFilters(ctx context.Context, userID recipe.UserID, category *recipe.Category, dietaryTags []recipe.DietaryTag) ([]*dto.RecipeDTO, error) {
	recipes, err := q.recipeRepo.FindByUserIDAndFilters(ctx, userID, category, dietaryTags)
	if err != nil {
		return nil, fmt.Errorf("failed to filter recipes: %w", err)
	}

	dtos := make([]*dto.RecipeDTO, len(recipes))
	for i, rec := range recipes {
		dtos[i] = convertToDTO(rec)
	}

	return dtos, nil
}

// convertToDTO converts a domain Recipe to a DTO
func convertToDTO(rec *recipe.Recipe) *dto.RecipeDTO {
	recipeDTO := &dto.RecipeDTO{
		ID:             rec.ID().String(),
		UserID:         rec.UserID().String(),
		Title:          rec.Title(),
		SourceURL:      rec.Source().URL(),
		SourcePlatform: string(rec.Source().Platform()),
		SourceAuthor:   rec.Source().Author(),
		Transcript:     rec.Transcript(),
		Captions:       rec.Captions(),
		CreatedAt:      rec.CreatedAt(),
		UpdatedAt:      rec.UpdatedAt(),
	}

	// Convert ingredients
	recipeDTO.Ingredients = make([]dto.IngredientDTO, len(rec.Ingredients()))
	for i, ing := range rec.Ingredients() {
		recipeDTO.Ingredients[i] = dto.IngredientDTO{
			Name:     ing.Name(),
			Quantity: ing.Quantity(),
			Unit:     ing.Unit(),
			Notes:    ing.Notes(),
		}
	}

	// Convert instructions
	recipeDTO.Instructions = make([]dto.InstructionDTO, len(rec.Instructions()))
	for i, inst := range rec.Instructions() {
		var durationMinutes *int
		if inst.Duration() != nil {
			minutes := int(inst.Duration().Minutes())
			durationMinutes = &minutes
		}

		recipeDTO.Instructions[i] = dto.InstructionDTO{
			StepNumber:      inst.StepNumber(),
			Text:            inst.Text(),
			DurationMinutes: durationMinutes,
		}
	}

	// Convert optional times
	if rec.PrepTime() != nil {
		minutes := int(rec.PrepTime().Minutes())
		recipeDTO.PrepTimeMinutes = &minutes
	}
	if rec.CookTime() != nil {
		minutes := int(rec.CookTime().Minutes())
		recipeDTO.CookTimeMinutes = &minutes
	}

	recipeDTO.Servings = rec.Servings()

	// Convert category fields
	recipeDTO.Category = string(rec.Category())
	recipeDTO.Cuisine = rec.Cuisine()

	recipeDTO.DietaryTags = make([]string, len(rec.DietaryTags()))
	for i, tag := range rec.DietaryTags() {
		recipeDTO.DietaryTags[i] = string(tag)
	}

	recipeDTO.Tags = rec.Tags()

	return recipeDTO
}
