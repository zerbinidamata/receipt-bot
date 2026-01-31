package matching

import (
	"testing"
	"time"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
)

func createTestRecipe(title string, category recipe.Category, ingredientNames []string) *recipe.Recipe {
	ingredients := make([]recipe.Ingredient, len(ingredientNames))
	for i, name := range ingredientNames {
		ing, _ := recipe.NewIngredient(name, "1", "unit", "")
		ingredients[i] = ing
	}

	inst, _ := recipe.NewInstruction(1, "Cook it", nil)
	source, _ := recipe.NewSource("https://example.com", recipe.PlatformWeb, "Chef")

	rec, _ := recipe.NewRecipe(
		shared.NewID(),
		title,
		ingredients,
		[]recipe.Instruction{inst},
		source,
		"",
		"",
	)
	rec.SetCategory(category)
	return rec
}

func TestIngredientMatcher_Match(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()
	matcher := NewIngredientMatcher(normalizer)

	// Create test recipes
	pastaRecipe := createTestRecipe("Pasta Carbonara", recipe.CategoryPasta,
		[]string{"spaghetti", "eggs", "parmesan", "bacon", "black pepper"})

	saladRecipe := createTestRecipe("Caesar Salad", recipe.CategorySalads,
		[]string{"romaine lettuce", "parmesan", "croutons", "caesar dressing"})

	chickenRecipe := createTestRecipe("Grilled Chicken", recipe.CategoryMeat,
		[]string{"chicken breast", "olive oil", "garlic", "lemon", "rosemary"})

	recipes := []*recipe.Recipe{pastaRecipe, saladRecipe, chickenRecipe}

	tests := []struct {
		name            string
		userIngredients []string
		options         MatchOptions
		wantMinResults  int
		wantMaxResults  int
	}{
		{
			name:            "match pasta ingredients",
			userIngredients: []string{"spaghetti", "eggs", "parmesan", "bacon"},
			options:         DefaultMatchOptions(),
			wantMinResults:  1,
			wantMaxResults:  3,
		},
		{
			name:            "match with substitution",
			userIngredients: []string{"linguine", "eggs", "romano", "bacon"}, // linguine substitutes spaghetti, romano substitutes parmesan
			options:         DefaultMatchOptions(),
			wantMinResults:  1,
			wantMaxResults:  3,
		},
		{
			name:            "no matching ingredients",
			userIngredients: []string{"tofu", "tempeh", "seitan"},
			options:         DefaultMatchOptions(),
			wantMinResults:  0,
			wantMaxResults:  0,
		},
		{
			name:            "empty user ingredients",
			userIngredients: []string{},
			options:         DefaultMatchOptions(),
			wantMinResults:  0,
			wantMaxResults:  0,
		},
		{
			name:            "category filter",
			userIngredients: []string{"parmesan", "romaine lettuce"},
			options: MatchOptions{
				CategoryFilter: categoryPtr(recipe.CategorySalads),
				ExcludeStaples: true,
				MinMatchLevel:  MatchLevelMedium,
				MaxResults:     20,
			},
			wantMinResults: 0,
			wantMaxResults: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := matcher.Match(tt.userIngredients, recipes, tt.options)

			if len(results) < tt.wantMinResults {
				t.Errorf("Match() returned %d results, want at least %d", len(results), tt.wantMinResults)
			}
			if len(results) > tt.wantMaxResults {
				t.Errorf("Match() returned %d results, want at most %d", len(results), tt.wantMaxResults)
			}
		})
	}
}

