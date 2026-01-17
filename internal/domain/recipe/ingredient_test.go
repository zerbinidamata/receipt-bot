package recipe

import (
	"testing"
)

func TestNewIngredient(t *testing.T) {
	tests := []struct {
		name        string
		ingName     string
		quantity    string
		unit        string
		notes       string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid ingredient",
			ingName:  "flour",
			quantity: "2",
			unit:     "cups",
			notes:    "all-purpose",
			wantErr:  false,
		},
		{
			name:     "valid without unit",
			ingName:  "eggs",
			quantity: "3",
			unit:     "",
			notes:    "",
			wantErr:  false,
		},
		{
			name:        "empty name",
			ingName:     "",
			quantity:    "2",
			unit:        "cups",
			notes:       "",
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "empty quantity",
			ingName:     "flour",
			quantity:    "",
			unit:        "cups",
			notes:       "",
			wantErr:     true,
			errContains: "quantity cannot be empty",
		},
		{
			name:     "whitespace trimmed",
			ingName:  "  flour  ",
			quantity: "  2  ",
			unit:     "  cups  ",
			notes:    "  all-purpose  ",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ing, err := NewIngredient(tt.ingName, tt.quantity, tt.unit, tt.notes)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewIngredient() expected error but got nil")
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("NewIngredient() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewIngredient() unexpected error = %v", err)
				return
			}

			// Verify fields are trimmed
			if ing.Name() == "" || ing.Quantity() == "" {
				t.Errorf("NewIngredient() name or quantity is empty after trimming")
			}
		})
	}
}

func TestIngredient_String(t *testing.T) {
	tests := []struct {
		name     string
		ing      Ingredient
		expected string
	}{
		{
			name: "full ingredient",
			ing: Ingredient{
				name:     "flour",
				quantity: "2",
				unit:     "cups",
				notes:    "all-purpose",
			},
			expected: "2 cups flour (all-purpose)",
		},
		{
			name: "no unit",
			ing: Ingredient{
				name:     "eggs",
				quantity: "3",
				unit:     "",
				notes:    "",
			},
			expected: "3 eggs",
		},
		{
			name: "with unit no notes",
			ing: Ingredient{
				name:     "sugar",
				quantity: "1",
				unit:     "cup",
				notes:    "",
			},
			expected: "1 cup sugar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ing.String()
			if got != tt.expected {
				t.Errorf("Ingredient.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
