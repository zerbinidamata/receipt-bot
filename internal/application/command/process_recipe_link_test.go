package command

import (
	"context"
	"testing"
	"time"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/ports"
)

// Mock implementations for testing

type mockScraperPort struct {
	result *ports.ScrapeResult
	err    error
}

func (m *mockScraperPort) Scrape(ctx context.Context, req ports.ScrapeRequest) (*ports.ScrapeResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

type mockLLMPort struct {
	extraction *ports.RecipeExtraction
	err        error
}

func (m *mockLLMPort) ExtractRecipe(ctx context.Context, text string) (*ports.RecipeExtraction, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.extraction, nil
}

type mockRecipeRepository struct {
	recipes map[string]*recipe.Recipe
}

func newMockRecipeRepository() *mockRecipeRepository {
	return &mockRecipeRepository{
		recipes: make(map[string]*recipe.Recipe),
	}
}

func (m *mockRecipeRepository) Save(ctx context.Context, rec *recipe.Recipe) error {
	m.recipes[rec.ID().String()] = rec
	return nil
}

func (m *mockRecipeRepository) FindByID(ctx context.Context, id recipe.RecipeID) (*recipe.Recipe, error) {
	if rec, ok := m.recipes[id.String()]; ok {
		return rec, nil
	}
	return nil, shared.ErrRecipeNotFound
}

func (m *mockRecipeRepository) FindByUserID(ctx context.Context, userID recipe.UserID) ([]*recipe.Recipe, error) {
	var results []*recipe.Recipe
	for _, rec := range m.recipes {
		if rec.UserID() == userID {
			results = append(results, rec)
		}
	}
	return results, nil
}

func (m *mockRecipeRepository) FindBySourceURL(ctx context.Context, sourceURL string) (*recipe.Recipe, error) {
	for _, rec := range m.recipes {
		if rec.Source().URL() == sourceURL {
			return rec, nil
		}
	}
	return nil, shared.ErrRecipeNotFound
}

func (m *mockRecipeRepository) Update(ctx context.Context, rec *recipe.Recipe) error {
	m.recipes[rec.ID().String()] = rec
	return nil
}

func (m *mockRecipeRepository) Delete(ctx context.Context, id recipe.RecipeID) error {
	delete(m.recipes, id.String())
	return nil
}

type mockMessengerPort struct {
	messages []string
}

func (m *mockMessengerPort) SendMessage(ctx context.Context, chatID int64, text string) error {
	m.messages = append(m.messages, text)
	return nil
}

func (m *mockMessengerPort) SendRecipe(ctx context.Context, chatID int64, rec *recipe.Recipe) error {
	m.messages = append(m.messages, "Recipe: "+rec.Title())
	return nil
}

func (m *mockMessengerPort) SendProgress(ctx context.Context, chatID int64, message string) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *mockMessengerPort) SendError(ctx context.Context, chatID int64, errorMsg string) error {
	m.messages = append(m.messages, "Error: "+errorMsg)
	return nil
}

func TestProcessRecipeLinkCommand_Execute(t *testing.T) {
	ctx := context.Background()
	userID := shared.NewID()

	mockScraper := &mockScraperPort{
		result: &ports.ScrapeResult{
			Captions:    "This is a chocolate cake recipe",
			Transcript:  "Mix flour, sugar, and eggs. Bake at 350F.",
			OriginalURL: "https://youtube.com/watch?v=abc",
			Metadata: map[string]string{
				"author": "Chef John",
			},
		},
	}

	prepTime := 15 * time.Minute
	cookTime := 30 * time.Minute
	servings := 8

	mockLLM := &mockLLMPort{
		extraction: &ports.RecipeExtraction{
			Title: "Chocolate Cake",
			Ingredients: []ports.IngredientData{
				{Name: "flour", Quantity: "2", Unit: "cups", Notes: ""},
				{Name: "sugar", Quantity: "1", Unit: "cup", Notes: ""},
			},
			Instructions: []ports.InstructionData{
				{StepNumber: 1, Text: "Mix ingredients", Duration: nil},
				{StepNumber: 2, Text: "Bake", Duration: nil},
			},
			PrepTime: &prepTime,
			CookTime: &cookTime,
			Servings: &servings,
		},
	}

	mockRepo := newMockRecipeRepository()
	mockMessenger := &mockMessengerPort{}
	recipeService := recipe.NewService()

	cmd := NewProcessRecipeLinkCommand(
		mockScraper,
		mockLLM,
		recipeService,
		mockRepo,
		mockMessenger,
	)

	// Execute command
	rec, err := cmd.Execute(ctx, "https://youtube.com/watch?v=abc", userID, 12345)

	// Assertions
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}

	if rec == nil {
		t.Fatal("Execute() returned nil recipe")
	}

	if rec.Title() != "Chocolate Cake" {
		t.Errorf("Title = %v, want %v", rec.Title(), "Chocolate Cake")
	}

	if len(rec.Ingredients()) != 2 {
		t.Errorf("Ingredients count = %v, want 2", len(rec.Ingredients()))
	}

	if len(rec.Instructions()) != 2 {
		t.Errorf("Instructions count = %v, want 2", len(rec.Instructions()))
	}

	if rec.PrepTime() == nil || *rec.PrepTime() != prepTime {
		t.Errorf("PrepTime = %v, want %v", rec.PrepTime(), prepTime)
	}

	// Verify recipe was saved
	if len(mockRepo.recipes) != 1 {
		t.Errorf("Recipe repository has %v recipes, want 1", len(mockRepo.recipes))
	}

	// Verify progress messages were sent
	if len(mockMessenger.messages) == 0 {
		t.Error("No progress messages were sent")
	}
}

func TestProcessRecipeLinkCommand_Execute_NoIngredients(t *testing.T) {
	ctx := context.Background()
	userID := shared.NewID()

	mockScraper := &mockScraperPort{
		result: &ports.ScrapeResult{
			Captions:    "Some text",
			Transcript:  "Some transcript",
			OriginalURL: "https://youtube.com/watch?v=abc",
			Metadata:    map[string]string{},
		},
	}

	mockLLM := &mockLLMPort{
		extraction: &ports.RecipeExtraction{
			Title:        "Not a recipe",
			Ingredients:  []ports.IngredientData{}, // Empty!
			Instructions: []ports.InstructionData{{StepNumber: 1, Text: "Do nothing"}},
		},
	}

	mockRepo := newMockRecipeRepository()
	recipeService := recipe.NewService()

	cmd := NewProcessRecipeLinkCommand(
		mockScraper,
		mockLLM,
		recipeService,
		mockRepo,
		nil, // No messenger
	)

	// Execute command
	_, err := cmd.Execute(ctx, "https://youtube.com/watch?v=abc", userID, 12345)

	// Should fail because no ingredients
	if err == nil {
		t.Error("Execute() expected error for empty ingredients, got nil")
	}
}
