package query

import (
	"context"
	"testing"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
)

// mockRecipeRepository is a mock implementation of recipe.Repository for testing
type mockRecipeRepository struct {
	recipes []*recipe.Recipe
	err     error
}

func newMockRepo(recipes []*recipe.Recipe) *mockRecipeRepository {
	return &mockRecipeRepository{recipes: recipes}
}

func (m *mockRecipeRepository) Save(ctx context.Context, rec *recipe.Recipe) error {
	m.recipes = append(m.recipes, rec)
	return m.err
}

func (m *mockRecipeRepository) FindByID(ctx context.Context, id recipe.RecipeID) (*recipe.Recipe, error) {
	for _, rec := range m.recipes {
		if rec.ID() == id {
			return rec, nil
		}
	}
	return nil, shared.ErrRecipeNotFound
}

func (m *mockRecipeRepository) FindByUserID(ctx context.Context, userID recipe.UserID) ([]*recipe.Recipe, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*recipe.Recipe
	for _, rec := range m.recipes {
		if rec.UserID() == userID {
			result = append(result, rec)
		}
	}
	return result, nil
}

func (m *mockRecipeRepository) FindByUserIDAndCategory(ctx context.Context, userID recipe.UserID, category recipe.Category) ([]*recipe.Recipe, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*recipe.Recipe
	for _, rec := range m.recipes {
		if rec.UserID() == userID && rec.Category() == category {
			result = append(result, rec)
		}
	}
	return result, nil
}

func (m *mockRecipeRepository) FindByUserIDAndFilters(ctx context.Context, userID recipe.UserID, category *recipe.Category, dietaryTags []recipe.DietaryTag) ([]*recipe.Recipe, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*recipe.Recipe
	for _, rec := range m.recipes {
		if rec.UserID() != userID {
			continue
		}
		if category != nil && rec.Category() != *category {
			continue
		}
		if len(dietaryTags) > 0 && !hasAllTags(rec, dietaryTags) {
			continue
		}
		result = append(result, rec)
	}
	return result, nil
}

