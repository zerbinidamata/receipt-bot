package command

import (
	"context"
	"fmt"

	"receipt-bot/internal/domain/matching"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/ports"
)

// ProcessRecipeLinkCommand orchestrates the entire recipe extraction flow
type ProcessRecipeLinkCommand struct {
	scraper       ports.ScraperPort
	llm           ports.LLMPort
	recipeService *recipe.Service
	recipeRepo    recipe.Repository
	messenger     ports.MessengerPort
}

// NewProcessRecipeLinkCommand creates a new command
func NewProcessRecipeLinkCommand(
	scraper ports.ScraperPort,
	llm ports.LLMPort,
	recipeService *recipe.Service,
	recipeRepo recipe.Repository,
	messenger ports.MessengerPort,
) *ProcessRecipeLinkCommand {
	return &ProcessRecipeLinkCommand{
		scraper:       scraper,
		llm:           llm,
		recipeService: recipeService,
		recipeRepo:    recipeRepo,
		messenger:     messenger,
	}
}

// Execute processes a recipe link end-to-end
func (c *ProcessRecipeLinkCommand) Execute(ctx context.Context, url string, userID recipe.UserID, chatID int64) (*recipe.Recipe, error) {
	// Step 1: Send progress update
	if c.messenger != nil {
		_ = c.messenger.SendProgress(ctx, chatID, "🔍 Analyzing link...")
	}

	// Step 2: Detect platform
	platform := recipe.DetectPlatform(url)

	// Step 3: Check if recipe already exists for this URL
	existingRecipe, err := c.recipeRepo.FindBySourceURL(ctx, url)
	if err == nil && existingRecipe != nil {
		// Recipe already processed
		if c.messenger != nil {
			_ = c.messenger.SendProgress(ctx, chatID, "✅ Found existing recipe!")
		}
		return existingRecipe, nil
	}

	// Step 4: Scrape content from URL
	if c.messenger != nil {
		_ = c.messenger.SendProgress(ctx, chatID, "📥 Downloading content...")
	}

	scrapeResult, err := c.scraper.Scrape(ctx, ports.ScrapeRequest{
		URL:      url,
		Platform: platform,
	})
	if err != nil {
		return nil, fmt.Errorf("scraping failed: %w", err)
	}

	// Step 5: Merge text sources
	if c.messenger != nil {
		_ = c.messenger.SendProgress(ctx, chatID, "🎤 Processing audio...")
	}

	combinedText := c.recipeService.MergeTextSources(scrapeResult.Captions, scrapeResult.Transcript)
	if combinedText == "" {
		return nil, fmt.Errorf("no content extracted from URL")
	}

	// Log what we're sending to LLM (first 500 chars for debugging)
	textPreview := combinedText
	if len(textPreview) > 500 {
		textPreview = textPreview[:500] + "..."
	}
	fmt.Printf("[DEBUG] Sending to LLM (preview): %s\n", textPreview)
	fmt.Printf("[DEBUG] Captions length: %d, Transcript length: %d\n", len(scrapeResult.Captions), len(scrapeResult.Transcript))

	// Step 6: Extract recipe using LLM
	if c.messenger != nil {
		_ = c.messenger.SendProgress(ctx, chatID, "🤖 Extracting recipe...")
	}

	extraction, err := c.llm.ExtractRecipe(ctx, combinedText)
	if err != nil {
		return nil, fmt.Errorf("recipe extraction failed: %w", err)
	}

	// Log what we got back
	fmt.Printf("[DEBUG] LLM returned: %d ingredients, %d instructions, title: %s\n", 
		len(extraction.Ingredients), len(extraction.Instructions), extraction.Title)

	// Step 7: Validate extraction
	if len(extraction.Ingredients) == 0 {
		// Provide more context in the error
		return nil, fmt.Errorf("no ingredients found in content. Captions had %d chars, transcript had %d chars. LLM may have failed to parse the format", 
			len(scrapeResult.Captions), len(scrapeResult.Transcript))
	}
	if len(extraction.Instructions) == 0 {
		return nil, fmt.Errorf("no instructions found in content")
	}

	// Step 8: Build domain objects
	ingredients := make([]recipe.Ingredient, 0, len(extraction.Ingredients))
	for _, ingData := range extraction.Ingredients {
		ing, err := recipe.NewIngredient(ingData.Name, ingData.Quantity, ingData.Unit, ingData.Notes)
		if err != nil {
			continue // Skip invalid ingredients
		}
		ingredients = append(ingredients, ing)
	}

	instructions := make([]recipe.Instruction, 0, len(extraction.Instructions))
	for _, instData := range extraction.Instructions {
		inst, err := recipe.NewInstruction(instData.StepNumber, instData.Text, instData.Duration)
		if err != nil {
			continue // Skip invalid instructions
		}
		instructions = append(instructions, inst)
	}

	// Get author from metadata
	author := scrapeResult.Metadata["author"]
	if author == "" {
		author = "Unknown"
	}

	// Create source
	source, err := recipe.NewSource(url, platform, author)
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	// Step 9: Create recipe entity
	if c.messenger != nil {
		_ = c.messenger.SendProgress(ctx, chatID, "💾 Saving recipe...")
	}

	rec, err := recipe.NewRecipe(
		userID,
		extraction.Title,
		ingredients,
		instructions,
		source,
		scrapeResult.Transcript,
		scrapeResult.Captions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

	// Set optional fields
	if extraction.PrepTime != nil {
		rec.SetPrepTime(*extraction.PrepTime)
	}
	if extraction.CookTime != nil {
		rec.SetCookTime(*extraction.CookTime)
	}
	if extraction.Servings != nil {
		rec.SetServings(*extraction.Servings)
	}

	// Set category fields from LLM extraction
	if extraction.Category != "" {
		rec.SetCategory(recipe.CategoryFromLLM(extraction.Category))
	}
	if extraction.Cuisine != "" {
		rec.SetCuisine(extraction.Cuisine)
	}
	if len(extraction.DietaryTags) > 0 {
		dietaryTags := recipe.ParseDietaryTags(extraction.DietaryTags)
		rec.SetDietaryTags(dietaryTags)
	}
	if len(extraction.Tags) > 0 {
		rec.SetTags(extraction.Tags)
	}

	// Set multilingual fields from LLM extraction
	if extraction.SourceLanguage != "" {
		rec.SetSourceLanguage(extraction.SourceLanguage)
	}
	if extraction.TranslatedTitle != nil || len(extraction.TranslatedIngredients) > 0 || len(extraction.TranslatedInstructions) > 0 {
		// Convert translated ingredients
		var translatedIngs []recipe.Ingredient
		if len(extraction.TranslatedIngredients) > 0 {
			translatedIngs = make([]recipe.Ingredient, 0, len(extraction.TranslatedIngredients))
			for _, ingData := range extraction.TranslatedIngredients {
				ing, err := recipe.NewIngredient(ingData.Name, ingData.Quantity, ingData.Unit, ingData.Notes)
				if err == nil {
					translatedIngs = append(translatedIngs, ing)
				}
			}
		}
		// Convert translated instructions
		var translatedInsts []recipe.Instruction
		if len(extraction.TranslatedInstructions) > 0 {
			translatedInsts = make([]recipe.Instruction, 0, len(extraction.TranslatedInstructions))
			for _, instData := range extraction.TranslatedInstructions {
				inst, err := recipe.NewInstruction(instData.StepNumber, instData.Text, instData.Duration)
				if err == nil {
					translatedInsts = append(translatedInsts, inst)
				}
			}
		}
		rec.SetTranslations(extraction.TranslatedTitle, translatedIngs, translatedInsts)
	}

	// Step 10: Normalize and cache ingredients for faster matching
	normalizer := matching.NewRuleBasedNormalizer()
	normalizedIngredients := make([]string, 0, len(ingredients))
	for _, ing := range ingredients {
		normalized := normalizer.Normalize(ing.Name())
		if normalized != "" {
			normalizedIngredients = append(normalizedIngredients, normalized)
		}
	}
	rec.SetNormalizedIngredients(normalizedIngredients)

	// Step 11: Validate recipe
	if err := c.recipeService.ValidateRecipe(rec); err != nil {
		return nil, fmt.Errorf("recipe validation failed: %w", err)
	}

	// Step 13: Save recipe
	if err := c.recipeRepo.Save(ctx, rec); err != nil {
		return nil, fmt.Errorf("failed to save recipe: %w", err)
	}

	// Step 14: Success!
	if c.messenger != nil {
		_ = c.messenger.SendProgress(ctx, chatID, "✨ Recipe extracted successfully!")
	}

	return rec, nil
}
