# API & Interface Specification

## Telegram Bot Commands

### Existing Commands (Unchanged)

| Command | Description | Example |
|---------|-------------|---------|
| `/start` | Welcome message and introduction | `/start` |
| `/help` | Show available commands | `/help` |
| `/recipes` | List all saved recipes | `/recipes` |
| `/recipe <n>` | Show specific recipe details | `/recipe 3` |

### New Commands

#### Category & Filtering Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/recipes <category>` | Filter recipes by category | `/recipes pasta` |
| `/recipes --tag <tag>` | Filter recipes by dietary tag | `/recipes --tag vegetarian` |
| `/categories` | Show all categories with recipe counts | `/categories` |
| `/recategorize <n>` | Re-run categorization for a recipe | `/recategorize 5` |

**Category Shortcuts:**

| Shortcut | Full Category |
|----------|---------------|
| `pasta` | Pasta & Noodles |
| `rice` | Rice & Grains |
| `soup`, `soups` | Soups & Stews |
| `salad`, `salads` | Salads |
| `meat` | Meat & Poultry |
| `seafood`, `fish` | Seafood |
| `vegetarian`, `veggie` | Vegetarian |
| `dessert`, `desserts`, `sweet` | Desserts & Sweets |
| `breakfast` | Breakfast |
| `appetizer`, `snack` | Appetizers & Snacks |
| `drinks`, `beverage` | Beverages |
| `sauce`, `sauces` | Sauces & Condiments |
| `bread`, `baking` | Bread & Baking |

**Tag Shortcuts:**

| Shortcut | Full Tag |
|----------|----------|
| `veg`, `vegetarian` | vegetarian |
| `vegan` | vegan |
| `gf`, `gluten-free` | gluten-free |
| `df`, `dairy-free` | dairy-free |
| `lc`, `low-carb`, `keto` | low-carb |
| `quick`, `fast` | quick |

#### Ingredient Matching Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/match <ingredients>` | Find recipes matching ingredients | `/match chicken, pasta, garlic, tomato` |
| `/match --strict <ingredients>` | Only show perfect matches | `/match --strict eggs, flour, milk` |
| `/match --category <cat> <ingredients>` | Filter matches by category | `/match --category dessert butter, sugar, eggs` |
| `/pantry` | Show saved pantry items | `/pantry` |
| `/pantry add <items>` | Add items to pantry | `/pantry add butter, eggs, milk, cheese` |
| `/pantry remove <items>` | Remove items from pantry | `/pantry remove milk` |
| `/pantry clear` | Clear all pantry items | `/pantry clear` |

#### Export Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/export obsidian <n>` | Export single recipe as markdown | `/export obsidian 3` |
| `/export obsidian all` | Export all recipes as ZIP | `/export obsidian all` |
| `/connect notion` | Start Notion OAuth connection | `/connect notion` |
| `/disconnect notion` | Remove Notion integration | `/disconnect notion` |
| `/export notion <n>` | Export single recipe to Notion | `/export notion 3` |
| `/export notion all` | Export all recipes to Notion | `/export notion all` |

---

## Response Formats

### Recipe Display (Updated)

```
🍳 Chicken Tikka Masala

📊 Info
⏱️ Prep: 20 min | 🔥 Cook: 35 min
🍽️ Servings: 4
📁 Category: Meat & Poultry
🌍 Cuisine: Indian
🏷️ Tags: #spicy #curry

📝 Ingredients
• 500g chicken breast, cubed
• 1 cup yogurt
• 2 tbsp tikka masala paste
• 1 can coconut milk
• ...

👨‍🍳 Instructions
1. Marinate chicken in yogurt and spices for 30 minutes
2. Cook chicken until browned
3. Add coconut milk and simmer
...

🔗 Source
TikTok by @indianfoodie
https://tiktok.com/...
```

### Recipe List (Updated)

```
📚 Your Recipes

🍝 Pasta & Noodles (3)
1. Homemade Lasagna
2. Chicken Alfredo
3. Spaghetti Carbonara

🥘 Soups & Stews (2)
4. Tomato Basil Soup
5. Chicken Noodle Soup

🍖 Meat & Poultry (1)
6. Chicken Tikka Masala

Reply with /recipe <number> to see details
```

### Categories Display

