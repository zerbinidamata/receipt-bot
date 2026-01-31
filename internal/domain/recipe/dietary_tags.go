package recipe

import "strings"

// DietaryTag represents a dietary classification tag
type DietaryTag string

const (
	TagVegetarian  DietaryTag = "vegetarian"
	TagVegan       DietaryTag = "vegan"
	TagGlutenFree  DietaryTag = "gluten-free"
	TagDairyFree   DietaryTag = "dairy-free"
	TagLowCarb     DietaryTag = "low-carb"
	TagQuick       DietaryTag = "quick"
	TagOnePot      DietaryTag = "one-pot"
	TagKidFriendly DietaryTag = "kid-friendly"
)

// AllDietaryTags returns all valid dietary tags
func AllDietaryTags() []DietaryTag {
	return []DietaryTag{
		TagVegetarian,
		TagVegan,
		TagGlutenFree,
		TagDairyFree,
		TagLowCarb,
		TagQuick,
		TagOnePot,
		TagKidFriendly,
	}
}

// IsValid checks if the dietary tag is valid
func (t DietaryTag) IsValid() bool {
	switch t {
	case TagVegetarian, TagVegan, TagGlutenFree, TagDairyFree,
		TagLowCarb, TagQuick, TagOnePot, TagKidFriendly:
		return true
	default:
		return false
	}
}

// String returns the string representation of the dietary tag
func (t DietaryTag) String() string {
	return string(t)
}

// ParseDietaryTag parses a string into a DietaryTag
// Returns the tag and a boolean indicating if it's valid
func ParseDietaryTag(s string) (DietaryTag, bool) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Handle aliases
	switch s {
	case "vegetarian", "veg", "veggie":
		return TagVegetarian, true
	case "vegan", "plant-based", "plant based":
		return TagVegan, true
	case "gluten-free", "gluten free", "gf", "no gluten":
		return TagGlutenFree, true
	case "dairy-free", "dairy free", "df", "no dairy", "lactose-free":
		return TagDairyFree, true
	case "low-carb", "low carb", "keto", "lc":
		return TagLowCarb, true
	case "quick", "fast", "30-min", "30 min", "under 30":
		return TagQuick, true
	case "one-pot", "one pot", "onepot", "single pot":
		return TagOnePot, true
	case "kid-friendly", "kid friendly", "kids", "family":
		return TagKidFriendly, true
	default:
		return DietaryTag(s), false
	}
}

// ParseDietaryTags parses a slice of strings into valid DietaryTags
// Invalid tags are filtered out
func ParseDietaryTags(tags []string) []DietaryTag {
	result := make([]DietaryTag, 0, len(tags))
	seen := make(map[DietaryTag]bool)

	for _, s := range tags {
		tag, valid := ParseDietaryTag(s)
		if valid && !seen[tag] {
			result = append(result, tag)
			seen[tag] = true
		}
	}

	return result
}

// DietaryTagsToStrings converts a slice of DietaryTags to strings
func DietaryTagsToStrings(tags []DietaryTag) []string {
	result := make([]string, len(tags))
	for i, tag := range tags {
		result[i] = string(tag)
	}
	return result
}

// StringsToDietaryTags converts a slice of strings to DietaryTags
// Only includes valid tags
func StringsToDietaryTags(strings []string) []DietaryTag {
	return ParseDietaryTags(strings)
}
