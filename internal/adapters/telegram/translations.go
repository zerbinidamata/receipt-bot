package telegram

import (
	"receipt-bot/internal/domain/user"
)

// Translations holds all translatable strings for a language
type Translations struct {
	// Welcome and help
	Welcome string
	Help    string

	// Common labels
	Info         string
	Prep         string
	Cook         string
	Servings     string
	Category     string
	Cuisine      string
	Tags         string
	Ingredients  string
	Instructions string
	Source       string
	By           string

	// Recipe list
	YourRecipes       string
	Recipes           string
	NoRecipesYet      string
	NoRecipesFound    string
	SendLinkToStart   string
	UseRecipeNumber   string
	UseRecipesFilter  string
	AndMoreRecipes    string
	ShowMoreHint      string
	DetailsHint       string
	FilterHint        string

	// Categories
	RecipeCategories string
	UseRecipesCmd    string
	Example          string

	// Match results
	HereWhatYouCanMake string
	PerfectMatches     string
	AlmostThere        string
	PartialMatches     string
	Missing            string
	NoMatchingRecipes  string
	TryAddingMore      string
	UseRecipeCmd       string

	// Pantry
	YourPantry        string
	PantryEmpty       string
	PantryAddHint     string
	PantryRemoveHint  string
	PantryClearHint   string
	MatchHint         string
	AddedToPantry     string
	RemovedFromPantry string
	PantryCleared     string
	PantryNowHas      string
	FindRecipesHint   string

	// Processing
	ProcessingLink string
	MayTakeMinute  string

	// Errors
	FailedToList       string
	FailedToGet        string
	FailedToProcess    string
	FailedToMatch      string
	FailedToAddPantry  string
	FailedToClear      string
	PleaseTryAgain     string
	InvalidRecipeNum   string
	SpecifyRecipeNum   string
	SpecifyItems       string

	// Commands
	UnknownCommand   string
	UseHelpCmd       string
	Commands         string
	StartCmd         string
	HelpCmd          string
	RecipesCmd       string
	RecipeCmd        string
	CategoriesCmd    string
	MatchCmd         string
	PantryCmd        string
	LanguageCmd      string

	// Greetings and fallbacks
	Greeting          string
	GreetingHint      string
	FallbackMessage   string
	NotSureWhatYouMean string

	// Language
	LanguageSet      string
	LanguageCurrent  string
	LanguageChoose   string
	LanguageEnglish  string
	LanguagePortuguese string

	// Natural language hints
	NLSendLink      string
	NLShowRecipes   string
	NLHaveIngredients string
	NLMyPantry      string

	// Category names (for display)
	CategoryPastaNoodles     string
	CategoryRiceGrains       string
	CategorySoupsStews       string
	CategorySalads           string
	CategoryMeatPoultry      string
	CategorySeafood          string
	CategoryVegetarian       string
	CategoryDessertsSweets   string
	CategoryBreakfast        string
	CategoryAppetizersSnacks string
	CategoryBeverages        string
	CategorySaucesCondiments string
	CategoryBreadBaking      string
	CategoryOther            string

	// Dietary tags (for display)
	TagVegetarian  string
	TagVegan       string
	TagGlutenFree  string
	TagDairyFree   string
	TagLowCarb     string
	TagQuick       string
	TagOnePot      string
	TagKidFriendly string

	// Export
	ExportCmd           string
	ExportHelp          string
	ExportUsage         string
	ExportObsidianHint  string
	ExportNotionHint    string
	ExportingRecipes    string
	ExportSuccess       string
	ExportFailed        string
	ExportNoRecipes     string
	ConnectCmd          string
	ConnectHelp         string
	ConnectNotionHint   string
	NotionConnected     string
	NotionDisconnected  string
	NotionNotConnected  string
	NotionAuthURL       string
	DisconnectCmd       string
}