```
📊 Recipe Categories

🍝 Pasta & Noodles: 5 recipes
🍚 Rice & Grains: 3 recipes
🥘 Soups & Stews: 2 recipes
🥗 Salads: 1 recipe
🍖 Meat & Poultry: 4 recipes
🐟 Seafood: 2 recipes
🥬 Vegetarian: 3 recipes
🍰 Desserts & Sweets: 2 recipes
🍳 Breakfast: 1 recipe

Total: 23 recipes

Use /recipes <category> to filter
Example: /recipes pasta
```

### Match Results Display

```
🍳 What You Can Make

Your ingredients: chicken, pasta, garlic, tomato, onion, cheese

✅ Perfect Matches (3)
1. Chicken Pasta Primavera
   🏷️ Pasta & Noodles | ⏱️ 25 min
2. One-Pot Chicken Alfredo
   🏷️ Pasta & Noodles | ⏱️ 30 min
3. Tomato Chicken Penne
   🏷️ Pasta & Noodles | ⏱️ 20 min

🔸 Almost There - Missing 1-2 items (2)
4. Chicken Parmesan
   🏷️ Meat & Poultry | ⏱️ 45 min
   Missing: breadcrumbs, egg
5. Tuscan Chicken Pasta
   🏷️ Pasta & Noodles | ⏱️ 35 min
   Missing: cream, spinach

Reply with number to see full recipe!
```

### Pantry Display

```
🧺 Your Pantry

Saved items (12):
butter, eggs, milk, cheese, flour, sugar,
olive oil, garlic, onion, salt, pepper, rice

These items are used to improve recipe matching.

Commands:
• /pantry add <items> - Add items
• /pantry remove <items> - Remove items
• /pantry clear - Clear all
```

### Export Success Messages

**Obsidian Single:**
```
✅ Recipe exported!

📄 File: chicken-tikka-masala.md

Save this file to your Obsidian vault.
```
(File attachment follows)

**Obsidian Bulk:**
```
✅ All recipes exported!

📦 File: recipes-export-2026-01-23.zip
📊 Contains: 23 recipes

Extract to your Obsidian vault folder.
```
(ZIP attachment follows)

**Notion Single:**
```
✅ Recipe exported to Notion!

📄 Chicken Tikka Masala
🔗 https://notion.so/...

Open in Notion to view.
```

**Notion Bulk:**
```
✅ All recipes exported to Notion!

📊 23 recipes added to your Recipes database
🔗 https://notion.so/...

Open database in Notion to browse.
```

### Error Messages

| Scenario | Message |
|----------|---------|
| Invalid category | "❌ Unknown category '{input}'. Use /categories to see available options." |
| No recipes in category | "📭 No recipes found in {category}. Try adding some recipes first!" |
| Invalid recipe number | "❌ Recipe #{n} not found. Use /recipes to see your recipes." |
| No matching recipes | "😕 No recipes match your ingredients. Try adding more items or use /recipes to browse." |
| Notion not connected | "🔗 Please connect Notion first: /connect notion" |
| Notion token expired | "⚠️ Your Notion connection expired. Please reconnect: /connect notion" |
| Export failed | "❌ Export failed: {reason}. Please try again later." |
| Invalid ingredients | "❌ Please provide ingredients separated by commas.\nExample: /match chicken, pasta, garlic" |

---

## LLM Extraction Schema

### Request (Updated Prompt)

```
Extract the recipe from the following content and categorize it.

CATEGORIES (choose exactly one):
- Pasta & Noodles
- Rice & Grains
- Soups & Stews
- Salads
- Meat & Poultry
- Seafood
- Vegetarian
- Desserts & Sweets
- Breakfast
- Appetizers & Snacks
- Beverages
- Sauces & Condiments
- Bread & Baking
- Other

DIETARY TAGS (choose all that apply):
- vegetarian (no meat or fish)
- vegan (no animal products)
- gluten-free (no wheat, barley, rye)
- dairy-free (no milk, cheese, butter)
- low-carb (minimal carbohydrates)
- quick (total time under 30 minutes)

CUISINE (identify if applicable):
Italian, Mexican, Chinese, Japanese, Indian, Thai, French, Greek, Mediterranean, American, Korean, Vietnamese, Middle Eastern, etc.

Content:
{transcript}
{captions}
{web_content}

Return JSON:
```

### Response Schema

