package recipe

import "strings"

// Category represents a recipe category type
type Category string

const (
	CategoryPasta       Category = "Pasta & Noodles"
	CategoryRice        Category = "Rice & Grains"
	CategorySoups       Category = "Soups & Stews"
	CategorySalads      Category = "Salads"
	CategoryMeat        Category = "Meat & Poultry"
	CategorySeafood     Category = "Seafood"
	CategoryVegetarian  Category = "Vegetarian"
	CategoryDesserts    Category = "Desserts & Sweets"
	CategoryBreakfast   Category = "Breakfast"
	CategoryAppetizers  Category = "Appetizers & Snacks"
	CategoryBeverages   Category = "Beverages"
	CategorySauces      Category = "Sauces & Condiments"
	CategoryBread       Category = "Bread & Baking"
	CategoryOther       Category = "Other"
)

// AllCategories returns all valid categories
func AllCategories() []Category {
	return []Category{
		CategoryPasta,
		CategoryRice,
		CategorySoups,
		CategorySalads,
		CategoryMeat,
		CategorySeafood,
		CategoryVegetarian,
		CategoryDesserts,
		CategoryBreakfast,
		CategoryAppetizers,
		CategoryBeverages,
		CategorySauces,
		CategoryBread,
		CategoryOther,
	}
}

// IsValid checks if the category is a valid category type
func (c Category) IsValid() bool {
	switch c {
	case CategoryPasta, CategoryRice, CategorySoups, CategorySalads,
		CategoryMeat, CategorySeafood, CategoryVegetarian, CategoryDesserts,
		CategoryBreakfast, CategoryAppetizers, CategoryBeverages, CategorySauces,
		CategoryBread, CategoryOther:
		return true
	default:
		return false
	}
}

// String returns the string representation of the category
func (c Category) String() string {
	return string(c)
}

// ParseCategory parses a string into a Category
// It handles shorthand aliases and case-insensitive matching
func ParseCategory(s string) Category {
	s = strings.ToLower(strings.TrimSpace(s))

	// Handle shorthand aliases
	switch s {
	case "pasta", "noodles", "pasta & noodles":
		return CategoryPasta
	case "rice", "grains", "rice & grains":
		return CategoryRice
	case "soup", "soups", "stew", "stews", "soups & stews":
		return CategorySoups
	case "salad", "salads":
		return CategorySalads
	case "meat", "poultry", "meat & poultry":
		return CategoryMeat
	case "seafood", "fish":
		return CategorySeafood
	case "vegetarian", "veggie", "veg":
		return CategoryVegetarian
	case "dessert", "desserts", "sweet", "sweets", "desserts & sweets":
		return CategoryDesserts
	case "breakfast", "brunch":
		return CategoryBreakfast
	case "appetizer", "appetizers", "snack", "snacks", "appetizers & snacks":
		return CategoryAppetizers
	case "beverage", "beverages", "drink", "drinks":
		return CategoryBeverages
	case "sauce", "sauces", "condiment", "condiments", "sauces & condiments":
		return CategorySauces
	case "bread", "baking", "bread & baking":
		return CategoryBread
	case "other":
		return CategoryOther
	default:
		return CategoryOther
	}
}

// CategoryFromLLM parses an LLM response category string into a Category
// This is more strict and expects the exact category name
func CategoryFromLLM(s string) Category {
	s = strings.TrimSpace(s)

	// Try exact match first
	for _, cat := range AllCategories() {
		if strings.EqualFold(string(cat), s) {
			return cat
		}
	}

	// Fall back to alias parsing
	return ParseCategory(s)
}
