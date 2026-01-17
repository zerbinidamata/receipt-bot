package recipe

import (
	"net/url"
	"receipt-bot/internal/domain/shared"
	"strings"
)

// Platform represents the source platform of a recipe
type Platform string

const (
	PlatformTikTok    Platform = "tiktok"
	PlatformYouTube   Platform = "youtube"
	PlatformInstagram Platform = "instagram"
	PlatformWeb       Platform = "web"
	PlatformUnknown   Platform = "unknown"
)

// Source represents the origin of a recipe (Value Object)
type Source struct {
	url      string
	platform Platform
	author   string
}

// NewSource creates a new Source
func NewSource(rawURL string, platform Platform, author string) (Source, error) {
	rawURL = strings.TrimSpace(rawURL)
	author = strings.TrimSpace(author)

	if rawURL == "" {
		return Source{}, shared.ErrInvalidURL
	}

	// Validate URL format
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return Source{}, shared.ErrInvalidURL
	}

	// Validate platform
	if !isValidPlatform(platform) {
		return Source{}, shared.ErrInvalidPlatform
	}

	return Source{
		url:      rawURL,
		platform: platform,
		author:   author,
	}, nil
}

// URL returns the source URL
func (s Source) URL() string {
	return s.url
}

// Platform returns the source platform
func (s Source) Platform() Platform {
	return s.platform
}

// Author returns the content author
func (s Source) Author() string {
	return s.author
}

// IsValid checks if the source is valid
func (s Source) IsValid() bool {
	return s.url != "" && isValidPlatform(s.platform)
}

// isValidPlatform checks if a platform is valid
func isValidPlatform(p Platform) bool {
	switch p {
	case PlatformTikTok, PlatformYouTube, PlatformInstagram, PlatformWeb:
		return true
	default:
		return false
	}
}

// DetectPlatform attempts to detect the platform from a URL
func DetectPlatform(rawURL string) Platform {
	rawURL = strings.ToLower(rawURL)

	if strings.Contains(rawURL, "tiktok.com") {
		return PlatformTikTok
	}
	if strings.Contains(rawURL, "youtube.com") || strings.Contains(rawURL, "youtu.be") {
		return PlatformYouTube
	}
	if strings.Contains(rawURL, "instagram.com") {
		return PlatformInstagram
	}

	return PlatformWeb
}
