package recipe

import (
	"reflect"
	"testing"
)

func TestDietaryTag_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tag  DietaryTag
		want bool
	}{
		{"valid vegetarian", TagVegetarian, true},
		{"valid vegan", TagVegan, true},
		{"valid gluten-free", TagGlutenFree, true},
		{"valid dairy-free", TagDairyFree, true},
		{"valid low-carb", TagLowCarb, true},
		{"valid quick", TagQuick, true},
		{"valid one-pot", TagOnePot, true},
		{"valid kid-friendly", TagKidFriendly, true},
		{"invalid tag", DietaryTag("invalid"), false},
		{"empty tag", DietaryTag(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.IsValid(); got != tt.want {
				t.Errorf("DietaryTag.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDietaryTag_String(t *testing.T) {
	tests := []struct {
		name string
		tag  DietaryTag
		want string
	}{
		{"vegetarian", TagVegetarian, "vegetarian"},
		{"vegan", TagVegan, "vegan"},
		{"gluten-free", TagGlutenFree, "gluten-free"},
		{"dairy-free", TagDairyFree, "dairy-free"},
		{"low-carb", TagLowCarb, "low-carb"},
		{"quick", TagQuick, "quick"},
		{"one-pot", TagOnePot, "one-pot"},
		{"kid-friendly", TagKidFriendly, "kid-friendly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.String(); got != tt.want {
				t.Errorf("DietaryTag.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllDietaryTags(t *testing.T) {
	tags := AllDietaryTags()

	// Check we have all 8 tags
	if len(tags) != 8 {
		t.Errorf("AllDietaryTags() returned %d tags, want 8", len(tags))
	}

	// Check all returned tags are valid
	for _, tag := range tags {
		if !tag.IsValid() {
			t.Errorf("AllDietaryTags() returned invalid tag: %v", tag)
		}
	}

	// Check for expected tags
	expected := map[DietaryTag]bool{
		TagVegetarian:  true,
		TagVegan:       true,
		TagGlutenFree:  true,
		TagDairyFree:   true,
		TagLowCarb:     true,
		TagQuick:       true,
		TagOnePot:      true,
		TagKidFriendly: true,
	}

	for _, tag := range tags {
		if !expected[tag] {
			t.Errorf("Unexpected tag in AllDietaryTags(): %v", tag)
		}
	}
}

func TestParseDietaryTag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTag   DietaryTag
		wantValid bool
	}{
		// Vegetarian aliases
		{"vegetarian", "vegetarian", TagVegetarian, true},
		{"veg", "veg", TagVegetarian, true},
		{"veggie", "veggie", TagVegetarian, true},

		// Vegan aliases
		{"vegan", "vegan", TagVegan, true},
		{"plant-based", "plant-based", TagVegan, true},
		{"plant based", "plant based", TagVegan, true},

		// Gluten-free aliases
		{"gluten-free", "gluten-free", TagGlutenFree, true},
		{"gluten free", "gluten free", TagGlutenFree, true},
		{"gf", "gf", TagGlutenFree, true},
		{"no gluten", "no gluten", TagGlutenFree, true},

		// Dairy-free aliases
		{"dairy-free", "dairy-free", TagDairyFree, true},
		{"dairy free", "dairy free", TagDairyFree, true},
		{"df", "df", TagDairyFree, true},
		{"no dairy", "no dairy", TagDairyFree, true},
		{"lactose-free", "lactose-free", TagDairyFree, true},

		// Low-carb aliases
		{"low-carb", "low-carb", TagLowCarb, true},
		{"low carb", "low carb", TagLowCarb, true},
		{"keto", "keto", TagLowCarb, true},
		{"lc", "lc", TagLowCarb, true},

		// Quick aliases
		{"quick", "quick", TagQuick, true},
		{"fast", "fast", TagQuick, true},
		{"30-min", "30-min", TagQuick, true},
		{"30 min", "30 min", TagQuick, true},
		{"under 30", "under 30", TagQuick, true},

		// One-pot aliases
		{"one-pot", "one-pot", TagOnePot, true},
		{"one pot", "one pot", TagOnePot, true},
		{"onepot", "onepot", TagOnePot, true},
		{"single pot", "single pot", TagOnePot, true},

		// Kid-friendly aliases
		{"kid-friendly", "kid-friendly", TagKidFriendly, true},
		{"kid friendly", "kid friendly", TagKidFriendly, true},
		{"kids", "kids", TagKidFriendly, true},
		{"family", "family", TagKidFriendly, true},

		// Invalid tags
		{"unknown tag", "unknown", DietaryTag("unknown"), false},
		{"random text", "xyz123", DietaryTag("xyz123"), false},

		// Case insensitivity
		{"VEGETARIAN uppercase", "VEGETARIAN", TagVegetarian, true},
		{"VeGaN mixed case", "VeGaN", TagVegan, true},

		// Whitespace handling
		{"leading space", " vegan", TagVegan, true},
		{"trailing space", "vegan ", TagVegan, true},
		{"both spaces", " vegan ", TagVegan, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTag, gotValid := ParseDietaryTag(tt.input)
			if gotTag != tt.wantTag {
				t.Errorf("ParseDietaryTag(%q) tag = %v, want %v", tt.input, gotTag, tt.wantTag)
			}
			if gotValid != tt.wantValid {
				t.Errorf("ParseDietaryTag(%q) valid = %v, want %v", tt.input, gotValid, tt.wantValid)
			}
		})
	}
}

func TestParseDietaryTags(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []DietaryTag
	}{
		{
			name:  "single valid tag",
			input: []string{"vegan"},
			want:  []DietaryTag{TagVegan},
		},
		{
			name:  "multiple valid tags",
			input: []string{"vegan", "gluten-free", "quick"},
			want:  []DietaryTag{TagVegan, TagGlutenFree, TagQuick},
		},
		{
			name:  "filters invalid tags",
			input: []string{"vegan", "invalid", "quick"},
			want:  []DietaryTag{TagVegan, TagQuick},
		},
		{
			name:  "all invalid tags",
			input: []string{"invalid1", "invalid2"},
			want:  []DietaryTag{},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  []DietaryTag{},
		},
		{
			name:  "removes duplicates",
			input: []string{"vegan", "vegan", "plant-based"},
			want:  []DietaryTag{TagVegan},
		},
		{
			name:  "removes duplicates with aliases",
			input: []string{"vegetarian", "veggie", "veg"},
			want:  []DietaryTag{TagVegetarian},
		},
		{
			name:  "mixed case input",
			input: []string{"VEGAN", "Gluten-Free"},
			want:  []DietaryTag{TagVegan, TagGlutenFree},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDietaryTags(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDietaryTags(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDietaryTagsToStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []DietaryTag
		want  []string
	}{
		{
			name:  "single tag",
			input: []DietaryTag{TagVegan},
			want:  []string{"vegan"},
		},
		{
			name:  "multiple tags",
			input: []DietaryTag{TagVegan, TagGlutenFree, TagQuick},
			want:  []string{"vegan", "gluten-free", "quick"},
		},
		{
			name:  "empty input",
			input: []DietaryTag{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DietaryTagsToStrings(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DietaryTagsToStrings(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStringsToDietaryTags(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []DietaryTag
	}{
		{
			name:  "valid strings",
			input: []string{"vegan", "quick"},
			want:  []DietaryTag{TagVegan, TagQuick},
		},
		{
			name:  "mixed valid and invalid",
			input: []string{"vegan", "invalid", "quick"},
			want:  []DietaryTag{TagVegan, TagQuick},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  []DietaryTag{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringsToDietaryTags(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringsToDietaryTags(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
