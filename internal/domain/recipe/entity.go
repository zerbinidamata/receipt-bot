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
	category     Category
	cuisine      string
	dietaryTags  []DietaryTag
	tags         []string
	createdAt    shared.Timestamp
	updatedAt    shared.Timestamp

	// Multilingual support
	sourceLanguage         string        // ISO 639-1 language code (en, pt, es, etc.)
	translatedTitle        *string       // English translation (nil if source is English)
	translatedIngredients  []Ingredient  // English translations (nil if source is English)
	translatedInstructions []Instruction // English translations (nil if source is English)

	// Cached normalized ingredients for faster matching
	normalizedIngredients []string
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
		id:             shared.NewID(),
		userID:         userID,
		title:          title,
		ingredients:    ingredients,
		instructions:   instructions,
		source:         source,
		transcript:     transcript,
		captions:       captions,
		category:       CategoryOther,
		cuisine:        "",
		dietaryTags:    []DietaryTag{},
		tags:           []string{},
		sourceLanguage: "en", // Default to English
		createdAt:      now,
		updatedAt:      now,
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
	category Category,
	cuisine string,
	dietaryTags []DietaryTag,
	tags []string,
	createdAt time.Time,
	updatedAt time.Time,
) *Recipe {
	return ReconstructRecipeWithTranslations(
		id, userID, title, ingredients, instructions, source,
		transcript, captions, prepTime, cookTime, servings,
		category, cuisine, dietaryTags, tags, createdAt, updatedAt,
		"en", nil, nil, nil,
	)
}

// ReconstructRecipeWithTranslations reconstructs a recipe with translation data
func ReconstructRecipeWithTranslations(
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
	category Category,
	cuisine string,
	dietaryTags []DietaryTag,
	tags []string,
	createdAt time.Time,
	updatedAt time.Time,
	sourceLanguage string,
	translatedTitle *string,
	translatedIngredients []Ingredient,
	translatedInstructions []Instruction,
) *Recipe {
	return ReconstructRecipeWithNormalizedIngredients(
		id, userID, title, ingredients, instructions, source,
		transcript, captions, prepTime, cookTime, servings,
		category, cuisine, dietaryTags, tags, createdAt, updatedAt,
		sourceLanguage, translatedTitle, translatedIngredients, translatedInstructions,
		nil,
	)
}

// ReconstructRecipeWithNormalizedIngredients reconstructs a recipe with all fields including normalized ingredients
func ReconstructRecipeWithNormalizedIngredients(
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
	category Category,
	cuisine string,
	dietaryTags []DietaryTag,
	tags []string,
	createdAt time.Time,
	updatedAt time.Time,
	sourceLanguage string,
	translatedTitle *string,
	translatedIngredients []Ingredient,
	translatedInstructions []Instruction,
	normalizedIngredients []string,
) *Recipe {
	// Default category to Other if empty
	if category == "" {
		category = CategoryOther
	}
	// Initialize empty slices if nil
	if dietaryTags == nil {
		dietaryTags = []DietaryTag{}
	}
	if tags == nil {
		tags = []string{}
	}
	// Default source language to English if empty
	if sourceLanguage == "" {
		sourceLanguage = "en"
	}
	if normalizedIngredients == nil {
		normalizedIngredients = []string{}
	}

	return &Recipe{
		id:                     id,
		userID:                 userID,
		title:                  title,
		ingredients:            ingredients,
		instructions:           instructions,
		source:                 source,
		transcript:             transcript,
		captions:               captions,
		prepTime:               prepTime,
		cookTime:               cookTime,
		servings:               servings,
		category:               category,
		cuisine:                cuisine,
		dietaryTags:            dietaryTags,
		tags:                   tags,
		createdAt:              shared.NewTimestampFromTime(createdAt),
		updatedAt:              shared.NewTimestampFromTime(updatedAt),
		sourceLanguage:         sourceLanguage,
		translatedTitle:        translatedTitle,
		translatedIngredients:  translatedIngredients,
		translatedInstructions: translatedInstructions,
		normalizedIngredients:  normalizedIngredients,
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

// Category returns the recipe category
func (r *Recipe) Category() Category {
	return r.category
}

// Cuisine returns the cuisine type
func (r *Recipe) Cuisine() string {
	return r.cuisine
}

// DietaryTags returns the dietary tags
func (r *Recipe) DietaryTags() []DietaryTag {
	return r.dietaryTags
}

// Tags returns the free-form tags
func (r *Recipe) Tags() []string {
	return r.tags
}

// SourceLanguage returns the source language code
func (r *Recipe) SourceLanguage() string {
	if r.sourceLanguage == "" {
		return "en"
	}
	return r.sourceLanguage
}

// TranslatedTitle returns the English translation of the title (nil if source is English)
func (r *Recipe) TranslatedTitle() *string {
	return r.translatedTitle
}

// TranslatedIngredients returns the English translations of ingredients (nil if source is English)
func (r *Recipe) TranslatedIngredients() []Ingredient {
	return r.translatedIngredients
}

// TranslatedInstructions returns the English translations of instructions (nil if source is English)
func (r *Recipe) TranslatedInstructions() []Instruction {
	return r.translatedInstructions
}

// HasTranslation returns true if the recipe has translation data
func (r *Recipe) HasTranslation() bool {
	return r.translatedTitle != nil || len(r.translatedIngredients) > 0 || len(r.translatedInstructions) > 0
}

// NormalizedIngredients returns the cached normalized ingredient names
func (r *Recipe) NormalizedIngredients() []string {
	return r.normalizedIngredients
}

// SetNormalizedIngredients sets the cached normalized ingredient names
func (r *Recipe) SetNormalizedIngredients(normalized []string) {
	if normalized == nil {
		normalized = []string{}
	}
	r.normalizedIngredients = normalized
	r.updatedAt = shared.NewTimestamp()
}

// HasNormalizedIngredients returns true if the recipe has cached normalized ingredients
func (r *Recipe) HasNormalizedIngredients() bool {
	return len(r.normalizedIngredients) > 0
}

// IsEnglish returns true if the source language is English
func (r *Recipe) IsEnglish() bool {
	return r.sourceLanguage == "" || r.sourceLanguage == "en"
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

// SetCategory sets the recipe category
func (r *Recipe) SetCategory(category Category) {
	r.category = category
	r.updatedAt = shared.NewTimestamp()
}

// SetCuisine sets the cuisine type
func (r *Recipe) SetCuisine(cuisine string) {
	r.cuisine = strings.TrimSpace(cuisine)
	r.updatedAt = shared.NewTimestamp()
}

// SetDietaryTags sets the dietary tags
func (r *Recipe) SetDietaryTags(tags []DietaryTag) {
	if tags == nil {
		tags = []DietaryTag{}
	}
	r.dietaryTags = tags
	r.updatedAt = shared.NewTimestamp()
}

// SetTags sets the free-form tags
func (r *Recipe) SetTags(tags []string) {
	if tags == nil {
		tags = []string{}
	}
	r.tags = tags
	r.updatedAt = shared.NewTimestamp()
}

// SetSourceLanguage sets the source language code
func (r *Recipe) SetSourceLanguage(lang string) {
	r.sourceLanguage = lang
	r.updatedAt = shared.NewTimestamp()
}

// SetTranslations sets the translation data
func (r *Recipe) SetTranslations(title *string, ingredients []Ingredient, instructions []Instruction) {
	r.translatedTitle = title
	r.translatedIngredients = ingredients
	r.translatedInstructions = instructions
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
