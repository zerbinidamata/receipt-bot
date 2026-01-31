package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"receipt-bot/internal/adapters/firebase"
	"receipt-bot/internal/adapters/llm"
	"receipt-bot/internal/adapters/notion"
	"receipt-bot/internal/adapters/obsidian"
	"receipt-bot/internal/adapters/python"
	"receipt-bot/internal/adapters/telegram"
	"receipt-bot/internal/application/command"
	"receipt-bot/internal/application/query"
	"receipt-bot/internal/config"
	"receipt-bot/internal/domain/recipe"
	"receipt-bot/internal/ports"
)

func main() {
	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize context
	ctx := context.Background()

	// Initialize Firebase
	log.Println("Initializing Firebase...")
	firebaseClient, err := firebase.NewClient(ctx, firebase.Config{
		ProjectID:       cfg.Firebase.ProjectID,
		CredentialsPath: cfg.Firebase.CredentialsPath,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	defer firebaseClient.Close()

	// Initialize repositories
	recipeRepo := firebase.NewRecipeRepository(firebaseClient.Firestore())
	userRepo := firebase.NewUserRepository(firebaseClient.Firestore())

	// Initialize Python service adapter
	log.Println("Connecting to Python service...")
	scraperAdapter, err := python.NewScraperAdapter(
		cfg.Python.URL,
		time.Duration(cfg.Python.Timeout)*time.Second,
	)
	if err != nil {
		log.Fatalf("Failed to initialize scraper adapter: %v", err)
	}
	defer scraperAdapter.Close()

	// Initialize LLM adapter
	log.Printf("Initializing LLM adapter (%s)...", cfg.LLM.Provider)
	llmAdapter, err := llm.NewLLMAdapter(llm.LLMConfig{
		Provider: cfg.LLM.Provider,
		APIKey:   cfg.LLM.APIKey,
		Model:    cfg.LLM.Model,
	})
	if err != nil {
		log.Fatalf("Failed to initialize LLM adapter: %v", err)
	}

	// Close Gemini client if needed
	if geminiAdapter, ok := llmAdapter.(*llm.GeminiAdapter); ok {
		defer geminiAdapter.Close()
	}

	// Initialize intent detector for conversational interface
	log.Println("Initializing intent detector...")
	intentDetector, err := llm.NewIntentDetector(llm.LLMConfig{
		Provider: cfg.LLM.Provider,
		APIKey:   cfg.LLM.APIKey,
		Model:    cfg.LLM.Model,
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize intent detector: %v", err)
		log.Println("Conversational interface will be disabled")
		intentDetector = nil
	}

	// Initialize Telegram bot
	log.Println("Initializing Telegram bot...")
	bot, err := telegram.NewBot(telegram.Config{
		BotToken: cfg.Telegram.BotToken,
		Debug:    cfg.Telegram.Debug,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Initialize domain services
	recipeService := recipe.NewService()

	// Initialize application layer
	log.Println("Initializing application layer...")

	processRecipeLinkCmd := command.NewProcessRecipeLinkCommand(
		scraperAdapter,
		llmAdapter,
		recipeService,
		recipeRepo,
		bot,
	)

	getOrCreateUserCmd := command.NewGetOrCreateUserCommand(userRepo)

	listRecipesQuery := query.NewListRecipesQuery(recipeRepo)

	matchIngredientsCmd := command.NewMatchIngredientsCommand(recipeRepo)

	managePantryCmd := command.NewManagePantryCommand(userRepo)

	// Initialize exporters
	obsidianExporter := obsidian.NewExporter()

	// Initialize Notion exporter (optional - only if configured)
	var notionExporter ports.NotionExporter
	if cfg.Notion.ClientID != "" && cfg.Notion.ClientSecret != "" {
		log.Println("Initializing Notion integration...")
		notionClient := notion.NewClient(notion.Config{
			ClientID:     cfg.Notion.ClientID,
			ClientSecret: cfg.Notion.ClientSecret,
			RedirectURI:  cfg.Notion.RedirectURI,
		})
		notionExporter = notion.NewExporter(notionClient, userRepo)
	} else {
		log.Println("Notion integration not configured (NOTION_CLIENT_ID and NOTION_CLIENT_SECRET not set)")
	}

	// Initialize export command
	exportRecipeCmd := command.NewExportRecipeCommand(
		recipeRepo,
		obsidianExporter,
		notionExporter,
	)

	// Initialize handler
	handler := telegram.NewHandler(telegram.HandlerConfig{
		Bot:                      bot,
		ProcessRecipeLinkCommand: processRecipeLinkCmd,
		GetOrCreateUserCommand:   getOrCreateUserCmd,
		ListRecipesQuery:         listRecipesQuery,
		MatchIngredientsCommand:  matchIngredientsCmd,
		ManagePantryCommand:      managePantryCmd,
		ExportRecipeCommand:      exportRecipeCmd,
		IntentDetector:           intentDetector,
		UserRepo:                 userRepo,
	})

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start receiving updates
	log.Println("Bot is running. Press Ctrl+C to stop.")
	log.Println("Waiting for updates...")

	updates := bot.GetUpdatesChan()

	// Main loop
	go func() {
		for update := range updates {
			handler.HandleUpdate(update)
		}
	}()

	// Wait for shutdown signal
	<-stop

	log.Println("Shutting down gracefully...")
	bot.Stop()
	log.Println("Goodbye!")
}
