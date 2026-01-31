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
	Category     string
	Cuisine      string
	DietaryTags  []string
	Tags         []string

	// Multilingual support
	SourceLanguage         string            // ISO 639-1 language code (en, pt, es, etc.)
	TranslatedTitle        *string           // English translation (nil if source is English)
	TranslatedIngredients  []IngredientData  // English translations (nil if source is English)
	TranslatedInstructions []InstructionData // English translations (nil if source is English)
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
