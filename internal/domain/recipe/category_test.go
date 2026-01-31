package recipe

import (
	"testing"
)

func TestCategory_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		want     bool
	}{
		{"valid Pasta", CategoryPasta, true},
		{"valid Rice", CategoryRice, true},
		{"valid Soups", CategorySoups, true},
		{"valid Salads", CategorySalads, true},
		{"valid Meat", CategoryMeat, true},
		{"valid Seafood", CategorySeafood, true},
		{"valid Vegetarian", CategoryVegetarian, true},
		{"valid Desserts", CategoryDesserts, true},
		{"valid Breakfast", CategoryBreakfast, true},
		{"valid Appetizers", CategoryAppetizers, true},
		{"valid Beverages", CategoryBeverages, true},
		{"valid Sauces", CategorySauces, true},
		{"valid Bread", CategoryBread, true},
		{"valid Other", CategoryOther, true},
		{"invalid category", Category("Invalid"), false},
		{"empty category", Category(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.IsValid(); got != tt.want {
				t.Errorf("Category.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		want     string
	}{
		{"Pasta", CategoryPasta, "Pasta & Noodles"},
		{"Rice", CategoryRice, "Rice & Grains"},
		{"Soups", CategorySoups, "Soups & Stews"},
		{"Salads", CategorySalads, "Salads"},
		{"Meat", CategoryMeat, "Meat & Poultry"},
		{"Seafood", CategorySeafood, "Seafood"},
		{"Vegetarian", CategoryVegetarian, "Vegetarian"},
		{"Desserts", CategoryDesserts, "Desserts & Sweets"},
		{"Breakfast", CategoryBreakfast, "Breakfast"},
		{"Appetizers", CategoryAppetizers, "Appetizers & Snacks"},
		{"Beverages", CategoryBeverages, "Beverages"},
		{"Sauces", CategorySauces, "Sauces & Condiments"},
		{"Bread", CategoryBread, "Bread & Baking"},
		{"Other", CategoryOther, "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.String(); got != tt.want {
				t.Errorf("Category.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllCategories(t *testing.T) {
	categories := AllCategories()

	// Check we have all 14 categories
	if len(categories) != 14 {
		t.Errorf("AllCategories() returned %d categories, want 14", len(categories))
	}

	// Check all returned categories are valid
	for _, cat := range categories {
		if !cat.IsValid() {
			t.Errorf("AllCategories() returned invalid category: %v", cat)
		}
	}

	// Check for expected categories
	expected := map[Category]bool{
		CategoryPasta:       true,
		CategoryRice:        true,
		CategorySoups:       true,
		CategorySalads:      true,
		CategoryMeat:        true,
		CategorySeafood:     true,
		CategoryVegetarian:  true,
		CategoryDesserts:    true,
		CategoryBreakfast:   true,
		CategoryAppetizers:  true,
		CategoryBeverages:   true,
		CategorySauces:      true,
		CategoryBread:       true,
		CategoryOther:       true,
	}

	for _, cat := range categories {
		if !expected[cat] {
			t.Errorf("Unexpected category in AllCategories(): %v", cat)
		}
	}
}

func TestParseCategory(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Category
	}{
		// Pasta aliases
		{"pasta lowercase", "pasta", CategoryPasta},
		{"noodles", "noodles", CategoryPasta},
		{"pasta & noodles", "pasta & noodles", CategoryPasta},
		{"PASTA uppercase", "PASTA", CategoryPasta},

		// Rice aliases
		{"rice", "rice", CategoryRice},
		{"grains", "grains", CategoryRice},
		{"rice & grains", "rice & grains", CategoryRice},

		// Soups aliases
		{"soup singular", "soup", CategorySoups},
		{"soups plural", "soups", CategorySoups},
		{"stew", "stew", CategorySoups},
		{"stews", "stews", CategorySoups},

		// Salads aliases
		{"salad singular", "salad", CategorySalads},
		{"salads plural", "salads", CategorySalads},

		// Meat aliases
		{"meat", "meat", CategoryMeat},
		{"poultry", "poultry", CategoryMeat},
		{"meat & poultry", "meat & poultry", CategoryMeat},

		// Seafood aliases
		{"seafood", "seafood", CategorySeafood},
		{"fish", "fish", CategorySeafood},

		// Vegetarian aliases
		{"vegetarian", "vegetarian", CategoryVegetarian},
		{"veggie", "veggie", CategoryVegetarian},
		{"veg", "veg", CategoryVegetarian},

		// Desserts aliases
		{"dessert singular", "dessert", CategoryDesserts},
		{"desserts plural", "desserts", CategoryDesserts},
		{"sweet", "sweet", CategoryDesserts},
		{"sweets", "sweets", CategoryDesserts},

		// Breakfast aliases
		{"breakfast", "breakfast", CategoryBreakfast},
		{"brunch", "brunch", CategoryBreakfast},

		// Appetizers aliases
		{"appetizer singular", "appetizer", CategoryAppetizers},
		{"appetizers plural", "appetizers", CategoryAppetizers},
		{"snack", "snack", CategoryAppetizers},
		{"snacks", "snacks", CategoryAppetizers},

		// Beverages aliases
		{"beverage singular", "beverage", CategoryBeverages},
		{"beverages plural", "beverages", CategoryBeverages},
		{"drink", "drink", CategoryBeverages},
		{"drinks", "drinks", CategoryBeverages},

		// Sauces aliases
		{"sauce singular", "sauce", CategorySauces},
		{"sauces plural", "sauces", CategorySauces},
		{"condiment", "condiment", CategorySauces},
		{"condiments", "condiments", CategorySauces},

		// Bread aliases
		{"bread", "bread", CategoryBread},
		{"baking", "baking", CategoryBread},
		{"bread & baking", "bread & baking", CategoryBread},

		// Other
		{"other", "other", CategoryOther},

		// Unknown defaults to Other
		{"unknown category", "unknown", CategoryOther},
		{"random text", "xyz123", CategoryOther},

		// Whitespace handling
		{"with leading space", " pasta", CategoryPasta},
		{"with trailing space", "pasta ", CategoryPasta},
		{"with both spaces", " pasta ", CategoryPasta},

		// Case insensitivity
		{"mixed case", "PaStA", CategoryPasta},
		{"all caps", "SEAFOOD", CategorySeafood},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCategory(tt.input); got != tt.want {
				t.Errorf("ParseCategory(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCategoryFromLLM(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Category
	}{
		// Exact matches (LLM output)
		{"exact Pasta & Noodles", "Pasta & Noodles", CategoryPasta},
		{"exact Rice & Grains", "Rice & Grains", CategoryRice},
		{"exact Soups & Stews", "Soups & Stews", CategorySoups},
		{"exact Salads", "Salads", CategorySalads},
		{"exact Meat & Poultry", "Meat & Poultry", CategoryMeat},
		{"exact Seafood", "Seafood", CategorySeafood},
		{"exact Vegetarian", "Vegetarian", CategoryVegetarian},
		{"exact Desserts & Sweets", "Desserts & Sweets", CategoryDesserts},
		{"exact Breakfast", "Breakfast", CategoryBreakfast},
		{"exact Appetizers & Snacks", "Appetizers & Snacks", CategoryAppetizers},
		{"exact Beverages", "Beverages", CategoryBeverages},
		{"exact Sauces & Condiments", "Sauces & Condiments", CategorySauces},
		{"exact Bread & Baking", "Bread & Baking", CategoryBread},
		{"exact Other", "Other", CategoryOther},

		// Case insensitive exact match
		{"lowercase seafood", "seafood", CategorySeafood},
		{"uppercase BREAKFAST", "BREAKFAST", CategoryBreakfast},

		// Falls back to alias parsing
		{"alias pasta", "pasta", CategoryPasta},
		{"alias fish", "fish", CategorySeafood},
		{"alias veggie", "veggie", CategoryVegetarian},

		// Unknown defaults to Other
		{"unknown", "SomethingRandom", CategoryOther},

		// Whitespace handling
		{"with spaces", "  Pasta & Noodles  ", CategoryPasta},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CategoryFromLLM(tt.input); got != tt.want {
				t.Errorf("CategoryFromLLM(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
