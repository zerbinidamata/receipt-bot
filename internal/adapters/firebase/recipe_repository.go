package firebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
)

// RecipeRepository implements the recipe.Repository interface using Firestore
type RecipeRepository struct {
	client *firestore.Client
}

// NewRecipeRepository creates a new Firebase recipe repository
func NewRecipeRepository(client *firestore.Client) *RecipeRepository {
	return &RecipeRepository{
		client: client,
	}
}

// recipeDoc represents the Firestore document structure
type recipeDoc struct {
	RecipeID        string           `firestore:"recipeId"`
	UserID          string           `firestore:"userId"`
	Title           string           `firestore:"title"`
	Ingredients     []ingredientDoc  `firestore:"ingredients"`
	Instructions    []instructionDoc `firestore:"instructions"`
	Source          sourceDoc        `firestore:"source"`
	Transcript      string           `firestore:"transcript"`
	Captions        string           `firestore:"captions"`
	PrepTimeMinutes *int             `firestore:"prepTimeMinutes,omitempty"`
	CookTimeMinutes *int             `firestore:"cookTimeMinutes,omitempty"`
	Servings        *int             `firestore:"servings,omitempty"`
	Category        string           `firestore:"category,omitempty"`
	Cuisine         string           `firestore:"cuisine,omitempty"`
	DietaryTags     []string         `firestore:"dietaryTags,omitempty"`
	Tags            []string         `firestore:"tags,omitempty"`
	CreatedAt       time.Time        `firestore:"createdAt"`
	UpdatedAt       time.Time        `firestore:"updatedAt"`

	// Multilingual support
	SourceLanguage         string           `firestore:"sourceLanguage,omitempty"`
	TranslatedTitle        *string          `firestore:"translatedTitle,omitempty"`
	TranslatedIngredients  []ingredientDoc  `firestore:"translatedIngredients,omitempty"`
	TranslatedInstructions []instructionDoc `firestore:"translatedInstructions,omitempty"`

	// Cached normalized ingredients for faster matching
	NormalizedIngredients []string `firestore:"normalizedIngredients,omitempty"`
}

type ingredientDoc struct {
	Name     string `firestore:"name"`
	Quantity string `firestore:"quantity"`
	Unit     string `firestore:"unit"`
	Notes    string `firestore:"notes"`
}

type instructionDoc struct {
	StepNumber      int  `firestore:"stepNumber"`
	Text            string `firestore:"text"`
	DurationMinutes *int   `firestore:"durationMinutes,omitempty"`
}

type sourceDoc struct {
	URL      string `firestore:"url"`
	Platform string `firestore:"platform"`
	Author   string `firestore:"author"`
}

// Save persists a recipe to Firestore
func (r *RecipeRepository) Save(ctx context.Context, rec *recipe.Recipe) error {
	doc := r.toDocument(rec)

	_, err := r.client.Collection("recipes").Doc(rec.ID().String()).Set(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to save recipe: %w", err)
	}

	return nil
}

// FindByID retrieves a recipe by its ID
func (r *RecipeRepository) FindByID(ctx context.Context, id recipe.RecipeID) (*recipe.Recipe, error) {
	doc, err := r.client.Collection("recipes").Doc(id.String()).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find recipe: %w", err)
	}

	var recipeDoc recipeDoc
	if err := doc.DataTo(&recipeDoc); err != nil {
		return nil, fmt.Errorf("failed to parse recipe document: %w", err)
	}

	return r.fromDocument(&recipeDoc), nil
}

// FindByUserID retrieves all recipes for a user
func (r *RecipeRepository) FindByUserID(ctx context.Context, userID recipe.UserID) ([]*recipe.Recipe, error) {
	iter := r.client.Collection("recipes").
		Where("userId", "==", userID.String()).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var recipes []*recipe.Recipe
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate recipes: %w", err)
		}

		var recipeDoc recipeDoc
		if err := doc.DataTo(&recipeDoc); err != nil {
			continue // Skip invalid documents
		}

		recipes = append(recipes, r.fromDocument(&recipeDoc))
	}

	return recipes, nil
}

