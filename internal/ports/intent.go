package ports

import (
	"context"
	"time"

	"receipt-bot/internal/domain/recipe"
)

// IntentDetector defines the interface for detecting user intent from natural language
type IntentDetector interface {
	// DetectIntent analyzes text and returns the detected intent
	DetectIntent(ctx context.Context, text string) (*Intent, error)
	// DetectIntentWithContext analyzes text with conversation history for context-aware detection
	DetectIntentWithContext(ctx context.Context, text string, history []ConversationTurn) (*Intent, error)
}

// ConversationTurn represents a single exchange in conversation history
type ConversationTurn struct {
	Role      string    // "user" or "assistant"
	Content   string    // The message text
	Timestamp time.Time
}

// ConversationAction tells the handler what to do next
type ConversationAction string

const (
	ActionExecute ConversationAction = "EXECUTE" // Execute the intent immediately
	ActionClarify ConversationAction = "CLARIFY" // Ask user for clarification
	ActionRefine  ConversationAction = "REFINE"  // Refine previous results
)

// IntentType represents the type of user intent
type IntentType string

const (
	IntentListRecipes      IntentType = "LIST_RECIPES"
	IntentFilterCategory   IntentType = "FILTER_CATEGORY"
	IntentFilterIngredient IntentType = "FILTER_INGREDIENT"
	IntentMatchIngredients IntentType = "MATCH_INGREDIENTS"
	IntentShowCategories   IntentType = "SHOW_CATEGORIES"
	IntentManagePantry     IntentType = "MANAGE_PANTRY"
	IntentHelp             IntentType = "HELP"
	IntentGreeting         IntentType = "GREETING"
	IntentUnknown          IntentType = "UNKNOWN"

	// Follow-up intents for conversation context
	IntentShowMore      IntentType = "SHOW_MORE"      // "show more", "next", "more recipes"
	IntentShowDetails   IntentType = "SHOW_DETAILS"   // "details on #3", "show me the first one"
	IntentRepeatLast    IntentType = "REPEAT_LAST"    // "show again", "repeat"
	IntentCompoundQuery IntentType = "COMPOUND_QUERY" // "quick pasta recipes", "vegan breakfast"

	// Complex search with multiple ingredients
	IntentComplexSearch IntentType = "COMPLEX_SEARCH" // "salmon and sriracha", "pasta without dairy"
)

// PantryAction represents the type of pantry management action
type PantryAction string

const (
	PantryActionShow   PantryAction = "SHOW"
	PantryActionAdd    PantryAction = "ADD"
	PantryActionRemove PantryAction = "REMOVE"
	PantryActionClear  PantryAction = "CLEAR"
)

// Intent represents the detected intent from user input
type Intent struct {
	// Type is the detected intent type
	Type IntentType

	// Category is set for FILTER_CATEGORY and COMPOUND_QUERY intents
	Category *recipe.Category

	// DietaryTags is set for COMPOUND_QUERY intent (e.g., "quick", "vegan")
	DietaryTags []recipe.DietaryTag

	// Ingredients is set for MATCH_INGREDIENTS intent (ingredients user has)
	Ingredients []string

	// SearchTerm is set for FILTER_INGREDIENT intent (specific ingredient to search for)
	SearchTerm string

	// IngredientFilter is set for COMPLEX_SEARCH intent (multiple ingredients with AND/OR/NOT)
	IngredientFilter *recipe.IngredientFilter

	// PantryAction is set for MANAGE_PANTRY intent
	PantryAction PantryAction

	// PantryItems are items to add/remove for MANAGE_PANTRY intent
	PantryItems []string

	// RecipeNumber is set for SHOW_DETAILS intent (1-based index)
	RecipeNumber int

	// Confidence is the confidence score (0.0 to 1.0)
	Confidence float64

	// RawResponse is the original text for debugging
	RawResponse string

	// === Conversation control fields ===

	// NextAction tells the handler what to do (EXECUTE, CLARIFY, or REFINE)
	NextAction ConversationAction

	// ClarifyingQuestion is the question to ask user if NextAction is CLARIFY
	ClarifyingQuestion string

	// ClarifyingOptions are suggested options for clarification
	ClarifyingOptions []string

	// RefersToLast indicates if this intent references previous results
	RefersToLast bool
}
