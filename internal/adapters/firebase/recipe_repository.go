package firebase

import (
	"context"
	"fmt"
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
	RecipeID        string                 `firestore:"recipeId"`
	UserID          string                 `firestore:"userId"`
	Title           string                 `firestore:"title"`
	Ingredients     []ingredientDoc        `firestore:"ingredients"`
	Instructions    []instructionDoc       `firestore:"instructions"`
	Source          sourceDoc              `firestore:"source"`
	Transcript      string                 `firestore:"transcript"`
	Captions        string                 `firestore:"captions"`
	PrepTimeMinutes *int                   `firestore:"prepTimeMinutes,omitempty"`
	CookTimeMinutes *int                   `firestore:"cookTimeMinutes,omitempty"`
	Servings        *int                   `firestore:"servings,omitempty"`
	CreatedAt       time.Time              `firestore:"createdAt"`
	UpdatedAt       time.Time              `firestore:"updatedAt"`
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

	// Reconstruct the recipe
	return recipe.ReconstructRecipe(
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
		doc.CreatedAt,
		doc.UpdatedAt,
	)
}
