package recipe

import "context"

// Repository defines the interface for recipe persistence (Port)
type Repository interface {
	// Save persists a recipe
	Save(ctx context.Context, recipe *Recipe) error

	// FindByID retrieves a recipe by its ID
	FindByID(ctx context.Context, id RecipeID) (*Recipe, error)

	// FindByUserID retrieves all recipes for a user
	FindByUserID(ctx context.Context, userID UserID) ([]*Recipe, error)

	// FindBySourceURL retrieves a recipe by its source URL (for duplicate detection)
	FindBySourceURL(ctx context.Context, sourceURL string) (*Recipe, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *Recipe) error

	// Delete removes a recipe
	Delete(ctx context.Context, id RecipeID) error
}
