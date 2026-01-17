# Receipt Bot - Recipe Extraction Telegram Bot

A Telegram bot that extracts recipes from TikTok, YouTube, Instagram, and recipe websites. Built with Go and Python using hexagonal architecture and Domain-Driven Design principles.

## Features

- 🎥 **Video Support**: Extract recipes from TikTok, YouTube, and Instagram videos
- 🌐 **Web Support**: Extract recipes from recipe websites with schema.org markup
- 🎯 **AI-Powered**: Uses LLM (Gemini/OpenAI) for intelligent recipe extraction
- 🗣️ **Transcription**: Automatically transcribes video audio to text
- 📝 **Structured Output**: Organizes recipes into ingredients and cooking instructions
- 💾 **Cloud Storage**: Saves recipes to Firebase Firestore
- 🔄 **Swappable Components**: Easy to switch between LLM providers, databases, and transcription services
- 💰 **Free Tier Optimized**: Designed to run on free tiers (Gemini, Google Speech-to-Text, Firebase)

## Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) and **Domain-Driven Design** principles:

**Tech Stack:**
- **Go Service**: Telegram bot, orchestration, and business logic
- **Python Service**: Web scraping, video processing, and transcription
- **Communication**: gRPC between services
- **Database**: Firebase Firestore (easily swappable)
- **LLM**: Google Gemini (free tier, easily swappable to OpenAI/Anthropic)
- **Transcription**: ElevenLabs Scribe (default, easily swappable to Google STT or Whisper)

## Project Structure

```
receipt-bot/
├── cmd/bot/                    # Main application entry point
├── internal/
│   ├── domain/                 # Domain layer (business logic)
│   │   ├── recipe/             # Recipe aggregate
│   │   ├── user/               # User entity
│   │   └── shared/             # Shared domain objects
│   ├── application/            # Application layer (use cases)
│   ├── ports/                  # Port interfaces
│   ├── adapters/               # Adapter implementations
│   └── config/                 # Configuration
├── python-service/             # Python microservice
│   └── src/
│       ├── grpc_server/        # gRPC server
│       ├── scrapers/           # Platform scrapers
│       ├── video/              # Video/audio processing
│       └── utils/              # Utilities
├── proto/                      # Protocol buffer definitions
└── deployments/                # Docker configurations
```

## Prerequisites

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed setup instructions.

**Required:**
- Go 1.23+
- Python 3.11+
- Poetry (Python dependency management)
- FFmpeg
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))
- Google Gemini API Key (free tier)
- ElevenLabs API Key (for transcription)
- Firebase project with Firestore

## Quick Start

### 1. Clone and Configure

```bash
cd receipt-bot

# Copy environment file
cp .env.example .env

# Edit .env with your credentials
nano .env
```

### 2. Option A: Run with Docker (Recommended)

```bash
# Make sure credentials are configured in .env
# Start services
./scripts/docker-start.sh

# View logs
./scripts/docker-logs.sh

# Stop services
./scripts/docker-stop.sh
```

### 2. Option B: Run Locally

**Terminal 1 - Python Service:**
```bash
cd python-service
poetry install
cp .env.example .env
nano .env  # Configure transcription settings
poetry run python run_server.py
```

**Terminal 2 - Go Bot:**
```bash
# From project root
go mod download
go run cmd/bot/main.go
```

## Configuration

Key environment variables in `.env`:

```bash
# Telegram
TELEGRAM_BOT_TOKEN=your_token_here

# LLM Provider (gemini or openai)
LLM_PROVIDER=gemini
LLM_MODEL=gemini-1.5-flash
GEMINI_API_KEY=your_gemini_key

# Firebase
FIREBASE_PROJECT_ID=your_project_id
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-credentials.json

# Python Service
PYTHON_SERVICE_URL=localhost:50051
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete configuration guide.

## Usage

1. Start a chat with your bot on Telegram
2. Send `/start` to see the welcome message
3. Send a recipe link:
   - TikTok: `https://tiktok.com/@user/video/123`
   - YouTube: `https://youtube.com/watch?v=VIDEO_ID`
   - Instagram: `https://instagram.com/p/POST_ID/`
   - Web: Any recipe website with proper markup

