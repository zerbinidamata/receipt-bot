package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"receipt-bot/internal/application/command"
	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/application/query"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/domain/user"
	"receipt-bot/internal/ports"
)

// Handler handles Telegram bot messages
type Handler struct {
	bot                      *Bot
	processRecipeLinkCommand *command.ProcessRecipeLinkCommand
	getOrCreateUserCommand   *command.GetOrCreateUserCommand
	listRecipesQuery         *query.ListRecipesQuery
	matchIngredientsCommand  *command.MatchIngredientsCommand
	managePantryCommand      *command.ManagePantryCommand
	exportRecipeCommand      *command.ExportRecipeCommand
	intentDetector           ports.IntentDetector
	conversationManager      *ConversationManager
	userRepo                 user.Repository
	llm                      ports.LLMPort
}

// HandlerConfig contains all dependencies for the Handler
type HandlerConfig struct {
	Bot                      *Bot
	ProcessRecipeLinkCommand *command.ProcessRecipeLinkCommand
	GetOrCreateUserCommand   *command.GetOrCreateUserCommand
	ListRecipesQuery         *query.ListRecipesQuery
	MatchIngredientsCommand  *command.MatchIngredientsCommand
	ManagePantryCommand      *command.ManagePantryCommand
	ExportRecipeCommand      *command.ExportRecipeCommand
	IntentDetector           ports.IntentDetector
	UserRepo                 user.Repository
	LLM                      ports.LLMPort
}

// NewHandler creates a new message handler
func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		bot:                      cfg.Bot,
		processRecipeLinkCommand: cfg.ProcessRecipeLinkCommand,
		getOrCreateUserCommand:   cfg.GetOrCreateUserCommand,
		listRecipesQuery:         cfg.ListRecipesQuery,
		matchIngredientsCommand:  cfg.MatchIngredientsCommand,
		managePantryCommand:      cfg.ManagePantryCommand,
		exportRecipeCommand:      cfg.ExportRecipeCommand,
		intentDetector:           cfg.IntentDetector,
		conversationManager:      NewConversationManager(),
		userRepo:                 cfg.UserRepo,
		llm:                      cfg.LLM,
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
	usr, err := h.getOrCreateUserCommand.Execute(ctx, telegramID, username)
	if err != nil {
		log.Printf("Error getting/creating user: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to get user information. Please try again.")
		return
	}

	// Detect language from Telegram settings for new users (first message)
	if usr.Language() == user.DefaultLanguage() && update.Message.From.LanguageCode != "" {
		detectedLang := user.ParseLanguage(update.Message.From.LanguageCode)
		if detectedLang != usr.Language() {
			usr.SetLanguage(detectedLang)
			if h.userRepo != nil {
				_ = h.userRepo.UpdateLanguage(ctx, usr.ID(), detectedLang)
			}
		}
	}

	// Handle commands
	if update.Message.IsCommand() {
		h.handleCommand(ctx, update.Message, usr)
		return
	}

	// Handle text messages (URLs)
	if update.Message.Text != "" {
		h.handleTextMessage(ctx, update.Message, usr)
		return
	}
}

// handleCommand handles bot commands
func (h *Handler) handleCommand(ctx context.Context, message *tgbotapi.Message, usr *user.User) {
	chatID := message.Chat.ID
	cmd := message.Command()
	userID := usr.ID()
	lang := usr.Language()
	t := GetTranslations(lang)

	switch cmd {
	case "start":
		_ = h.bot.SendMessage(ctx, chatID, t.Welcome)

	case "help":
		_ = h.bot.SendMessage(ctx, chatID, t.Help)

	case "recipes":
		h.handleListRecipes(ctx, message, userID)

	case "recipe":
		h.handleGetRecipe(ctx, message, userID, lang)

	case "categories":
		h.handleCategories(ctx, chatID, userID)

	case "match":
		h.handleMatch(ctx, message, userID)

	case "pantry":
		h.handlePantry(ctx, message, userID)

	case "language", "lang", "idioma":
		h.handleLanguage(ctx, message, usr)

	case "export":
		h.handleExport(ctx, message, userID)

	case "connect":
		h.handleConnect(ctx, message, userID)

	case "disconnect":
		h.handleDisconnect(ctx, message, userID)

	default:
		_ = h.bot.SendMessage(ctx, chatID, t.UnknownCommand+" "+t.UseHelpCmd)
	}
}

// handleTextMessage handles text messages (URLs or natural language)
func (h *Handler) handleTextMessage(ctx context.Context, message *tgbotapi.Message, usr *user.User) {
	chatID := message.Chat.ID
	userID := usr.ID()
	text := strings.TrimSpace(message.Text)
	t := GetTranslations(usr.Language())

	// Check if it looks like a URL first
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		h.handleRecipeLink(ctx, chatID, userID, text)
		return
	}

	// Check conversation state first - handle clarification responses
	state := h.conversationManager.GetState(userID)
	if state == StateAwaitingClarification {
		h.handleClarificationResponse(ctx, chatID, userID, text, usr.Language())
		return
	}

	// Try to detect intent from natural language
	if h.intentDetector != nil {
		// Get conversation history for context-aware detection
		history := h.conversationManager.GetHistory(userID)

		var intent *ports.Intent
		var err error

		// Use context-aware detection if we have history
		if len(history) > 0 {
			intent, err = h.intentDetector.DetectIntentWithContext(ctx, text, history)
		} else {
			intent, err = h.intentDetector.DetectIntent(ctx, text)
		}

		if err != nil {
			log.Printf("Intent detection error: %v", err)
			// Fall through to default message
		} else if intent != nil && intent.Type != ports.IntentUnknown && intent.Confidence >= 0.6 {
			// Check NextAction to determine how to proceed
			switch intent.NextAction {
			case ports.ActionClarify:
				h.handleClarification(ctx, chatID, userID, text, intent)
				return
			case ports.ActionRefine:
				h.handleRefine(ctx, chatID, userID, intent, usr.Language())
				return
			default: // ActionExecute or empty
				h.handleIntent(ctx, chatID, userID, intent, usr.Language())
				return
			}
		}
	}

	// Default fallback message
	_ = h.bot.SendMessage(ctx, chatID,
		"👋 "+t.FallbackMessage+"\n\n"+
			"• "+t.NLSendLink+"\n"+
			"• "+t.NLShowRecipes+"\n"+
			"• "+t.NLHaveIngredients+"\n"+
			"• "+t.NLMyPantry+"\n\n"+
			t.UseHelpCmd)
}

