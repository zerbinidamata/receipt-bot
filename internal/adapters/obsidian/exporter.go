package obsidian

import (
	"archive/zip"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/ports"
)

// Exporter implements the ObsidianExporter interface
type Exporter struct{}

// NewExporter creates a new Obsidian exporter
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportRecipe exports a single recipe as Obsidian-compatible markdown
func (e *Exporter) ExportRecipe(rec *recipe.Recipe) (*ports.ExportResult, error) {
	markdown := e.generateMarkdown(rec)
	filename := e.sanitizeFilename(rec.Title()) + ".md"

	return &ports.ExportResult{
		Success:  true,
		Format:   "obsidian",
		Filename: filename,
		Data:     []byte(markdown),
		Message:  fmt.Sprintf("Recipe exported: %s", rec.Title()),
	}, nil
}

// ExportRecipes exports multiple recipes as a ZIP file
func (e *Exporter) ExportRecipes(recipes []*recipe.Recipe) (*ports.ExportResult, error) {
	if len(recipes) == 0 {
		return &ports.ExportResult{
			Success: false,
			Format:  "obsidian",
			Message: "No recipes to export",
		}, nil
	}

	// Create ZIP in memory
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, rec := range recipes {
		markdown := e.generateMarkdown(rec)
		filename := e.sanitizeFilename(rec.Title()) + ".md"

		writer, err := zipWriter.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create zip entry: %w", err)
		}

		if _, err := writer.Write([]byte(markdown)); err != nil {
			return nil, fmt.Errorf("failed to write zip entry: %w", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip: %w", err)
	}

	return &ports.ExportResult{
		Success:  true,
		Format:   "obsidian",
		Filename: fmt.Sprintf("recipes_%s.zip", time.Now().Format("2006-01-02")),
		Data:     buf.Bytes(),
		Message:  fmt.Sprintf("Exported %d recipes", len(recipes)),
	}, nil
}

// generateMarkdown generates Obsidian-compatible markdown with YAML frontmatter
func (e *Exporter) generateMarkdown(rec *recipe.Recipe) string {
	var sb strings.Builder

	// YAML Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: \"%s\"\n", escapeYAML(rec.Title())))
	sb.WriteString(fmt.Sprintf("category: %s\n", rec.Category()))

	if rec.Cuisine() != "" {
		sb.WriteString(fmt.Sprintf("cuisine: %s\n", rec.Cuisine()))
	}

	if len(rec.DietaryTags()) > 0 {
		sb.WriteString("dietary_tags:\n")
		for _, tag := range rec.DietaryTags() {
			sb.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}

	if len(rec.Tags()) > 0 {
		sb.WriteString("tags:\n")
		for _, tag := range rec.Tags() {
			sb.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}

	if rec.PrepTime() != nil {
		sb.WriteString(fmt.Sprintf("prep_time: %d\n", int(rec.PrepTime().Minutes())))
	}

	if rec.CookTime() != nil {
		sb.WriteString(fmt.Sprintf("cook_time: %d\n", int(rec.CookTime().Minutes())))
	}

	if rec.Servings() != nil {
		sb.WriteString(fmt.Sprintf("servings: %d\n", *rec.Servings()))
	}

	sb.WriteString(fmt.Sprintf("source_url: %s\n", rec.Source().URL()))

	if rec.Source().Platform() != "" {
		sb.WriteString(fmt.Sprintf("source_platform: %s\n", rec.Source().Platform()))
	}

	if rec.Source().Author() != "" {
		sb.WriteString(fmt.Sprintf("source_author: \"%s\"\n", escapeYAML(rec.Source().Author())))
	}

	sb.WriteString(fmt.Sprintf("created: %s\n", rec.CreatedAt().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("updated: %s\n", rec.UpdatedAt().Format("2006-01-02")))
	sb.WriteString("---\n\n")

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", rec.Title()))

	// Metadata summary
	var metaParts []string
	if rec.Category() != "" {
		metaParts = append(metaParts, fmt.Sprintf("**Category:** %s", rec.Category()))
	}
	if rec.Cuisine() != "" {
		metaParts = append(metaParts, fmt.Sprintf("**Cuisine:** %s", rec.Cuisine()))
	}
	if len(metaParts) > 0 {
		sb.WriteString(strings.Join(metaParts, " | ") + "\n\n")
	}

	// Time info
	var timeParts []string
	if rec.PrepTime() != nil {
		timeParts = append(timeParts, fmt.Sprintf("Prep: %d min", int(rec.PrepTime().Minutes())))
	}
	if rec.CookTime() != nil {
		timeParts = append(timeParts, fmt.Sprintf("Cook: %d min", int(rec.CookTime().Minutes())))
	}
	if rec.Servings() != nil {
		timeParts = append(timeParts, fmt.Sprintf("Servings: %d", *rec.Servings()))
	}
	if len(timeParts) > 0 {
		sb.WriteString("*" + strings.Join(timeParts, " | ") + "*\n\n")
	}

	// Dietary tags as badges
	if len(rec.DietaryTags()) > 0 {
		var tags []string
		for _, tag := range rec.DietaryTags() {
			tags = append(tags, fmt.Sprintf("`%s`", tag))
		}
		sb.WriteString(strings.Join(tags, " ") + "\n\n")
	}

	// Ingredients
	sb.WriteString("## Ingredients\n\n")
	for _, ing := range rec.Ingredients() {
		ingredient := formatIngredient(ing)
		sb.WriteString(fmt.Sprintf("- %s\n", ingredient))
	}
	sb.WriteString("\n")

	// Instructions
	sb.WriteString("## Instructions\n\n")
	for i, inst := range rec.Instructions() {
		stepNum := inst.StepNumber()
		if stepNum == 0 {
			stepNum = i + 1
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", stepNum, inst.Text()))
	}
	sb.WriteString("\n")

	// Source
	sb.WriteString("## Source\n\n")
	sb.WriteString(fmt.Sprintf("[Original Recipe](%s)", rec.Source().URL()))
	if rec.Source().Author() != "" {
		sb.WriteString(fmt.Sprintf(" by %s", rec.Source().Author()))
	}
	if rec.Source().Platform() != "" {
		sb.WriteString(fmt.Sprintf(" on %s", rec.Source().Platform()))
	}
	sb.WriteString("\n")

	return sb.String()
}

// formatIngredient formats an ingredient for display
func formatIngredient(ing recipe.Ingredient) string {
	var parts []string

	if ing.Quantity() != "" {
		parts = append(parts, ing.Quantity())
	}
	if ing.Unit() != "" {
		parts = append(parts, ing.Unit())
	}
	parts = append(parts, ing.Name())

	result := strings.Join(parts, " ")

	if ing.Notes() != "" {
		result += fmt.Sprintf(" (%s)", ing.Notes())
	}

	return result
}

// sanitizeFilename creates a safe filename from a recipe title
func (e *Exporter) sanitizeFilename(title string) string {
	// Remove or replace unsafe characters
	reg := regexp.MustCompile(`[<>:"/\\|?*]`)
	safe := reg.ReplaceAllString(title, "")

	// Replace spaces with underscores
	safe = strings.ReplaceAll(safe, " ", "_")

	// Limit length
	if len(safe) > 100 {
		safe = safe[:100]
	}

	// Trim leading/trailing underscores
	safe = strings.Trim(safe, "_")

	if safe == "" {
		safe = "recipe"
	}

	return safe
}

// escapeYAML escapes special characters in YAML string values
func escapeYAML(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
