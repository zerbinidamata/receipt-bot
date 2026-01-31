package matching

import (
	"regexp"
	"strings"
)

// IngredientNormalizer normalizes ingredient names for matching
type IngredientNormalizer interface {
	// Normalize extracts the base ingredient from a full ingredient string
	// "2 cups all-purpose flour" -> "flour"
	// "fresh basil leaves, chopped" -> "basil"
	Normalize(raw string) string

	// AreSimilar checks if two ingredients can substitute each other
	// "butter" and "margarine" -> true
	// "parmesan" and "romano" -> true
	AreSimilar(a, b string) bool
}

// RuleBasedNormalizer implements IngredientNormalizer using rule-based logic
type RuleBasedNormalizer struct {
	substitutionIndex map[string]int // maps ingredient to group index
}

// NewRuleBasedNormalizer creates a new rule-based normalizer
func NewRuleBasedNormalizer() *RuleBasedNormalizer {
	n := &RuleBasedNormalizer{
		substitutionIndex: make(map[string]int),
	}

	// Build substitution index from groups
	for i, group := range substitutionGroups {
		for _, item := range group {
			n.substitutionIndex[item] = i
		}
	}

	return n
}

// substitutionGroups defines ingredients that can substitute each other
var substitutionGroups = [][]string{
	{"milk", "cream", "half-and-half", "heavy cream", "whole milk", "skim milk"},
	{"parmesan", "romano", "pecorino", "grana padano", "asiago"},
	{"butter", "margarine", "unsalted butter", "salted butter"},
	{"chicken breast", "chicken thigh", "chicken", "chicken leg"},
	{"sugar", "honey", "maple syrup", "agave", "brown sugar", "white sugar"},
	{"spaghetti", "linguine", "fettuccine", "penne", "rigatoni", "pasta"},
	{"olive oil", "vegetable oil", "canola oil", "cooking oil", "oil"},
	{"garlic", "garlic cloves", "minced garlic", "garlic powder"},
	{"onion", "yellow onion", "white onion", "red onion", "shallot"},
	{"tomato", "tomatoes", "cherry tomatoes", "roma tomatoes", "diced tomatoes"},
	{"ground beef", "ground turkey", "ground pork", "ground meat"},
	{"cheddar", "monterey jack", "colby", "american cheese"},
	{"mozzarella", "provolone", "fontina"},
	{"basil", "fresh basil", "dried basil", "basil leaves"},
	{"oregano", "dried oregano", "fresh oregano"},
	{"thyme", "fresh thyme", "dried thyme"},
	{"lemon juice", "lime juice", "citrus juice"},
	{"chicken broth", "chicken stock", "vegetable broth", "vegetable stock", "broth", "stock"},
	{"sour cream", "greek yogurt", "plain yogurt"},
	{"bread crumbs", "breadcrumbs", "panko", "panko breadcrumbs"},
}

// quantityPattern matches common quantity patterns
var quantityPattern = regexp.MustCompile(`^[\d\s/½⅓¼⅔¾⅛⅜⅝⅞]+\s*`)

// unitPattern matches common units
var unitPattern = regexp.MustCompile(`(?i)^(cups?|tbsps?|tsps?|tablespoons?|teaspoons?|oz|ounces?|lbs?|pounds?|g|grams?|kg|kilograms?|ml|milliliters?|l|liters?|pinch(?:es)?|dash(?:es)?|cloves?|slices?|pieces?|cans?|packages?|bunche?s?|heads?|stalks?|sprigs?|handfuls?)\s+`)

// prepWordsPattern matches preparation words to remove
var prepWordsPattern = regexp.MustCompile(`(?i)\b(fresh|freshly|chopped|minced|diced|sliced|grated|shredded|crushed|ground|whole|large|medium|small|thin|thick|finely|coarsely|roughly|lightly|well|very|room temperature|cold|warm|hot|frozen|thawed|dried|canned|jarred|packed|loosely|firmly|about|approximately|optional|to taste|for garnish|for serving|divided|plus more|as needed|or more|or less)\b`)

// trailingPunctPattern removes trailing punctuation and parenthetical notes
var trailingPunctPattern = regexp.MustCompile(`[,;:]+.*$|\s*\([^)]*\)\s*$`)

// pluralSuffixes for simple depluralization
var pluralSuffixes = []struct {
	suffix string
	replace string
}{
	{"ies", "y"},      // berries -> berry
	{"ves", "f"},      // leaves -> leaf
	{"oes", "o"},      // tomatoes -> tomato
	{"es", ""},        // dishes -> dish
	{"s", ""},         // carrots -> carrot
}

// Normalize extracts the base ingredient name
func (n *RuleBasedNormalizer) Normalize(raw string) string {
	if raw == "" {
		return ""
	}

	// Convert to lowercase
	result := strings.ToLower(strings.TrimSpace(raw))

	// Remove quantities (numbers, fractions)
	result = quantityPattern.ReplaceAllString(result, "")

	// Remove units
	result = unitPattern.ReplaceAllString(result, "")

	// Remove trailing punctuation and parenthetical notes
	result = trailingPunctPattern.ReplaceAllString(result, "")

	// Remove preparation words
	result = prepWordsPattern.ReplaceAllString(result, " ")

	// Clean up extra whitespace
	result = strings.Join(strings.Fields(result), " ")

	// Simple depluralization
	result = n.depluralize(result)

	return strings.TrimSpace(result)
}

// depluralize attempts to convert plural to singular
func (n *RuleBasedNormalizer) depluralize(word string) string {
	// Don't depluralize very short words
	if len(word) <= 3 {
		return word
	}

	// Check each suffix pattern
	for _, p := range pluralSuffixes {
		if strings.HasSuffix(word, p.suffix) {
			candidate := strings.TrimSuffix(word, p.suffix) + p.replace
			// Don't return empty or very short results
			if len(candidate) >= 2 {
				return candidate
			}
		}
	}

	return word
}

// AreSimilar checks if two normalized ingredients can substitute each other
func (n *RuleBasedNormalizer) AreSimilar(a, b string) bool {
	// Normalize both inputs
	normA := n.Normalize(a)
	normB := n.Normalize(b)

	// Exact match after normalization
	if normA == normB {
		return true
	}

	// Check if one contains the other (for compound ingredients)
	if strings.Contains(normA, normB) || strings.Contains(normB, normA) {
		return true
	}

	// Check substitution groups
	groupA, okA := n.substitutionIndex[normA]
	groupB, okB := n.substitutionIndex[normB]

	if okA && okB && groupA == groupB {
		return true
	}

	return false
}

// CommonPantryStaples are ingredients typically found in most kitchens
// These are excluded from matching calculations
var CommonPantryStaples = map[string]bool{
	"salt":          true,
	"pepper":        true,
	"black pepper":  true,
	"white pepper":  true,
	"oil":           true,
	"olive oil":     true,
	"vegetable oil": true,
	"canola oil":    true,
	"cooking oil":   true,
	"water":         true,
	"sugar":         true,
	"flour":         true,
	"all-purpose flour": true,
	"butter":        true,
	"garlic powder": true,
	"onion powder":  true,
	"baking soda":   true,
	"baking powder": true,
}

// IsPantryStaple checks if an ingredient is a common pantry staple
func IsPantryStaple(ingredient string) bool {
	normalizer := NewRuleBasedNormalizer()
	normalized := normalizer.Normalize(ingredient)
	return CommonPantryStaples[normalized]
}
