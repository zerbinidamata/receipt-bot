package command

import (
	"context"
	"fmt"

	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/domain/matching"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
)

// MatchIngredientsCommand handles matching user ingredients to recipes
type MatchIngredientsCommand struct {
	recipeRepo recipe.Repository
	normalizer matching.IngredientNormalizer
	matcher    *matching.IngredientMatcher
}

// NewMatchIngredientsCommand creates a new command
func NewMatchIngredientsCommand(recipeRepo recipe.Repository) *MatchIngredientsCommand {
	normalizer := matching.NewRuleBasedNormalizer()
	return &MatchIngredientsCommand{
		recipeRepo: recipeRepo,
		normalizer: normalizer,
		matcher:    matching.NewIngredientMatcher(normalizer),
	}
}

// MatchIngredientsInput holds the input parameters
type MatchIngredientsInput struct {
	UserID         shared.ID
	Ingredients    []string
	CategoryFilter *recipe.Category
	StrictMatch    bool
}

// Execute finds recipes matching the given ingredients
func (c *MatchIngredientsCommand) Execute(ctx context.Context, input MatchIngredientsInput) (*dto.MatchIngredientsResultDTO, error) {
	// Fetch user's recipes
	recipes, err := c.recipeRepo.FindByUserID(ctx, recipe.UserID(input.UserID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}

	if len(recipes) == 0 {
		return &dto.MatchIngredientsResultDTO{
			PerfectMatches: []dto.MatchResultDTO{},
			HighMatches:    []dto.MatchResultDTO{},
			MediumMatches:  []dto.MatchResultDTO{},
			TotalMatches:   0,
		}, nil
	}

	// Configure matching options
	options := matching.DefaultMatchOptions()
	options.StrictMatch = input.StrictMatch
	options.CategoryFilter = input.CategoryFilter

	// Perform matching
	results := c.matcher.Match(input.Ingredients, recipes, options)

	// Group by match level
	grouped := matching.GroupByMatchLevel(results)

	// Convert to DTOs
	resultDTO := &dto.MatchIngredientsResultDTO{
		PerfectMatches: convertMatchResults(grouped[matching.MatchLevelPerfect]),
		HighMatches:    convertMatchResults(grouped[matching.MatchLevelHigh]),
		MediumMatches:  convertMatchResults(grouped[matching.MatchLevelMedium]),
		TotalMatches:   len(results),
	}

	return resultDTO, nil
}

// convertMatchResults converts domain match results to DTOs
func convertMatchResults(results []matching.MatchResult) []dto.MatchResultDTO {
	if results == nil {
		return []dto.MatchResultDTO{}
	}

	dtos := make([]dto.MatchResultDTO, len(results))
	for i, result := range results {
		dtos[i] = dto.MatchResultDTO{
			Recipe:          convertRecipeToDTO(result.Recipe),
			MatchPercentage: result.MatchPercentage,
			MatchedItems:    result.MatchedItems,
			MissingItems:    result.MissingItems,
			MatchLevel:      matching.MatchLevelString(result.MatchLevel),
		}
	}
	return dtos
}

// convertRecipeToDTO converts a recipe to DTO (simplified version)
func convertRecipeToDTO(rec *recipe.Recipe) *dto.RecipeDTO {
	recipeDTO := &dto.RecipeDTO{
		ID:             rec.ID().String(),
		UserID:         rec.UserID().String(),
		Title:          rec.Title(),
		SourceURL:      rec.Source().URL(),
		SourcePlatform: string(rec.Source().Platform()),
		SourceAuthor:   rec.Source().Author(),
		Category:       string(rec.Category()),
		Cuisine:        rec.Cuisine(),
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

	// Convert dietary tags
	recipeDTO.DietaryTags = make([]string, len(rec.DietaryTags()))
	for i, tag := range rec.DietaryTags() {
		recipeDTO.DietaryTags[i] = string(tag)
	}

	recipeDTO.Tags = rec.Tags()

	return recipeDTO
}
