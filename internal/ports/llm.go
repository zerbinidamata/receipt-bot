package ports

import (
	"context"
	"time"
)

// LLMPort defines the interface for LLM-based recipe extraction
type LLMPort interface {
	// ExtractRecipe parses text into structured recipe format
	ExtractRecipe(ctx context.Context, text string) (*RecipeExtraction, error)
}

// RecipeExtraction contains the structured recipe data extracted by LLM
type RecipeExtraction struct {
	Title        string
	Ingredients  []IngredientData
	Instructions []InstructionData
	PrepTime     *time.Duration
	CookTime     *time.Duration
	Servings     *int
}

// IngredientData represents ingredient information from LLM
type IngredientData struct {
	Name     string
	Quantity string
	Unit     string
	Notes    string
}

// InstructionData represents instruction information from LLM
type InstructionData struct {
	StepNumber int
	Text       string
	Duration   *time.Duration
}
