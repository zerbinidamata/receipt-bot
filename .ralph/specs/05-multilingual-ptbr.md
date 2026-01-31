# Specification: PT-BR Multilingual Support

## Overview
Support Portuguese (Brazilian) language throughout the bot. Recipes can be saved in any language but are stored with translations to both English and Portuguese. Users interact in their preferred language.

## Requirements

### Language Detection & Storage Strategy

```
Recipe Input (any language) → Detect Language →
  If English:    Store EN original + translate to PT-BR
  If Portuguese: Store PT-BR original + translate to EN
  If Other:      Translate to both EN and PT-BR, store all

User Output → Based on user's language preference (Telegram setting or /language)
```

### Domain Model Updates

#### Recipe Entity
Update `internal/domain/recipe/entity.go`:
```go
type Recipe struct {
    // ... existing fields

    // Language metadata
    OriginalLanguage string  // "en", "pt", "es", etc. (ISO 639-1)

    // English version (always populated - primary storage)
    Title        string
    Ingredients  []Ingredient
    Instructions []string

    // Portuguese translation
    TitlePT        string
    IngredientsPT  []Ingredient
    InstructionsPT []string
}

// Localized returns the recipe content in the specified language
func (r *Recipe) Localized(lang string) LocalizedRecipe

type LocalizedRecipe struct {
    Title        string
    Ingredients  []Ingredient
    Instructions []string
}
```

#### User Entity
Update `internal/domain/user/entity.go`:
```go
type User struct {
    // ... existing fields
    LanguagePreference string  // "en", "pt" - defaults from Telegram
}

func (u *User) SetLanguagePreference(lang string) error
func (u *User) GetLanguagePreference() string
```

### LLM Prompt Updates

Update `internal/adapters/llm/prompts.go`:
```go
const extractionPromptMultilingual = `
Extract the recipe from the following content and provide translations.

First, detect the original language of the recipe.

Return a JSON object with:
{
  "originalLanguage": "en|pt|es|fr|...",

  // English version (translate if not originally English)
  "title": "Recipe title in English",
  "ingredients": [
    {"name": "ingredient in English", "quantity": "amount", "unit": "unit", "notes": "optional notes"}
  ],
  "instructions": ["Step 1 in English", "Step 2 in English", ...],

  // Portuguese (Brazilian) version (translate if not originally Portuguese)
  "titlePT": "Titulo da receita em Portugues",
  "ingredientsPT": [
    {"name": "ingrediente em portugues", "quantity": "quantidade", "unit": "unidade", "notes": "notas opcionais"}
  ],
  "instructionsPT": ["Passo 1 em Portugues", "Passo 2 em Portugues", ...],

  // ... other fields (category, cuisine, dietaryTags, etc.)
}

Translation guidelines:
- Translate ingredient names naturally (e.g., "chicken breast" → "peito de frango")
- Translate cooking terms appropriately (e.g., "saute" → "refogar", "simmer" → "cozinhar em fogo baixo")
- Use metric measurements for PT-BR when possible
- Preserve original cooking times and temperatures
- Keep brand names unchanged
- Translate "to taste" → "a gosto"

Content to extract:
%s
`
```

### Category Translations

