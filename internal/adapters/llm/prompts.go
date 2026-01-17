package llm

import "fmt"

// SystemPrompt is the system prompt for recipe extraction
const SystemPrompt = `You are a recipe extraction assistant. Your task is to extract recipe information from video transcripts, captions, and web content.

You must respond with ONLY valid JSON in the following format:
{
  "title": "Recipe name",
  "ingredients": [
    {"name": "ingredient name", "quantity": "amount", "unit": "unit", "notes": "optional notes"}
  ],
  "instructions": [
    {"step_number": 1, "text": "instruction text", "duration_minutes": null}
  ],
  "prep_time_minutes": null,
  "cook_time_minutes": null,
  "servings": null
}

Rules:
- Extract ALL ingredients mentioned in the text, even if they're in different sections (e.g., "Ingredients (Dough)", "Ingredients (Filling)")
- Ingredients may be listed with bullets (-), numbers, or plain text - extract them all
- Parse ingredient lines that contain: quantity, unit, and name (e.g., "500g Self Rising Flour" or "1 Cup Diced White Onions")
- If an ingredient line has parentheses with additional info (e.g., "Greek Yogurt (0% Fat)"), put the parenthetical info in the "notes" field
- Preserve instruction order exactly as given
- Instructions may be numbered or use bullets - extract step numbers sequentially
- Include time estimates if mentioned (in minutes)
- Use null for missing information
- Standardize units (cups, tsp, tbsp, g, ml, oz, lb, etc.)
- If quantities are ranges (e.g., "2-3 cups"), use the average or pick one
- Keep instruction text concise but complete
- Extract prep time, cook time, and servings if mentioned
- If macros or nutritional info is present, ignore it - focus only on recipe ingredients and instructions
- IMPORTANT: Even if ingredients are split into multiple sections, extract ALL of them into a single ingredients array
- If the text contains a recipe, you MUST extract at least some ingredients - do not return an empty array unless there truly are no ingredients mentioned`

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
    "servings": {"type": ["integer", "null"]}
  },
  "required": ["title", "ingredients", "instructions"]
}`
