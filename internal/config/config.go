package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Telegram TelegramConfig
	Firebase FirebaseConfig
	LLM      LLMConfig
	Python   PythonServiceConfig
	App      AppConfig
	Notion   NotionConfig
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken string
	Debug    bool
}

// FirebaseConfig holds Firebase configuration
type FirebaseConfig struct {
	ProjectID       string
	CredentialsPath string
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	Provider string // "gemini", "openai", "anthropic"
	APIKey   string
	Model    string
}

// PythonServiceConfig holds Python service configuration
type PythonServiceConfig struct {
	URL     string
	Timeout int // in seconds
}

// AppConfig holds general application configuration
type AppConfig struct {
	LogLevel string
	Port     int
}

// NotionConfig holds Notion OAuth configuration
type NotionConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_LOG_LEVEL", "info")
	viper.SetDefault("APP_PORT", 8080)
	viper.SetDefault("LLM_PROVIDER", "gemini")
	viper.SetDefault("LLM_MODEL", "gemini-pro")
	viper.SetDefault("PYTHON_SERVICE_URL", "localhost:50051")
	viper.SetDefault("PYTHON_SERVICE_TIMEOUT", 300)
	viper.SetDefault("TELEGRAM_DEBUG", false)

	// Read config file (optional, won't error if not found)
	_ = viper.ReadInConfig()

	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken: viper.GetString("TELEGRAM_BOT_TOKEN"),
			Debug:    viper.GetBool("TELEGRAM_DEBUG"),
		},
		Firebase: FirebaseConfig{
			ProjectID:       viper.GetString("FIREBASE_PROJECT_ID"),
			CredentialsPath: viper.GetString("FIREBASE_CREDENTIALS_PATH"),
		},
		LLM: LLMConfig{
			Provider: viper.GetString("LLM_PROVIDER"),
			APIKey:   getLLMAPIKey(viper.GetString("LLM_PROVIDER")),
			Model:    viper.GetString("LLM_MODEL"),
		},
		Python: PythonServiceConfig{
			URL:     viper.GetString("PYTHON_SERVICE_URL"),
			Timeout: viper.GetInt("PYTHON_SERVICE_TIMEOUT"),
		},
		App: AppConfig{
			LogLevel: viper.GetString("APP_LOG_LEVEL"),
			Port:     viper.GetInt("APP_PORT"),
		},
		Notion: NotionConfig{
			ClientID:     viper.GetString("NOTION_CLIENT_ID"),
			ClientSecret: viper.GetString("NOTION_CLIENT_SECRET"),
			RedirectURI:  viper.GetString("NOTION_REDIRECT_URI"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// getLLMAPIKey gets the appropriate API key based on the provider
func getLLMAPIKey(provider string) string {
	switch provider {
	case "gemini":
		return viper.GetString("GEMINI_API_KEY")
	case "openai":
		return viper.GetString("OPENAI_API_KEY")
	case "anthropic":
		return viper.GetString("ANTHROPIC_API_KEY")
	default:
		return viper.GetString("GEMINI_API_KEY")
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if c.Firebase.ProjectID == "" {
		return fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}

	// Firebase credentials can come from either file path or JSON environment variable
	if c.Firebase.CredentialsPath == "" && viper.GetString("GOOGLE_APPLICATION_CREDENTIALS_JSON") == "" {
		return fmt.Errorf("FIREBASE_CREDENTIALS_PATH or GOOGLE_APPLICATION_CREDENTIALS_JSON is required")
	}

	if c.LLM.APIKey == "" {
		return fmt.Errorf("LLM API key is required (GEMINI_API_KEY, OPENAI_API_KEY, or ANTHROPIC_API_KEY)")
	}

	if c.Python.URL == "" {
		return fmt.Errorf("PYTHON_SERVICE_URL is required")
	}

	return nil
}