// FindBySourceURL retrieves a recipe by its source URL
func (r *RecipeRepository) FindBySourceURL(ctx context.Context, sourceURL string) (*recipe.Recipe, error) {
	iter := r.client.Collection("recipes").
		Where("source.url", "==", sourceURL).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, shared.ErrRecipeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find recipe by source URL: %w", err)
	}

	var recipeDoc recipeDoc
	if err := doc.DataTo(&recipeDoc); err != nil {
		return nil, fmt.Errorf("failed to parse recipe document: %w", err)
	}

	return r.fromDocument(&recipeDoc), nil
}

// FindByUserIDAndCategory retrieves recipes for a user filtered by category
func (r *RecipeRepository) FindByUserIDAndCategory(ctx context.Context, userID recipe.UserID, category recipe.Category) ([]*recipe.Recipe, error) {
	iter := r.client.Collection("recipes").
		Where("userId", "==", userID.String()).
		Where("category", "==", string(category)).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var recipes []*recipe.Recipe
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate recipes: %w", err)
		}

		var recipeDoc recipeDoc
		if err := doc.DataTo(&recipeDoc); err != nil {
			continue // Skip invalid documents
		}

		recipes = append(recipes, r.fromDocument(&recipeDoc))
	}

	return recipes, nil
}

// GetCategoryCounts returns the count of recipes per category for a user
func (r *RecipeRepository) GetCategoryCounts(ctx context.Context, userID recipe.UserID) (map[recipe.Category]int, error) {
	// Fetch all recipes for user and count locally
	// (Firestore doesn't support GROUP BY, so we count in-memory)
	recipes, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	counts := make(map[recipe.Category]int)
	for _, rec := range recipes {
		counts[rec.Category()]++
	}

	return counts, nil
}

// SearchByIngredient searches recipes containing a specific ingredient in title or ingredients
func (r *RecipeRepository) SearchByIngredient(ctx context.Context, userID recipe.UserID, ingredient string) ([]*recipe.Recipe, error) {
	// Firestore doesn't support full-text search, so we fetch all and filter in-memory
	allRecipes, err := r.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Normalize search term
	searchTerm := strings.ToLower(strings.TrimSpace(ingredient))
	if searchTerm == "" {
		return allRecipes, nil
	}

	var matchingRecipes []*recipe.Recipe
	for _, rec := range allRecipes {
		// Check title
		if strings.Contains(strings.ToLower(rec.Title()), searchTerm) {
			matchingRecipes = append(matchingRecipes, rec)
			continue
		}

		// Check ingredients
		for _, ing := range rec.Ingredients() {
			if strings.Contains(strings.ToLower(ing.Name()), searchTerm) {
				matchingRecipes = append(matchingRecipes, rec)
				break
			}
		}
	}

	return matchingRecipes, nil
}

// FindByUserIDAndFilters retrieves recipes for a user with optional category and dietary tag filters
func (r *RecipeRepository) FindByUserIDAndFilters(ctx context.Context, userID recipe.UserID, category *recipe.Category, dietaryTags []recipe.DietaryTag) ([]*recipe.Recipe, error) {
	// Start with all user recipes
	var recipes []*recipe.Recipe
	var err error

	if category != nil {
		// Use category filter if provided
		recipes, err = r.FindByUserIDAndCategory(ctx, userID, *category)
	} else {
		recipes, err = r.FindByUserID(ctx, userID)
	}

	if err != nil {
		return nil, err
	}

	// If no dietary tags, return as-is
	if len(dietaryTags) == 0 {
		return recipes, nil
	}

	// Filter by dietary tags in-memory
	var filtered []*recipe.Recipe
	for _, rec := range recipes {
		if r.hasAllTags(rec, dietaryTags) {
			filtered = append(filtered, rec)
		}
	}

	return filtered, nil
}

