package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"receipt-bot/internal/application/command"
	"receipt-bot/internal/application/query"
	"receipt-bot/internal/domain/shared"
)

// Handler handles Telegram bot messages
type Handler struct {
	bot                      *Bot
	processRecipeLinkCommand *command.ProcessRecipeLinkCommand
	getOrCreateUserCommand   *command.GetOrCreateUserCommand
	listRecipesQuery         *query.ListRecipesQuery
}

// NewHandler creates a new message handler
func NewHandler(
	bot *Bot,
	processRecipeLinkCommand *command.ProcessRecipeLinkCommand,
	getOrCreateUserCommand *command.GetOrCreateUserCommand,
	listRecipesQuery *query.ListRecipesQuery,
) *Handler {
	return &Handler{
		bot:                      bot,
		processRecipeLinkCommand: processRecipeLinkCommand,
		getOrCreateUserCommand:   getOrCreateUserCommand,
		listRecipesQuery:         listRecipesQuery,
	}
}

// HandleUpdate handles a single Telegram update
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	// Only process messages
	if update.Message == nil {
		return
	}

	// Get user information
	chatID := update.Message.Chat.ID
	telegramID := update.Message.From.ID
	username := update.Message.From.UserName

	// Get or create user
	user, err := h.getOrCreateUserCommand.Execute(ctx, telegramID, username)
	if err != nil {
		log.Printf("Error getting/creating user: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to get user information. Please try again.")
		return
	}

	// Handle commands
	if update.Message.IsCommand() {
		h.handleCommand(ctx, update.Message, user.ID())
		return
	}

	// Handle text messages (URLs)
	if update.Message.Text != "" {
		h.handleTextMessage(ctx, update.Message, user.ID())
		return
	}
}

// handleCommand handles bot commands
func (h *Handler) handleCommand(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	command := message.Command()

	switch command {
	case "start":
		_ = h.bot.SendMessage(ctx, chatID, FormatWelcome())

	case "help":
		_ = h.bot.SendMessage(ctx, chatID, FormatHelp())

	case "recipes":
		h.handleListRecipes(ctx, chatID, userID)

	case "recipe":
		h.handleGetRecipe(ctx, message, userID)

	default:
		_ = h.bot.SendMessage(ctx, chatID, "Unknown command. Use /help to see available commands.")
	}
}

// handleTextMessage handles text messages (expecting URLs)
func (h *Handler) handleTextMessage(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	text := strings.TrimSpace(message.Text)

	// Check if it looks like a URL
	if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
		_ = h.bot.SendMessage(ctx, chatID,
			"👋 Please send me a link to a recipe video or webpage.\n\n"+
				"Supported platforms:\n"+
				"• TikTok\n"+
				"• YouTube\n"+
				"• Instagram\n"+
				"• Recipe websites\n\n"+
				"Use /help for more information.")
		return
	}

	// Process the recipe link
	h.handleRecipeLink(ctx, chatID, userID, text)
}

// handleRecipeLink processes a recipe link
func (h *Handler) handleRecipeLink(ctx context.Context, chatID int64, userID shared.ID, url string) {
	// Send initial acknowledgment
	_ = h.bot.SendMessage(ctx, chatID, "🔍 Processing your recipe link...\n\nThis may take a minute.")

	// Process the recipe
	recipe, err := h.processRecipeLinkCommand.Execute(ctx, url, userID, chatID)
	if err != nil {
		log.Printf("Error processing recipe: %v", err)
		errorMsg := h.formatError(err)
		_ = h.bot.SendError(ctx, chatID, errorMsg)
		return
	}

	// Send the formatted recipe
	if err := h.bot.SendRecipe(ctx, chatID, recipe); err != nil {
		log.Printf("Error sending recipe: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to send recipe. Please try again.")
	}
}

// handleGetRecipe shows a specific recipe by number
func (h *Handler) handleGetRecipe(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := message.CommandArguments()

	if args == "" {
		_ = h.bot.SendMessage(ctx, chatID, "Please specify a recipe number.\n\nUsage: /recipe <number>\nExample: /recipe 1\n\nUse /recipes to see your recipe list.")
		return
	}

	index, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil {
		_ = h.bot.SendMessage(ctx, chatID, "Invalid recipe number. Please use a number like: /recipe 1")
		return
	}

	recipeDTO, err := h.listRecipesQuery.ExecuteByIndex(ctx, userID, index)
	if err != nil {
		log.Printf("Error getting recipe: %v", err)
		_ = h.bot.SendError(ctx, chatID, err.Error())
		return
	}

	// Format and send the recipe
	messageText := FormatRecipeDTO(recipeDTO)
	_ = h.bot.SendMessage(ctx, chatID, messageText)
}

// handleListRecipes lists user's recipes
func (h *Handler) handleListRecipes(ctx context.Context, chatID int64, userID shared.ID) {
	recipes, err := h.listRecipesQuery.Execute(ctx, userID)
	if err != nil {
		log.Printf("Error listing recipes: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to list recipes. Please try again.")
		return
	}

	// Convert DTOs to domain recipes for formatting
	// (In a real app, you might want to create a separate formatter for DTOs)
	message := fmt.Sprintf("📚 *Your Recipes* (%d total)\n\n", len(recipes))

	if len(recipes) == 0 {
		message = "📭 You don't have any saved recipes yet.\n\nSend me a link to get started!"
	} else {
		for i, recipeDTO := range recipes {
			if i >= 10 {
				message += fmt.Sprintf("\n... and %d more recipes", len(recipes)-10)
				break
			}

			message += fmt.Sprintf("%d\\. %s\n", i+1, escapeMarkdown(recipeDTO.Title))
			message += fmt.Sprintf("   _From %s_\n", recipeDTO.SourcePlatform)
		}
	}

	_ = h.bot.SendMessage(ctx, chatID, message)
}

// formatError formats an error message for the user
func (h *Handler) formatError(err error) string {
	errMsg := err.Error()

	// Provide user-friendly error messages
	if strings.Contains(errMsg, "scraping failed") {
		return "Failed to download content from the URL. Please check:\n" +
			"• The link is valid and accessible\n" +
			"• The content is publicly available\n" +
			"• The platform is supported"
	}

	if strings.Contains(errMsg, "no content extracted") {
		return "Could not extract any content from the URL.\n" +
			"Please make sure the link contains a recipe."
	}

	if strings.Contains(errMsg, "no ingredients found") {
		return "Could not find any ingredients in the content.\n" +
			"Please make sure the link contains a recipe with ingredients."
	}

	if strings.Contains(errMsg, "no instructions found") {
		return "Could not find any cooking instructions in the content.\n" +
			"Please make sure the link contains a recipe with steps."
	}

	if strings.Contains(errMsg, "extraction failed") {
		return "Failed to extract recipe from the content.\n" +
			"The AI had trouble understanding this content. Please try a different recipe."
	}

	// Generic error
	return "An error occurred while processing your recipe.\n" +
		"Please try again or use /help for assistance."
}