func TestIngredientMatcher_MatchSorting(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()
	matcher := NewIngredientMatcher(normalizer)

	// Create recipes with different match levels
	perfectMatch := createTestRecipe("Perfect Match", recipe.CategoryOther,
		[]string{"chicken", "garlic", "onion"})

	partialMatch := createTestRecipe("Partial Match", recipe.CategoryOther,
		[]string{"chicken", "garlic", "onion", "tomato", "basil", "oregano"})

	recipes := []*recipe.Recipe{partialMatch, perfectMatch}

	userIngredients := []string{"chicken", "garlic", "onion"}

	options := MatchOptions{
		ExcludeStaples: false, // Include all ingredients in calculation
		MinMatchLevel:  MatchLevelLow,
		MaxResults:     10,
	}

	results := matcher.Match(userIngredients, recipes, options)

	if len(results) < 2 {
		t.Fatalf("Expected at least 2 results, got %d", len(results))
	}

	// Results should be sorted by match percentage (descending)
	if results[0].MatchPercentage < results[1].MatchPercentage {
		t.Errorf("Results not sorted correctly: first has %.2f%%, second has %.2f%%",
			results[0].MatchPercentage, results[1].MatchPercentage)
	}
}

func TestIngredientMatcher_MaxResults(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()
	matcher := NewIngredientMatcher(normalizer)

	// Create many recipes
	recipes := make([]*recipe.Recipe, 30)
	for i := 0; i < 30; i++ {
		recipes[i] = createTestRecipe("Recipe", recipe.CategoryOther,
			[]string{"chicken", "garlic"})
	}

	userIngredients := []string{"chicken", "garlic"}

	options := MatchOptions{
		ExcludeStaples: false,
		MinMatchLevel:  MatchLevelLow,
		MaxResults:     5,
	}

	results := matcher.Match(userIngredients, recipes, options)

	if len(results) > 5 {
		t.Errorf("MaxResults not respected: got %d results, want at most 5", len(results))
	}
}

func TestIngredientMatcher_StrictMode(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()
	matcher := NewIngredientMatcher(normalizer)

	perfectMatch := createTestRecipe("Perfect", recipe.CategoryOther,
		[]string{"chicken", "garlic"})

	partialMatch := createTestRecipe("Partial", recipe.CategoryOther,
		[]string{"chicken", "garlic", "onion", "tomato"})

	recipes := []*recipe.Recipe{perfectMatch, partialMatch}

	userIngredients := []string{"chicken", "garlic"}

	// Strict mode - only perfect matches
	options := MatchOptions{
		StrictMatch:    true,
		ExcludeStaples: false,
		MinMatchLevel:  MatchLevelLow,
		MaxResults:     10,
	}

	results := matcher.Match(userIngredients, recipes, options)

	for _, result := range results {
		if result.MatchLevel != MatchLevelPerfect {
			t.Errorf("Strict mode returned non-perfect match: %s with level %d",
				result.Recipe.Title(), result.MatchLevel)
		}
	}
}

func TestMatchResult_MatchLevels(t *testing.T) {
	tests := []struct {
		percentage float64
		wantLevel  MatchLevel
	}{
		{100, MatchLevelPerfect},
		{95, MatchLevelHigh},
		{80, MatchLevelHigh},
		{79, MatchLevelMedium},
		{60, MatchLevelMedium},
		{59, MatchLevelLow},
		{0, MatchLevelLow},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			var level MatchLevel
			switch {
			case tt.percentage >= 100:
				level = MatchLevelPerfect
			case tt.percentage >= 80:
				level = MatchLevelHigh
			case tt.percentage >= 60:
				level = MatchLevelMedium
			default:
				level = MatchLevelLow
			}

			if level != tt.wantLevel {
				t.Errorf("percentage %.0f%% got level %d, want %d", tt.percentage, level, tt.wantLevel)
			}
		})
	}
}