4. The bot will:
   - Download the video (if applicable)
   - Extract captions and transcribe audio
   - Use AI to extract recipe details
   - Send you a formatted recipe with ingredients and instructions

5. Use `/recipes` to list your saved recipes

## Architecture Highlights

### Hexagonal Architecture

The application is structured with clear separation of concerns:

- **Domain Layer**: Pure business logic (Recipe, Ingredient, Instruction entities)
- **Application Layer**: Use cases orchestrating domain logic
- **Ports**: Interface definitions for external dependencies
- **Adapters**: Concrete implementations (Telegram, Firebase, Gemini, Python gRPC)

### Easy Component Swapping

Thanks to the port/adapter pattern, you can easily switch:

**LLM Provider:**
```bash
# Switch from Gemini to OpenAI
LLM_PROVIDER=openai
OPENAI_API_KEY=your_key
```

**Transcription Provider:**
```bash
# In python-service/.env
TRANSCRIPTION_PROVIDER=elevenlabs  # default
# TRANSCRIPTION_PROVIDER=google-stt  # requires poetry install --with google-stt
# TRANSCRIPTION_PROVIDER=whisper  # requires poetry install --with whisper
```

**Database:**
Implement the `RecipeRepository` interface for any database (PostgreSQL, MongoDB, etc.)

## Testing

Run unit tests:

```bash
# Go tests
go test ./...

# Python tests (when implemented)
cd python-service
pytest
```

See [TESTING.md](TESTING.md) for detailed testing documentation.

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for:
- Detailed setup instructions
- Production deployment options
- Docker configuration
- Monitoring and troubleshooting
- Cost optimization tips

## Development Status

### ✅ Phase 1: Foundation - COMPLETED
- Go module initialization
- Project structure
- Protocol buffer definitions
- Domain model (entities, value objects)
- Port interfaces
- Configuration structure

### ✅ Phase 2: Python Scraping Service - COMPLETED
- YouTube, TikTok, Instagram scrapers
- Web scraper (BeautifulSoup + schema.org)
- Audio extraction (FFmpeg)
- ElevenLabs Scribe transcription (default)
- Google STT and Whisper (alternative providers)
- gRPC server
- Automatic file cleanup

### ✅ Phase 3: Go Adapters - COMPLETED
- gRPC client for Python service
- Gemini & OpenAI LLM adapters
- Firebase repositories
- All port implementations

### ✅ Phase 4: Application Layer - COMPLETED
- ProcessRecipeLinkCommand (main orchestration)
- User management commands
- Recipe queries
- Comprehensive unit tests

### ✅ Phase 5: Telegram Bot - COMPLETED
- Bot handlers and message routing
- Recipe formatting
- Command handlers (/start, /help, /recipes)
- Main application entry point with DI

### ✅ Phase 6: Deployment - COMPLETED
- Docker configuration
- Docker Compose setup
- Startup scripts
- Complete documentation

## Cost & Free Tiers

- **Google Gemini**: Free tier - 15 requests/min, 1M tokens/day
- **ElevenLabs**: Pay-per-use - affordable per-hour transcription pricing
- **Firebase Firestore**: Free tier - 50K reads, 20K writes/day

With Gemini's free tier and ElevenLabs' affordable pricing, this bot runs at **minimal cost**.

## Cost Optimization

1. **No Video Storage**: Videos are downloaded temporarily, transcribed, then deleted
2. **Transcript Caching**: Store transcripts in Firestore to avoid re-transcribing
3. **Free-Tier Services**: All services have generous free tiers
4. **Efficient LLM Usage**: Optimized prompts to reduce token usage

## Contributing

This is a personal project, but feel free to fork and customize!

## License

MIT License - See LICENSE file for details

## Support

For setup help and troubleshooting, see:
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment guide
- [TESTING.md](TESTING.md) - Testing guide

## Acknowledgments

Built with:
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Google Gemini](https://ai.google.dev/)
- [Firebase](https://firebase.google.com/)
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- [FFmpeg](https://ffmpeg.org/)
