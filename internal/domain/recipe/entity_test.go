package recipe

import (
	"testing"

	"receipt-bot/internal/domain/shared"
)

func TestNewRecipe(t *testing.T) {
	userID := shared.NewID()
	ingredient, _ := NewIngredient("flour", "2", "cups", "")
	instruction, _ := NewInstruction(1, "Mix ingredients", nil)
	source, _ := NewSource("https://example.com", PlatformWeb, "Chef")

	tests := []struct {
		name         string
		userID       shared.ID
		title        string
		ingredients  []Ingredient
		instructions []Instruction
		source       Source
		transcript   string
		captions     string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "valid recipe",
			userID:       userID,
			title:        "Chocolate Cake",
			ingredients:  []Ingredient{ingredient},
			instructions: []Instruction{instruction},
			source:       source,
			transcript:   "Video transcript",
			captions:     "Video captions",
			wantErr:      false,
		},
		{
			name:         "empty user ID",
			userID:       "",
			title:        "Cake",
			ingredients:  []Ingredient{ingredient},
			instructions: []Instruction{instruction},
			source:       source,
			transcript:   "",
			captions:     "",
			wantErr:      true,
			errContains:  "invalid input",
		},
		{
			name:         "empty title",
			userID:       userID,
			title:        "",
			ingredients:  []Ingredient{ingredient},
			instructions: []Instruction{instruction},
			source:       source,
			transcript:   "",
			captions:     "",
			wantErr:      true,
			errContains:  "title cannot be empty",
		},
		{
			name:         "no ingredients",
			userID:       userID,
			title:        "Cake",
			ingredients:  []Ingredient{},
			instructions: []Instruction{instruction},
			source:       source,
			transcript:   "",
			captions:     "",
			wantErr:      true,
			errContains:  "must have at least one ingredient",
		},
		{
			name:         "no instructions",
			userID:       userID,
			title:        "Cake",
			ingredients:  []Ingredient{ingredient},
			instructions: []Instruction{},
			source:       source,
			transcript:   "",
			captions:     "",
			wantErr:      true,
			errContains:  "must have at least one instruction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := NewRecipe(
				tt.userID,
				tt.title,
				tt.ingredients,
				tt.instructions,
				tt.source,
				tt.transcript,
				tt.captions,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRecipe() expected error but got nil")
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("NewRecipe() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewRecipe() unexpected error = %v", err)
				return
			}

			if recipe.ID().IsEmpty() {
				t.Errorf("Recipe ID is empty")
			}

			if recipe.Title() != tt.title {
				t.Errorf("Title() = %v, want %v", recipe.Title(), tt.title)
			}

			if len(recipe.Ingredients()) != len(tt.ingredients) {
				t.Errorf("Ingredients count = %v, want %v", len(recipe.Ingredients()), len(tt.ingredients))
			}

			if len(recipe.Instructions()) != len(tt.instructions) {
				t.Errorf("Instructions count = %v, want %v", len(recipe.Instructions()), len(tt.instructions))
			}
		})
	}
}

func TestRecipe_AddIngredient(t *testing.T) {
	userID := shared.NewID()
	ingredient, _ := NewIngredient("flour", "2", "cups", "")
	instruction, _ := NewInstruction(1, "Mix", nil)
	source, _ := NewSource("https://example.com", PlatformWeb, "Chef")

	recipe, _ := NewRecipe(userID, "Cake", []Ingredient{ingredient}, []Instruction{instruction}, source, "", "")

	newIngredient, _ := NewIngredient("sugar", "1", "cup", "")
	err := recipe.AddIngredient(newIngredient)

	if err != nil {
		t.Errorf("AddIngredient() unexpected error = %v", err)
	}

	if len(recipe.Ingredients()) != 2 {
		t.Errorf("Ingredients count = %v, want 2", len(recipe.Ingredients()))
	}
}

func TestRecipe_Validate(t *testing.T) {
	userID := shared.NewID()
	ingredient, _ := NewIngredient("flour", "2", "cups", "")
	instruction, _ := NewInstruction(1, "Mix", nil)
	source, _ := NewSource("https://example.com", PlatformWeb, "Chef")

	recipe, _ := NewRecipe(userID, "Cake", []Ingredient{ingredient}, []Instruction{instruction}, source, "", "")

	if err := recipe.Validate(); err != nil {
		t.Errorf("Validate() unexpected error = %v", err)
	}
}