func hasAllTags(rec *recipe.Recipe, requiredTags []recipe.DietaryTag) bool {
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

func (m *mockRecipeRepository) SearchByIngredient(ctx context.Context, userID recipe.UserID, ingredient string) ([]*recipe.Recipe, error) {
	return m.FindByUserID(ctx, userID) // Simplified for testing
}

func (m *mockRecipeRepository) FindBySourceURL(ctx context.Context, sourceURL string) (*recipe.Recipe, error) {
	return nil, shared.ErrRecipeNotFound
}

func (m *mockRecipeRepository) GetCategoryCounts(ctx context.Context, userID recipe.UserID) (map[recipe.Category]int, error) {
	if m.err != nil {
		return nil, m.err
	}
	counts := make(map[recipe.Category]int)
	for _, rec := range m.recipes {
		if rec.UserID() == userID {
			counts[rec.Category()]++
		}
	}
	return counts, nil
}

func (m *mockRecipeRepository) Update(ctx context.Context, rec *recipe.Recipe) error {
	return m.err
}

func (m *mockRecipeRepository) Delete(ctx context.Context, id recipe.RecipeID) error {
	return m.err
}

func createTestRecipe(userID recipe.UserID, title string, category recipe.Category, tags []recipe.DietaryTag) *recipe.Recipe {
	ing, _ := recipe.NewIngredient("flour", "2", "cups", "")
	inst, _ := recipe.NewInstruction(1, "Mix", nil)
	source, _ := recipe.NewSource("https://example.com", recipe.PlatformWeb, "Chef")

	rec, _ := recipe.NewRecipe(userID, title, []recipe.Ingredient{ing}, []recipe.Instruction{inst}, source, "", "")
	rec.SetCategory(category)
	rec.SetDietaryTags(tags)
	return rec
}

func TestListRecipesQuery_Execute(t *testing.T) {
	userID := shared.NewID()

	recipes := []*recipe.Recipe{
		createTestRecipe(userID, "Pasta Recipe", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Salad Recipe", recipe.CategorySalads, nil),
	}

	repo := newMockRepo(recipes)
	query := NewListRecipesQuery(repo)

	result, err := query.Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Execute() returned %d recipes, want 2", len(result))
	}
}

func TestListRecipesQuery_ExecuteByCategory(t *testing.T) {
	userID := shared.NewID()

	recipes := []*recipe.Recipe{
		createTestRecipe(userID, "Spaghetti", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Lasagna", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Caesar Salad", recipe.CategorySalads, nil),
		createTestRecipe(userID, "Grilled Salmon", recipe.CategorySeafood, nil),
	}

	repo := newMockRepo(recipes)
	query := NewListRecipesQuery(repo)

	tests := []struct {
		name      string
		category  recipe.Category
		wantCount int
	}{
		{"Pasta recipes", recipe.CategoryPasta, 2},
		{"Salad recipes", recipe.CategorySalads, 1},
		{"Seafood recipes", recipe.CategorySeafood, 1},
		{"Dessert recipes (none)", recipe.CategoryDesserts, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := query.ExecuteByCategory(context.Background(), userID, tt.category)
			if err != nil {
				t.Fatalf("ExecuteByCategory() error = %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("ExecuteByCategory(%v) returned %d recipes, want %d", tt.category, len(result), tt.wantCount)
			}

			// Verify all returned recipes have the correct category
			for _, r := range result {
				if r.Category != string(tt.category) {
					t.Errorf("Recipe has category %s, want %s", r.Category, tt.category)
				}
			}
		})
	}
}

func TestListRecipesQuery_GetCategoryCounts(t *testing.T) {
	userID := shared.NewID()

	recipes := []*recipe.Recipe{
		createTestRecipe(userID, "Spaghetti", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Lasagna", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Fettuccine", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Caesar Salad", recipe.CategorySalads, nil),
		createTestRecipe(userID, "Grilled Salmon", recipe.CategorySeafood, nil),
	}

	repo := newMockRepo(recipes)
	query := NewListRecipesQuery(repo)

	counts, err := query.GetCategoryCounts(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCategoryCounts() error = %v", err)
	}

	expectedCounts := map[string]int{
		"Pasta & Noodles": 3,
		"Salads":          1,
		"Seafood":         1,
	}

	for cat, expected := range expectedCounts {
		if counts[cat] != expected {
			t.Errorf("GetCategoryCounts()[%s] = %d, want %d", cat, counts[cat], expected)
		}
	}
}

func TestListRecipesQuery_ExecuteByFilters(t *testing.T) {
	userID := shared.NewID()

	recipes := []*recipe.Recipe{
		createTestRecipe(userID, "Vegan Pasta", recipe.CategoryPasta, []recipe.DietaryTag{recipe.TagVegan}),
		createTestRecipe(userID, "Regular Pasta", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Vegan Salad", recipe.CategorySalads, []recipe.DietaryTag{recipe.TagVegan, recipe.TagGlutenFree}),
		createTestRecipe(userID, "Quick Chicken", recipe.CategoryMeat, []recipe.DietaryTag{recipe.TagQuick}),
	}

	repo := newMockRepo(recipes)
	query := NewListRecipesQuery(repo)

	tests := []struct {
		name        string
		category    *recipe.Category
		dietaryTags []recipe.DietaryTag
		wantCount   int
	}{
		{
			name:        "all pasta",
			category:    categoryPtr(recipe.CategoryPasta),
			dietaryTags: nil,
			wantCount:   2,
		},
		{
			name:        "vegan only",
			category:    nil,
			dietaryTags: []recipe.DietaryTag{recipe.TagVegan},
			wantCount:   2,
		},
		{
			name:        "vegan pasta",
			category:    categoryPtr(recipe.CategoryPasta),
			dietaryTags: []recipe.DietaryTag{recipe.TagVegan},
			wantCount:   1,
		},
		{
			name:        "vegan gluten-free",
			category:    nil,
			dietaryTags: []recipe.DietaryTag{recipe.TagVegan, recipe.TagGlutenFree},
			wantCount:   1,
		},
		{
			name:        "no filters",
			category:    nil,
			dietaryTags: nil,
			wantCount:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := query.ExecuteByFilters(context.Background(), userID, tt.category, tt.dietaryTags)
			if err != nil {
				t.Fatalf("ExecuteByFilters() error = %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("ExecuteByFilters() returned %d recipes, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestListRecipesQuery_ExecuteByIndex(t *testing.T) {
	userID := shared.NewID()

	recipes := []*recipe.Recipe{
		createTestRecipe(userID, "First Recipe", recipe.CategoryPasta, nil),
		createTestRecipe(userID, "Second Recipe", recipe.CategorySalads, nil),
		createTestRecipe(userID, "Third Recipe", recipe.CategorySeafood, nil),
	}

	repo := newMockRepo(recipes)
	query := NewListRecipesQuery(repo)

	tests := []struct {
		name      string
		index     int
		wantTitle string
		wantErr   bool
	}{
		{"first recipe", 1, "First Recipe", false},
		{"second recipe", 2, "Second Recipe", false},
		{"third recipe", 3, "Third Recipe", false},
		{"index too high", 4, "", true},
		{"index zero", 0, "", true},
		{"negative index", -1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := query.ExecuteByIndex(context.Background(), userID, tt.index)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExecuteByIndex(%d) expected error but got nil", tt.index)
				}
				return
			}

			if err != nil {
				t.Fatalf("ExecuteByIndex(%d) error = %v", tt.index, err)
			}

			if result.Title != tt.wantTitle {
				t.Errorf("ExecuteByIndex(%d) returned title %s, want %s", tt.index, result.Title, tt.wantTitle)
			}
		})
	}
}

func categoryPtr(c recipe.Category) *recipe.Category {
	return &c
}
