package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"receipt-bot/internal/domain/recipe"
)

// Bot wraps the Telegram bot API
type Bot struct {
	api   *tgbotapi.BotAPI
	debug bool
}

// Config holds Telegram bot configuration
type Config struct {
	BotToken string
	Debug    bool
}

// NewBot creates a new Telegram bot
func NewBot(config Config) (*Bot, error) {
	if config.BotToken == "" {
		return nil, fmt.Errorf("bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = config.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Bot{
		api:   bot,
		debug: config.Debug,
	}, nil
}

// GetUpdatesChan returns a channel for receiving updates
func (b *Bot) GetUpdatesChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.api.GetUpdatesChan(u)
}

// SendMessage sends a text message to a chat
func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendRecipe sends a formatted recipe to a chat
func (b *Bot) SendRecipe(ctx context.Context, chatID int64, rec *recipe.Recipe) error {
	text := FormatRecipe(rec)
	return b.SendMessage(ctx, chatID, text)
}

// SendProgress sends a progress update message
func (b *Bot) SendProgress(ctx context.Context, chatID int64, message string) error {
	return b.SendMessage(ctx, chatID, message)
}

// SendError sends an error message to a chat
func (b *Bot) SendError(ctx context.Context, chatID int64, errorMsg string) error {
	text := fmt.Sprintf("❌ *Error*\n\n%s", errorMsg)
	return b.SendMessage(ctx, chatID, text)
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
}