// handleIntent routes detected intents to appropriate handlers
func (h *Handler) handleIntent(ctx context.Context, chatID int64, userID shared.ID, intent *ports.Intent, lang user.Language) {
	t := GetTranslations(lang)

	switch intent.Type {
	case ports.IntentListRecipes:
		h.handleListRecipesNatural(ctx, chatID, userID, nil, "")

	case ports.IntentFilterCategory:
		h.handleListRecipesNatural(ctx, chatID, userID, intent.Category, "")

	case ports.IntentFilterIngredient:
		h.handleSearchByIngredient(ctx, chatID, userID, intent.SearchTerm)

	case ports.IntentMatchIngredients:
		h.handleMatchNatural(ctx, chatID, userID, intent.Ingredients)

	case ports.IntentShowCategories:
		h.handleCategories(ctx, chatID, userID)

	case ports.IntentManagePantry:
		h.handlePantryNatural(ctx, chatID, userID, intent.PantryAction, intent.PantryItems)

	case ports.IntentHelp:
		_ = h.bot.SendMessage(ctx, chatID, t.Help)

	case ports.IntentGreeting:
		_ = h.bot.SendMessage(ctx, chatID,
			"👋 "+t.Greeting+"\n\n"+
				t.GreetingHint+"\n"+
				"• "+t.NLShowRecipes+"\n"+
				"• "+t.NLHaveIngredients+"\n\n"+
				t.UseHelpCmd)

	case ports.IntentShowMore:
		h.handleShowMore(ctx, chatID, userID)

	case ports.IntentShowDetails:
		h.handleShowDetails(ctx, chatID, userID, intent.RecipeNumber, lang)

	case ports.IntentRepeatLast:
		h.handleRepeatLast(ctx, chatID, userID)

	case ports.IntentCompoundQuery:
		h.handleCompoundQuery(ctx, chatID, userID, intent.Category, intent.DietaryTags)

	case ports.IntentComplexSearch:
		h.handleComplexSearch(ctx, chatID, userID, intent.IngredientFilter, intent.DietaryTags)

	default:
		_ = h.bot.SendMessage(ctx, chatID,
			t.NotSureWhatYouMean+"\n"+
				"• "+t.NLSendLink+"\n"+
				"• "+t.NLShowRecipes+"\n"+
				"• "+t.NLHaveIngredients)
	}
}

