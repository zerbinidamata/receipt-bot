package user

import (
	"strings"
	"time"

	"receipt-bot/internal/domain/shared"
)

// UserID represents a unique user identifier
type UserID = shared.ID

// Language represents a supported language code
type Language string

const (
	LanguageEnglish    Language = "en"
	LanguagePortuguese Language = "pt-BR"
)

// IsValid checks if the language is supported
func (l Language) IsValid() bool {
	return l == LanguageEnglish || l == LanguagePortuguese
}

// DefaultLanguage returns the default language
func DefaultLanguage() Language {
	return LanguageEnglish
}

// ParseLanguage parses a language code string to a Language
func ParseLanguage(code string) Language {
	// Normalize: lowercase and trim any trailing special characters
	code = strings.ToLower(strings.TrimSpace(code))
	code = strings.TrimRight(code, "\\/ ")

	switch code {
	case "pt", "pt-br", "pt_br", "portuguese", "português":
		return LanguagePortuguese
	default:
		return LanguageEnglish
	}
}

// User represents a bot user (Entity)
type User struct {
	id              UserID
	telegramID      int64
	username        string
	language        Language
	createdAt       shared.Timestamp
	pantryItems     []string
	pantryUpdatedAt *time.Time

	// Notion integration
	notionAccessToken  string
	notionWorkspaceID  string
	notionDatabaseID   string
	notionConnectedAt  *time.Time
}

// NewUser creates a new User
func NewUser(telegramID int64, username string) (*User, error) {
	if telegramID <= 0 {
		return nil, shared.ErrInvalidTelegramID
	}

	return &User{
		id:         shared.NewID(),
		telegramID: telegramID,
		username:   username,
		language:   DefaultLanguage(),
		createdAt:  shared.NewTimestamp(),
	}, nil
}

// UserData contains data for reconstructing a user from storage
type UserData struct {
	ID              UserID
	TelegramID      int64
	Username        string
	Language        Language
	CreatedAt       shared.Timestamp
	PantryItems     []string
	PantryUpdatedAt *time.Time

	// Notion integration (optional)
	NotionAccessToken string
	NotionWorkspaceID string
	NotionDatabaseID  string
	NotionConnectedAt *time.Time
}

// ReconstructUser reconstructs a user from stored data (for repository)
func ReconstructUser(id UserID, telegramID int64, username string, language Language, createdAt shared.Timestamp, pantryItems []string, pantryUpdatedAt *time.Time) *User {
	return ReconstructUserFromData(UserData{
		ID:              id,
		TelegramID:      telegramID,
		Username:        username,
		Language:        language,
		CreatedAt:       createdAt,
		PantryItems:     pantryItems,
		PantryUpdatedAt: pantryUpdatedAt,
	})
}

// ReconstructUserFromData reconstructs a user from a UserData struct
func ReconstructUserFromData(data UserData) *User {
	lang := data.Language
	if !lang.IsValid() {
		lang = DefaultLanguage()
	}
	return &User{
		id:                 data.ID,
		telegramID:         data.TelegramID,
		username:           data.Username,
		language:           lang,
		createdAt:          data.CreatedAt,
		pantryItems:        data.PantryItems,
		pantryUpdatedAt:    data.PantryUpdatedAt,
		notionAccessToken:  data.NotionAccessToken,
		notionWorkspaceID:  data.NotionWorkspaceID,
		notionDatabaseID:   data.NotionDatabaseID,
		notionConnectedAt:  data.NotionConnectedAt,
	}
}

// ID returns the user ID
func (u *User) ID() UserID {
	return u.id
}

// TelegramID returns the Telegram user ID
func (u *User) TelegramID() int64 {
	return u.telegramID
}

// Username returns the username
func (u *User) Username() string {
	return u.username
}

// CreatedAt returns the creation timestamp
func (u *User) CreatedAt() shared.Timestamp {
	return u.createdAt
}

// UpdateUsername updates the user's username
func (u *User) UpdateUsername(username string) {
	u.username = username
}

// Language returns the user's preferred language
func (u *User) Language() Language {
	if u.language == "" {
		return DefaultLanguage()
	}
	return u.language
}

// SetLanguage sets the user's preferred language
func (u *User) SetLanguage(lang Language) {
	if lang.IsValid() {
		u.language = lang
	}
}

// PantryItems returns the user's pantry items
func (u *User) PantryItems() []string {
	return u.pantryItems
}

// PantryUpdatedAt returns when the pantry was last updated
func (u *User) PantryUpdatedAt() *time.Time {
	return u.pantryUpdatedAt
}

// SetPantryItems sets the user's pantry items
func (u *User) SetPantryItems(items []string) {
	u.pantryItems = items
	now := time.Now()
	u.pantryUpdatedAt = &now
}

// AddPantryItems adds items to the user's pantry
func (u *User) AddPantryItems(items []string) {
	// Create a map for deduplication
	existing := make(map[string]bool)
	for _, item := range u.pantryItems {
		existing[item] = true
	}

	// Add new items if not already present
	for _, item := range items {
		if !existing[item] {
			u.pantryItems = append(u.pantryItems, item)
			existing[item] = true
		}
	}

	now := time.Now()
	u.pantryUpdatedAt = &now
}

// RemovePantryItems removes items from the user's pantry
func (u *User) RemovePantryItems(items []string) {
	toRemove := make(map[string]bool)
	for _, item := range items {
		toRemove[item] = true
	}

	newItems := make([]string, 0, len(u.pantryItems))
	for _, item := range u.pantryItems {
		if !toRemove[item] {
			newItems = append(newItems, item)
		}
	}

	u.pantryItems = newItems
	now := time.Now()
	u.pantryUpdatedAt = &now
}

// ClearPantry clears all pantry items
func (u *User) ClearPantry() {
	u.pantryItems = nil
	now := time.Now()
	u.pantryUpdatedAt = &now
}

// NotionAccessToken returns the Notion access token
func (u *User) NotionAccessToken() string {
	return u.notionAccessToken
}

// NotionWorkspaceID returns the Notion workspace ID
func (u *User) NotionWorkspaceID() string {
	return u.notionWorkspaceID
}

// NotionDatabaseID returns the Notion database ID
func (u *User) NotionDatabaseID() string {
	return u.notionDatabaseID
}

// NotionConnectedAt returns when Notion was connected
func (u *User) NotionConnectedAt() *time.Time {
	return u.notionConnectedAt
}

// HasNotionConnection returns true if the user has a Notion connection
func (u *User) HasNotionConnection() bool {
	return u.notionAccessToken != "" && u.notionConnectedAt != nil
}

// SetNotionConnection sets the Notion connection details
func (u *User) SetNotionConnection(accessToken, workspaceID, databaseID string) {
	u.notionAccessToken = accessToken
	u.notionWorkspaceID = workspaceID
	u.notionDatabaseID = databaseID
	now := time.Now()
	u.notionConnectedAt = &now
}

// ClearNotionConnection removes the Notion connection
func (u *User) ClearNotionConnection() {
	u.notionAccessToken = ""
	u.notionWorkspaceID = ""
	u.notionDatabaseID = ""
	u.notionConnectedAt = nil
}
