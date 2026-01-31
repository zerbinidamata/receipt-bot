package user

import "context"

// Repository defines the interface for user persistence (Port)
type Repository interface {
	// Save persists a user
	Save(ctx context.Context, user *User) error

	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id UserID) (*User, error)

	// FindByTelegramID retrieves a user by their Telegram ID
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// UpdatePantry updates only the pantry items for a user
	UpdatePantry(ctx context.Context, userID UserID, items []string) error

	// GetPantry retrieves the pantry items for a user
	GetPantry(ctx context.Context, userID UserID) ([]string, error)

	// UpdateLanguage updates the user's language preference
	UpdateLanguage(ctx context.Context, userID UserID, language Language) error
}