// hasAllTags checks if a recipe has all the specified dietary tags
func (r *RecipeRepository) hasAllTags(rec *recipe.Recipe, requiredTags []recipe.DietaryTag) bool {
	recipeTags := make(map[recipe.DietaryTag]bool)
	for _, tag := range rec.DietaryTags() {
		recipeTags[tag] = true
	}

	for _, required := range requiredTags {
		if !recipeTags[required] {
			return false
		}
	}

	return true
}

// Update updates an existing recipe
func (r *RecipeRepository) Update(ctx context.Context, rec *recipe.Recipe) error {
	return r.Save(ctx, rec) // In Firestore, Set with merge accomplishes update
}

// Delete removes a recipe
func (r *RecipeRepository) Delete(ctx context.Context, id recipe.RecipeID) error {
	_, err := r.client.Collection("recipes").Doc(id.String()).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}
	return nil
}

// toDocument converts a domain Recipe to a Firestore document
func (r *RecipeRepository) toDocument(rec *recipe.Recipe) *recipeDoc {
	doc := &recipeDoc{
		RecipeID:   rec.ID().String(),
		UserID:     rec.UserID().String(),
		Title:      rec.Title(),
		Transcript: rec.Transcript(),
		Captions:   rec.Captions(),
		CreatedAt:  rec.CreatedAt(),
		UpdatedAt:  rec.UpdatedAt(),
	}

	// Convert ingredients
	doc.Ingredients = make([]ingredientDoc, len(rec.Ingredients()))
	for i, ing := range rec.Ingredients() {
		doc.Ingredients[i] = ingredientDoc{
			Name:     ing.Name(),
			Quantity: ing.Quantity(),
			Unit:     ing.Unit(),
			Notes:    ing.Notes(),
		}
	}

	// Convert instructions
	doc.Instructions = make([]instructionDoc, len(rec.Instructions()))
	for i, inst := range rec.Instructions() {
		var durationMinutes *int
		if inst.Duration() != nil {
			minutes := int(inst.Duration().Minutes())
			durationMinutes = &minutes
		}

		doc.Instructions[i] = instructionDoc{
			StepNumber:      inst.StepNumber(),
			Text:            inst.Text(),
			DurationMinutes: durationMinutes,
		}
	}

	// Convert source
	doc.Source = sourceDoc{
		URL:      rec.Source().URL(),
		Platform: string(rec.Source().Platform()),
		Author:   rec.Source().Author(),
	}

	// Convert optional times
	if rec.PrepTime() != nil {
		minutes := int(rec.PrepTime().Minutes())
		doc.PrepTimeMinutes = &minutes
	}

	if rec.CookTime() != nil {
		minutes := int(rec.CookTime().Minutes())
		doc.CookTimeMinutes = &minutes
	}

	doc.Servings = rec.Servings()

	// Convert category fields
	doc.Category = string(rec.Category())
	doc.Cuisine = rec.Cuisine()

	// Convert dietary tags to strings
	doc.DietaryTags = make([]string, len(rec.DietaryTags()))
	for i, tag := range rec.DietaryTags() {
		doc.DietaryTags[i] = string(tag)
	}

	doc.Tags = rec.Tags()

	// Convert multilingual fields
	doc.SourceLanguage = rec.SourceLanguage()
	doc.TranslatedTitle = rec.TranslatedTitle()

	// Convert normalized ingredients
	doc.NormalizedIngredients = rec.NormalizedIngredients()

	// Convert translated ingredients
	if rec.TranslatedIngredients() != nil {
		doc.TranslatedIngredients = make([]ingredientDoc, len(rec.TranslatedIngredients()))
		for i, ing := range rec.TranslatedIngredients() {
			doc.TranslatedIngredients[i] = ingredientDoc{
				Name:     ing.Name(),
				Quantity: ing.Quantity(),
				Unit:     ing.Unit(),
				Notes:    ing.Notes(),
			}
		}
	}

	// Convert translated instructions
	if rec.TranslatedInstructions() != nil {
		doc.TranslatedInstructions = make([]instructionDoc, len(rec.TranslatedInstructions()))
		for i, inst := range rec.TranslatedInstructions() {
			var durationMinutes *int
			if inst.Duration() != nil {
				minutes := int(inst.Duration().Minutes())
				durationMinutes = &minutes
			}

			doc.TranslatedInstructions[i] = instructionDoc{
				StepNumber:      inst.StepNumber(),
				Text:            inst.Text(),
				DurationMinutes: durationMinutes,
			}
		}
	}

	return doc
}

