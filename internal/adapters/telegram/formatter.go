package telegram

import (
	"fmt"
	"strings"

	"receipt-bot/internal/application/dto"
	"receipt-bot/internal/domain/recipe"
)

// FormatRecipe formats a recipe for Telegram display
func FormatRecipe(rec *recipe.Recipe) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("🍳 *%s*\n\n", escapeMarkdown(rec.Title())))

	// Metadata
	sb.WriteString("📊 *Info*\n")

	if rec.PrepTime() != nil {
		sb.WriteString(fmt.Sprintf("⏱️ Prep: %d min\n", int(rec.PrepTime().Minutes())))
	}

	if rec.CookTime() != nil {
		sb.WriteString(fmt.Sprintf("🔥 Cook: %d min\n", int(rec.CookTime().Minutes())))
	}

	if rec.Servings() != nil {
		sb.WriteString(fmt.Sprintf("🍽️ Servings: %d\n", *rec.Servings()))
	}

	// Category info
	sb.WriteString(fmt.Sprintf("📁 Category: %s\n", escapeMarkdown(string(rec.Category()))))

	if rec.Cuisine() != "" {
		sb.WriteString(fmt.Sprintf("🌍 Cuisine: %s\n", escapeMarkdown(rec.Cuisine())))
	}

	if len(rec.DietaryTags()) > 0 {
		tags := make([]string, len(rec.DietaryTags()))
		for i, t := range rec.DietaryTags() {
			tags[i] = "#" + string(t)
		}
		sb.WriteString(fmt.Sprintf("🏷️ Tags: %s\n", escapeMarkdown(strings.Join(tags, " "))))
	}

	sb.WriteString("\n")

	// Ingredients
	sb.WriteString("📝 *Ingredients*\n")
	for _, ing := range rec.Ingredients() {
		sb.WriteString(fmt.Sprintf("• %s\n", escapeMarkdown(ing.String())))
	}
	sb.WriteString("\n")

	// Instructions
	sb.WriteString("👨‍🍳 *Instructions*\n")
	for _, inst := range rec.Instructions() {
		sb.WriteString(fmt.Sprintf("%s\n", escapeMarkdown(inst.String())))
	}
	sb.WriteString("\n")

	// Source
	sb.WriteString("🔗 *Source*\n")
	sb.WriteString(fmt.Sprintf("[%s](%s)\n",
		escapeMarkdown(string(rec.Source().Platform())),
		rec.Source().URL()))

	if rec.Source().Author() != "" {
		sb.WriteString(fmt.Sprintf("By: %s\n", escapeMarkdown(rec.Source().Author())))
	}

	return sb.String()
}

// FormatRecipeDTO formats a recipe DTO for Telegram display
func FormatRecipeDTO(rec *dto.RecipeDTO) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("🍳 *%s*\n\n", escapeMarkdown(rec.Title)))

	// Metadata
	sb.WriteString("📊 *Info*\n")

	if rec.PrepTimeMinutes != nil {
		sb.WriteString(fmt.Sprintf("⏱️ Prep: %d min\n", *rec.PrepTimeMinutes))
	}

	if rec.CookTimeMinutes != nil {
		sb.WriteString(fmt.Sprintf("🔥 Cook: %d min\n", *rec.CookTimeMinutes))
	}

	if rec.Servings != nil {
		sb.WriteString(fmt.Sprintf("🍽️ Servings: %d\n", *rec.Servings))
	}

	// Category info
	if rec.Category != "" {
		sb.WriteString(fmt.Sprintf("📁 Category: %s\n", escapeMarkdown(rec.Category)))
	}

	if rec.Cuisine != "" {
		sb.WriteString(fmt.Sprintf("🌍 Cuisine: %s\n", escapeMarkdown(rec.Cuisine)))
	}

	if len(rec.DietaryTags) > 0 {
		tags := make([]string, len(rec.DietaryTags))
		for i, t := range rec.DietaryTags {
			tags[i] = "#" + t
		}
		sb.WriteString(fmt.Sprintf("🏷️ Tags: %s\n", escapeMarkdown(strings.Join(tags, " "))))
	}

	sb.WriteString("\n")

	// Ingredients
	sb.WriteString("📝 *Ingredients*\n")
	for _, ing := range rec.Ingredients {
		ingStr := ing.Name
		if ing.Quantity != "" {
			ingStr = ing.Quantity + " " + ing.Unit + " " + ing.Name
		}
		if ing.Notes != "" {
			ingStr += " (" + ing.Notes + ")"
		}
		sb.WriteString(fmt.Sprintf("• %s\n", escapeMarkdown(ingStr)))
	}
	sb.WriteString("\n")

	// Instructions
	sb.WriteString("👨‍🍳 *Instructions*\n")
	for _, inst := range rec.Instructions {
		sb.WriteString(fmt.Sprintf("%d\\. %s\n", inst.StepNumber, escapeMarkdown(inst.Text)))
	}
	sb.WriteString("\n")

	// Source
	sb.WriteString("🔗 *Source*\n")
	sb.WriteString(fmt.Sprintf("[%s](%s)\n",
		escapeMarkdown(rec.SourcePlatform),
		rec.SourceURL))

	if rec.SourceAuthor != "" {
		sb.WriteString(fmt.Sprintf("By: %s\n", escapeMarkdown(rec.SourceAuthor)))
	}

	return sb.String()
}

