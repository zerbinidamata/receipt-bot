# Cloud Run Deployment Guide

This guide explains how to deploy the Recipe Bot to Google Cloud Run.

## Prerequisites

1. Google Cloud project with billing enabled
2. `gcloud` CLI installed and configured
3. APIs enabled:
   - Cloud Run API
   - Cloud Build API
   - Container Registry API
   - Secret Manager API

```bash
gcloud services enable run.googleapis.com cloudbuild.googleapis.com containerregistry.googleapis.com secretmanager.googleapis.com
```

## Step 1: Create Secrets in Secret Manager

Store your sensitive configuration in Secret Manager:

```bash
# Telegram Bot Token
echo -n "YOUR_TELEGRAM_BOT_TOKEN" | gcloud secrets create telegram-bot-token --data-file=-

# Gemini API Key
echo -n "YOUR_GEMINI_API_KEY" | gcloud secrets create gemini-api-key --data-file=-

# ElevenLabs API Key (for Python service)
echo -n "YOUR_ELEVENLABS_API_KEY" | gcloud secrets create elevenlabs-api-key --data-file=-

# Firebase credentials (as JSON)
gcloud secrets create firebase-credentials --data-file=path/to/firebase-credentials.json
```

## Step 2: Grant Cloud Run Access to Secrets

```bash
PROJECT_ID=$(gcloud config get-value project)
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')

# Grant the Cloud Run service account access to secrets
gcloud secrets add-iam-policy-binding telegram-bot-token \
    --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding gemini-api-key \
    --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding elevenlabs-api-key \
    --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding firebase-credentials \
    --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
```

## Step 3: Deploy Using Cloud Build

### Option A: Manual Deployment

```bash
# Deploy from local machine
gcloud builds submit --config cloudbuild.yaml
```

### Option B: Set Up Continuous Deployment

Connect your repository to Cloud Build for automatic deployments on push:

```bash
# Create a Cloud Build trigger
gcloud builds triggers create github \
    --repo-name=receipt-bot \
    --repo-owner=YOUR_GITHUB_USERNAME \
    --branch-pattern="^main$" \
    --build-config=cloudbuild.yaml
```

## Step 4: Configure Environment Variables

After initial deployment, update the services with secrets:

### Python Service (Scraper)

```bash
gcloud run services update recipe-bot-scraper \
    --region us-central1 \
    --set-secrets=ELEVENLABS_API_KEY=elevenlabs-api-key:latest
```

### Go Service (Telegram Bot)

```bash
# Get the Python service URL
PYTHON_SERVICE_URL=$(gcloud run services describe recipe-bot-scraper \
    --region us-central1 \
    --format='value(status.url)')

# Update Go service with all required config
gcloud run services update recipe-bot \
    --region us-central1 \
    --set-secrets=TELEGRAM_BOT_TOKEN=telegram-bot-token:latest,GEMINI_API_KEY=gemini-api-key:latest \
    --set-env-vars="PYTHON_SERVICE_URL=${PYTHON_SERVICE_URL}:443,LLM_PROVIDER=gemini,LLM_MODEL=gemini-1.5-flash,FIREBASE_PROJECT_ID=YOUR_PROJECT_ID" \
    --set-secrets=/app/firebase-credentials.json=firebase-credentials:latest \
    --update-env-vars="FIREBASE_CREDENTIALS_PATH=/app/firebase-credentials.json"
```

## Step 5: Set Up Service-to-Service Authentication

The Go service needs to authenticate with the Python service:

```bash
# Grant the Go service permission to invoke the Python service
gcloud run services add-iam-policy-binding recipe-bot-scraper \
    --region us-central1 \
    --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
    --role="roles/run.invoker"
```

## Step 6: Configure Telegram Webhook (Optional)

For production, use webhooks instead of polling:

```bash
# Get the Go service URL
BOT_URL=$(gcloud run services describe recipe-bot \
    --region us-central1 \
    --format='value(status.url)')

# Set webhook
curl "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook?url=${BOT_URL}/webhook"
```

## Architecture Notes

```
                    ┌─────────────────┐
                    │    Telegram     │
                    │      API        │
                    └────────┬────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────┐
│                  Cloud Run                       │
│  ┌──────────────────────────────────────────┐   │
│  │         recipe-bot (Go)                   │   │
│  │  - Telegram bot handler                   │   │
│  │  - LLM recipe extraction                  │   │
│  │  - Firebase persistence                   │   │
│  └──────────────────┬───────────────────────┘   │
│                     │ gRPC                       │
│  ┌──────────────────▼───────────────────────┐   │
│  │     recipe-bot-scraper (Python)          │   │
│  │  - Video downloading (yt-dlp)            │   │
│  │  - Audio extraction (ffmpeg)             │   │
│  │  - Transcription (ElevenLabs)            │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │    Firestore    │
                    │    Database     │
                    └─────────────────┘
```

## Troubleshooting

### View Logs

```bash
# Go service logs
gcloud run services logs read recipe-bot --region us-central1

# Python service logs
gcloud run services logs read recipe-bot-scraper --region us-central1
```

### Check Service Status

```bash
gcloud run services describe recipe-bot --region us-central1
gcloud run services describe recipe-bot-scraper --region us-central1
```

### Test Python Service Locally

```bash
# Build and run locally
docker build -t recipe-bot-scraper -f python-service/Dockerfile python-service
docker run -p 50051:50051 -e ELEVENLABS_API_KEY=xxx recipe-bot-scraper
```

## Cost Optimization

- Cloud Run charges only when handling requests
- Set `--min-instances=0` to scale to zero when idle
- Use `--max-instances=5` to limit concurrent instances
- Consider regional endpoints to reduce latency
