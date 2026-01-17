package recipe

import (
	"strings"
)

// Service contains domain logic for recipes
type Service struct{}

// NewService creates a new recipe domain service
func NewService() *Service {
	return &Service{}
}

// MergeTextSources combines captions and transcript into a single text for LLM processing
func (s *Service) MergeTextSources(captions, transcript string) string {
	var parts []string

	if captions = strings.TrimSpace(captions); captions != "" {
		parts = append(parts, "CAPTIONS/DESCRIPTION:", captions, "")
	}

	if transcript = strings.TrimSpace(transcript); transcript != "" {
		parts = append(parts, "VIDEO TRANSCRIPT:", transcript)
	}

	return strings.Join(parts, "\n")
}

// ValidateRecipe validates a recipe according to domain rules
func (s *Service) ValidateRecipe(recipe *Recipe) error {
	return recipe.Validate()
}
