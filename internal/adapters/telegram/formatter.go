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
	if rec.PrepTime() != nil || rec.CookTime() != nil || rec.Servings() != nil {
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

		sb.WriteString("\n")
	}

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
	if rec.PrepTimeMinutes != nil || rec.CookTimeMinutes != nil || rec.Servings != nil {
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

		sb.WriteString("\n")
	}

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
	sb.WriteString(fmt.Sprintf("📚 *Your Recipes* (%d total)\n\n", len(recipes)))

	for i, rec := range recipes {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("\n... and %d more recipes", len(recipes)-10))
			break
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, escapeMarkdown(rec.Title())))
		sb.WriteString(fmt.Sprintf("   _From %s_\n", string(rec.Source().Platform())))
	}

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
• TikTok (tiktok.com)
• YouTube (youtube.com, youtu.be)
• Instagram (instagram.com)
• Recipe websites (with schema.org markup)

*How it works:*
1. Send me a recipe link
2. I'll download and transcribe the video
3. AI extracts ingredients & instructions
4. You get a formatted recipe!

*Tips:*
• Make sure the link contains a recipe
• Videos with clear audio work best
• Written recipes are also supported

*Commands:*
/start - Welcome message
/help - This help message
/recipes - Your saved recipes
/recipe <number> - View a specific recipe

*Having issues?*
Make sure:
• The link is valid
• The content contains a recipe
• The video has clear audio (if applicable)

Happy cooking! 🍳`
}
