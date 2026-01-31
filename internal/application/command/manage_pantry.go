package command

import (
	"context"
	"fmt"
	"strings"

	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/domain/matching"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/domain/user"
)

// ManagePantryCommand handles pantry operations
type ManagePantryCommand struct {
	userRepo   user.Repository
	normalizer matching.IngredientNormalizer
}

// NewManagePantryCommand creates a new command
func NewManagePantryCommand(userRepo user.Repository) *ManagePantryCommand {
	return &ManagePantryCommand{
		userRepo:   userRepo,
		normalizer: matching.NewRuleBasedNormalizer(),
	}
}

// GetPantry retrieves the user's pantry items
func (c *ManagePantryCommand) GetPantry(ctx context.Context, userID shared.ID) (*dto.PantryDTO, error) {
	items, err := c.userRepo.GetPantry(ctx, user.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get pantry: %w", err)
	}

	return &dto.PantryDTO{
		Items: items,
	}, nil
}

// AddItems adds items to the user's pantry
func (c *ManagePantryCommand) AddItems(ctx context.Context, userID shared.ID, items []string) (*dto.PantryDTO, error) {
	// Get current pantry
	currentItems, err := c.userRepo.GetPantry(ctx, user.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get current pantry: %w", err)
	}

	// Normalize and deduplicate new items
	normalized := c.normalizeItems(items)

	// Create a set of existing items for deduplication
	existing := make(map[string]bool)
	for _, item := range currentItems {
		existing[item] = true
	}

	// Add new items
	for _, item := range normalized {
		if !existing[item] {
			currentItems = append(currentItems, item)
			existing[item] = true
		}
	}

	// Save updated pantry
	if err := c.userRepo.UpdatePantry(ctx, user.UserID(userID), currentItems); err != nil {
		return nil, fmt.Errorf("failed to update pantry: %w", err)
	}

	return &dto.PantryDTO{
		Items: currentItems,
	}, nil
}

// RemoveItems removes items from the user's pantry
func (c *ManagePantryCommand) RemoveItems(ctx context.Context, userID shared.ID, items []string) (*dto.PantryDTO, error) {
	// Get current pantry
	currentItems, err := c.userRepo.GetPantry(ctx, user.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get current pantry: %w", err)
	}

	// Normalize items to remove
	toRemove := make(map[string]bool)
	for _, item := range c.normalizeItems(items) {
		toRemove[item] = true
	}

	// Filter out removed items
	newItems := make([]string, 0, len(currentItems))
	for _, item := range currentItems {
		if !toRemove[item] {
			newItems = append(newItems, item)
		}
	}

	// Save updated pantry
	if err := c.userRepo.UpdatePantry(ctx, user.UserID(userID), newItems); err != nil {
		return nil, fmt.Errorf("failed to update pantry: %w", err)
	}

	return &dto.PantryDTO{
		Items: newItems,
	}, nil
}

// ClearPantry removes all items from the user's pantry
func (c *ManagePantryCommand) ClearPantry(ctx context.Context, userID shared.ID) error {
	if err := c.userRepo.UpdatePantry(ctx, user.UserID(userID), []string{}); err != nil {
		return fmt.Errorf("failed to clear pantry: %w", err)
	}
	return nil
}

// SetPantry replaces the entire pantry with new items
func (c *ManagePantryCommand) SetPantry(ctx context.Context, userID shared.ID, items []string) (*dto.PantryDTO, error) {
	// Normalize and deduplicate items
	normalized := c.normalizeItems(items)

	// Deduplicate
	seen := make(map[string]bool)
	unique := make([]string, 0, len(normalized))
	for _, item := range normalized {
		if !seen[item] {
			unique = append(unique, item)
			seen[item] = true
		}
	}

	// Save pantry
	if err := c.userRepo.UpdatePantry(ctx, user.UserID(userID), unique); err != nil {
		return nil, fmt.Errorf("failed to set pantry: %w", err)
	}

	return &dto.PantryDTO{
		Items: unique,
	}, nil
}

// normalizeItems normalizes a list of ingredient items
func (c *ManagePantryCommand) normalizeItems(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		// Normalize the ingredient
		norm := c.normalizer.Normalize(item)
		if norm != "" {
			normalized = append(normalized, norm)
		}
	}
	return normalized
}
