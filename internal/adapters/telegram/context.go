package telegram

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/ports"
)

// ConversationState represents the current state of conversation
type ConversationState string

const (
	StateIdle                  ConversationState = "idle"
	StateAwaitingClarification ConversationState = "awaiting_clarification"
)

// PendingClarification tracks what we're asking the user about
type PendingClarification struct {
	OriginalMessage string   // What the user originally said
	Question        string   // What we asked them
	Options         []string // Suggested options (if any)
}

// ActiveFilters tracks current search filters for refinement
type ActiveFilters struct {
	Category         *recipe.Category
	DietaryTags      []recipe.DietaryTag
	IngredientFilter *recipe.IngredientFilter
	SearchTerm       string
}

// ConversationContext stores the context of a user's conversation
type ConversationContext struct {
	// LastAction is the last action performed
	LastAction ActionType
	// LastRecipes is the list of recipes from the last query
	LastRecipes []*dto.RecipeDTO
	// LastCategory is the category from the last filter
	LastCategory *recipe.Category
	// LastSearchTerm is the search term from the last search
	LastSearchTerm string
	// LastMatchIngredients is the ingredients from the last match
	LastMatchIngredients []string
	// CurrentOffset is the pagination offset for "show more"
	CurrentOffset int
	// UpdatedAt is when the context was last updated
	UpdatedAt time.Time

	// === NEW: Conversation State Machine ===
	// State is the current conversation state
	State ConversationState
	// PendingClarification tracks pending clarification if State is StateAwaitingClarification
	PendingClarification *PendingClarification

	// === NEW: History for LLM Context (last 5 turns) ===
	// History stores recent conversation turns for context-aware intent detection
	History []ports.ConversationTurn

	// === NEW: Active Filters for Refinement ===
	// ActiveFilters stores current filters that can be refined
	ActiveFilters *ActiveFilters
}

const maxHistorySize = 5

// ActionType represents the type of last action
type ActionType string

const (
	ActionNone            ActionType = ""
	ActionListRecipes     ActionType = "list_recipes"
	ActionFilterCategory  ActionType = "filter_category"
	ActionFilterIngredient ActionType = "filter_ingredient"
	ActionMatchIngredients ActionType = "match_ingredients"
	ActionShowCategories  ActionType = "show_categories"
	ActionViewRecipe      ActionType = "view_recipe"
)

// ConversationManager manages conversation contexts for users
type ConversationManager struct {
	mu       sync.RWMutex
	contexts map[shared.ID]*ConversationContext
	ttl      time.Duration
}

// NewConversationManager creates a new conversation manager
func NewConversationManager() *ConversationManager {
	cm := &ConversationManager{
		contexts: make(map[shared.ID]*ConversationContext),
		ttl:      30 * time.Minute, // Context expires after 30 minutes of inactivity
	}

	// Start cleanup goroutine
	go cm.cleanupLoop()

	return cm
}

// GetContext returns the conversation context for a user
func (cm *ConversationManager) GetContext(userID shared.ID) *ConversationContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[userID]
	if !exists || time.Since(ctx.UpdatedAt) > cm.ttl {
		return nil
	}

	return ctx
}

// SetContext sets the conversation context for a user
func (cm *ConversationManager) SetContext(userID shared.ID, ctx *ConversationContext) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx.UpdatedAt = time.Now()
	cm.contexts[userID] = ctx
}

// UpdateLastRecipes updates the last recipes in the context
func (cm *ConversationManager) UpdateLastRecipes(userID shared.ID, action ActionType, recipes []*dto.RecipeDTO) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		ctx = &ConversationContext{}
	}

	ctx.LastAction = action
	ctx.LastRecipes = recipes
	ctx.CurrentOffset = 0
	ctx.UpdatedAt = time.Now()
	cm.contexts[userID] = ctx
}

// UpdateCategoryFilter updates the category filter context
func (cm *ConversationManager) UpdateCategoryFilter(userID shared.ID, category *recipe.Category, recipes []*dto.RecipeDTO) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		ctx = &ConversationContext{}
	}

	ctx.LastAction = ActionFilterCategory
	ctx.LastCategory = category
	ctx.LastRecipes = recipes
	ctx.CurrentOffset = 0
	ctx.UpdatedAt = time.Now()
	cm.contexts[userID] = ctx
}

// UpdateIngredientSearch updates the ingredient search context
func (cm *ConversationManager) UpdateIngredientSearch(userID shared.ID, searchTerm string, recipes []*dto.RecipeDTO) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		ctx = &ConversationContext{}
	}

	ctx.LastAction = ActionFilterIngredient
	ctx.LastSearchTerm = searchTerm
	ctx.LastRecipes = recipes
	ctx.CurrentOffset = 0
	ctx.UpdatedAt = time.Now()
	cm.contexts[userID] = ctx
}

// UpdateMatchIngredients updates the match ingredients context
func (cm *ConversationManager) UpdateMatchIngredients(userID shared.ID, ingredients []string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		ctx = &ConversationContext{}
	}

	ctx.LastAction = ActionMatchIngredients
	ctx.LastMatchIngredients = ingredients
	ctx.UpdatedAt = time.Now()
	cm.contexts[userID] = ctx
}

