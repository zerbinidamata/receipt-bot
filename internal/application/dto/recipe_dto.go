package dto

import "time"

// RecipeDTO is a data transfer object for recipes
type RecipeDTO struct {
	ID              string
	UserID          string
	Title           string
	Ingredients     []IngredientDTO
	Instructions    []InstructionDTO
	SourceURL       string
	SourcePlatform  string
	SourceAuthor    string
	Transcript      string
	Captions        string
	PrepTimeMinutes *int
	CookTimeMinutes *int
	Servings        *int
	Category        string
	Cuisine         string
	DietaryTags     []string
	Tags            []string
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Multilingual support
	SourceLanguage         string
	TranslatedTitle        *string
	TranslatedIngredients  []IngredientDTO
	TranslatedInstructions []InstructionDTO
}

// IngredientDTO represents an ingredient
type IngredientDTO struct {
	Name     string
	Quantity string
	Unit     string
	Notes    string
}

// InstructionDTO represents a cooking instruction
type InstructionDTO struct {
	StepNumber      int
	Text            string
	DurationMinutes *int
}

// ProcessRecipeLinkRequest is the request for processing a recipe link
type ProcessRecipeLinkRequest struct {
	URL       string
	UserID    string
	TelegramChatID int64
}

// ProcessRecipeLinkResponse is the response after processing
type ProcessRecipeLinkResponse struct {
	Recipe  *RecipeDTO
	Success bool
	Error   string
}

// MatchResultDTO represents a recipe match result
type MatchResultDTO struct {
	Recipe          *RecipeDTO
	MatchPercentage float64
	MatchedItems    []string
	MissingItems    []string
	MatchLevel      string
}

// MatchIngredientsResultDTO contains grouped match results
type MatchIngredientsResultDTO struct {
	PerfectMatches []MatchResultDTO
	HighMatches    []MatchResultDTO
	MediumMatches  []MatchResultDTO
	TotalMatches   int
}

// PantryDTO represents user pantry data
type PantryDTO struct {
	Items     []string
	UpdatedAt *time.Time
}
