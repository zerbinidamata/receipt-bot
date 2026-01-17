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
	UserID     string    `firestore:"userId"`
	TelegramID int64     `firestore:"telegramId"`
	Username   string    `firestore:"username"`
	CreatedAt  time.Time `firestore:"createdAt"`
}

// Save persists a user to Firestore
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	doc := &userDoc{
		UserID:     u.ID().String(),
		TelegramID: u.TelegramID(),
		Username:   u.Username(),
		CreatedAt:  u.CreatedAt().Time(),
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
	return user.ReconstructUser(
		user.UserID(doc.UserID),
		doc.TelegramID,
		doc.Username,
		shared.NewTimestampFromTime(doc.CreatedAt),
	)
}
