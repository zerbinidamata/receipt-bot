package recipe

import (
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want Platform
	}{
		{
			name: "YouTube full URL",
			url:  "https://www.youtube.com/watch?v=abc123",
			want: PlatformYouTube,
		},
		{
			name: "YouTube short URL",
			url:  "https://youtu.be/abc123",
			want: PlatformYouTube,
		},
		{
			name: "TikTok URL",
			url:  "https://www.tiktok.com/@user/video/123",
			want: PlatformTikTok,
		},
		{
			name: "Instagram URL",
			url:  "https://www.instagram.com/p/abc123/",
			want: PlatformInstagram,
		},
		{
			name: "Generic web URL",
			url:  "https://www.example.com/recipe",
			want: PlatformWeb,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPlatform(tt.url)
			if got != tt.want {
				t.Errorf("DetectPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSource(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		platform    Platform
		author      string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid source",
			url:      "https://www.youtube.com/watch?v=abc",
			platform: PlatformYouTube,
			author:   "Chef John",
			wantErr:  false,
		},
		{
			name:        "empty URL",
			url:         "",
			platform:    PlatformYouTube,
			author:      "Chef John",
			wantErr:     true,
			errContains: "invalid URL",
		},
		{
			name:        "invalid URL format",
			url:         "not-a-url",
			platform:    PlatformYouTube,
			author:      "Chef John",
			wantErr:     true,
			errContains: "invalid URL",
		},
		{
			name:        "invalid platform",
			url:         "https://www.youtube.com/watch?v=abc",
			platform:    PlatformUnknown,
			author:      "Chef John",
			wantErr:     true,
			errContains: "invalid platform",
		},
		{
			name:     "empty author allowed",
			url:      "https://www.youtube.com/watch?v=abc",
			platform: PlatformYouTube,
			author:   "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := NewSource(tt.url, tt.platform, tt.author)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewSource() expected error but got nil")
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("NewSource() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewSource() unexpected error = %v", err)
				return
			}

			if source.URL() != tt.url {
				t.Errorf("URL() = %v, want %v", source.URL(), tt.url)
			}

			if source.Platform() != tt.platform {
				t.Errorf("Platform() = %v, want %v", source.Platform(), tt.platform)
			}

			if !source.IsValid() {
				t.Errorf("IsValid() = false, want true")
			}
		})
	}
}