// fromDocument converts a Firestore document to a domain Recipe
func (r *RecipeRepository) fromDocument(doc *recipeDoc) *recipe.Recipe {
	// Convert ingredients
	ingredients := make([]recipe.Ingredient, len(doc.Ingredients))
	for i, ingDoc := range doc.Ingredients {
		ing, _ := recipe.NewIngredient(ingDoc.Name, ingDoc.Quantity, ingDoc.Unit, ingDoc.Notes)
		ingredients[i] = ing
	}

	// Convert instructions
	instructions := make([]recipe.Instruction, len(doc.Instructions))
	for i, instDoc := range doc.Instructions {
		var duration *time.Duration
		if instDoc.DurationMinutes != nil {
			d := time.Duration(*instDoc.DurationMinutes) * time.Minute
			duration = &d
		}

		inst, _ := recipe.NewInstruction(instDoc.StepNumber, instDoc.Text, duration)
		instructions[i] = inst
	}

	// Convert source
	platform := recipe.Platform(doc.Source.Platform)
	source, _ := recipe.NewSource(doc.Source.URL, platform, doc.Source.Author)

	// Convert optional times
	var prepTime, cookTime *time.Duration
	if doc.PrepTimeMinutes != nil {
		d := time.Duration(*doc.PrepTimeMinutes) * time.Minute
		prepTime = &d
	}
	if doc.CookTimeMinutes != nil {
		d := time.Duration(*doc.CookTimeMinutes) * time.Minute
		cookTime = &d
	}

	// Convert category
	category := recipe.CategoryFromLLM(doc.Category)

	// Convert dietary tags
	dietaryTags := make([]recipe.DietaryTag, 0, len(doc.DietaryTags))
	for _, tagStr := range doc.DietaryTags {
		tag, valid := recipe.ParseDietaryTag(tagStr)
		if valid {
			dietaryTags = append(dietaryTags, tag)
		}
	}

	// Convert translated ingredients
	var translatedIngredients []recipe.Ingredient
	if len(doc.TranslatedIngredients) > 0 {
		translatedIngredients = make([]recipe.Ingredient, len(doc.TranslatedIngredients))
		for i, ingDoc := range doc.TranslatedIngredients {
			ing, _ := recipe.NewIngredient(ingDoc.Name, ingDoc.Quantity, ingDoc.Unit, ingDoc.Notes)
			translatedIngredients[i] = ing
		}
	}

	// Convert translated instructions
	var translatedInstructions []recipe.Instruction
	if len(doc.TranslatedInstructions) > 0 {
		translatedInstructions = make([]recipe.Instruction, len(doc.TranslatedInstructions))
		for i, instDoc := range doc.TranslatedInstructions {
			var duration *time.Duration
			if instDoc.DurationMinutes != nil {
				d := time.Duration(*instDoc.DurationMinutes) * time.Minute
				duration = &d
			}

			inst, _ := recipe.NewInstruction(instDoc.StepNumber, instDoc.Text, duration)
			translatedInstructions[i] = inst
		}
	}

	// Reconstruct the recipe with all fields including normalized ingredients
	return recipe.ReconstructRecipeWithNormalizedIngredients(
		recipe.RecipeID(doc.RecipeID),
		recipe.UserID(doc.UserID),
		doc.Title,
		ingredients,
		instructions,
		source,
		doc.Transcript,
		doc.Captions,
		prepTime,
		cookTime,
		doc.Servings,
		category,
		doc.Cuisine,
		dietaryTags,
		doc.Tags,
		doc.CreatedAt,
		doc.UpdatedAt,
		doc.SourceLanguage,
		doc.TranslatedTitle,
		translatedIngredients,
		translatedInstructions,
		doc.NormalizedIngredients,
	)
}
