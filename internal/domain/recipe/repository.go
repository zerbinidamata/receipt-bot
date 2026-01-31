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

	// FindByUserIDAndCategory retrieves recipes for a user filtered by category
	FindByUserIDAndCategory(ctx context.Context, userID UserID, category Category) ([]*Recipe, error)

	// FindByUserIDAndFilters retrieves recipes for a user with optional category and dietary tag filters
	FindByUserIDAndFilters(ctx context.Context, userID UserID, category *Category, dietaryTags []DietaryTag) ([]*Recipe, error)

	// SearchByIngredient searches recipes containing a specific ingredient in title or ingredients
	SearchByIngredient(ctx context.Context, userID UserID, ingredient string) ([]*Recipe, error)

	// FindBySourceURL retrieves a recipe by its source URL (for duplicate detection)
	FindBySourceURL(ctx context.Context, sourceURL string) (*Recipe, error)

	// GetCategoryCounts returns the count of recipes per category for a user
	GetCategoryCounts(ctx context.Context, userID UserID) (map[Category]int, error)

	// Update updates an existing recipe
	Update(ctx context.Context, recipe *Recipe) error

	// Delete removes a recipe
	Delete(ctx context.Context, id RecipeID) error
}
