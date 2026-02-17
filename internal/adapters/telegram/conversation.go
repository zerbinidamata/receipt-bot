package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"receipt-bot/internal/application/query"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/domain/user"
	"receipt-bot/internal/ports"
)

// ConversationResult represents the result of conversation handling
type ConversationResult struct {
	// Message is the response message to send to the user
	Message string
	// Intent is the intent to execute (may be modified from original)
	Intent *ports.Intent
	// ShouldExecute indicates whether the main handler should execute the intent
	ShouldExecute bool
}

// ConversationHandler handles conversation state machine logic
type ConversationHandler struct {
	conversationManager *ConversationManager
	intentDetector      ports.IntentDetector
	listRecipesQuery    *query.ListRecipesQuery
	bot                 *Bot
}

// NewConversationHandler creates a new ConversationHandler
func NewConversationHandler(
	conversationManager *ConversationManager,
	intentDetector ports.IntentDetector,
	listRecipesQuery *query.ListRecipesQuery,
	bot *Bot,
) *ConversationHandler {
	return &ConversationHandler{
		conversationManager: conversationManager,
		intentDetector:      intentDetector,
		listRecipesQuery:    listRecipesQuery,
		bot:                 bot,
	}
}

// HandleMessage processes a user message through the conversation state machine
func (h *ConversationHandler) HandleMessage(ctx context.Context, chatID int64, userID shared.ID, text string, lang user.Language) (string, error) {
	// Check if we're awaiting clarification
	state := h.conversationManager.GetState(userID)
	if state == StateAwaitingClarification {
		return h.handleClarificationResponse(ctx, chatID, userID, text, lang)
	}

	// Add user message to conversation history
	h.conversationManager.AddTurn(userID, "user", text)

	// Get conversation history for context-aware detection
	history := h.conversationManager.GetHistory(userID)

	// Detect intent with conversation context
	intent, err := h.intentDetector.DetectIntentWithContext(ctx, text, history)
	if err != nil {
		return "", fmt.Errorf("failed to detect intent: %w", err)
	}

	// Route based on NextAction
	switch intent.NextAction {
	case ports.ActionClarify:
		return h.askClarification(ctx, userID, text, intent, lang)

	case ports.ActionRefine:
		return h.handleRefine(ctx, userID, intent, lang)

	case ports.ActionExecute:
		// For execute, we just return the intent info and let the main handler execute
		// The main handler will add the assistant response to history after execution
		return "", nil

	default:
		// Default to execute
		return "", nil
	}
}

// handleClarificationResponse processes the user's answer to a clarification question
func (h *ConversationHandler) handleClarificationResponse(ctx context.Context, chatID int64, userID shared.ID, text string, lang user.Language) (string, error) {
	pending := h.conversationManager.GetPendingClarification(userID)
	if pending == nil {
		// No pending clarification, reset state and process normally
		h.conversationManager.SetState(userID, StateIdle)
		return h.HandleMessage(ctx, chatID, userID, text, lang)
	}

	// Clear the pending clarification
	h.conversationManager.ClearPendingClarification(userID)

	// Add the clarification response to history
	h.conversationManager.AddTurn(userID, "user", text)

	// Check if user selected from options
	selectedOption := h.parseOptionSelection(text, pending.Options)

	// Combine original message with clarification response
	var combinedQuery string
	if selectedOption != "" {
		combinedQuery = fmt.Sprintf("%s - %s", pending.OriginalMessage, selectedOption)
	} else {
		combinedQuery = fmt.Sprintf("%s - %s", pending.OriginalMessage, text)
	}

	// Re-detect intent with the combined context
	history := h.conversationManager.GetHistory(userID)
	intent, err := h.intentDetector.DetectIntentWithContext(ctx, combinedQuery, history)
	if err != nil {
		return "", fmt.Errorf("failed to detect intent after clarification: %w", err)
	}

	// If still needs clarification, ask again
	if intent.NextAction == ports.ActionClarify {
		return h.askClarification(ctx, userID, combinedQuery, intent, lang)
	}

	// Return empty to signal main handler should execute
	return "", nil
}

// askClarification formats and stores a clarification question
func (h *ConversationHandler) askClarification(ctx context.Context, userID shared.ID, originalMessage string, intent *ports.Intent, lang user.Language) (string, error) {
	// Build clarification message
	var msg strings.Builder
	msg.WriteString(intent.ClarifyingQuestion)

	// Add options if available
	if len(intent.ClarifyingOptions) > 0 {
		msg.WriteString("\n\n")
		for i, option := range intent.ClarifyingOptions {
			msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, option))
		}
		msg.WriteString("\nReply with a number or describe what you want.")
	}

	// Store pending clarification
	h.conversationManager.SetPendingClarification(userID, &PendingClarification{
		OriginalMessage: originalMessage,
		Question:        intent.ClarifyingQuestion,
		Options:         intent.ClarifyingOptions,
	})

	// Add assistant message to history
	h.conversationManager.AddTurn(userID, "assistant", msg.String())

	return msg.String(), nil
}