Create `internal/domain/recipe/category_i18n.go`:
```go
package recipe

var CategoryTranslations = map[Category]map[string]string{
    CategoryPasta: {
        "en": "Pasta & Noodles",
        "pt": "Massas & Macarrao",
    },
    CategoryRice: {
        "en": "Rice & Grains",
        "pt": "Arroz & Graos",
    },
    CategorySoups: {
        "en": "Soups & Stews",
        "pt": "Sopas & Ensopados",
    },
    CategorySalads: {
        "en": "Salads",
        "pt": "Saladas",
    },
    CategoryMeat: {
        "en": "Meat & Poultry",
        "pt": "Carnes & Aves",
    },
    CategorySeafood: {
        "en": "Seafood",
        "pt": "Frutos do Mar",
    },
    CategoryVegetarian: {
        "en": "Vegetarian",
        "pt": "Vegetariano",
    },
    CategoryDesserts: {
        "en": "Desserts & Sweets",
        "pt": "Sobremesas & Doces",
    },
    CategoryBreakfast: {
        "en": "Breakfast",
        "pt": "Cafe da Manha",
    },
    CategoryAppetizers: {
        "en": "Appetizers & Snacks",
        "pt": "Aperitivos & Petiscos",
    },
    CategoryBeverages: {
        "en": "Beverages",
        "pt": "Bebidas",
    },
    CategorySauces: {
        "en": "Sauces & Condiments",
        "pt": "Molhos & Condimentos",
    },
    CategoryBread: {
        "en": "Bread & Baking",
        "pt": "Paes & Assados",
    },
    CategoryOther: {
        "en": "Other",
        "pt": "Outros",
    },
}

// Localized returns the category name in the specified language
func (c Category) Localized(lang string) string {
    if translations, ok := CategoryTranslations[c]; ok {
        if name, ok := translations[lang]; ok {
            return name
        }
    }
    return string(c) // fallback to English
}

// ParseCategoryLocalized parses a category from any supported language
func ParseCategoryLocalized(s string) Category
```

### UI String Translations

Create `internal/adapters/telegram/i18n.go`:
```go
package telegram

var UIStrings = map[string]map[string]string{
    "welcome": {
        "en": "Welcome to Recipe Bot! Send me a link to save a recipe.",
        "pt": "Bem-vindo ao Recipe Bot! Envie um link para salvar uma receita.",
    },
    "processing": {
        "en": "Processing your recipe link...",
        "pt": "Processando o link da receita...",
    },
    "recipes_title": {
        "en": "Your Recipes",
        "pt": "Suas Receitas",
    },
    "no_recipes": {
        "en": "You don't have any saved recipes yet.",
        "pt": "Voce ainda nao tem receitas salvas.",
    },
    "ingredients": {
        "en": "Ingredients",
        "pt": "Ingredientes",
    },
    "instructions": {
        "en": "Instructions",
        "pt": "Modo de Preparo",
    },
    "prep_time": {
        "en": "Prep Time",
        "pt": "Tempo de Preparo",
    },
    "cook_time": {
        "en": "Cook Time",
        "pt": "Tempo de Cozimento",
    },
    "servings": {
        "en": "Servings",
        "pt": "Porcoes",
    },
    // ... more strings
}

func T(key, lang string) string {
    if translations, ok := UIStrings[key]; ok {
        if text, ok := translations[lang]; ok {
            return text
        }
    }
    return UIStrings[key]["en"] // fallback to English
}
```

### Firestore Schema Updates

```
recipes collection (add fields):
  originalLanguage: string      // "en", "pt", etc.
  titlePT: string
  ingredientsPT: []map{
    name: string,
    quantity: string,
    unit: string,
    notes: string
  }
  instructionsPT: []string

users collection (add fields):
  languagePreference: string    // "en", "pt"
```

### Repository Updates

Update `internal/adapters/firebase/recipe_repository.go`:
```go
func (r *RecipeRepository) toFirestoreMap(recipe *recipe.Recipe) map[string]interface{} {
    // ... existing fields
    data["originalLanguage"] = recipe.OriginalLanguage
    data["titlePT"] = recipe.TitlePT
    data["ingredientsPT"] = toIngredientMaps(recipe.IngredientsPT)
    data["instructionsPT"] = recipe.InstructionsPT
    return data
}

func (r *RecipeRepository) fromFirestoreDoc(doc *firestore.DocumentSnapshot) (*recipe.Recipe, error) {
    // ... existing fields
    recipe.OriginalLanguage = data["originalLanguage"].(string)
    recipe.TitlePT = data["titlePT"].(string)
    // ... etc
}
```

