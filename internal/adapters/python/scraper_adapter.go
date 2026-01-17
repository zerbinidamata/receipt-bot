package python

import (
	"context"
	"fmt"
	"time"

	"receipt-bot/internal/adapters/python/pb"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/ports"
)

// ScraperAdapter implements the ScraperPort interface using the Python gRPC service
type ScraperAdapter struct {
	client *GRPCClient
}

// NewScraperAdapter creates a new scraper adapter
func NewScraperAdapter(pythonServiceURL string, timeout time.Duration) (*ScraperAdapter, error) {
	client, err := NewGRPCClient(pythonServiceURL, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &ScraperAdapter{
		client: client,
	}, nil
}

// Close closes the underlying gRPC connection
func (a *ScraperAdapter) Close() error {
	return a.client.Close()
}

// Scrape implements the ScraperPort interface
func (a *ScraperAdapter) Scrape(ctx context.Context, req ports.ScrapeRequest) (*ports.ScrapeResult, error) {
	// Convert domain platform to proto platform
	protoPlatform := convertPlatform(req.Platform)

	// Create gRPC request
	grpcReq := &pb.ScrapeRequest{
		Url:           req.URL,
		Platform:      protoPlatform,
		DownloadVideo: true,  // Always download for transcription
		Transcribe:    true,  // Always transcribe
	}

	// Log the request
	fmt.Printf("[DEBUG] Scraper request - URL: %s, Platform: %v, DownloadVideo: %v, Transcribe: %v\n", 
		grpcReq.Url, grpcReq.Platform, grpcReq.DownloadVideo, grpcReq.Transcribe)

	// Call Python service
	resp, err := a.client.ScrapeContent(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("scraping failed: %w", err)
	}

	// Log the response
	fmt.Printf("[DEBUG] Scraper response - Captions length: %d, Transcript length: %d, Has error: %v\n",
		len(resp.Captions), len(resp.Transcript), resp.Error != nil)
	if resp.Error != nil {
		fmt.Printf("[DEBUG] Scraper error: %s (code: %s)\n", resp.Error.Message, resp.Error.Code)
	}
	if len(resp.Captions) > 0 {
		captionsPreview := resp.Captions
		if len(captionsPreview) > 200 {
			captionsPreview = captionsPreview[:200] + "..."
		}
		fmt.Printf("[DEBUG] Captions preview: %s\n", captionsPreview)
	}
	if len(resp.Transcript) > 0 {
		transcriptPreview := resp.Transcript
		if len(transcriptPreview) > 200 {
			transcriptPreview = transcriptPreview[:200] + "..."
		}
		fmt.Printf("[DEBUG] Transcript preview: %s\n", transcriptPreview)
	}

	// Check for service-level errors
	if resp.Error != nil && resp.Error.Message != "" {
		return nil, fmt.Errorf("scraping error: %s (code: %s)", resp.Error.Message, resp.Error.Code)
	}

	// Convert response to domain result
	result := &ports.ScrapeResult{
		Captions:    resp.Captions,
		Transcript:  resp.Transcript,
		OriginalURL: resp.OriginalUrl,
		Metadata:    resp.Metadata,
	}

	return result, nil
}

// convertPlatform converts domain Platform to proto Platform
func convertPlatform(p recipe.Platform) pb.Platform {
	switch p {
	case recipe.PlatformTikTok:
		return pb.Platform_PLATFORM_TIKTOK
	case recipe.PlatformYouTube:
		return pb.Platform_PLATFORM_YOUTUBE
	case recipe.PlatformInstagram:
		return pb.Platform_PLATFORM_INSTAGRAM
	case recipe.PlatformWeb:
		return pb.Platform_PLATFORM_WEB
	default:
		return pb.Platform_PLATFORM_UNKNOWN
	}
}
