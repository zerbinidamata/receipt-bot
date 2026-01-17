package shared

import "errors"

// Domain errors
var (
	// Recipe errors
	ErrRecipeNotFound       = errors.New("recipe not found")
	ErrInvalidRecipeTitle   = errors.New("recipe title cannot be empty")
	ErrNoIngredients        = errors.New("recipe must have at least one ingredient")
	ErrNoInstructions       = errors.New("recipe must have at least one instruction")
	ErrInvalidSource        = errors.New("invalid recipe source")

	// Ingredient errors
	ErrInvalidIngredientName = errors.New("ingredient name cannot be empty")
	ErrInvalidQuantity       = errors.New("ingredient quantity cannot be empty")

	// Instruction errors
	ErrInvalidInstructionText = errors.New("instruction text cannot be empty")
	ErrInvalidStepNumber      = errors.New("instruction step number must be positive")

	// Source errors
	ErrInvalidURL      = errors.New("invalid URL")
	ErrInvalidPlatform = errors.New("invalid platform")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidTelegramID  = errors.New("invalid telegram ID")
	ErrInvalidUsername    = errors.New("invalid username")

	// General errors
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)