// FormatRecipeList formats a list of recipes for Telegram display
func FormatRecipeList(recipes []recipe.Recipe) string {
	if len(recipes) == 0 {
		return "📭 You don't have any saved recipes yet.\n\nSend me a link to a recipe video or webpage to get started!"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📚 *Your Recipes* \\(%d total\\)\n\n", len(recipes)))

	for i, rec := range recipes {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("\n\\.\\.\\. and %d more recipes", len(recipes)-10))
			break
		}

		sb.WriteString(fmt.Sprintf("%d\\. %s\n", i+1, escapeMarkdown(rec.Title())))
		sb.WriteString(fmt.Sprintf("   _%s_ \\| %s\n", escapeMarkdown(string(rec.Category())), string(rec.Source().Platform())))
	}

	sb.WriteString("\nUse /recipe <number> to view details")
	sb.WriteString("\nUse /recipes <category> to filter")

	return sb.String()
}

// FormatRecipeListDTO formats a list of recipe DTOs for Telegram display
func FormatRecipeListDTO(recipes []*dto.RecipeDTO) string {
	if len(recipes) == 0 {
		return "📭 No recipes found\\."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📚 *Recipes* \\(%d found\\)\n\n", len(recipes)))

	for i, rec := range recipes {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("\n\\.\\.\\. and %d more recipes", len(recipes)-10))
			break
		}

		sb.WriteString(fmt.Sprintf("%d\\. %s\n", i+1, escapeMarkdown(rec.Title)))
		sb.WriteString(fmt.Sprintf("   _%s_ \\| %s\n", escapeMarkdown(rec.Category), rec.SourcePlatform))
	}

	sb.WriteString("\nUse /recipe <number> to view details")

	return sb.String()
}

// FormatCategories formats category counts for Telegram display
func FormatCategories(counts map[string]int, total int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *Recipe Categories* \\(%d total\\)\n\n", total))

	// Category emoji mapping
	categoryEmoji := map[string]string{
		"Pasta & Noodles":      "🍝",
		"Rice & Grains":        "🍚",
		"Soups & Stews":        "🥘",
		"Salads":               "🥗",
		"Meat & Poultry":       "🍖",
		"Seafood":              "🐟",
		"Vegetarian":           "🥬",
		"Desserts & Sweets":    "🍰",
		"Breakfast":            "🍳",
		"Appetizers & Snacks":  "🍿",
		"Beverages":            "🥤",
		"Sauces & Condiments":  "🫙",
		"Bread & Baking":       "🍞",
		"Other":                "📦",
	}

	// Sort categories by count (descending)
	type catCount struct {
		name  string
		count int
	}
	var sorted []catCount
	for cat, count := range counts {
		sorted = append(sorted, catCount{cat, count})
	}
	// Simple bubble sort since we have max 14 categories
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for _, cc := range sorted {
		emoji := categoryEmoji[cc.name]
		if emoji == "" {
			emoji = "📁"
		}
		sb.WriteString(fmt.Sprintf("%s %s: %d recipes\n", emoji, escapeMarkdown(cc.name), cc.count))
	}

	sb.WriteString("\nUse /recipes <category> to filter\n")
	sb.WriteString("Example: /recipes pasta")

	return sb.String()
}

