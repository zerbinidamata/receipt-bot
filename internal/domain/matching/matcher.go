package matching

import (
	"sort"

	"receipt-bot/internal/domain/recipe"
)

// MatchLevel represents the quality of a recipe match
type MatchLevel int

const (
	MatchLevelPerfect MatchLevel = iota // 100% match
	MatchLevelHigh                      // 80%+ match
	MatchLevelMedium                    // 60-80% match
	MatchLevelLow                       // <60% (typically not shown)
)

// MatchResult represents a recipe match with details
type MatchResult struct {
	Recipe          *recipe.Recipe
	MatchPercentage float64
	MatchedItems    []string
	MissingItems    []string
	MatchLevel      MatchLevel
}

// MatchOptions configures the matching behavior
type MatchOptions struct {
	StrictMatch      bool             // Only return perfect matches
	CategoryFilter   *recipe.Category // Filter by category
	ExcludeStaples   bool             // Exclude common pantry staples from calculation
	MinMatchLevel    MatchLevel       // Minimum match level to include
	MaxResults       int              // Maximum number of results (0 = unlimited)
}

// DefaultMatchOptions returns sensible defaults
func DefaultMatchOptions() MatchOptions {
	return MatchOptions{
		StrictMatch:    false,
		CategoryFilter: nil,
		ExcludeStaples: true,
		MinMatchLevel:  MatchLevelMedium,
		MaxResults:     20,
	}
}

// IngredientMatcher matches user ingredients against recipes
type IngredientMatcher struct {
	normalizer IngredientNormalizer
}

// NewIngredientMatcher creates a new matcher
func NewIngredientMatcher(normalizer IngredientNormalizer) *IngredientMatcher {
	return &IngredientMatcher{
		normalizer: normalizer,
	}
}

// Match finds recipes that match the given ingredients
func (m *IngredientMatcher) Match(
	userIngredients []string,
	recipes []*recipe.Recipe,
	options MatchOptions,
) []MatchResult {
	// Normalize user ingredients
	normalizedUser := make(map[string]bool)
	for _, ing := range userIngredients {
		normalized := m.normalizer.Normalize(ing)
		if normalized != "" {
			normalizedUser[normalized] = true
		}
	}

	if len(normalizedUser) == 0 {
		return nil
	}

	var results []MatchResult

	for _, rec := range recipes {
		// Apply category filter if specified
		if options.CategoryFilter != nil && rec.Category() != *options.CategoryFilter {
			continue
		}

		result := m.matchRecipe(rec, normalizedUser, options.ExcludeStaples)

		// Apply minimum match level filter
		if result.MatchLevel > options.MinMatchLevel {
			continue
		}

		// For strict mode, only include perfect matches
		if options.StrictMatch && result.MatchLevel != MatchLevelPerfect {
			continue
		}

		results = append(results, result)
	}

	// Sort by match percentage (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchPercentage > results[j].MatchPercentage
	})

	// Apply max results limit
	if options.MaxResults > 0 && len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results
}

// matchRecipe calculates the match score for a single recipe
func (m *IngredientMatcher) matchRecipe(
	rec *recipe.Recipe,
	normalizedUser map[string]bool,
	excludeStaples bool,
) MatchResult {
	result := MatchResult{
		Recipe:       rec,
		MatchedItems: make([]string, 0),
		MissingItems: make([]string, 0),
	}

	recipeIngredients := rec.Ingredients()
	totalRequired := 0

	for _, ing := range recipeIngredients {
		normalized := m.normalizer.Normalize(ing.Name())

		// Skip pantry staples if configured
		if excludeStaples && IsPantryStaple(normalized) {
			continue
		}

		totalRequired++

		if m.hasIngredient(normalized, normalizedUser) {
			result.MatchedItems = append(result.MatchedItems, ing.Name())
		} else {
			result.MissingItems = append(result.MissingItems, ing.Name())
		}
	}

	// Calculate match percentage
	if totalRequired > 0 {
		result.MatchPercentage = float64(len(result.MatchedItems)) / float64(totalRequired) * 100
	} else {
		// If all ingredients are staples, consider it a perfect match
		result.MatchPercentage = 100
	}

	// Determine match level
	switch {
	case result.MatchPercentage >= 100:
		result.MatchLevel = MatchLevelPerfect
	case result.MatchPercentage >= 80:
		result.MatchLevel = MatchLevelHigh
	case result.MatchPercentage >= 60:
		result.MatchLevel = MatchLevelMedium
	default:
		result.MatchLevel = MatchLevelLow
	}

	return result
}

// hasIngredient checks if the user has an ingredient (exact or similar)
func (m *IngredientMatcher) hasIngredient(recipeIng string, userIngredients map[string]bool) bool {
	// Direct match
	if userIngredients[recipeIng] {
		return true
	}

	// Check for similar ingredients
	for userIng := range userIngredients {
		if m.normalizer.AreSimilar(recipeIng, userIng) {
			return true
		}
	}

	return false
}

// GroupByMatchLevel groups results by their match level
func GroupByMatchLevel(results []MatchResult) map[MatchLevel][]MatchResult {
	grouped := make(map[MatchLevel][]MatchResult)

	for _, result := range results {
		grouped[result.MatchLevel] = append(grouped[result.MatchLevel], result)
	}

	return grouped
}

// MatchLevelString returns a human-readable match level description
func MatchLevelString(level MatchLevel) string {
	switch level {
	case MatchLevelPerfect:
		return "Perfect Match"
	case MatchLevelHigh:
		return "Almost There"
	case MatchLevelMedium:
		return "Partial Match"
	case MatchLevelLow:
		return "Low Match"
	default:
		return "Unknown"
	}
}