Update `internal/adapters/firebase/user_repository.go`:
```go
func (r *UserRepository) UpdateLanguagePreference(ctx context.Context, userID shared.ID, lang string) error
```

### Telegram Commands

```
/language           - Show current language preference
/language pt        - Set language to Portuguese
/language en        - Set language to English
```

### Handler Updates

Update `internal/adapters/telegram/handlers.go`:
```go
func (h *Handler) HandleUpdate(update tgbotapi.Update) {
    // ... existing user creation

    // Detect language from Telegram if not set
    if user.LanguagePreference == "" {
        telegramLang := update.Message.From.LanguageCode
        if telegramLang == "pt" || strings.HasPrefix(telegramLang, "pt-") {
            user.SetLanguagePreference("pt")
        } else {
            user.SetLanguagePreference("en")
        }
    }

    // ... rest of handling
}

func (h *Handler) handleLanguage(ctx context.Context, message *tgbotapi.Message, userID shared.ID) {
    chatID := message.Chat.ID
    args := strings.TrimSpace(message.CommandArguments())
    lang := getUserLanguage(ctx, userID)

    if args == "" {
        // Show current language
        msg := T("current_language", lang) + ": " + lang
        h.bot.SendMessage(ctx, chatID, msg)
        return
    }

    // Set language
    newLang := strings.ToLower(args)
    if newLang != "en" && newLang != "pt" {
        h.bot.SendMessage(ctx, chatID, T("invalid_language", lang))
        return
    }

    h.userRepo.UpdateLanguagePreference(ctx, userID, newLang)
    h.bot.SendMessage(ctx, chatID, T("language_updated", newLang))
}
```

### Formatter Updates

Update `internal/adapters/telegram/formatter.go`:
```go
func FormatRecipeDTOLocalized(recipe *dto.RecipeDTO, lang string) string {
    var title, ingredients, instructions string

    if lang == "pt" && recipe.TitlePT != "" {
        title = recipe.TitlePT
        // ... use PT fields
    } else {
        title = recipe.Title
        // ... use EN fields
    }

    msg := fmt.Sprintf("*%s*\n\n", escapeMarkdown(title))
    msg += fmt.Sprintf("*%s:*\n", T("ingredients", lang))
    // ... rest of formatting
}
```

## Example Flows

### Flow 1: English user saves English recipe
1. User (lang: en) sends TikTok link with English recipe
2. LLM extracts: `{originalLanguage: "en", title: "Garlic Butter Shrimp", titlePT: "Camarao na Manteiga de Alho", ...}`
3. Stored with both versions
4. Displayed in English to user

### Flow 2: Brazilian user saves Portuguese recipe
1. User (lang: pt) sends link with Portuguese recipe
2. LLM extracts: `{originalLanguage: "pt", title: "Butter Garlic Shrimp", titlePT: "Camarao na Manteiga de Alho", ...}`
3. Stored with both versions
4. Displayed in Portuguese to user

### Flow 3: Brazilian user queries in Portuguese
1. User (lang: pt) sends: "Receitas de frango"
2. Intent detection understands Portuguese
3. Returns chicken recipes with Portuguese titles and content

## Acceptance Criteria
- [ ] New recipes stored with both EN and PT-BR translations
- [ ] User language detected from Telegram settings on first use
- [ ] /language command allows changing preference
- [ ] Recipe output respects user language preference
- [ ] Category names translated in lists
- [ ] UI strings (help, errors, etc.) translated
- [ ] Portuguese natural language queries work
- [ ] Unit tests for translation fields
- [ ] Integration tests for multilingual flow

## Technical Notes
- Default language is English if Telegram doesn't provide language_code
- LLM handles translation during extraction (single call, not separate translation step)
- Category parsing accepts both English and Portuguese names
- Ingredient matching works with both language versions
- Search queries normalized to work across languages
