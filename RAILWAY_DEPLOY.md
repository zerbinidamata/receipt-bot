# Railway Deployment Guide

This guide explains how to deploy the Recipe Bot to Railway with both services.

## Architecture

```
┌─────────────────────┐     ┌─────────────────────┐
│   recipe-bot (Go)   │────▶│ python-scraper      │
│   Port: 8080        │gRPC │ Port: 50051         │
│   (Public)          │     │ (Private)           │
└─────────────────────┘     └─────────────────────┘
           │                          │
           ▼                          ▼
    ┌──────────────┐          ┌──────────────┐
    │   Firebase   │          │  ElevenLabs  │
    │  Firestore   │          │  (or other)  │
    └──────────────┘          └──────────────┘
```

## Setup Steps

### 1. Create Railway Project

1. Go to [railway.app](https://railway.app) and create a new project
2. Connect your GitHub repository

### 2. Create Two Services

You need to create two separate services from the same repository:

#### Service 1: recipe-bot (Go Bot)
- Click "New Service" → "GitHub Repo"
- Select your repository
- Railway will auto-detect the `Dockerfile` in the root
- Set the service name to `recipe-bot`

#### Service 2: python-scraper (Python gRPC)
- Click "New Service" → "GitHub Repo"
- Select the same repository
- In service settings, set:
  - **Root Directory**: `python-service`
  - **Service Name**: `python-scraper`

### 3. Configure Environment Variables

#### For `recipe-bot` (Go service):

| Variable | Description | Example |
|----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram bot token | `123456:ABC-DEF...` |
| `FIREBASE_PROJECT_ID` | Firebase project ID | `my-project-id` |
| `GOOGLE_APPLICATION_CREDENTIALS_JSON` | Firebase service account JSON (entire content) | `{"type":"service_account",...}` |
| `LLM_PROVIDER` | LLM provider to use | `gemini` |
| `GEMINI_API_KEY` | Gemini API key (if using Gemini) | `AIza...` |
| `PYTHON_SERVICE_URL` | Internal URL to Python service | `python-scraper.railway.internal:50051` |
| `APP_PORT` | Port for the service | `8080` |

#### For `python-scraper` (Python service):

| Variable | Description | Example |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `50051` |
| `TRANSCRIPTION_PROVIDER` | Transcription provider | `elevenlabs` |
| `ELEVENLABS_API_KEY` | ElevenLabs API key | `sk_...` |
| `LOG_LEVEL` | Logging level | `INFO` |

### 4. Configure Private Networking

The Python service should be **private** (no public URL):

1. Go to `python-scraper` service settings
2. Under "Networking" → disable public networking
3. Note the internal DNS name: `python-scraper.railway.internal`

The Go service uses this internal URL to communicate with the Python service.

### 5. Set PYTHON_SERVICE_URL

In the `recipe-bot` service, set:
```
PYTHON_SERVICE_URL=python-scraper.railway.internal:50051
```

## Firebase Credentials

For Railway deployment, you cannot use a credentials file. Instead:

1. Go to Firebase Console → Project Settings → Service Accounts
2. Generate a new private key (downloads JSON file)
3. Copy the entire JSON content
4. Paste it as the value of `GOOGLE_APPLICATION_CREDENTIALS_JSON` in Railway

The JSON should look like:
```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...",
  ...
}
```

## Deployment

Railway will automatically deploy when you push to your main branch.

To manually deploy:
1. Go to your Railway project
2. Click on the service
3. Click "Deploy" or push a new commit

## Monitoring

- View logs in the Railway dashboard under each service
- Both services log to stdout/stderr which Railway captures

## Troubleshooting

### Python service connection failed
- Verify `PYTHON_SERVICE_URL` uses `.railway.internal` domain
- Check that both services are in the same Railway project
- Ensure Python service is healthy (check logs)

### Firebase authentication error
- Verify `GOOGLE_APPLICATION_CREDENTIALS_JSON` contains valid JSON
- Ensure the service account has Firestore permissions
- Check `FIREBASE_PROJECT_ID` matches the credentials

### Telegram bot not responding
- Verify `TELEGRAM_BOT_TOKEN` is correct
- Check Go service logs for errors
- Ensure the service is running (check Railway dashboard)

## Cost Estimation

Railway pricing (as of 2024):
- Hobby: $5/month includes $5 credits
- Pro: $20/month includes $20 credits

Resource usage:
- Go service: ~256MB RAM, minimal CPU
- Python service: ~512MB-1GB RAM (video processing)

Estimated cost: ~$5-15/month depending on usage
