package ports

import "context"

// StoragePort defines the interface for file storage operations
// Note: Currently not used for video storage, but kept for future extensibility
type StoragePort interface {
	// Upload stores a file and returns its URL
	Upload(ctx context.Context, filePath string, destination string) (string, error)

	// Download retrieves a file
	Download(ctx context.Context, url string, destination string) error

	// Delete removes a file
	Delete(ctx context.Context, url string) error

	// GetURL returns a public URL for a stored file
	GetURL(ctx context.Context, path string) (string, error)
}
