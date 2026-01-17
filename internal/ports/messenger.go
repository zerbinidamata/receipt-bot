package ports

import (
	"context"
	"receipt-bot/internal/domain/recipe"
)

// MessengerPort defines the interface for sending messages to users
type MessengerPort interface {
	// SendMessage sends a text message to a chat
	SendMessage(ctx context.Context, chatID int64, text string) error

	// SendRecipe sends a formatted recipe to a chat
	SendRecipe(ctx context.Context, chatID int64, recipe *recipe.Recipe) error

	// SendProgress sends a progress update message
	SendProgress(ctx context.Context, chatID int64, message string) error

	// SendError sends an error message to a chat
	SendError(ctx context.Context, chatID int64, errorMsg string) error
}