```json
{
  "title": "string (required)",
  "category": "string (required, from CATEGORIES list)",
  "cuisine": "string (optional)",
  "dietaryTags": ["string"] ,
  "tags": ["string"],
  "ingredients": [
    {
      "name": "string (required)",
      "quantity": "string (required)",
      "unit": "string (optional)",
      "notes": "string (optional)"
    }
  ],
  "instructions": [
    {
      "stepNumber": "integer (required, starts at 1)",
      "text": "string (required)",
      "durationMinutes": "integer (optional)"
    }
  ],
  "prepTimeMinutes": "integer (optional)",
  "cookTimeMinutes": "integer (optional)",
  "servings": "integer (optional)",
  "author": "string (optional)"
}
```

### Example Response

```json
{
  "title": "Homemade Lasagna",
  "category": "Pasta & Noodles",
  "cuisine": "Italian",
  "dietaryTags": [],
  "tags": ["comfort-food", "bake", "family-meal"],
  "ingredients": [
    {"name": "lasagna noodles", "quantity": "12", "unit": "sheets"},
    {"name": "ground beef", "quantity": "1", "unit": "lb"},
    {"name": "ricotta cheese", "quantity": "2", "unit": "cups"},
    {"name": "mozzarella cheese", "quantity": "2", "unit": "cups", "notes": "shredded"},
    {"name": "marinara sauce", "quantity": "24", "unit": "oz"},
    {"name": "egg", "quantity": "1"},
    {"name": "Italian seasoning", "quantity": "1", "unit": "tsp"}
  ],
  "instructions": [
    {"stepNumber": 1, "text": "Preheat oven to 375°F (190°C)"},
    {"stepNumber": 2, "text": "Cook lasagna noodles according to package directions", "durationMinutes": 10},
    {"stepNumber": 3, "text": "Brown ground beef in a skillet, drain fat", "durationMinutes": 8},
    {"stepNumber": 4, "text": "Mix ricotta, egg, and half the mozzarella in a bowl"},
    {"stepNumber": 5, "text": "Layer: sauce, noodles, meat, cheese mixture. Repeat 3 times"},
    {"stepNumber": 6, "text": "Top with remaining mozzarella"},
    {"stepNumber": 7, "text": "Cover with foil and bake for 25 minutes", "durationMinutes": 25},
    {"stepNumber": 8, "text": "Remove foil and bake until bubbly, about 15 more minutes", "durationMinutes": 15}
  ],
  "prepTimeMinutes": 20,
  "cookTimeMinutes": 50,
  "servings": 8,
  "author": "@italianfoodlover"
}
```

---

## Firestore Schema

### recipes Collection

```
{
  // Existing fields
  recipeId: string,
  userId: string,
  title: string,
  ingredients: [
    {name: string, quantity: string, unit: string, notes: string}
  ],
  instructions: [
    {stepNumber: int, text: string, durationMinutes: int}
  ],
  source: {url: string, platform: string, author: string},
  transcript: string,
  captions: string,
  prepTimeMinutes: int,
  cookTimeMinutes: int,
  servings: int,
  createdAt: timestamp,
  updatedAt: timestamp,

  // NEW fields
  category: string,          // "Pasta & Noodles"
  cuisine: string,           // "Italian"
  dietaryTags: [string],     // ["vegetarian", "quick"]
  tags: [string]             // ["comfort-food", "bake"]
}
```

### users Collection

```
{
  // Existing fields
  userId: string,
  telegramId: int64,
  username: string,
  createdAt: timestamp,

  // NEW fields
  pantryItems: [string],           // ["butter", "eggs", "milk"]
  pantryUpdatedAt: timestamp,
  notionAccessToken: string,       // encrypted
  notionWorkspaceId: string,
  notionDatabaseId: string,
  notionConnectedAt: timestamp
}
```

### Firestore Indexes Required

```
Collection: recipes
Fields: userId (ASC), category (ASC)
Query scope: Collection

Collection: recipes
Fields: userId (ASC), dietaryTags (ARRAY_CONTAINS)
Query scope: Collection

Collection: recipes
Fields: userId (ASC), createdAt (DESC)
Query scope: Collection
```

---

## gRPC Service (Unchanged)

The Python scraper service interface remains unchanged. Recipe categorization is handled by the Go service's LLM adapter after content is scraped.

```protobuf
// proto/scraper.proto (unchanged)
service ScraperService {
  rpc ScrapeURL(ScrapeRequest) returns (ScrapeResponse);
  rpc Health(HealthRequest) returns (HealthResponse);
}
```
