# Deployment Guide

## Prerequisites

### Required Services
1. **Telegram Bot Token**
   - Create a bot via [@BotFather](https://t.me/botfather)
   - Use `/newbot` command
   - Save the token

2. **Google Gemini API Key**
   - Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Create API key
   - Free tier: 15 requests/min, 1M tokens/day

3. **ElevenLabs API Key** (for transcription - recommended)
   - Sign up at [ElevenLabs](https://elevenlabs.io)
   - Go to Profile Settings > API Keys
   - Create and copy your API key
   - Affordable per-hour pricing

4. **Firebase Project**
   - Create project in [Firebase Console](https://console.firebase.google.com/)
   - **Enable Cloud Firestore API**:
     - Go to [Google Cloud Console](https://console.cloud.google.com/)
     - Select your project (or create one)
     - Navigate to [APIs & Services > Library](https://console.cloud.google.com/apis/library)
     - Search for "Cloud Firestore API"
     - Click "Enable"
     - Alternatively, use the direct link: `https://console.developers.google.com/apis/api/firestore.googleapis.com/overview?project=YOUR_PROJECT_ID`
   - Enable Firestore Database in [Firebase Console](https://console.firebase.google.com/)
     - Go to Firestore Database
     - Click "Create database"
     - Choose "Start in test mode" (you can add security rules later)
   - Create service account and download JSON key:
     - Go to [Google Cloud Console > IAM & Admin > Service Accounts](https://console.cloud.google.com/iam-admin/serviceaccounts)
     - Click "Create Service Account"
     - Give it a name (e.g., "receipt-bot-service")
     - **Grant role: "Firebase Admin SDK Administrator Service Agent"** (REQUIRED - this role provides full access to Firestore)
       - ⚠️ "Cloud Datastore User" is NOT sufficient - you need the Admin SDK role
     - Click "Done"
     - Click on the created service account
     - Go to "Keys" tab
     - Click "Add Key" > "Create new key"
     - Choose JSON format
     - Save the downloaded JSON file securely
     - **Important**: Verify the `project_id` in the JSON file matches your `FIREBASE_PROJECT_ID` in `.env`
   - Free tier: 50K reads, 20K writes/day

### Required Software
- **Go 1.23+** - [Download](https://golang.org/dl/)
- **Python 3.11+** - [Download](https://www.python.org/downloads/)
- **Poetry** - Python dependency management - [Install](https://python-poetry.org/docs/#installation)
- **FFmpeg** - For audio extraction
  - macOS: `brew install ffmpeg`
  - Ubuntu: `sudo apt-get install ffmpeg`
  - Windows: [Download](https://ffmpeg.org/download.html)

---

## Setup Steps

### 1. Clone and Configure

```bash
# Clone repository (or use existing directory)
cd receipt-bot

# Copy environment file
cp .env.example .env

# Edit .env with your credentials
nano .env
```

### 2. Configure Environment Variables

Edit `.env` file:

```bash
# Telegram
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_DEBUG=false

# LLM (Gemini recommended for free tier)
LLM_PROVIDER=gemini
LLM_MODEL=gemini-1.5-flash
GEMINI_API_KEY=your_gemini_api_key_here

# Firebase
FIREBASE_PROJECT_ID=your_project_id
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-credentials.json

# Python Service
PYTHON_SERVICE_URL=localhost:50051
PYTHON_SERVICE_TIMEOUT=300

# Application
APP_LOG_LEVEL=info
APP_PORT=8080
```

### 3. Setup Python Service

```bash
cd python-service

# Install Poetry (if not already installed)
curl -sSL https://install.python-poetry.org | python3 -

# Install dependencies
poetry install

# Copy Python service env file
cp .env.example .env

# Edit Python service .env
nano .env
```

**Alternative transcription providers:**
```bash
# Install Google Cloud Speech-to-Text (optional)
poetry install --with google-stt

# Install Whisper (optional)
poetry install --with whisper
```

Python service `.env`:
```bash
# Transcription (ElevenLabs is default and recommended)
TRANSCRIPTION_PROVIDER=elevenlabs
ELEVENLABS_API_KEY=your_elevenlabs_api_key_here

# Alternative: Google Cloud Speech-to-Text
# TRANSCRIPTION_PROVIDER=google-stt
# GOOGLE_CLOUD_CREDENTIALS_PATH=/path/to/google-cloud-credentials.json

GRPC_PORT=50051
LOG_LEVEL=INFO
TEMP_DIR=/tmp/recipe-bot
```

### 4. Generate Protocol Buffers

```bash
cd ../proto

# Install protoc if needed
# macOS: brew install protobuf
# Ubuntu: sudo apt-get install protobuf-compiler

# Generate code
make generate
```

### 5. Install Go Dependencies

```bash
cd ..
go mod download
```

---

## Running the Application

### Development Mode (Two Terminals)

**Terminal 1 - Python Service:**
```bash
cd python-service
poetry run python run_server.py
```

You should see:
```
INFO:root:Initialized transcriber with provider: google-stt
INFO:root:ScraperServicer initialized
INFO:root:gRPC server started on port 50051
```

**Terminal 2 - Go Bot:**
```bash
cd receipt-bot
go run cmd/bot/main.go
```

You should see:
```
Loading configuration...
Initializing Firebase...
Connecting to Python service...
Initializing LLM adapter (gemini)...
Initializing Telegram bot...
Authorized on account YourBotName
Bot is running. Press Ctrl+C to stop.
Waiting for updates...
```

---

## Testing the Bot

1. **Open Telegram** and find your bot
2. **Send /start** - You should get a welcome message
3. **Send a recipe URL**, for example:
   - YouTube: `https://youtube.com/watch?v=VIDEO_ID`
   - TikTok: `https://tiktok.com/@user/video/123`
   - Instagram: `https://instagram.com/p/POST_ID/`
   - Web: Any recipe website with schema.org markup

4. **Watch the magic!** The bot will:
   - Download the video
   - Transcribe the audio
   - Extract the recipe
   - Send you a formatted recipe

---

## Production Deployment

### Option 1: Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  python-service:
    build:
      context: .
      dockerfile: deployments/python-service.Dockerfile
    ports:
      - "50051:50051"
    env_file:
      - python-service/.env
    volumes:
      - /tmp/recipe-bot:/tmp/recipe-bot

  go-service:
    build:
      context: .
      dockerfile: deployments/go-service.Dockerfile
    depends_on:
      - python-service
    env_file:
      - .env
    environment:
      - PYTHON_SERVICE_URL=python-service:50051
```

Run:
```bash
docker-compose up -d
```

### Option 2: Separate Servers

**Server 1 - Python Service (any cloud provider):**
```bash
# Install Poetry
curl -sSL https://install.python-poetry.org | python3 -

# Install dependencies
poetry install --only main

# Run with process manager (e.g., supervisor, systemd)
poetry run python run_server.py
```

**Server 2 - Go Bot (any cloud provider):**
```bash
# Build
go build -o receipt-bot cmd/bot/main.go

# Run with process manager
./receipt-bot
```

Update `.env`:
```bash
PYTHON_SERVICE_URL=python-server-ip:50051
```

---

## Monitoring

### Logs

**Python Service:**
```bash
# Development
poetry run python run_server.py

# Production (with log file)
poetry run python run_server.py >> logs/python-service.log 2>&1
```

**Go Service:**
```bash
# Development
go run cmd/bot/main.go

# Production (with log file)
./receipt-bot >> logs/go-service.log 2>&1
```

### Health Checks

**Python Service:**
```bash
# Check if gRPC server is running
grpcurl -plaintext localhost:50051 list
```

**Go Service:**
```bash
# Check if bot is running
ps aux | grep receipt-bot
```

---

## Troubleshooting

### Python Service Won't Start

**Issue**: `ModuleNotFoundError`
```bash
# Solution: Install dependencies
cd python-service
poetry install
```

**Issue**: `FFmpeg not found`
```bash
# macOS
brew install ffmpeg

# Ubuntu
sudo apt-get install ffmpeg
```

### Go Service Won't Start

**Issue**: `Failed to connect to Python service`
```bash
# Check if Python service is running
nc -z localhost 50051

# Start Python service first
cd python-service
python run_server.py
```

**Issue**: `Firebase credentials not found`
```bash
# Check path in .env
FIREBASE_CREDENTIALS_PATH=/absolute/path/to/credentials.json
```

**Issue**: `Cloud Firestore API has not been used in project before or it is disabled` or `PermissionDenied`
```bash
# If you already enabled the API but still get PermissionDenied, check:

# 1. Verify API is enabled:
#    https://console.developers.google.com/apis/api/firestore.googleapis.com/overview?project=YOUR_PROJECT_ID

# 2. Check service account role (MOST COMMON ISSUE):
#    - Go to: https://console.cloud.google.com/iam-admin/serviceaccounts
#    - Click on your service account
#    - Go to "Permissions" tab
#    - Verify it has "Firebase Admin SDK Administrator Service Agent" role
#    - If not, click "Grant Access" and add this role
#    - ⚠️ "Cloud Datastore User" is NOT sufficient!

# 3. Verify project ID matches:
#    # Check project ID in credentials file
#    cat $FIREBASE_CREDENTIALS_PATH | grep project_id
#    # Should match FIREBASE_PROJECT_ID in your .env file

# 4. Wait a few minutes after making changes for them to propagate

# 5. Verify credentials file path is correct and file is readable
```

### Bot Not Responding

**Issue**: Bot token invalid
```bash
# Get new token from @BotFather
# Update TELEGRAM_BOT_TOKEN in .env
```

**Issue**: No updates received
```bash
# Check if bot is running
# Check Telegram bot logs
# Try sending /start command
```

### Recipe Extraction Fails

**Issue**: Video too long
```bash
# Google STT free tier: 60 min/month
# Solution: Use shorter videos or upgrade
```

**Issue**: LLM API quota exceeded
```bash
# Gemini free tier: 15 req/min, 1M tokens/day
# Solution: Wait or switch to OpenAI
```

**Issue**: `models/gemini-1.5-flash is not found for API version v1beta`
```bash
# This means the model name isn't available in the API version being used
# Solution: Update your LLM_MODEL in .env to one of these:
#   - gemini-pro (most reliable, recommended)
#   - gemini-1.5-pro
#   - gemini-1.5-flash-latest
#
# Example:
# LLM_MODEL=gemini-pro
#
# Then restart your bot
```

---

## Performance Optimization

### Python Service

```python
# Increase gRPC workers
GRPC_MAX_WORKERS=20

# Use faster Whisper model
TRANSCRIPTION_PROVIDER=whisper
WHISPER_MODEL=tiny  # Faster but less accurate
```

### Go Service

```go
# Increase timeout for long videos
PYTHON_SERVICE_TIMEOUT=600  # 10 minutes
```

### Database

```
# Add Firestore indexes
# In Firebase Console → Firestore → Indexes
# Index on: userId + createdAt (descending)
```

---

## Cost Monitoring

### Free Tier Limits

- **Gemini**: 15 requests/min, 1M tokens/day
- **Google STT**: 60 minutes/month
- **Firebase**: 50K reads, 20K writes/day

### Monitoring Usage

**Gemini**:
- [Google AI Studio](https://makersuite.google.com/) → Usage

**Firebase**:
- [Firebase Console](https://console.firebase.google.com/) → Usage

**Google Cloud STT**:
- [Google Cloud Console](https://console.cloud.google.com/) → Billing

---

## Security Best Practices

1. **Never commit credentials**
   ```bash
   # Already in .gitignore:
   .env
   *-firebase-adminsdk-*.json
   ```

2. **Use environment variables**
   - Never hardcode API keys
   - Use `.env` files

3. **Rotate keys regularly**
   - Telegram bot token
   - API keys

4. **Restrict Firebase access**
   - Use Firebase Security Rules
   - Limit read/write access

5. **Rate limiting**
   - Implement per-user rate limits
   - Prevent spam/abuse

---

## Backup & Recovery

### Firestore Backup

```bash
# Automated backups (Firebase Console)
# Firestore → Backups → Schedule backups
```

### Manual Export

```bash
gcloud firestore export gs://YOUR_BUCKET/backups
```

---

## Scaling

### Horizontal Scaling

**Python Service**:
- Run multiple instances
- Use load balancer
- Share /tmp/recipe-bot via network storage

**Go Service**:
- Run multiple instances
- Each connects to load-balanced Python service
- Telegram handles routing

### Vertical Scaling

- Increase server resources
- More CPU for transcription
- More RAM for video processing

---

## Support

For issues and questions:
1. Check [TESTING.md](TESTING.md) for test failures
2. Check logs for errors
3. Verify all credentials are correct
4. Ensure services are running

Happy cooking! 🍳