// englishTranslations contains all English strings
var englishTranslations = &Translations{
	// Welcome and help
	Welcome: `Welcome to Recipe Bot!

I can help you extract recipes from:
• TikTok videos
• YouTube videos
• Instagram posts/reels
• Recipe websites

*How to use:*
Just send me a link to any recipe video or webpage, and I'll extract the ingredients and cooking instructions for you!

*Commands:*
/start - Show this message
/help - Get help
/recipes - List your saved recipes
/recipe <number> - View a specific recipe
/language - Change language

Let's get cooking!`,

	Help: `*Recipe Bot Help*

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
/recipes <category> - Filter by category
/recipe <number> - View a specific recipe
/categories - Show recipe categories
/match <ingredients> - Find recipes by ingredients
/pantry - Manage your pantry items
/language - Change language

*Having issues?*
Make sure:
• The link is valid
• The content contains a recipe
• The video has clear audio (if applicable)

Happy cooking!`,

	// Common labels
	Info:         "Info",
	Prep:         "Prep",
	Cook:         "Cook",
	Servings:     "Servings",
	Category:     "Category",
	Cuisine:      "Cuisine",
	Tags:         "Tags",
	Ingredients:  "Ingredients",
	Instructions: "Instructions",
	Source:       "Source",
	By:           "By",

	// Recipe list
	YourRecipes:      "Your Recipes",
	Recipes:          "Recipes",
	NoRecipesYet:     "You don't have any saved recipes yet.",
	NoRecipesFound:   "No recipes found.",
	SendLinkToStart:  "Send me a link to a recipe video or webpage to get started!",
	UseRecipeNumber:  "Use /recipe <number> to view details",
	UseRecipesFilter: "Use /recipes <category> to filter",
	AndMoreRecipes:   "... and %d more recipes",
	ShowMoreHint:     "Say \"show more\" to see them.",
	DetailsHint:      "Say \"details on #X\" to view a recipe",
	FilterHint:       "Or try \"quick pasta recipes\" to filter",

	// Categories
	RecipeCategories: "Recipe Categories",
	UseRecipesCmd:    "Use /recipes <category> to filter",
	Example:          "Example: /recipes pasta",

	// Match results
	HereWhatYouCanMake: "Here's what you can make:",
	PerfectMatches:     "Perfect Matches",
	AlmostThere:        "Almost There",
	PartialMatches:     "Partial Matches",
	Missing:            "Missing",
	NoMatchingRecipes:  "No matching recipes found.",
	TryAddingMore:      "Try adding more ingredients or use /recipes to see all your recipes.",
	UseRecipeCmd:       "Use /recipe <number> to view full recipe!",

	// Pantry
	YourPantry:        "Your Pantry",
	PantryEmpty:       "Your pantry is empty.",
	PantryAddHint:     "Use /pantry add <items> to add ingredients.",
	PantryRemoveHint:  "/pantry remove <items> - Remove items",
	PantryClearHint:   "/pantry clear - Clear all items",
	MatchHint:         "/match - Find recipes with pantry items",
	AddedToPantry:     "Added %d item(s) to your pantry.",
	RemovedFromPantry: "Removed item(s) from your pantry.",
	PantryCleared:     "Your pantry has been cleared.",
	PantryNowHas:      "Your pantry now has %d items.",
	FindRecipesHint:   "Use /match to find recipes!",

	// Processing
	ProcessingLink: "Processing your recipe link...",
	MayTakeMinute:  "This may take a minute.",

	// Errors
	FailedToList:      "Failed to list recipes.",
	FailedToGet:       "Failed to get recipe.",
	FailedToProcess:   "Failed to process recipe.",
	FailedToMatch:     "Failed to match ingredients.",
	FailedToAddPantry: "Failed to add items.",
	FailedToClear:     "Failed to clear pantry.",
	PleaseTryAgain:    "Please try again.",
	InvalidRecipeNum:  "Invalid recipe number. Please use a number like: /recipe 1",
	SpecifyRecipeNum:  "Please specify a recipe number.",
	SpecifyItems:      "Please specify items.",

	// Commands
	UnknownCommand: "Unknown command.",
	UseHelpCmd:     "Use /help to see available commands.",
	Commands:       "Commands:",
	StartCmd:       "/start - Welcome message",
	HelpCmd:        "/help - This help message",
	RecipesCmd:     "/recipes - Your saved recipes",
	RecipeCmd:      "/recipe <number> - View a specific recipe",
	CategoriesCmd:  "/categories - Show recipe categories",
	MatchCmd:       "/match <ingredients> - Find recipes by ingredients",
	PantryCmd:      "/pantry - Manage your pantry items",
	LanguageCmd:    "/language - Change language",

	// Greetings and fallbacks
	Greeting:          "Hello! I'm your recipe assistant.",
	GreetingHint:      "Send me a recipe link to save it, or try:",
	FallbackMessage:   "I can help you with recipes! Try:",
	NotSureWhatYouMean: "I'm not sure what you mean. Try:",

	// Language
	LanguageSet:        "Language set to English.",
	LanguageCurrent:    "Current language: English",
	LanguageChoose:     "Choose your language:",
	LanguageEnglish:    "English",
	LanguagePortuguese: "Português (BR)",

	// Natural language hints
	NLSendLink:        "Send me a recipe link to save it",
	NLShowRecipes:     "\"Show my recipes\" or \"seafood recipes\"",
	NLHaveIngredients: "\"I have chicken and pasta\" to find matching recipes",
	NLMyPantry:        "\"My pantry\" or \"add eggs to pantry\"",

	// Category names
	CategoryPastaNoodles:     "Pasta & Noodles",
	CategoryRiceGrains:       "Rice & Grains",
	CategorySoupsStews:       "Soups & Stews",
	CategorySalads:           "Salads",
	CategoryMeatPoultry:      "Meat & Poultry",
	CategorySeafood:          "Seafood",
	CategoryVegetarian:       "Vegetarian",
	CategoryDessertsSweets:   "Desserts & Sweets",
	CategoryBreakfast:        "Breakfast",
	CategoryAppetizersSnacks: "Appetizers & Snacks",
	CategoryBeverages:        "Beverages",
	CategorySaucesCondiments: "Sauces & Condiments",
	CategoryBreadBaking:      "Bread & Baking",
	CategoryOther:            "Other",

	// Dietary tags
	TagVegetarian:  "vegetarian",
	TagVegan:       "vegan",
	TagGlutenFree:  "gluten-free",
	TagDairyFree:   "dairy-free",
	TagLowCarb:     "low-carb",
	TagQuick:       "quick",
	TagOnePot:      "one-pot",
	TagKidFriendly: "kid-friendly",

	// Export
	ExportCmd:           "/export - Export recipes",
	ExportHelp:          "Export your recipes to other apps",
	ExportUsage:         "Usage: /export <format> [recipe_number]",
	ExportObsidianHint:  "/export obsidian - Export as Markdown file (for Obsidian)",
	ExportNotionHint:    "/export notion - Export to Notion database",
	ExportingRecipes:    "Exporting recipes...",
	ExportSuccess:       "Export successful!",
	ExportFailed:        "Export failed. Please try again.",
	ExportNoRecipes:     "No recipes to export.",
	ConnectCmd:          "/connect - Connect external services",
	ConnectHelp:         "Connect your account to external services",
	ConnectNotionHint:   "/connect notion - Connect to Notion",
	NotionConnected:     "Notion connected successfully!",
	NotionDisconnected:  "Notion disconnected.",
	NotionNotConnected:  "Not connected to Notion. Use /connect notion to authorize.",
	NotionAuthURL:       "Click here to authorize Notion access:",
	DisconnectCmd:       "/disconnect notion - Disconnect Notion",
}

