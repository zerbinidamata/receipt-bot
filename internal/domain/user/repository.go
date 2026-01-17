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
}