// handleListRecipesNatural handles natural language recipe listing
func (h *Handler) handleListRecipesNatural(ctx context.Context, chatID int64, userID shared.ID, category *recipe.Category, _ string) {
	var recipes []*dto.RecipeDTO
	var err error
	var categoryFilter string
	var action ActionType

	if category != nil {
		categoryFilter = string(*category)
		recipes, err = h.listRecipesQuery.ExecuteByCategory(ctx, userID, *category)
		action = ActionFilterCategory
	} else {
		recipes, err = h.listRecipesQuery.Execute(ctx, userID)
		action = ActionListRecipes
	}

	if err != nil {
		log.Printf("Error listing recipes: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to list recipes. Please try again.")
		return
	}

	// Store results in conversation context
	if category != nil {
		h.conversationManager.UpdateCategoryFilter(userID, category, recipes)
	} else {
		h.conversationManager.UpdateLastRecipes(userID, action, recipes)
	}

	var msg string
	if categoryFilter != "" {
		msg = fmt.Sprintf("📚 *%s Recipes* (%d found)\n\n", categoryFilter, len(recipes))
	} else {
		msg = fmt.Sprintf("📚 *Your Recipes* (%d total)\n\n", len(recipes))
	}

	if len(recipes) == 0 {
		if categoryFilter != "" {
			msg = fmt.Sprintf("📭 No recipes found in category: %s\n\nUse /categories to see available categories.", categoryFilter)
		} else {
			msg = "📭 You don't have any saved recipes yet.\n\nSend me a link to get started!"
		}
	} else {
		for i, recipeDTO := range recipes {
			if i >= 10 {
				msg += fmt.Sprintf("\n... and %d more recipes. Say \"show more\" to see them.", len(recipes)-10)
				break
			}

			msg += fmt.Sprintf("%d. %s\n", i+1, recipeDTO.Title)
			msg += fmt.Sprintf("   _%s_ | %s\n", recipeDTO.Category, recipeDTO.SourcePlatform)
		}

		msg += "\nSay \"details on #X\" to view a recipe"
		if categoryFilter == "" {
			msg += "\nOr try \"quick pasta recipes\" to filter"
		}
	}

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleSearchByIngredient handles searching recipes by a specific ingredient
func (h *Handler) handleSearchByIngredient(ctx context.Context, chatID int64, userID shared.ID, ingredient string) {
	recipes, err := h.listRecipesQuery.SearchByIngredient(ctx, userID, ingredient)
	if err != nil {
		log.Printf("Error searching recipes: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to search recipes. Please try again.")
		return
	}

	// Store results in conversation context
	h.conversationManager.UpdateIngredientSearch(userID, ingredient, recipes)

	msg := fmt.Sprintf("🔍 *Recipes with %s* (%d found)\n\n", ingredient, len(recipes))

	if len(recipes) == 0 {
		msg = fmt.Sprintf("📭 No recipes found containing \"%s\".\n\nTry a different ingredient or use /recipes to see all your recipes.", ingredient)
	} else {
		for i, recipeDTO := range recipes {
			if i >= 10 {
				msg += fmt.Sprintf("\n... and %d more recipes. Say \"show more\" to see them.", len(recipes)-10)
				break
			}

			msg += fmt.Sprintf("%d. %s\n", i+1, recipeDTO.Title)
			msg += fmt.Sprintf("   _%s_ | %s\n", recipeDTO.Category, recipeDTO.SourcePlatform)
		}

		msg += "\nSay \"details on #X\" to view a recipe"
	}

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleMatchNatural handles natural language ingredient matching
func (h *Handler) handleMatchNatural(ctx context.Context, chatID int64, userID shared.ID, ingredients []string) {
	if len(ingredients) == 0 {
		// Check if user has pantry items
		pantry, err := h.managePantryCommand.GetPantry(ctx, userID)
		if err != nil || len(pantry.Items) == 0 {
			_ = h.bot.SendMessage(ctx, chatID,
				"Tell me what ingredients you have!\n\n"+
					"Example: \"I have chicken, pasta, and garlic\"\n"+
					"Or add items to your pantry: \"add chicken to pantry\"")
			return
		}
		ingredients = pantry.Items
	}

	// Store ingredients in conversation context
	h.conversationManager.UpdateMatchIngredients(userID, ingredients)

	input := command.MatchIngredientsInput{
		UserID:      userID,
		Ingredients: ingredients,
	}

	result, err := h.matchIngredientsCommand.Execute(ctx, input)
	if err != nil {
		log.Printf("Error matching ingredients: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to match ingredients. Please try again.")
		return
	}

	msg := FormatMatchResults(result)
	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleShowMore shows more results from the previous query
func (h *Handler) handleShowMore(ctx context.Context, chatID int64, userID shared.ID) {
	convCtx := h.conversationManager.GetContext(userID)
	if convCtx == nil || len(convCtx.LastRecipes) == 0 {
		_ = h.bot.SendMessage(ctx, chatID,
			"I don't have any previous results to show more of.\n\n"+
				"Try searching for recipes first, like:\n"+
				"• \"Show my recipes\"\n"+
				"• \"Pasta recipes\"")
		return
	}

	// Increment offset and get next page
	pageSize := 10
	newOffset := h.conversationManager.IncrementOffset(userID, pageSize)
	recipes, hasMore := h.conversationManager.GetRemainingRecipes(userID, pageSize)

	if len(recipes) == 0 {
		_ = h.bot.SendMessage(ctx, chatID,
			"You've seen all the recipes from your last search.\n\n"+
				"Try a different search or say \"show again\" to repeat.")
		return
	}

	msg := fmt.Sprintf("📚 *More Recipes* (showing %d-%d of %d)\n\n",
		newOffset-pageSize+1, newOffset-pageSize+len(recipes), len(convCtx.LastRecipes))

	for i, recipeDTO := range recipes {
		idx := newOffset - pageSize + i + 1
		msg += fmt.Sprintf("%d. %s\n", idx, recipeDTO.Title)
		msg += fmt.Sprintf("   _%s_ | %s\n", recipeDTO.Category, recipeDTO.SourcePlatform)
	}

	if hasMore {
		msg += "\nSay \"show more\" for additional recipes"
	} else {
		msg += "\nThat's all! Say \"show again\" to see them from the beginning"
	}
	msg += "\nOr say \"details on #X\" to view a specific recipe"

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleShowDetails shows details of a specific recipe from the last results
func (h *Handler) handleShowDetails(ctx context.Context, chatID int64, userID shared.ID, recipeNumber int, lang user.Language) {
	convCtx := h.conversationManager.GetContext(userID)
	if convCtx == nil || len(convCtx.LastRecipes) == 0 {
		_ = h.bot.SendMessage(ctx, chatID,
			"I don't have any recent recipe results.\n\n"+
				"Try searching for recipes first, like:\n"+
				"• \"Show my recipes\"\n"+
				"• \"Pasta recipes\"")
		return
	}

	if recipeNumber < 1 || recipeNumber > len(convCtx.LastRecipes) {
		_ = h.bot.SendMessage(ctx, chatID,
			fmt.Sprintf("Recipe #%d not found. I have %d recipes from your last search.\n\n"+
				"Try \"details on #1\" through \"details on #%d\"",
				recipeNumber, len(convCtx.LastRecipes), len(convCtx.LastRecipes)))
		return
	}

	recipeDTO := convCtx.LastRecipes[recipeNumber-1]

	// Translate recipe if user language is Portuguese and we have LLM
	var translation *TranslatedRecipeDTO
	if lang == user.LanguagePortuguese && h.llm != nil {
		translated, err := h.translateRecipe(ctx, recipeDTO, "Portuguese")
		if err != nil {
			log.Printf("Translation error (showing original): %v", err)
		} else {
			translation = translated
		}
	}

	messageText := FormatRecipeDTOWithTranslation(recipeDTO, translation, lang)
	_ = h.bot.SendMessage(ctx, chatID, messageText)

	// Update context to track that user viewed a recipe
	h.conversationManager.SetContext(userID, &ConversationContext{
		LastAction:  ActionViewRecipe,
		LastRecipes: convCtx.LastRecipes,
	})
}

// translateRecipe translates a recipe DTO to the target language using LLM
func (h *Handler) translateRecipe(ctx context.Context, rec *dto.RecipeDTO, targetLang string) (*TranslatedRecipeDTO, error) {
	// Build input for translation
	input := &ports.RecipeTranslationInput{
		Title:        rec.Title,
		Ingredients:  make([]ports.IngredientData, len(rec.Ingredients)),
		Instructions: make([]ports.InstructionData, len(rec.Instructions)),
	}

	for i, ing := range rec.Ingredients {
		input.Ingredients[i] = ports.IngredientData{
			Name:     ing.Name,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
			Notes:    ing.Notes,
		}
	}

	for i, inst := range rec.Instructions {
		input.Instructions[i] = ports.InstructionData{
			StepNumber: inst.StepNumber,
			Text:       inst.Text,
		}
	}

	// Call LLM for translation
	output, err := h.llm.TranslateRecipe(ctx, input, targetLang)
	if err != nil {
		return nil, err
	}

	// Convert to TranslatedRecipeDTO
	result := &TranslatedRecipeDTO{
		Title:        output.Title,
		Ingredients:  make([]dto.IngredientDTO, len(output.Ingredients)),
		Instructions: make([]dto.InstructionDTO, len(output.Instructions)),
	}

	for i, ing := range output.Ingredients {
		result.Ingredients[i] = dto.IngredientDTO{
			Name:     ing.Name,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
			Notes:    ing.Notes,
		}
	}

	for i, inst := range output.Instructions {
		result.Instructions[i] = dto.InstructionDTO{
			StepNumber: inst.StepNumber,
			Text:       inst.Text,
		}
	}

	return result, nil
}

// handleRepeatLast repeats the last action/query
func (h *Handler) handleRepeatLast(ctx context.Context, chatID int64, userID shared.ID) {
	convCtx := h.conversationManager.GetContext(userID)
	if convCtx == nil {
		_ = h.bot.SendMessage(ctx, chatID,
			"I don't have any previous action to repeat.\n\n"+
				"Try something like:\n"+
				"• \"Show my recipes\"\n"+
				"• \"Pasta recipes\"")
		return
	}

	// Reset offset for repeat
	h.conversationManager.SetContext(userID, &ConversationContext{
		LastAction:           convCtx.LastAction,
		LastRecipes:          convCtx.LastRecipes,
		LastCategory:         convCtx.LastCategory,
		LastSearchTerm:       convCtx.LastSearchTerm,
		LastMatchIngredients: convCtx.LastMatchIngredients,
		CurrentOffset:        0,
	})

	switch convCtx.LastAction {
	case ActionListRecipes:
		h.handleListRecipesNatural(ctx, chatID, userID, nil, "")
	case ActionFilterCategory:
		h.handleListRecipesNatural(ctx, chatID, userID, convCtx.LastCategory, "")
	case ActionFilterIngredient:
		h.handleSearchByIngredient(ctx, chatID, userID, convCtx.LastSearchTerm)
	case ActionMatchIngredients:
		h.handleMatchNatural(ctx, chatID, userID, convCtx.LastMatchIngredients)
	default:
		_ = h.bot.SendMessage(ctx, chatID,
			"I'm not sure what to repeat.\n\n"+
				"Try a new search like \"pasta recipes\"")
	}
}

// handleClarification sends a clarifying question to the user
func (h *Handler) handleClarification(ctx context.Context, chatID int64, userID shared.ID, originalMessage string, intent *ports.Intent) {
	// Set pending clarification in conversation manager
	h.conversationManager.SetPendingClarification(userID, &PendingClarification{
		OriginalMessage: originalMessage,
		Question:        intent.ClarifyingQuestion,
		Options:         intent.ClarifyingOptions,
	})

	// Add the user's message to history
	h.conversationManager.AddTurn(userID, "user", originalMessage)

	// Build the clarification message with options
	msg := intent.ClarifyingQuestion
	if len(intent.ClarifyingOptions) > 0 {
		msg += "\n\nOptions:\n"
		for i, option := range intent.ClarifyingOptions {
			msg += fmt.Sprintf("%d. %s\n", i+1, option)
		}
		msg += "\nYou can reply with a number or type your preference."
	}

	// Add the assistant's clarifying question to history
	h.conversationManager.AddTurn(userID, "assistant", msg)

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleClarificationResponse handles the user's response to a clarifying question
func (h *Handler) handleClarificationResponse(ctx context.Context, chatID int64, userID shared.ID, text string, lang user.Language) {
	pending := h.conversationManager.GetPendingClarification(userID)
	if pending == nil {
		// No pending clarification, treat as normal message
		h.conversationManager.SetState(userID, StateIdle)
		return
	}

	// Clear the pending clarification
	h.conversationManager.ClearPendingClarification(userID)

	// Add the user's response to history
	h.conversationManager.AddTurn(userID, "user", text)

	// Check if the user selected an option by number
	selectedText := text
	if num, err := strconv.Atoi(strings.TrimSpace(text)); err == nil && num > 0 && num <= len(pending.Options) {
		selectedText = pending.Options[num-1]
	}

	// Combine the original message with the clarification response for intent detection
	combinedQuery := pending.OriginalMessage + " " + selectedText

	// Re-run intent detection with the combined context
	if h.intentDetector != nil {
		history := h.conversationManager.GetHistory(userID)
		intent, err := h.intentDetector.DetectIntentWithContext(ctx, combinedQuery, history)
		if err != nil {
			log.Printf("Intent detection error after clarification: %v", err)
		} else if intent != nil && intent.Type != ports.IntentUnknown && intent.Confidence >= 0.5 {
			h.handleIntent(ctx, chatID, userID, intent, lang)
			return
		}
	}

	// Fallback: couldn't understand the clarification response
	t := GetTranslations(lang)
	_ = h.bot.SendMessage(ctx, chatID,
		"I still couldn't understand that. Let's start over.\n\n"+
			t.UseHelpCmd)
}

// handleRefine refines previous search results with new filters
func (h *Handler) handleRefine(ctx context.Context, chatID int64, userID shared.ID, intent *ports.Intent, lang user.Language) {
	// Get active filters from conversation manager
	activeFilters := h.conversationManager.GetActiveFilters(userID)

	// Merge new intent filters with existing active filters
	mergedFilters := &ActiveFilters{}
	if activeFilters != nil {
		// Copy existing filters
		mergedFilters.Category = activeFilters.Category
		mergedFilters.DietaryTags = append([]recipe.DietaryTag{}, activeFilters.DietaryTags...)
		mergedFilters.IngredientFilter = activeFilters.IngredientFilter
		mergedFilters.SearchTerm = activeFilters.SearchTerm
	}

	// Apply new filters from intent
	if intent.Category != nil {
		mergedFilters.Category = intent.Category
	}
	if len(intent.DietaryTags) > 0 {
		// Add new dietary tags without duplicates
		existingTags := make(map[recipe.DietaryTag]bool)
		for _, tag := range mergedFilters.DietaryTags {
			existingTags[tag] = true
		}
		for _, tag := range intent.DietaryTags {
			if !existingTags[tag] {
				mergedFilters.DietaryTags = append(mergedFilters.DietaryTags, tag)
			}
		}
	}
	if intent.IngredientFilter != nil {
		// Merge ingredient filters
		if mergedFilters.IngredientFilter == nil {
			mergedFilters.IngredientFilter = intent.IngredientFilter
		} else {
			// Combine Include (AND logic)
			mergedFilters.IngredientFilter.Include = append(
				mergedFilters.IngredientFilter.Include,
				intent.IngredientFilter.Include...,
			)
			// Combine Optional (OR logic)
			mergedFilters.IngredientFilter.Optional = append(
				mergedFilters.IngredientFilter.Optional,
				intent.IngredientFilter.Optional...,
			)
			// Combine Exclude (NOT logic)
			mergedFilters.IngredientFilter.Exclude = append(
				mergedFilters.IngredientFilter.Exclude,
				intent.IngredientFilter.Exclude...,
			)
		}
	}
	if intent.SearchTerm != "" {
		mergedFilters.SearchTerm = intent.SearchTerm
	}

	// Update active filters
	h.conversationManager.SetActiveFilters(userID, mergedFilters)

	// Re-execute the search with merged filters
	if mergedFilters.IngredientFilter != nil {
		h.handleComplexSearch(ctx, chatID, userID, mergedFilters.IngredientFilter, mergedFilters.DietaryTags)
	} else if mergedFilters.Category != nil || len(mergedFilters.DietaryTags) > 0 {
		h.handleCompoundQuery(ctx, chatID, userID, mergedFilters.Category, mergedFilters.DietaryTags)
	} else if mergedFilters.SearchTerm != "" {
		h.handleSearchByIngredient(ctx, chatID, userID, mergedFilters.SearchTerm)
	} else {
		// No filters to refine, just list recipes
		h.handleListRecipesNatural(ctx, chatID, userID, nil, "")
	}
}

// handleCompoundQuery handles queries combining category and dietary tags
func (h *Handler) handleCompoundQuery(ctx context.Context, chatID int64, userID shared.ID, category *recipe.Category, dietaryTags []recipe.DietaryTag) {
	recipes, err := h.listRecipesQuery.ExecuteByFilters(ctx, userID, category, dietaryTags)
	if err != nil {
		log.Printf("Error filtering recipes: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to filter recipes. Please try again.")
		return
	}

	// Build filter description
	var filterParts []string
	if len(dietaryTags) > 0 {
		for _, tag := range dietaryTags {
			filterParts = append(filterParts, string(tag))
		}
	}
	if category != nil {
		filterParts = append(filterParts, string(*category))
	}
	filterDesc := strings.Join(filterParts, " ")
	if filterDesc == "" {
		filterDesc = "filtered"
	}

	// Store in conversation context
	h.conversationManager.UpdateCategoryFilter(userID, category, recipes)

	msg := fmt.Sprintf("📚 *%s Recipes* (%d found)\n\n", strings.Title(filterDesc), len(recipes))

	if len(recipes) == 0 {
		msg = fmt.Sprintf("📭 No recipes found matching: %s\n\n"+
			"Try a different combination or use /categories to see what you have.", filterDesc)
	} else {
		for i, recipeDTO := range recipes {
			if i >= 10 {
				msg += fmt.Sprintf("\n... and %d more recipes. Say \"show more\" to see them.", len(recipes)-10)
				break
			}

			msg += fmt.Sprintf("%d. %s\n", i+1, recipeDTO.Title)
			msg += fmt.Sprintf("   _%s_ | %s\n", recipeDTO.Category, recipeDTO.SourcePlatform)
		}

		if len(recipes) <= 10 {
			msg += "\nSay \"details on #X\" to view a recipe"
		}
	}

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleComplexSearch handles complex ingredient searches with filters and dietary tags
func (h *Handler) handleComplexSearch(ctx context.Context, chatID int64, userID shared.ID, filter *recipe.IngredientFilter, dietaryTags []recipe.DietaryTag) {
	recipes, err := h.listRecipesQuery.SearchByIngredientFilterWithTags(ctx, userID, filter, dietaryTags)
	if err != nil {
		log.Printf("Error searching recipes with filter: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to search recipes. Please try again.")
		return
	}

	// Build filter description
	var filterParts []string
	if filter != nil {
		if len(filter.Include) > 0 {
			filterParts = append(filterParts, "with "+strings.Join(filter.Include, " and "))
		}
		if len(filter.Optional) > 0 {
			filterParts = append(filterParts, "with any of: "+strings.Join(filter.Optional, ", "))
		}
		if len(filter.Exclude) > 0 {
			filterParts = append(filterParts, "without "+strings.Join(filter.Exclude, ", "))
		}
	}
	if len(dietaryTags) > 0 {
		for _, tag := range dietaryTags {
			filterParts = append(filterParts, string(tag))
		}
	}
	filterDesc := strings.Join(filterParts, ", ")
	if filterDesc == "" {
		filterDesc = "filtered"
	}

	// Store in conversation context
	h.conversationManager.UpdateLastRecipes(userID, ActionFilterIngredient, recipes)
	h.conversationManager.SetActiveFilters(userID, &ActiveFilters{
		DietaryTags:      dietaryTags,
		IngredientFilter: filter,
	})

	msg := fmt.Sprintf("🔍 *Recipes %s* (%d found)\n\n", filterDesc, len(recipes))

	if len(recipes) == 0 {
		msg = fmt.Sprintf("📭 No recipes found matching: %s\n\n"+
			"Try a different combination or use /recipes to see all your recipes.", filterDesc)
	} else {
		for i, recipeDTO := range recipes {
			if i >= 10 {
				msg += fmt.Sprintf("\n... and %d more recipes. Say \"show more\" to see them.", len(recipes)-10)
				break
			}

			msg += fmt.Sprintf("%d. %s\n", i+1, recipeDTO.Title)
			msg += fmt.Sprintf("   _%s_ | %s\n", recipeDTO.Category, recipeDTO.SourcePlatform)
		}

		if len(recipes) <= 10 {
			msg += "\nSay \"details on #X\" to view a recipe"
		}
	}

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handlePantryNatural handles natural language pantry management
func (h *Handler) handlePantryNatural(ctx context.Context, chatID int64, userID shared.ID, action ports.PantryAction, items []string) {
	switch action {
	case ports.PantryActionAdd:
		if len(items) == 0 {
			_ = h.bot.SendMessage(ctx, chatID,
				"What would you like to add to your pantry?\n\n"+
					"Example: \"add chicken and rice to pantry\"")
			return
		}
		pantry, err := h.managePantryCommand.AddItems(ctx, userID, items)
		if err != nil {
			log.Printf("Error adding pantry items: %v", err)
			_ = h.bot.SendError(ctx, chatID, "Failed to add items. Please try again.")
			return
		}
		_ = h.bot.SendMessage(ctx, chatID,
			fmt.Sprintf("✅ Added %d item(s) to your pantry.\n\nYour pantry now has %d items.\nSay \"what can I make\" to find recipes!",
				len(items), len(pantry.Items)))

	case ports.PantryActionRemove:
		if len(items) == 0 {
			_ = h.bot.SendMessage(ctx, chatID,
				"What would you like to remove from your pantry?\n\n"+
					"Example: \"remove chicken from pantry\"")
			return
		}
		pantry, err := h.managePantryCommand.RemoveItems(ctx, userID, items)
		if err != nil {
			log.Printf("Error removing pantry items: %v", err)
			_ = h.bot.SendError(ctx, chatID, "Failed to remove items. Please try again.")
			return
		}
		_ = h.bot.SendMessage(ctx, chatID,
			fmt.Sprintf("✅ Removed item(s) from your pantry.\n\nYour pantry now has %d items.",
				len(pantry.Items)))

	case ports.PantryActionClear:
		err := h.managePantryCommand.ClearPantry(ctx, userID)
		if err != nil {
			log.Printf("Error clearing pantry: %v", err)
			_ = h.bot.SendError(ctx, chatID, "Failed to clear pantry. Please try again.")
			return
		}
		_ = h.bot.SendMessage(ctx, chatID, "✅ Your pantry has been cleared.")

	default: // PantryActionShow
		h.handlePantryShow(ctx, chatID, userID)
	}
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
func (h *Handler) handleGetRecipe(ctx context.Context, message *tgbotapi.Message, userID shared.ID, lang user.Language) {
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

	// Translate recipe if user language is Portuguese and we have LLM
	var translation *TranslatedRecipeDTO
	if lang == user.LanguagePortuguese && h.llm != nil {
		translated, err := h.translateRecipe(ctx, recipeDTO, "Portuguese")
		if err != nil {
			log.Printf("Translation error (showing original): %v", err)
		} else {
			translation = translated
		}
	}

	// Format and send the recipe
	messageText := FormatRecipeDTOWithTranslation(recipeDTO, translation, lang)
	_ = h.bot.SendMessage(ctx, chatID, messageText)
}

// handleListRecipes lists user's recipes, optionally filtered by category
func (h *Handler) handleListRecipes(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	var recipes []*dto.RecipeDTO
	var err error
	var categoryFilter string

	if args != "" {
		// Filter by category
		category := recipe.ParseCategory(args)
		categoryFilter = string(category)
		recipes, err = h.listRecipesQuery.ExecuteByCategory(ctx, userID, category)
	} else {
		// List all recipes
		recipes, err = h.listRecipesQuery.Execute(ctx, userID)
	}

	if err != nil {
		log.Printf("Error listing recipes: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to list recipes\\. Please try again\\.")
		return
	}

	var msg string
	if categoryFilter != "" {
		msg = fmt.Sprintf("📚 *%s Recipes* \\(%d found\\)\n\n", escapeMarkdown(categoryFilter), len(recipes))
	} else {
		msg = fmt.Sprintf("📚 *Your Recipes* \\(%d total\\)\n\n", len(recipes))
	}

	if len(recipes) == 0 {
		if categoryFilter != "" {
			msg = fmt.Sprintf("📭 No recipes found in category: %s\n\nUse /categories to see available categories\\.", escapeMarkdown(categoryFilter))
		} else {
			msg = "📭 You don't have any saved recipes yet\\.\n\nSend me a link to get started\\!"
		}
	} else {
		for i, recipeDTO := range recipes {
			if i >= 10 {
				msg += fmt.Sprintf("\n\\.\\.\\. and %d more recipes", len(recipes)-10)
				break
			}

			msg += fmt.Sprintf("%d\\. %s\n", i+1, escapeMarkdown(recipeDTO.Title))
			msg += fmt.Sprintf("   _%s_ \\| %s\n", escapeMarkdown(recipeDTO.Category), recipeDTO.SourcePlatform)
		}

		msg += "\nUse /recipe <number> to view details"
		if categoryFilter == "" {
			msg += "\nUse /recipes <category> to filter"
		}
	}

	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handleCategories shows recipe category counts
func (h *Handler) handleCategories(ctx context.Context, chatID int64, userID shared.ID) {
	counts, err := h.listRecipesQuery.GetCategoryCounts(ctx, userID)
	if err != nil {
		log.Printf("Error getting category counts: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to get categories\\. Please try again\\.")
		return
	}

	// Calculate total
	total := 0
	for _, count := range counts {
		total += count
	}

	if total == 0 {
		_ = h.bot.SendMessage(ctx, chatID, "📭 You don't have any saved recipes yet\\.\n\nSend me a link to get started\\!")
		return
	}

	message := FormatCategories(counts, total)
	_ = h.bot.SendMessage(ctx, chatID, message)
}

// handleMatch handles the /match command for ingredient matching
func (h *Handler) handleMatch(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	// If no ingredients provided, check if user has pantry items
	if args == "" {
		pantry, err := h.managePantryCommand.GetPantry(ctx, userID)
		if err != nil || len(pantry.Items) == 0 {
			_ = h.bot.SendMessage(ctx, chatID,
				"Please provide ingredients to match\\.\n\n"+
					"*Usage:* /match chicken, pasta, garlic\n"+
					"*Or:* Add items to your pantry with /pantry add")
			return
		}
		args = strings.Join(pantry.Items, ", ")
	}

	// Parse ingredients from comma-separated list
	ingredients := parseIngredientList(args)
	if len(ingredients) == 0 {
		_ = h.bot.SendMessage(ctx, chatID, "Please provide at least one ingredient\\.\n\nExample: /match chicken, pasta, garlic")
		return
	}

	// Check for flags
	strictMatch := strings.Contains(args, "--strict")
	var categoryFilter *recipe.Category
	if strings.Contains(args, "--category") {
		// Extract category from args (simple parsing)
		parts := strings.Split(args, "--category")
		if len(parts) > 1 {
			catParts := strings.Fields(parts[1])
			if len(catParts) > 0 {
				cat := recipe.ParseCategory(catParts[0])
				categoryFilter = &cat
			}
		}
	}

	// Execute matching
	input := command.MatchIngredientsInput{
		UserID:         userID,
		Ingredients:    ingredients,
		CategoryFilter: categoryFilter,
		StrictMatch:    strictMatch,
	}

	result, err := h.matchIngredientsCommand.Execute(ctx, input)
	if err != nil {
		log.Printf("Error matching ingredients: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to match ingredients\\. Please try again\\.")
		return
	}

	// Format and send results
	msg := FormatMatchResults(result)
	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handlePantry handles the /pantry command for pantry management
func (h *Handler) handlePantry(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	// Parse subcommand
	parts := strings.SplitN(args, " ", 2)
	subcommand := ""
	itemsArg := ""
	if len(parts) > 0 {
		subcommand = strings.ToLower(parts[0])
	}
	if len(parts) > 1 {
		itemsArg = parts[1]
	}

	switch subcommand {
	case "":
		// Show current pantry
		h.handlePantryShow(ctx, chatID, userID)

	case "add":
		h.handlePantryAdd(ctx, chatID, userID, itemsArg)

	case "remove":
		h.handlePantryRemove(ctx, chatID, userID, itemsArg)

	case "clear":
		h.handlePantryClear(ctx, chatID, userID)

	default:
		// Treat as items to add if no recognized subcommand
		h.handlePantryAdd(ctx, chatID, userID, args)
	}
}

// handlePantryShow shows the user's pantry
func (h *Handler) handlePantryShow(ctx context.Context, chatID int64, userID shared.ID) {
	pantry, err := h.managePantryCommand.GetPantry(ctx, userID)
	if err != nil {
		log.Printf("Error getting pantry: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to get pantry\\. Please try again\\.")
		return
	}

	msg := FormatPantry(pantry.Items)
	_ = h.bot.SendMessage(ctx, chatID, msg)
}

// handlePantryAdd adds items to the pantry
func (h *Handler) handlePantryAdd(ctx context.Context, chatID int64, userID shared.ID, itemsArg string) {
	if itemsArg == "" {
		_ = h.bot.SendMessage(ctx, chatID,
			"Please specify items to add\\.\n\n"+
				"*Usage:* /pantry add butter, eggs, milk")
		return
	}

	items := parseIngredientList(itemsArg)
	if len(items) == 0 {
		_ = h.bot.SendMessage(ctx, chatID, "Please provide at least one item to add\\.")
		return
	}

	pantry, err := h.managePantryCommand.AddItems(ctx, userID, items)
	if err != nil {
		log.Printf("Error adding pantry items: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to add items\\. Please try again\\.")
		return
	}

	_ = h.bot.SendMessage(ctx, chatID,
		fmt.Sprintf("✅ Added %d item\\(s\\) to your pantry\\.\n\nYour pantry now has %d items\\.\nUse /match to find recipes\\!",
			len(items), len(pantry.Items)))
}

// handlePantryRemove removes items from the pantry
func (h *Handler) handlePantryRemove(ctx context.Context, chatID int64, userID shared.ID, itemsArg string) {
	if itemsArg == "" {
		_ = h.bot.SendMessage(ctx, chatID,
			"Please specify items to remove\\.\n\n"+
				"*Usage:* /pantry remove butter, eggs")
		return
	}

	items := parseIngredientList(itemsArg)
	if len(items) == 0 {
		_ = h.bot.SendMessage(ctx, chatID, "Please provide at least one item to remove\\.")
		return
	}

	pantry, err := h.managePantryCommand.RemoveItems(ctx, userID, items)
	if err != nil {
		log.Printf("Error removing pantry items: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to remove items\\. Please try again\\.")
		return
	}

	_ = h.bot.SendMessage(ctx, chatID,
		fmt.Sprintf("✅ Removed item\\(s\\) from your pantry\\.\n\nYour pantry now has %d items\\.",
			len(pantry.Items)))
}

// handlePantryClear clears all pantry items
func (h *Handler) handlePantryClear(ctx context.Context, chatID int64, userID shared.ID) {
	err := h.managePantryCommand.ClearPantry(ctx, userID)
	if err != nil {
		log.Printf("Error clearing pantry: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Failed to clear pantry\\. Please try again\\.")
		return
	}

	_ = h.bot.SendMessage(ctx, chatID, "✅ Your pantry has been cleared\\.")
}

// handleLanguage handles the /language command for changing user language preference
func (h *Handler) handleLanguage(ctx context.Context, message *tgbotapi.Message, usr *user.User) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())
	t := GetTranslations(usr.Language())

	// If no argument, show current language and options
	if args == "" {
		_ = h.bot.SendMessage(ctx, chatID,
			t.LanguageCurrent+"\n\n"+
				t.LanguageChoose+"\n"+
				"• /language en \\- "+t.LanguageEnglish+"\n"+
				"• /language pt \\- "+t.LanguagePortuguese)
		return
	}

	// Parse the requested language
	newLang := user.ParseLanguage(args)

	// Update user's language preference
	usr.SetLanguage(newLang)
	if h.userRepo != nil {
		if err := h.userRepo.UpdateLanguage(ctx, usr.ID(), newLang); err != nil {
			log.Printf("Error updating language: %v", err)
			_ = h.bot.SendError(ctx, chatID, "Failed to update language\\. Please try again\\.")
			return
		}
	}

	// Send confirmation in the NEW language
	newT := GetTranslations(newLang)
	_ = h.bot.SendMessage(ctx, chatID, "✅ "+newT.LanguageSet)
}

// parseIngredientList parses a comma-separated list of ingredients
func parseIngredientList(input string) []string {
	// Remove any flags
	input = strings.ReplaceAll(input, "--strict", "")
	if idx := strings.Index(input, "--category"); idx != -1 {
		// Remove --category and its argument
		endIdx := strings.Index(input[idx:], ",")
		if endIdx == -1 {
			input = input[:idx]
		} else {
			input = input[:idx] + input[idx+endIdx:]
		}
	}

	// Split by comma
	parts := strings.Split(input, ",")
	var ingredients []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ingredients = append(ingredients, part)
		}
	}
	return ingredients
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

// handleExport handles the /export command
func (h *Handler) handleExport(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	if h.exportRecipeCommand == nil {
		_ = h.bot.SendError(ctx, chatID, "Export functionality is not available.")
		return
	}

	// Parse arguments: /export <format> [recipe_number]
	parts := strings.Fields(args)
	if len(parts) == 0 {
		// Show export help
		_ = h.bot.SendMessage(ctx, chatID,
			"*Export Recipes*\n\n"+
				"*Usage:*\n"+
				"/export obsidian \\- Export all recipes as Markdown\n"+
				"/export obsidian <number> \\- Export a specific recipe\n"+
				"/export notion \\- Export all to Notion\n"+
				"/export notion <number> \\- Export specific recipe to Notion\n\n"+
				"*Obsidian:* Downloads a \\.md file with YAML frontmatter\n"+
				"*Notion:* Requires /connect notion first")
		return
	}

	format := strings.ToLower(parts[0])
	var recipeID *shared.ID

	// Check if a recipe number was specified
	if len(parts) > 1 {
		recipeNum, err := strconv.Atoi(parts[1])
		if err != nil {
			_ = h.bot.SendError(ctx, chatID, "Invalid recipe number\\. Use /export <format> <number>")
			return
		}

		// Get recipe by index to find its ID
		recipeDTO, err := h.listRecipesQuery.ExecuteByIndex(ctx, userID, recipeNum)
		if err != nil {
			_ = h.bot.SendError(ctx, chatID, fmt.Sprintf("Recipe #%d not found\\.", recipeNum))
			return
		}

		id := shared.ID(recipeDTO.ID)
		recipeID = &id
	}

	// Execute export
	var exportFormat command.ExportFormat
	switch format {
	case "obsidian", "md", "markdown":
		exportFormat = command.ExportFormatObsidian
	case "notion":
		exportFormat = command.ExportFormatNotion
	default:
		_ = h.bot.SendError(ctx, chatID, "Unknown format\\. Use 'obsidian' or 'notion'\\.")
		return
	}

	_ = h.bot.SendMessage(ctx, chatID, "📤 Exporting recipes...")

	input := command.ExportRecipeInput{
		UserID:   userID,
		RecipeID: recipeID,
		Format:   exportFormat,
	}

	result, err := h.exportRecipeCommand.Execute(ctx, input)
	if err != nil {
		log.Printf("Export error: %v", err)
		_ = h.bot.SendError(ctx, chatID, "Export failed\\. Please try again\\.")
		return
	}

	if !result.Success {
		_ = h.bot.SendMessage(ctx, chatID, result.Message)
		return
	}

	// Handle result based on format
	switch exportFormat {
	case command.ExportFormatObsidian:
		// Send file as document
		caption := fmt.Sprintf("✅ %s", result.Message)
		if err := h.bot.SendDocument(ctx, chatID, result.Filename, result.Data, caption); err != nil {
			log.Printf("Failed to send document: %v", err)
			_ = h.bot.SendError(ctx, chatID, "Failed to send file\\. Please try again\\.")
		}
	case command.ExportFormatNotion:
		// Send success message with link
		msg := fmt.Sprintf("✅ %s", result.Message)
		if result.URL != "" {
			msg += fmt.Sprintf("\n\n[View in Notion](%s)", result.URL)
		}
		_ = h.bot.SendMessage(ctx, chatID, msg)
	}
}

// handleConnect handles the /connect command
func (h *Handler) handleConnect(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		_ = h.bot.SendMessage(ctx, chatID,
			"*Connect External Services*\n\n"+
				"*Usage:*\n"+
				"/connect notion \\- Connect to Notion\n\n"+
				"*Connected services:*\n"+
				"• Notion \\- Sync recipes to your Notion database")
		return
	}

	service := strings.ToLower(args)
	switch service {
	case "notion":
		h.handleConnectNotion(ctx, chatID, userID)
	default:
		_ = h.bot.SendError(ctx, chatID, "Unknown service\\. Currently supported: notion")
	}
}

// handleConnectNotion handles Notion OAuth connection
func (h *Handler) handleConnectNotion(ctx context.Context, chatID int64, userID shared.ID) {
	if h.exportRecipeCommand == nil || !h.exportRecipeCommand.HasNotionExporter() {
		_ = h.bot.SendError(ctx, chatID, "Notion integration is not configured\\.")
		return
	}

	// TODO: Implement OAuth flow - for now, show a placeholder message
	_ = h.bot.SendMessage(ctx, chatID,
		"*Connect to Notion*\n\n"+
			"Notion integration requires OAuth authentication\\.\n\n"+
			"This feature is coming soon\\! For now, use /export obsidian to export your recipes as Markdown files\\.")
}

// handleDisconnect handles the /disconnect command
func (h *Handler) handleDisconnect(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
	chatID := message.Chat.ID
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		_ = h.bot.SendMessage(ctx, chatID,
			"*Disconnect Services*\n\n"+
				"*Usage:*\n"+
				"/disconnect notion \\- Disconnect from Notion")
		return
	}

	service := strings.ToLower(args)
	switch service {
	case "notion":
		h.handleDisconnectNotion(ctx, chatID, userID)
	default:
		_ = h.bot.SendError(ctx, chatID, "Unknown service\\. Currently supported: notion")
	}
}

// handleDisconnectNotion handles Notion disconnection
func (h *Handler) handleDisconnectNotion(ctx context.Context, chatID int64, userID shared.ID) {
	if h.exportRecipeCommand == nil || !h.exportRecipeCommand.HasNotionExporter() {
		_ = h.bot.SendError(ctx, chatID, "Notion integration is not configured\\.")
		return
	}

	// TODO: Implement disconnection
	_ = h.bot.SendMessage(ctx, chatID, "Notion integration is not yet connected\\.")
}
