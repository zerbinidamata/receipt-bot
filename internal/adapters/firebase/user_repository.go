package firebase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"receipt-bot/internal/domain/shared"
	"receipt-bot/internal/domain/user"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// UserRepository implements the user.Repository interface using Firestore
type UserRepository struct {
	client *firestore.Client
}

// NewUserRepository creates a new Firebase user repository
func NewUserRepository(client *firestore.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

// userDoc represents the Firestore document structure for users
type userDoc struct {
	UserID          string     `firestore:"userId"`
	TelegramID      int64      `firestore:"telegramId"`
	Username        string     `firestore:"username"`
	Language        string     `firestore:"language,omitempty"`
	CreatedAt       time.Time  `firestore:"createdAt"`
	PantryItems     []string   `firestore:"pantryItems,omitempty"`
	PantryUpdatedAt *time.Time `firestore:"pantryUpdatedAt,omitempty"`

	// Notion integration
	NotionAccessToken string     `firestore:"notionAccessToken,omitempty"`
	NotionWorkspaceID string     `firestore:"notionWorkspaceId,omitempty"`
	NotionDatabaseID  string     `firestore:"notionDatabaseId,omitempty"`
	NotionConnectedAt *time.Time `firestore:"notionConnectedAt,omitempty"`
}

// Save persists a user to Firestore
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	doc := &userDoc{
		UserID:            u.ID().String(),
		TelegramID:        u.TelegramID(),
		Username:          u.Username(),
		Language:          string(u.Language()),
		CreatedAt:         u.CreatedAt().Time(),
		PantryItems:       u.PantryItems(),
		PantryUpdatedAt:   u.PantryUpdatedAt(),
		NotionAccessToken: u.NotionAccessToken(),
		NotionWorkspaceID: u.NotionWorkspaceID(),
		NotionDatabaseID:  u.NotionDatabaseID(),
		NotionConnectedAt: u.NotionConnectedAt(),
	}

	_, err := r.client.Collection("users").Doc(u.ID().String()).Set(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// FindByID retrieves a user by their ID
func (r *UserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	doc, err := r.client.Collection("users").Doc(id.String()).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	var userDoc userDoc
	if err := doc.DataTo(&userDoc); err != nil {
		return nil, fmt.Errorf("failed to parse user document: %w", err)
	}

	return r.fromDocument(&userDoc), nil
}

// FindByTelegramID retrieves a user by their Telegram ID
func (r *UserRepository) FindByTelegramID(ctx context.Context, telegramID int64) (*user.User, error) {
	iter := r.client.Collection("users").
		Where("telegramId", "==", telegramID).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, shared.ErrUserNotFound
	}
	if err != nil {
		errStr := err.Error()
		// Provide helpful error message for permission issues
		if contains(errStr, "PermissionDenied") || contains(errStr, "Cloud Firestore API has not been used") {
			return nil, fmt.Errorf("failed to find user by Telegram ID: %w\n\n"+
				"Troubleshooting:\n"+
				"1. Verify the Cloud Firestore API is enabled: https://console.developers.google.com/apis/api/firestore.googleapis.com/overview\n"+
				"2. Check your service account has the 'Firebase Admin SDK Administrator Service Agent' role\n"+
				"3. Verify FIREBASE_PROJECT_ID matches the project in your credentials file\n"+
				"4. Wait a few minutes after enabling the API for changes to propagate", err)
		}
		return nil, fmt.Errorf("failed to find user by Telegram ID: %w", err)
	}

	var userDoc userDoc
	if err := doc.DataTo(&userDoc); err != nil {
		return nil, fmt.Errorf("failed to parse user document: %w", err)
	}

	return r.fromDocument(&userDoc), nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	return r.Save(ctx, u) // In Firestore, Set accomplishes update
}

// fromDocument converts a Firestore document to a domain User
func (r *UserRepository) fromDocument(doc *userDoc) *user.User {
	return user.ReconstructUserFromData(user.UserData{
		ID:                user.UserID(doc.UserID),
		TelegramID:        doc.TelegramID,
		Username:          doc.Username,
		Language:          user.Language(doc.Language),
		CreatedAt:         shared.NewTimestampFromTime(doc.CreatedAt),
		PantryItems:       doc.PantryItems,
		PantryUpdatedAt:   doc.PantryUpdatedAt,
		NotionAccessToken: doc.NotionAccessToken,
		NotionWorkspaceID: doc.NotionWorkspaceID,
		NotionDatabaseID:  doc.NotionDatabaseID,
		NotionConnectedAt: doc.NotionConnectedAt,
	})
}

// UpdateLanguage updates only the language preference for a user
func (r *UserRepository) UpdateLanguage(ctx context.Context, userID user.UserID, language user.Language) error {
	_, err := r.client.Collection("users").Doc(userID.String()).Update(ctx, []firestore.Update{
		{Path: "language", Value: string(language)},
	})
	if err != nil {
		return fmt.Errorf("failed to update language: %w", err)
	}
	return nil
}

// UpdatePantry updates only the pantry items for a user
func (r *UserRepository) UpdatePantry(ctx context.Context, userID user.UserID, items []string) error {
	now := time.Now()
	_, err := r.client.Collection("users").Doc(userID.String()).Update(ctx, []firestore.Update{
		{Path: "pantryItems", Value: items},
		{Path: "pantryUpdatedAt", Value: now},
	})
	if err != nil {
		return fmt.Errorf("failed to update pantry: %w", err)
	}
	return nil
}

// GetPantry retrieves the pantry items for a user
func (r *UserRepository) GetPantry(ctx context.Context, userID user.UserID) ([]string, error) {
	doc, err := r.client.Collection("users").Doc(userID.String()).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pantry: %w", err)
	}

	var userDoc userDoc
	if err := doc.DataTo(&userDoc); err != nil {
		return nil, fmt.Errorf("failed to parse user document: %w", err)
	}

	return userDoc.PantryItems, nil
}