// escapeMarkdown escapes special characters for Telegram Markdown
func escapeMarkdown(text string) string {
	// Escape special Markdown characters
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

// FormatWelcome formats the welcome message
func FormatWelcome() string {
	return `👋 *Welcome to Recipe Bot!*

I can help you extract recipes from:
• 🎵 TikTok videos
• 📺 YouTube videos
• 📸 Instagram posts/reels
• 🌐 Recipe websites

*How to use:*
Just send me a link to any recipe video or webpage, and I'll extract the ingredients and cooking instructions for you!

*Commands:*
/start - Show this message
/help - Get help
/recipes - List your saved recipes
/recipe <number> - View a specific recipe

Let's get cooking! 👨‍🍳`
}

// FormatHelp formats the help message
func FormatHelp() string {
	return `🤖 *Recipe Bot Help*

*Supported Platforms:*
• TikTok \(tiktok\.com\)
• YouTube \(youtube\.com, youtu\.be\)
• Instagram \(instagram\.com\)
• Recipe websites \(with schema\.org markup\)

*How it works:*
1\. Send me a recipe link
2\. I'll download and transcribe the video
3\. AI extracts ingredients \& instructions
4\. You get a formatted recipe\!

*Tips:*
• Make sure the link contains a recipe
• Videos with clear audio work best
• Written recipes are also supported

*Commands:*
/start \- Welcome message
/help \- This help message
/recipes \- Your saved recipes
/recipes <category> \- Filter by category
/recipe <number> \- View a specific recipe
/categories \- Show recipe categories
/match <ingredients> \- Find recipes by ingredients
/pantry \- Manage your pantry items

*Having issues?*
Make sure:
• The link is valid
• The content contains a recipe
• The video has clear audio \(if applicable\)

Happy cooking\! 🍳`
}

// FormatMatchResults formats ingredient match results for Telegram display
func FormatMatchResults(result *dto.MatchIngredientsResultDTO) string {
	var sb strings.Builder

	if result.TotalMatches == 0 {
		return "📭 No matching recipes found\\.\n\nTry adding more ingredients or use /recipes to see all your recipes\\."
	}

	sb.WriteString("🍳 *Here's what you can make:*\n\n")

	// Perfect matches
	if len(result.PerfectMatches) > 0 {
		sb.WriteString(fmt.Sprintf("✅ *Perfect Matches* \\(%d recipes\\):\n", len(result.PerfectMatches)))
		for i, match := range result.PerfectMatches {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("   \\.\\.\\. and %d more\n", len(result.PerfectMatches)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("%d\\. %s\n", i+1, escapeMarkdown(match.Recipe.Title)))
		}
		sb.WriteString("\n")
	}

	// High matches (missing 1-2 items)
	if len(result.HighMatches) > 0 {
		sb.WriteString(fmt.Sprintf("🔸 *Almost There* \\(%d recipes\\):\n", len(result.HighMatches)))
		startIndex := len(result.PerfectMatches)
		for i, match := range result.HighMatches {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("   \\.\\.\\. and %d more\n", len(result.HighMatches)-5))
				break
			}
			missing := formatMissingItems(match.MissingItems, 3)
			sb.WriteString(fmt.Sprintf("%d\\. %s\n", startIndex+i+1, escapeMarkdown(match.Recipe.Title)))
			sb.WriteString(fmt.Sprintf("   _Missing: %s_\n", escapeMarkdown(missing)))
		}
		sb.WriteString("\n")
	}

	// Medium matches
	if len(result.MediumMatches) > 0 {
		sb.WriteString(fmt.Sprintf("🔹 *Partial Matches* \\(%d recipes\\):\n", len(result.MediumMatches)))
		startIndex := len(result.PerfectMatches) + len(result.HighMatches)
		for i, match := range result.MediumMatches {
			if i >= 3 {
				sb.WriteString(fmt.Sprintf("   \\.\\.\\. and %d more\n", len(result.MediumMatches)-3))
				break
			}
			missing := formatMissingItems(match.MissingItems, 3)
			sb.WriteString(fmt.Sprintf("%d\\. %s \\(%.0f%% match\\)\n", startIndex+i+1, escapeMarkdown(match.Recipe.Title), match.MatchPercentage))
			sb.WriteString(fmt.Sprintf("   _Missing: %s_\n", escapeMarkdown(missing)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Use /recipe <number> to view full recipe\\!")

	return sb.String()
}

// formatMissingItems formats missing items list
func formatMissingItems(items []string, max int) string {
	if len(items) == 0 {
		return "none"
	}

	if len(items) <= max {
		return strings.Join(items, ", ")
	}

	shown := items[:max]
	remaining := len(items) - max
	return fmt.Sprintf("%s +%d more", strings.Join(shown, ", "), remaining)
}

// FormatPantry formats pantry items for Telegram display
func FormatPantry(items []string) string {
	if len(items) == 0 {
		return "📭 Your pantry is empty\\.\n\nUse /pantry add <items> to add ingredients\\.\nExample: /pantry add butter, eggs, milk"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🥫 *Your Pantry* \\(%d items\\)\n\n", len(items)))

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("• %s\n", escapeMarkdown(item)))
	}

	sb.WriteString("\n*Commands:*\n")
	sb.WriteString("/pantry add <items> \\- Add items\n")
	sb.WriteString("/pantry remove <items> \\- Remove items\n")
	sb.WriteString("/pantry clear \\- Clear all items\n")
	sb.WriteString("/match \\- Find recipes with pantry items")

	return sb.String()
}
