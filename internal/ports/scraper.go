package ports

import (
	"context"
	"receipt-bot/internal/domain/recipe"
)

// ScraperPort defines the interface for content scraping
type ScraperPort interface {
	// Scrape extracts content from a URL
	Scrape(ctx context.Context, req ScrapeRequest) (*ScrapeResult, error)
}

// ScrapeRequest contains the parameters for scraping
type ScrapeRequest struct {
	URL      string
	Platform recipe.Platform
}

// ScrapeResult contains the extracted content
type ScrapeResult struct {
	Captions    string
	Transcript  string
	OriginalURL string
	Metadata    map[string]string
}