// UpdateNotionConnection updates the Notion connection for a user
func (r *UserRepository) UpdateNotionConnection(ctx context.Context, userID user.UserID, accessToken, workspaceID, databaseID string) error {
	now := time.Now()
	_, err := r.client.Collection("users").Doc(userID.String()).Update(ctx, []firestore.Update{
		{Path: "notionAccessToken", Value: accessToken},
		{Path: "notionWorkspaceId", Value: workspaceID},
		{Path: "notionDatabaseId", Value: databaseID},
		{Path: "notionConnectedAt", Value: now},
	})
	if err != nil {
		return fmt.Errorf("failed to update Notion connection: %w", err)
	}
	return nil
}

// ClearNotionConnection removes the Notion connection for a user
func (r *UserRepository) ClearNotionConnection(ctx context.Context, userID user.UserID) error {
	_, err := r.client.Collection("users").Doc(userID.String()).Update(ctx, []firestore.Update{
		{Path: "notionAccessToken", Value: ""},
		{Path: "notionWorkspaceId", Value: ""},
		{Path: "notionDatabaseId", Value: ""},
		{Path: "notionConnectedAt", Value: nil},
	})
	if err != nil {
		return fmt.Errorf("failed to clear Notion connection: %w", err)
	}
	return nil
}

// GetNotionConnection retrieves Notion connection details for a user
func (r *UserRepository) GetNotionConnection(ctx context.Context, userID user.UserID) (accessToken, workspaceID, databaseID string, connectedAt *time.Time, err error) {
	doc, err := r.client.Collection("users").Doc(userID.String()).Get(ctx)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	var userDoc userDoc
	if err := doc.DataTo(&userDoc); err != nil {
		return "", "", "", nil, fmt.Errorf("failed to parse user document: %w", err)
	}

	return userDoc.NotionAccessToken, userDoc.NotionWorkspaceID, userDoc.NotionDatabaseID, userDoc.NotionConnectedAt, nil
}
