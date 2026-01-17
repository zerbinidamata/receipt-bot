package recipe

import (
	"receipt-bot/internal/domain/shared"
	"strings"
)

// Ingredient represents a recipe ingredient (Value Object)
type Ingredient struct {
	name     string
	quantity string
	unit     string
	notes    string
}

// NewIngredient creates a new Ingredient
func NewIngredient(name, quantity, unit, notes string) (Ingredient, error) {
	name = strings.TrimSpace(name)
	quantity = strings.TrimSpace(quantity)
	unit = strings.TrimSpace(unit)
	notes = strings.TrimSpace(notes)

	if name == "" {
		return Ingredient{}, shared.ErrInvalidIngredientName
	}

	if quantity == "" {
		return Ingredient{}, shared.ErrInvalidQuantity
	}

	return Ingredient{
		name:     name,
		quantity: quantity,
		unit:     unit,
		notes:    notes,
	}, nil
}

// Name returns the ingredient name
func (i Ingredient) Name() string {
	return i.name
}

// Quantity returns the ingredient quantity
func (i Ingredient) Quantity() string {
	return i.quantity
}

// Unit returns the ingredient unit
func (i Ingredient) Unit() string {
	return i.unit
}

// Notes returns the ingredient notes
func (i Ingredient) Notes() string {
	return i.notes
}

// String returns a formatted string representation
func (i Ingredient) String() string {
	result := i.quantity
	if i.unit != "" {
		result += " " + i.unit
	}
	result += " " + i.name
	if i.notes != "" {
		result += " (" + i.notes + ")"
	}
	return result
}
