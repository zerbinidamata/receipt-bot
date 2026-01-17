package dto

import "time"

// RecipeDTO is a data transfer object for recipes
type RecipeDTO struct {
	ID               string
	UserID           string
	Title            string
	Ingredients      []IngredientDTO
	Instructions     []InstructionDTO
	SourceURL        string
	SourcePlatform   string
	SourceAuthor     string
	Transcript       string
	Captions         string
	PrepTimeMinutes  *int
	CookTimeMinutes  *int
	Servings         *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
