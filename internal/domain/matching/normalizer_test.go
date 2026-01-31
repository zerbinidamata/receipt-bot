package matching

import (
	"testing"
)

func TestRuleBasedNormalizer_Normalize(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Basic normalization
		{"simple ingredient", "flour", "flour"},
		{"ingredient with quantity", "2 cups flour", "flour"},
		{"ingredient with fraction", "1/2 cup sugar", "sugar"},
		{"ingredient with unicode fraction", "½ cup milk", "milk"},

		// Unit removal
		{"with tablespoon", "2 tbsp olive oil", "olive oil"},
		{"with teaspoon", "1 tsp salt", "salt"},
		{"with grams", "500g chicken breast", "chicken breast"},
		{"with ounces", "8 oz cream cheese", "cream cheese"},
		{"with pounds", "2 lbs ground beef", "beef"},
		{"with ml", "250ml water", "water"},
		{"with cloves", "3 cloves garlic", "garlic"},

		// Preparation word removal
		{"chopped", "chopped onion", "onion"},
		{"minced", "minced garlic", "garlic"},
		{"diced", "diced tomatoes", "tomato"},
		{"sliced", "thinly sliced carrots", "thinly carrot"},
		{"fresh", "fresh basil", "basil"},
		{"dried", "dried oregano", "oregano"},
		{"grated", "grated parmesan", "parmesan"},

		// Compound preparations
		{"freshly chopped", "freshly chopped parsley", "parsley"},
		{"finely minced", "finely minced ginger", "ginger"},

		// Parenthetical notes
		{"with parentheses", "butter (room temperature)", "butter"},
		{"with optional note", "cilantro (optional)", "cilantro"},

		// Trailing punctuation
		{"with comma", "salt, to taste", "salt"},
		{"with semicolon", "pepper; freshly ground", "pepper"},

		// Pluralization
		{"plural tomatoes", "tomatoes", "tomato"},
		{"plural berries", "berries", "berry"},
		{"plural leaves", "leaves", "leaf"},
		{"plural carrots", "carrots", "carrot"},

		// Edge cases
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
		{"with extra spaces", "  flour  ", "flour"},
		{"mixed case", "OLIVE OIL", "olive oil"},

		// Complex examples
		{"full ingredient line", "2 cups all-purpose flour, sifted", "all-purpose flour"},
		{"ingredient with notes", "1 lb boneless skinless chicken breast, cut into cubes", "boneless skinless chicken breast"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRuleBasedNormalizer_AreSimilar(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()

	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		// Exact matches after normalization
		{"exact match", "flour", "flour", true},
		{"case insensitive", "Flour", "flour", true},
		{"with quantity", "2 cups flour", "flour", true},

		// Substitution groups
		{"milk and cream", "milk", "cream", true},
		{"milk and heavy cream", "milk", "heavy cream", true},
		{"butter and margarine", "butter", "margarine", true},
		{"unsalted butter and butter", "unsalted butter", "butter", true},
		{"parmesan and romano", "parmesan", "romano", true},
		{"parmesan and pecorino", "parmesan", "pecorino", true},
		{"spaghetti and linguine", "spaghetti", "linguine", true},
		{"pasta and penne", "pasta", "penne", true},
		{"olive oil and vegetable oil", "olive oil", "vegetable oil", true},
		{"garlic and garlic cloves", "garlic", "garlic cloves", true},
		{"onion and shallot", "onion", "shallot", true},
		{"tomato and cherry tomatoes", "tomato", "cherry tomatoes", true},
		{"ground beef and ground turkey", "ground beef", "ground turkey", false}, // "ground" is removed by normalizer, so they become "beef" and "turkey"
		{"chicken broth and vegetable stock", "chicken broth", "vegetable stock", true},
		{"sour cream and greek yogurt", "sour cream", "greek yogurt", true},
		{"breadcrumbs and panko", "breadcrumbs", "panko breadcrumbs", true},

		// Contains match
		{"chicken breast contains chicken", "chicken breast", "chicken", true},
		{"contains both ways", "chicken", "chicken thigh", true},

		// Not similar
		{"different ingredients", "flour", "sugar", false},
		{"unrelated", "chicken", "fish", false},
		{"not in same group", "milk", "olive oil", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.AreSimilar(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("AreSimilar(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsPantryStaple(t *testing.T) {
	tests := []struct {
		name       string
		ingredient string
		want       bool
	}{
		// Staples
		{"salt", "salt", true},
		{"pepper", "pepper", true},
		{"black pepper", "black pepper", true},
		{"olive oil", "olive oil", true},
		{"vegetable oil", "vegetable oil", true},
		{"water", "water", true},
		{"sugar", "sugar", true},
		{"flour", "flour", true},
		{"all-purpose flour", "all-purpose flour", true},
		{"butter", "butter", true},
		{"garlic powder", "garlic powder", true},
		{"baking soda", "baking soda", true},
		{"baking powder", "baking powder", true},

		// With quantities/preparation
		{"salt with quantity", "1 tsp salt", true},
		{"pepper ground", "freshly ground pepper", true},

		// Not staples
		{"chicken", "chicken", false},
		{"tomatoes", "tomatoes", false},
		{"cheese", "cheese", false},
		{"pasta", "pasta", false},
		{"onion", "onion", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPantryStaple(tt.ingredient)
			if got != tt.want {
				t.Errorf("IsPantryStaple(%q) = %v, want %v", tt.ingredient, got, tt.want)
			}
		})
	}
}

func TestNewRuleBasedNormalizer(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()

	if normalizer == nil {
		t.Fatal("NewRuleBasedNormalizer() returned nil")
	}

	if normalizer.substitutionIndex == nil {
		t.Error("substitutionIndex is nil")
	}

	// Check that substitution index is populated
	if len(normalizer.substitutionIndex) == 0 {
		t.Error("substitutionIndex is empty")
	}

	// Verify specific entries
	if _, ok := normalizer.substitutionIndex["milk"]; !ok {
		t.Error("milk not found in substitution index")
	}

	if _, ok := normalizer.substitutionIndex["parmesan"]; !ok {
		t.Error("parmesan not found in substitution index")
	}
}

func TestDepluralize(t *testing.T) {
	normalizer := NewRuleBasedNormalizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"tomatoes", "tomatoes", "tomato"},
		{"berries", "berries", "berry"},
		{"leaves", "leaves", "leaf"},
		{"dishes", "dishes", "dish"},
		{"carrots", "carrots", "carrot"},
		{"short word", "as", "as"},           // Don't depluralize very short words
		{"already singular", "chicken", "chicken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.depluralize(tt.input)
			if got != tt.want {
				t.Errorf("depluralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
