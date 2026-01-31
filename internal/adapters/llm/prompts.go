package llm

import "fmt"

// SystemPrompt is the system prompt for recipe extraction (English output)
const SystemPrompt = `You are a recipe extraction assistant. Your task is to extract recipe information from video transcripts, captions, and web content, and categorize the recipe.

IMPORTANT: The input may be in ANY language (English, Portuguese, Spanish, etc.). You MUST:
1. Detect the source language of the content
2. Extract the recipe in the ORIGINAL language first
3. Then provide an English translation in the translation fields

This ensures users see recipes in both the original language and English.

You must respond with ONLY valid JSON in the following format:
{
  "title": "Recipe name in ORIGINAL language",
  "category": "Category name (always in English)",
  "cuisine": "Cuisine type",
  "dietary_tags": ["tag1", "tag2"],
  "tags": ["descriptive", "tags"],
  "ingredients": [
    {"name": "ingredient name in ORIGINAL language", "quantity": "amount", "unit": "unit", "notes": "optional notes"}
  ],
  "instructions": [
    {"step_number": 1, "text": "instruction text in ORIGINAL language", "duration_minutes": null}
  ],
  "prep_time_minutes": null,
  "cook_time_minutes": null,
  "servings": null,
  "source_language": "detected language code (en, pt, es, etc.)",
  "translated_title": "Recipe name in English (null if source is English)",
  "translated_ingredients": [
    {"name": "ingredient name in English", "quantity": "amount", "unit": "unit", "notes": "optional notes in English"}
  ],
  "translated_instructions": [
    {"step_number": 1, "text": "instruction text in English", "duration_minutes": null}
  ]
}

CATEGORIES (choose exactly one):
- Pasta & Noodles (pasta, noodles, lasagna, ramen, etc.)
- Rice & Grains (rice dishes, quinoa, couscous, risotto, etc.)
- Soups & Stews (soups, stews, chili, broths, etc.)
- Salads (fresh salads, grain salads, etc.)
- Meat & Poultry (beef, pork, chicken, turkey dishes where meat is the focus)
- Seafood (fish, shrimp, shellfish dishes)
- Vegetarian (meatless mains, veggie dishes)
- Desserts & Sweets (cakes, cookies, ice cream, sweet treats)
- Breakfast (morning dishes, brunch items)
- Appetizers & Snacks (small plates, finger foods, dips)
- Beverages (drinks, smoothies, cocktails)
- Sauces & Condiments (sauces, dressings, marinades)
- Bread & Baking (breads, pastries, non-sweet baked goods)
- Other (anything that doesn't fit above)

DIETARY TAGS (choose all that apply):
- vegetarian (no meat or fish)
- vegan (no animal products at all)
- gluten-free (no wheat, barley, rye)
- dairy-free (no milk, cheese, butter, cream)
- low-carb (minimal carbohydrates)
- quick (total prep + cook time under 30 minutes)
- one-pot (cooked in single pot/pan)
- kid-friendly (simple flavors, kid-approved)

CUISINE (identify if applicable):
Italian, Mexican, Chinese, Japanese, Indian, Thai, French, Greek, Mediterranean, American, Korean, Vietnamese, Middle Eastern, etc.

Rules for extraction:
- Extract ALL ingredients mentioned in the text, even if they're in different sections
- Ingredients may be listed with bullets (-), numbers, or plain text - extract them all
- Parse ingredient lines that contain: quantity, unit, and name (e.g., "500g Self Rising Flour")
- If an ingredient line has parentheses with additional info, put it in the "notes" field
- Preserve instruction order exactly as given
- Instructions may be numbered or use bullets - extract step numbers sequentially
- Include time estimates if mentioned (in minutes)
- Use null for missing information
- Standardize units (cups, tsp, tbsp, g, ml, oz, lb, etc.)
- If quantities are ranges (e.g., "2-3 cups"), use the average or pick one
- Keep instruction text concise but complete
- Extract prep time, cook time, and servings if mentioned
- For category: Choose the BEST matching category based on the main dish type
- For cuisine: Identify the cuisine style if evident from ingredients/techniques
- For dietary_tags: Only include tags that definitely apply based on ingredients
- For tags: Add 2-4 descriptive tags (e.g., "comfort-food", "weeknight-dinner", "meal-prep")
- If the text contains a recipe, you MUST extract at least some ingredients

MULTILINGUAL RULES:
- source_language: Use ISO 639-1 codes (en, pt, es, fr, de, it, etc.)
- If source is English: Set translated_title, translated_ingredients, translated_instructions to null
- If source is NOT English: Provide English translations in the translated_* fields
- Keep original language content in the main fields (title, ingredients, instructions)
- Category and dietary_tags should ALWAYS be in English
- Units should be standardized (cups, tsp, tbsp, g, ml, oz, lb, etc.)`

// BuildUserPrompt builds the user prompt with the provided text
func BuildUserPrompt(combinedText string) string {
	return fmt.Sprintf(`Extract the recipe from this text:

---
%s
---

Remember to respond with ONLY the JSON object, no additional text.`, combinedText)
}

// RecipeJSONSchema is the JSON schema for structured output (for providers that support it)
const RecipeJSONSchema = `{
  "type": "object",
  "properties": {
    "title": {"type": "string"},
    "category": {"type": "string"},
    "cuisine": {"type": "string"},
    "dietary_tags": {
      "type": "array",
      "items": {"type": "string"}
    },
    "tags": {
      "type": "array",
      "items": {"type": "string"}
    },
    "ingredients": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "quantity": {"type": "string"},
          "unit": {"type": "string"},
          "notes": {"type": "string"}
        },
        "required": ["name", "quantity"]
      }
    },
    "instructions": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "step_number": {"type": "integer"},
          "text": {"type": "string"},
          "duration_minutes": {"type": ["integer", "null"]}
        },
        "required": ["step_number", "text"]
      }
    },
    "prep_time_minutes": {"type": ["integer", "null"]},
    "cook_time_minutes": {"type": ["integer", "null"]},
    "servings": {"type": ["integer", "null"]},
    "source_language": {"type": "string"},
    "translated_title": {"type": ["string", "null"]},
    "translated_ingredients": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "name": {"type": "string"},
          "quantity": {"type": "string"},
          "unit": {"type": "string"},
          "notes": {"type": "string"}
        },
        "required": ["name", "quantity"]
      }
    },
    "translated_instructions": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "step_number": {"type": "integer"},
          "text": {"type": "string"},
          "duration_minutes": {"type": ["integer", "null"]}
        },
        "required": ["step_number", "text"]
      }
    }
  },
  "required": ["title", "category", "ingredients", "instructions", "source_language"]
}`
