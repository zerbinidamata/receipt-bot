package recipe

import (
	"receipt-bot/internal/domain/shared"
	"strings"
	"time"
)

// RecipeID represents a unique recipe identifier
type RecipeID = shared.ID

// UserID represents a unique user identifier
type UserID = shared.ID

// Recipe represents a cooking recipe (Aggregate Root)
type Recipe struct {
	id           RecipeID
	userID       UserID
	title        string
	ingredients  []Ingredient
	instructions []Instruction
	source       Source
	transcript   string
	captions     string
	prepTime     *time.Duration
	cookTime     *time.Duration
	servings     *int
	createdAt    shared.Timestamp
	updatedAt    shared.Timestamp
}

// NewRecipe creates a new Recipe
func NewRecipe(
	userID UserID,
	title string,
	ingredients []Ingredient,
	instructions []Instruction,
	source Source,
	transcript string,
	captions string,
) (*Recipe, error) {
	title = strings.TrimSpace(title)

	if userID.IsEmpty() {
		return nil, shared.ErrInvalidInput
	}

	if title == "" {
		return nil, shared.ErrInvalidRecipeTitle
	}

	if len(ingredients) == 0 {
		return nil, shared.ErrNoIngredients
	}

	if len(instructions) == 0 {
		return nil, shared.ErrNoInstructions
	}

	if !source.IsValid() {
		return nil, shared.ErrInvalidSource
	}

	now := shared.NewTimestamp()

	return &Recipe{
		id:           shared.NewID(),
		userID:       userID,
		title:        title,
		ingredients:  ingredients,
		instructions: instructions,
		source:       source,
		transcript:   transcript,
		captions:     captions,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstructRecipe reconstructs a recipe from stored data (for repository)
func ReconstructRecipe(
	id RecipeID,
	userID UserID,
	title string,
	ingredients []Ingredient,
	instructions []Instruction,
	source Source,
	transcript string,
	captions string,
	prepTime *time.Duration,
	cookTime *time.Duration,
	servings *int,
	createdAt time.Time,
	updatedAt time.Time,
) *Recipe {
	return &Recipe{
		id:           id,
		userID:       userID,
		title:        title,
		ingredients:  ingredients,
		instructions: instructions,
		source:       source,
		transcript:   transcript,
		captions:     captions,
		prepTime:     prepTime,
		cookTime:     cookTime,
		servings:     servings,
		createdAt:    shared.NewTimestampFromTime(createdAt),
		updatedAt:    shared.NewTimestampFromTime(updatedAt),
	}
}

// ID returns the recipe ID
func (r *Recipe) ID() RecipeID {
	return r.id
}

// UserID returns the user ID
func (r *Recipe) UserID() UserID {
	return r.userID
}

// Title returns the recipe title
func (r *Recipe) Title() string {
	return r.title
}

// Ingredients returns the recipe ingredients
func (r *Recipe) Ingredients() []Ingredient {
	return r.ingredients
}

// Instructions returns the recipe instructions
func (r *Recipe) Instructions() []Instruction {
	return r.instructions
}

// Source returns the recipe source
func (r *Recipe) Source() Source {
	return r.source
}

// Transcript returns the video transcript
func (r *Recipe) Transcript() string {
	return r.transcript
}

// Captions returns the video captions
func (r *Recipe) Captions() string {
	return r.captions
}

// PrepTime returns the preparation time
func (r *Recipe) PrepTime() *time.Duration {
	return r.prepTime
}

// CookTime returns the cooking time
func (r *Recipe) CookTime() *time.Duration {
	return r.cookTime
}

// Servings returns the number of servings
func (r *Recipe) Servings() *int {
	return r.servings
}

// CreatedAt returns the creation timestamp
func (r *Recipe) CreatedAt() time.Time {
	return r.createdAt.Time()
}

// UpdatedAt returns the last update timestamp
func (r *Recipe) UpdatedAt() time.Time {
	return r.updatedAt.Time()
}

// SetPrepTime sets the preparation time
func (r *Recipe) SetPrepTime(duration time.Duration) {
	r.prepTime = &duration
	r.updatedAt = shared.NewTimestamp()
}

// SetCookTime sets the cooking time
func (r *Recipe) SetCookTime(duration time.Duration) {
	r.cookTime = &duration
	r.updatedAt = shared.NewTimestamp()
}

// SetServings sets the number of servings
func (r *Recipe) SetServings(servings int) {
	r.servings = &servings
	r.updatedAt = shared.NewTimestamp()
}

// AddIngredient adds an ingredient to the recipe
func (r *Recipe) AddIngredient(ingredient Ingredient) error {
	r.ingredients = append(r.ingredients, ingredient)
	r.updatedAt = shared.NewTimestamp()
	return nil
}

// AddInstruction adds an instruction to the recipe
func (r *Recipe) AddInstruction(instruction Instruction) error {
	r.instructions = append(r.instructions, instruction)
	r.updatedAt = shared.NewTimestamp()
	return nil
}

// Validate validates the recipe according to domain rules
func (r *Recipe) Validate() error {
	if r.title == "" {
		return shared.ErrInvalidRecipeTitle
	}

	if len(r.ingredients) == 0 {
		return shared.ErrNoIngredients
	}

	if len(r.instructions) == 0 {
		return shared.ErrNoInstructions
	}

	if !r.source.IsValid() {
		return shared.ErrInvalidSource
	}

	return nil
}
