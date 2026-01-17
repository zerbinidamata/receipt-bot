package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/db"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// Client wraps Firebase services
type Client struct {
	app       *firebase.App
	auth      *auth.Client
	firestore *firestore.Client
	db        *db.Client
}

// Config holds Firebase configuration
type Config struct {
	ProjectID       string
	CredentialsPath string
	DatabaseURL     string // Optional, for Realtime Database
}

// NewClient creates a new Firebase client
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if config.ProjectID == "" {
		return nil, fmt.Errorf("Firebase project ID is required")
	}

	if config.CredentialsPath == "" {
		return nil, fmt.Errorf("Firebase credentials path is required")
	}

	// Initialize Firebase app
	conf := &firebase.Config{
		ProjectID:   config.ProjectID,
		DatabaseURL: config.DatabaseURL,
	}

	opt := option.WithCredentialsFile(config.CredentialsPath)
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	// Initialize Firestore
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firestore: %w", err)
	}

	// Initialize Auth (optional, for future use)
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Auth: %w", err)
	}

	return &Client{
		app:       app,
		auth:      authClient,
		firestore: firestoreClient,
	}, nil
}

// Close closes all Firebase connections
func (c *Client) Close() error {
	if c.firestore != nil {
		return c.firestore.Close()
	}
	return nil
}

// Firestore returns the Firestore client
func (c *Client) Firestore() *firestore.Client {
	return c.firestore
}

// Auth returns the Auth client
func (c *Client) Auth() *auth.Client {
	return c.auth
}