// portugueseTranslations contains all Portuguese (BR) strings
var portugueseTranslations = &Translations{
	// Welcome and help
	Welcome: `Bem-vindo ao Recipe Bot!

Posso te ajudar a extrair receitas de:
• Vídeos do TikTok
• Vídeos do YouTube
• Posts/reels do Instagram
• Sites de receitas

*Como usar:*
Basta me enviar um link de qualquer vídeo ou página de receita, e eu vou extrair os ingredientes e instruções de preparo para você!

*Comandos:*
/start - Mostrar esta mensagem
/help - Obter ajuda
/recipes - Listar suas receitas salvas
/recipe <número> - Ver uma receita específica
/language - Mudar idioma

Vamos cozinhar!`,

	Help: `*Ajuda do Recipe Bot*

*Plataformas Suportadas:*
• TikTok (tiktok.com)
• YouTube (youtube.com, youtu.be)
• Instagram (instagram.com)
• Sites de receitas (com marcação schema.org)

*Como funciona:*
1. Me envie um link de receita
2. Vou baixar e transcrever o vídeo
3. IA extrai ingredientes e instruções
4. Você recebe a receita formatada!

*Dicas:*
• Certifique-se de que o link contém uma receita
• Vídeos com áudio claro funcionam melhor
• Receitas escritas também são suportadas

*Comandos:*
/start - Mensagem de boas-vindas
/help - Esta mensagem de ajuda
/recipes - Suas receitas salvas
/recipes <categoria> - Filtrar por categoria
/recipe <número> - Ver uma receita específica
/categories - Mostrar categorias
/match <ingredientes> - Encontrar receitas por ingredientes
/pantry - Gerenciar sua despensa
/language - Mudar idioma

*Tendo problemas?*
Verifique:
• O link é válido
• O conteúdo contém uma receita
• O vídeo tem áudio claro (se aplicável)

Bom apetite!`,

	// Common labels
	Info:         "Info",
	Prep:         "Preparo",
	Cook:         "Cozimento",
	Servings:     "Porções",
	Category:     "Categoria",
	Cuisine:      "Cozinha",
	Tags:         "Tags",
	Ingredients:  "Ingredientes",
	Instructions: "Modo de Preparo",
	Source:       "Fonte",
	By:           "Por",

	// Recipe list
	YourRecipes:      "Suas Receitas",
	Recipes:          "Receitas",
	NoRecipesYet:     "Você ainda não tem receitas salvas.",
	NoRecipesFound:   "Nenhuma receita encontrada.",
	SendLinkToStart:  "Me envie um link de receita para começar!",
	UseRecipeNumber:  "Use /recipe <número> para ver detalhes",
	UseRecipesFilter: "Use /recipes <categoria> para filtrar",
	AndMoreRecipes:   "... e mais %d receitas",
	ShowMoreHint:     "Diga \"mostrar mais\" para ver.",
	DetailsHint:      "Diga \"detalhes do #X\" para ver uma receita",
	FilterHint:       "Ou tente \"receitas rápidas de massa\" para filtrar",

	// Categories
	RecipeCategories: "Categorias de Receitas",
	UseRecipesCmd:    "Use /recipes <categoria> para filtrar",
	Example:          "Exemplo: /recipes massa",

	// Match results
	HereWhatYouCanMake: "Veja o que você pode fazer:",
	PerfectMatches:     "Combinações Perfeitas",
	AlmostThere:        "Quase Lá",
	PartialMatches:     "Combinações Parciais",
	Missing:            "Faltando",
	NoMatchingRecipes:  "Nenhuma receita encontrada.",
	TryAddingMore:      "Tente adicionar mais ingredientes ou use /recipes para ver todas suas receitas.",
	UseRecipeCmd:       "Use /recipe <número> para ver a receita completa!",

	// Pantry
	YourPantry:        "Sua Despensa",
	PantryEmpty:       "Sua despensa está vazia.",
	PantryAddHint:     "Use /pantry add <itens> para adicionar ingredientes.",
	PantryRemoveHint:  "/pantry remove <itens> - Remover itens",
	PantryClearHint:   "/pantry clear - Limpar tudo",
	MatchHint:         "/match - Encontrar receitas com itens da despensa",
	AddedToPantry:     "Adicionado(s) %d item(ns) à sua despensa.",
	RemovedFromPantry: "Item(ns) removido(s) da sua despensa.",
	PantryCleared:     "Sua despensa foi limpa.",
	PantryNowHas:      "Sua despensa agora tem %d itens.",
	FindRecipesHint:   "Use /match para encontrar receitas!",

	// Processing
	ProcessingLink: "Processando seu link de receita...",
	MayTakeMinute:  "Isso pode levar um minuto.",

	// Errors
	FailedToList:      "Falha ao listar receitas.",
	FailedToGet:       "Falha ao obter receita.",
	FailedToProcess:   "Falha ao processar receita.",
	FailedToMatch:     "Falha ao combinar ingredientes.",
	FailedToAddPantry: "Falha ao adicionar itens.",
	FailedToClear:     "Falha ao limpar despensa.",
	PleaseTryAgain:    "Por favor, tente novamente.",
	InvalidRecipeNum:  "Número de receita inválido. Use um número como: /recipe 1",
	SpecifyRecipeNum:  "Por favor, especifique um número de receita.",
	SpecifyItems:      "Por favor, especifique os itens.",

	// Commands
	UnknownCommand: "Comando desconhecido.",
	UseHelpCmd:     "Use /help para ver os comandos disponíveis.",
	Commands:       "Comandos:",
	StartCmd:       "/start - Mensagem de boas-vindas",
	HelpCmd:        "/help - Esta mensagem de ajuda",
	RecipesCmd:     "/recipes - Suas receitas salvas",
	RecipeCmd:      "/recipe <número> - Ver uma receita específica",
	CategoriesCmd:  "/categories - Mostrar categorias",
	MatchCmd:       "/match <ingredientes> - Encontrar receitas por ingredientes",
	PantryCmd:      "/pantry - Gerenciar sua despensa",
	LanguageCmd:    "/language - Mudar idioma",

	// Greetings and fallbacks
	Greeting:          "Olá! Sou seu assistente de receitas.",
	GreetingHint:      "Me envie um link de receita para salvar, ou tente:",
	FallbackMessage:   "Posso te ajudar com receitas! Tente:",
	NotSureWhatYouMean: "Não tenho certeza do que você quer dizer. Tente:",

	// Language
	LanguageSet:        "Idioma definido para Português (BR).",
	LanguageCurrent:    "Idioma atual: Português (BR)",
	LanguageChoose:     "Escolha seu idioma:",
	LanguageEnglish:    "English",
	LanguagePortuguese: "Português (BR)",

	// Natural language hints
	NLSendLink:        "Me envie um link de receita para salvar",
	NLShowRecipes:     "\"Mostrar minhas receitas\" ou \"receitas de frutos do mar\"",
	NLHaveIngredients: "\"Tenho frango e macarrão\" para encontrar receitas",
	NLMyPantry:        "\"Minha despensa\" ou \"adicionar ovos à despensa\"",

	// Category names
	CategoryPastaNoodles:     "Massas",
	CategoryRiceGrains:       "Arroz e Grãos",
	CategorySoupsStews:       "Sopas e Ensopados",
	CategorySalads:           "Saladas",
	CategoryMeatPoultry:      "Carnes e Aves",
	CategorySeafood:          "Frutos do Mar",
	CategoryVegetarian:       "Vegetariano",
	CategoryDessertsSweets:   "Sobremesas e Doces",
	CategoryBreakfast:        "Café da Manhã",
	CategoryAppetizersSnacks: "Aperitivos e Lanches",
	CategoryBeverages:        "Bebidas",
	CategorySaucesCondiments: "Molhos e Condimentos",
	CategoryBreadBaking:      "Pães e Assados",
	CategoryOther:            "Outros",

	// Dietary tags
	TagVegetarian:  "vegetariano",
	TagVegan:       "vegano",
	TagGlutenFree:  "sem glúten",
	TagDairyFree:   "sem lactose",
	TagLowCarb:     "low-carb",
	TagQuick:       "rápido",
	TagOnePot:      "panela única",
	TagKidFriendly: "para crianças",

	// Export
	ExportCmd:           "/export - Exportar receitas",
	ExportHelp:          "Exporte suas receitas para outros apps",
	ExportUsage:         "Uso: /export <formato> [número_receita]",
	ExportObsidianHint:  "/export obsidian - Exportar como arquivo Markdown (para Obsidian)",
	ExportNotionHint:    "/export notion - Exportar para banco de dados Notion",
	ExportingRecipes:    "Exportando receitas...",
	ExportSuccess:       "Exportação concluída!",
	ExportFailed:        "Falha na exportação. Tente novamente.",
	ExportNoRecipes:     "Nenhuma receita para exportar.",
	ConnectCmd:          "/connect - Conectar serviços externos",
	ConnectHelp:         "Conecte sua conta a serviços externos",
	ConnectNotionHint:   "/connect notion - Conectar ao Notion",
	NotionConnected:     "Notion conectado com sucesso!",
	NotionDisconnected:  "Notion desconectado.",
	NotionNotConnected:  "Não conectado ao Notion. Use /connect notion para autorizar.",
	NotionAuthURL:       "Clique aqui para autorizar acesso ao Notion:",
	DisconnectCmd:       "/disconnect notion - Desconectar Notion",
}