// handleRefine handles refining previous results with new filters
func (h *ConversationHandler) handleRefine(ctx context.Context, userID shared.ID, intent *ports.Intent, lang user.Language) (string, error) {
	// Get active filters
	activeFilters := h.conversationManager.GetActiveFilters(userID)

	// Merge new intent with active filters
	mergedIntent := h.mergeFilters(intent, activeFilters)

	// Update active filters with merged result
	newFilters := h.intentToFilters(mergedIntent)
	h.conversationManager.SetActiveFilters(userID, newFilters)

	// Return empty to signal main handler should execute with merged intent
	return "", nil
}

// mergeFilters merges a new intent with existing active filters
func (h *ConversationHandler) mergeFilters(intent *ports.Intent, activeFilters *ActiveFilters) *ports.Intent {
	if activeFilters == nil {
		return intent
	}

	// Create a copy of the intent to avoid modifying the original
	merged := *intent

	// Merge category - new intent takes precedence if set
	if merged.Category == nil && activeFilters.Category != nil {
		merged.Category = activeFilters.Category
	}

	// Merge dietary tags - combine both sets
	if len(activeFilters.DietaryTags) > 0 {
		existingTags := make(map[recipe.DietaryTag]bool)
		for _, tag := range merged.DietaryTags {
			existingTags[tag] = true
		}
		for _, tag := range activeFilters.DietaryTags {
			if !existingTags[tag] {
				merged.DietaryTags = append(merged.DietaryTags, tag)
			}
		}
	}

	// Merge ingredient filter
	if merged.IngredientFilter == nil && activeFilters.IngredientFilter != nil {
		merged.IngredientFilter = activeFilters.IngredientFilter
	} else if merged.IngredientFilter != nil && activeFilters.IngredientFilter != nil {
		// Combine ingredient filters
		mergedFilter := &recipe.IngredientFilter{
			Include:  append(activeFilters.IngredientFilter.Include, merged.IngredientFilter.Include...),
			Exclude:  append(activeFilters.IngredientFilter.Exclude, merged.IngredientFilter.Exclude...),
			Optional: append(activeFilters.IngredientFilter.Optional, merged.IngredientFilter.Optional...),
		}
		// Deduplicate
		mergedFilter.Include = deduplicateStrings(mergedFilter.Include)
		mergedFilter.Exclude = deduplicateStrings(mergedFilter.Exclude)
		mergedFilter.Optional = deduplicateStrings(mergedFilter.Optional)
		merged.IngredientFilter = mergedFilter
	}

	// Merge search term - new intent takes precedence if set
	if merged.SearchTerm == "" && activeFilters.SearchTerm != "" {
		merged.SearchTerm = activeFilters.SearchTerm
	}

	return &merged
}

// intentToFilters converts an intent to ActiveFilters
func (h *ConversationHandler) intentToFilters(intent *ports.Intent) *ActiveFilters {
	if intent == nil {
		return nil
	}

	return &ActiveFilters{
		Category:         intent.Category,
		DietaryTags:      intent.DietaryTags,
		IngredientFilter: intent.IngredientFilter,
		SearchTerm:       intent.SearchTerm,
	}
}

// parseOptionSelection checks if the user selected one of the provided options
func (h *ConversationHandler) parseOptionSelection(text string, options []string) string {
	text = strings.TrimSpace(text)

	// Check if user replied with a number
	if num, err := strconv.Atoi(text); err == nil {
		if num >= 1 && num <= len(options) {
			return options[num-1]
		}
	}

	// Check if user replied with the option text (case-insensitive partial match)
	textLower := strings.ToLower(text)
	for _, option := range options {
		if strings.Contains(strings.ToLower(option), textLower) ||
			strings.Contains(textLower, strings.ToLower(option)) {
			return option
		}
	}

	return ""
}

// GetIntentWithHistory detects intent using conversation history
func (h *ConversationHandler) GetIntentWithHistory(ctx context.Context, userID shared.ID, text string) (*ports.Intent, error) {
	history := h.conversationManager.GetHistory(userID)
	return h.intentDetector.DetectIntentWithContext(ctx, text, history)
}

// AddAssistantResponse adds an assistant response to the conversation history
func (h *ConversationHandler) AddAssistantResponse(userID shared.ID, response string) {
	h.conversationManager.AddTurn(userID, "assistant", response)
}

// SetActiveFiltersFromIntent updates active filters from an executed intent
func (h *ConversationHandler) SetActiveFiltersFromIntent(userID shared.ID, intent *ports.Intent) {
	filters := h.intentToFilters(intent)
	h.conversationManager.SetActiveFilters(userID, filters)
}

// ClearActiveFilters clears the active filters for a user
func (h *ConversationHandler) ClearActiveFilters(userID shared.ID) {
	h.conversationManager.SetActiveFilters(userID, nil)
}

// IsAwaitingClarification checks if the user is awaiting clarification
func (h *ConversationHandler) IsAwaitingClarification(userID shared.ID) bool {
	return h.conversationManager.GetState(userID) == StateAwaitingClarification
}

// deduplicateStrings removes duplicate strings from a slice
func deduplicateStrings(input []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(input))
	for _, s := range input {
		lower := strings.ToLower(s)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, s)
		}
	}
	return result
}