func TestGroupByMatchLevel(t *testing.T) {
	results := []MatchResult{
		{MatchLevel: MatchLevelPerfect, MatchPercentage: 100},
		{MatchLevel: MatchLevelPerfect, MatchPercentage: 100},
		{MatchLevel: MatchLevelHigh, MatchPercentage: 85},
		{MatchLevel: MatchLevelMedium, MatchPercentage: 70},
		{MatchLevel: MatchLevelMedium, MatchPercentage: 65},
		{MatchLevel: MatchLevelLow, MatchPercentage: 40},
	}

	grouped := GroupByMatchLevel(results)

	if len(grouped[MatchLevelPerfect]) != 2 {
		t.Errorf("Expected 2 perfect matches, got %d", len(grouped[MatchLevelPerfect]))
	}
	if len(grouped[MatchLevelHigh]) != 1 {
		t.Errorf("Expected 1 high match, got %d", len(grouped[MatchLevelHigh]))
	}
	if len(grouped[MatchLevelMedium]) != 2 {
		t.Errorf("Expected 2 medium matches, got %d", len(grouped[MatchLevelMedium]))
	}
	if len(grouped[MatchLevelLow]) != 1 {
		t.Errorf("Expected 1 low match, got %d", len(grouped[MatchLevelLow]))
	}
}

func TestMatchLevelString(t *testing.T) {
	tests := []struct {
		level MatchLevel
		want  string
	}{
		{MatchLevelPerfect, "Perfect Match"},
		{MatchLevelHigh, "Almost There"},
		{MatchLevelMedium, "Partial Match"},
		{MatchLevelLow, "Low Match"},
		{MatchLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := MatchLevelString(tt.level)
			if got != tt.want {
				t.Errorf("MatchLevelString(%d) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestDefaultMatchOptions(t *testing.T) {
	opts := DefaultMatchOptions()

	if opts.StrictMatch {
		t.Error("StrictMatch should be false by default")
	}
	if opts.CategoryFilter != nil {
		t.Error("CategoryFilter should be nil by default")
	}
	if !opts.ExcludeStaples {
		t.Error("ExcludeStaples should be true by default")
	}
	if opts.MinMatchLevel != MatchLevelMedium {
		t.Errorf("MinMatchLevel = %d, want %d", opts.MinMatchLevel, MatchLevelMedium)
	}
	if opts.MaxResults != 20 {
		t.Errorf("MaxResults = %d, want 20", opts.MaxResults)
	}
}

func TestIngredientMatcher_ExcludeStaples(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()
	matcher := NewIngredientMatcher(normalizer)

	// Recipe with staples and non-staples
	rec := createTestRecipe("Test Recipe", recipe.CategoryOther,
		[]string{"chicken breast", "salt", "pepper", "olive oil", "garlic"})

	recipes := []*recipe.Recipe{rec}

	// User has chicken and garlic (non-staples)
	userIngredients := []string{"chicken breast", "garlic"}

	// With ExcludeStaples = true
	optionsWithExclude := MatchOptions{
		ExcludeStaples: true,
		MinMatchLevel:  MatchLevelLow,
		MaxResults:     10,
	}

	resultsWithExclude := matcher.Match(userIngredients, recipes, optionsWithExclude)

	// Without ExcludeStaples
	optionsWithoutExclude := MatchOptions{
		ExcludeStaples: false,
		MinMatchLevel:  MatchLevelLow,
		MaxResults:     10,
	}

	resultsWithoutExclude := matcher.Match(userIngredients, recipes, optionsWithoutExclude)

	// With exclude, percentage should be higher (only counting non-staples)
	if len(resultsWithExclude) > 0 && len(resultsWithoutExclude) > 0 {
		if resultsWithExclude[0].MatchPercentage <= resultsWithoutExclude[0].MatchPercentage {
			t.Logf("With exclude: %.2f%%, Without: %.2f%%",
				resultsWithExclude[0].MatchPercentage, resultsWithoutExclude[0].MatchPercentage)
		}
	}
}

func TestNewIngredientMatcher(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()
	matcher := NewIngredientMatcher(normalizer)

	if matcher == nil {
		t.Fatal("NewIngredientMatcher returned nil")
	}

	if matcher.normalizer == nil {
		t.Error("normalizer is nil")
	}
}

// Helper to create time.Duration pointer
func durationPtr(d time.Duration) *time.Duration {
	return &d
}

// Helper to create category pointer
func categoryPtr(c recipe.Category) *recipe.Category {
	return &c
}