// GetTranslations returns the translations for the given language
func GetTranslations(lang user.Language) *Translations {
	switch lang {
	case user.LanguagePortuguese:
		return portugueseTranslations
	default:
		return englishTranslations
	}
}

// TranslateCategory translates a category name to the given language
func TranslateCategory(category string, lang user.Language) string {
	t := GetTranslations(lang)
	switch category {
	case "Pasta & Noodles":
		return t.CategoryPastaNoodles
	case "Rice & Grains":
		return t.CategoryRiceGrains
	case "Soups & Stews":
		return t.CategorySoupsStews
	case "Salads":
		return t.CategorySalads
	case "Meat & Poultry":
		return t.CategoryMeatPoultry
	case "Seafood":
		return t.CategorySeafood
	case "Vegetarian":
		return t.CategoryVegetarian
	case "Desserts & Sweets":
		return t.CategoryDessertsSweets
	case "Breakfast":
		return t.CategoryBreakfast
	case "Appetizers & Snacks":
		return t.CategoryAppetizersSnacks
	case "Beverages":
		return t.CategoryBeverages
	case "Sauces & Condiments":
		return t.CategorySaucesCondiments
	case "Bread & Baking":
		return t.CategoryBreadBaking
	case "Other":
		return t.CategoryOther
	default:
		return category
	}
}

// TranslateDietaryTag translates a dietary tag to the given language
func TranslateDietaryTag(tag string, lang user.Language) string {
	t := GetTranslations(lang)
	switch tag {
	case "vegetarian":
		return t.TagVegetarian
	case "vegan":
		return t.TagVegan
	case "gluten-free":
		return t.TagGlutenFree
	case "dairy-free":
		return t.TagDairyFree
	case "low-carb":
		return t.TagLowCarb
	case "quick":
		return t.TagQuick
	case "one-pot":
		return t.TagOnePot
	case "kid-friendly":
		return t.TagKidFriendly
	default:
		return tag
	}
}