// IncrementOffset increments the pagination offset
func (cm *ConversationManager) IncrementOffset(userID shared.ID, amount int) int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		return 0
	}

	ctx.CurrentOffset += amount
	ctx.UpdatedAt = time.Now()
	return ctx.CurrentOffset
}

// ClearContext clears the conversation context for a user
func (cm *ConversationManager) ClearContext(userID shared.ID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.contexts, userID)
}

// cleanupLoop periodically removes expired contexts
func (cm *ConversationManager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.cleanup()
	}
}

// cleanup removes expired contexts
func (cm *ConversationManager) cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for userID, ctx := range cm.contexts {
		if now.Sub(ctx.UpdatedAt) > cm.ttl {
			delete(cm.contexts, userID)
		}
	}
}

// HasRecentResults checks if the user has recent recipe results
func (cm *ConversationManager) HasRecentResults(userID shared.ID) bool {
	ctx := cm.GetContext(userID)
	return ctx != nil && len(ctx.LastRecipes) > 0
}

// GetRemainingRecipes returns recipes after the current offset
func (cm *ConversationManager) GetRemainingRecipes(userID shared.ID, pageSize int) ([]*dto.RecipeDTO, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[userID]
	if !exists || len(ctx.LastRecipes) == 0 {
		return nil, false
	}

	start := ctx.CurrentOffset
	if start >= len(ctx.LastRecipes) {
		return nil, false
	}

	end := start + pageSize
	if end > len(ctx.LastRecipes) {
		end = len(ctx.LastRecipes)
	}

	hasMore := end < len(ctx.LastRecipes)
	return ctx.LastRecipes[start:end], hasMore
}

// === NEW: Conversation State and History Methods ===

// AddTurn adds a conversation turn to history
func (cm *ConversationManager) AddTurn(userID shared.ID, role, content string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx := cm.getOrCreateContext(userID)

	turn := ports.ConversationTurn{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	ctx.History = append(ctx.History, turn)

	// Trim to max size (keep most recent turns)
	if len(ctx.History) > maxHistorySize {
		ctx.History = ctx.History[len(ctx.History)-maxHistorySize:]
	}

	ctx.UpdatedAt = time.Now()
}

// GetHistory returns the conversation history for a user
func (cm *ConversationManager) GetHistory(userID shared.ID) []ports.ConversationTurn {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	history := make([]ports.ConversationTurn, len(ctx.History))
	copy(history, ctx.History)
	return history
}

// GetHistoryForPrompt returns formatted history for LLM context
func (cm *ConversationManager) GetHistoryForPrompt(userID shared.ID) string {
	history := cm.GetHistory(userID)
	if len(history) == 0 {
		return "(No previous conversation)"
	}

	var sb strings.Builder
	for _, turn := range history {
		role := "User"
		if turn.Role == "assistant" {
			role = "Assistant"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", role, turn.Content))
	}
	return sb.String()
}

// SetState sets the conversation state for a user
func (cm *ConversationManager) SetState(userID shared.ID, state ConversationState) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx := cm.getOrCreateContext(userID)
	ctx.State = state
	ctx.UpdatedAt = time.Now()
}

// GetState returns the current conversation state for a user
func (cm *ConversationManager) GetState(userID shared.ID) ConversationState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		return StateIdle
	}
	if ctx.State == "" {
		return StateIdle
	}
	return ctx.State
}

// SetPendingClarification sets the pending clarification for a user
func (cm *ConversationManager) SetPendingClarification(userID shared.ID, pending *PendingClarification) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx := cm.getOrCreateContext(userID)
	ctx.PendingClarification = pending
	ctx.State = StateAwaitingClarification
	ctx.UpdatedAt = time.Now()
}

// GetPendingClarification returns the pending clarification for a user
func (cm *ConversationManager) GetPendingClarification(userID shared.ID) *PendingClarification {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		return nil
	}
	return ctx.PendingClarification
}

// ClearPendingClarification clears the pending clarification for a user
func (cm *ConversationManager) ClearPendingClarification(userID shared.ID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		return
	}
	ctx.PendingClarification = nil
	ctx.State = StateIdle
	ctx.UpdatedAt = time.Now()
}

// SetActiveFilters sets the active filters for a user
func (cm *ConversationManager) SetActiveFilters(userID shared.ID, filters *ActiveFilters) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx := cm.getOrCreateContext(userID)
	ctx.ActiveFilters = filters
	ctx.UpdatedAt = time.Now()
}

// GetActiveFilters returns the active filters for a user
func (cm *ConversationManager) GetActiveFilters(userID shared.ID) *ActiveFilters {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[userID]
	if !exists {
		return nil
	}
	return ctx.ActiveFilters
}

// getOrCreateContext gets or creates a conversation context (must be called with lock held)
func (cm *ConversationManager) getOrCreateContext(userID shared.ID) *ConversationContext {
	ctx, exists := cm.contexts[userID]
	if !exists {
		ctx = &ConversationContext{
			State:     StateIdle,
			UpdatedAt: time.Now(),
		}
		cm.contexts[userID] = ctx
	}
	return ctx
}
