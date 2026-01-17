terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs (these are free)
resource "google_project_service" "run" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

# Secrets
resource "google_secret_manager_secret" "telegram_token" {
  secret_id = "telegram-bot-token"

  replication {
    auto {}
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "telegram_token" {
  secret      = google_secret_manager_secret.telegram_token.id
  secret_data = var.telegram_bot_token
}

resource "google_secret_manager_secret" "gemini_api_key" {
  secret_id = "gemini-api-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "gemini_api_key" {
  secret      = google_secret_manager_secret.gemini_api_key.id
  secret_data = var.gemini_api_key
}

resource "google_secret_manager_secret" "elevenlabs_api_key" {
  secret_id = "elevenlabs-api-key"

  replication {
    auto {}
  }

  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_version" "elevenlabs_api_key" {
  secret      = google_secret_manager_secret.elevenlabs_api_key.id
  secret_data = var.elevenlabs_api_key
}

# Service account for Cloud Run services
resource "google_service_account" "recipe_bot" {
  account_id   = "recipe-bot-sa"
  display_name = "Recipe Bot Service Account"
}

# Grant service account access to secrets
resource "google_secret_manager_secret_iam_member" "telegram_token_access" {
  secret_id = google_secret_manager_secret.telegram_token.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.recipe_bot.email}"
}

resource "google_secret_manager_secret_iam_member" "gemini_api_key_access" {
  secret_id = google_secret_manager_secret.gemini_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.recipe_bot.email}"
}

resource "google_secret_manager_secret_iam_member" "elevenlabs_api_key_access" {
  secret_id = google_secret_manager_secret.elevenlabs_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.recipe_bot.email}"
}

# Python Scraper Service (internal only)
resource "google_cloud_run_v2_service" "scraper" {
  name     = "recipe-bot-scraper"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

  template {
    service_account = google_service_account.recipe_bot.email

    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }

    containers {
      image = var.scraper_image

      ports {
        container_port = 50051
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "1Gi"
        }
      }

      env {
        name  = "TRANSCRIPTION_PROVIDER"
        value = "elevenlabs"
      }

      env {
        name = "ELEVENLABS_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.elevenlabs_api_key.secret_id
            version = "latest"
          }
        }
      }
    }

    timeout = "300s"
  }

  depends_on = [
    google_project_service.run,
    google_secret_manager_secret_iam_member.elevenlabs_api_key_access
  ]
}

# Allow Go service to invoke Python service
resource "google_cloud_run_v2_service_iam_member" "scraper_invoker" {
  name     = google_cloud_run_v2_service.scraper.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.recipe_bot.email}"
}

# Go Bot Service (public for Telegram webhook)
resource "google_cloud_run_v2_service" "bot" {
  name     = "recipe-bot"
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.recipe_bot.email

    scaling {
      min_instance_count = 0
      max_instance_count = 3
    }

    containers {
      image = var.bot_image

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "LLM_PROVIDER"
        value = "gemini"
      }

      env {
        name  = "LLM_MODEL"
        value = "gemini-1.5-flash"
      }

      env {
        name  = "FIREBASE_PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "PYTHON_SERVICE_URL"
        value = "${google_cloud_run_v2_service.scraper.uri}:443"
      }

      env {
        name = "TELEGRAM_BOT_TOKEN"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.telegram_token.secret_id
            version = "latest"
          }
        }
      }

      env {
        name = "GEMINI_API_KEY"
        value_source {
          secret_key_ref {
            secret  = google_secret_manager_secret.gemini_api_key.secret_id
            version = "latest"
          }
        }
      }
    }

    timeout = "300s"
  }

  depends_on = [
    google_project_service.run,
    google_cloud_run_v2_service.scraper,
    google_secret_manager_secret_iam_member.telegram_token_access,
    google_secret_manager_secret_iam_member.gemini_api_key_access
  ]
}

# Allow unauthenticated access to bot (for Telegram webhook)
resource "google_cloud_run_v2_service_iam_member" "bot_public" {
  name     = google_cloud_run_v2_service.bot.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Outputs
output "bot_url" {
  value       = google_cloud_run_v2_service.bot.uri
  description = "URL of the Recipe Bot service"
}

output "scraper_url" {
  value       = google_cloud_run_v2_service.scraper.uri
  description = "URL of the Scraper service (internal)"
}

output "webhook_setup_command" {
  value       = "curl 'https://api.telegram.org/bot<YOUR_TOKEN>/setWebhook?url=${google_cloud_run_v2_service.bot.uri}/webhook'"
  description = "Command to set up Telegram webhook"
}
