package command

import (
	"context"
	"fmt"

	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/domain/user"
)

// GetOrCreateUserCommand handles getting or creating a user
type GetOrCreateUserCommand struct {
	userRepo user.Repository
}

// NewGetOrCreateUserCommand creates a new command
func NewGetOrCreateUserCommand(userRepo user.Repository) *GetOrCreateUserCommand {
	return &GetOrCreateUserCommand{
		userRepo: userRepo,
	}
}

// Execute gets an existing user or creates a new one
func (c *GetOrCreateUserCommand) Execute(ctx context.Context, telegramID int64, username string) (*user.User, error) {
	// Try to find existing user
	existingUser, err := c.userRepo.FindByTelegramID(ctx, telegramID)
	if err == nil {
		// User exists, update username if changed
		if existingUser.Username() != username && username != "" {
			existingUser.UpdateUsername(username)
			if err := c.userRepo.Update(ctx, existingUser); err != nil {
				// Log error but don't fail - username update is not critical
				fmt.Printf("Failed to update username: %v\n", err)
			}
		}
		return existingUser, nil
	}

	// Check if error is "not found" - if so, create new user
	if err == shared.ErrUserNotFound {
		newUser, err := user.NewUser(telegramID, username)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		if err := c.userRepo.Save(ctx, newUser); err != nil {
			return nil, fmt.Errorf("failed to save user: %w", err)
		}

		return newUser, nil
	}

	// Other error occurred
	return nil, fmt.Errorf("failed to find user: %w", err)
}
