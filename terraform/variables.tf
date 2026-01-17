variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP region for Cloud Run"
  type        = string
  default     = "us-central1"
}

variable "telegram_bot_token" {
  description = "Telegram Bot API token"
  type        = string
  sensitive   = true
}

variable "gemini_api_key" {
  description = "Google Gemini API key"
  type        = string
  sensitive   = true
}

variable "elevenlabs_api_key" {
  description = "ElevenLabs API key for transcription"
  type        = string
  sensitive   = true
}

variable "bot_image" {
  description = "Docker image for the Go bot service"
  type        = string
  default     = "gcr.io/PROJECT_ID/recipe-bot:latest"
}

variable "scraper_image" {
  description = "Docker image for the Python scraper service"
  type        = string
  default     = "gcr.io/PROJECT_ID/recipe-bot-scraper:latest"
}
