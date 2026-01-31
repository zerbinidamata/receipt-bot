# Receipt-Bot Agent Build Instructions

## Project Overview
Receipt-Bot is a Telegram bot that extracts recipes from TikTok, YouTube, Instagram, and web pages using AI.

**Architecture:** Hexagonal Architecture (Go + Python microservices + Firebase)

## Project Setup

### Go Service (Main Bot)
```bash
# Install Go dependencies
go mod download

# Build the bot
go build -o main ./cmd/bot

# Run locally (requires .env file)
./main
```

### Python Service (Scraping/Transcription)
```bash
cd python-service

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run gRPC server
python run_server.py
```

### Docker (Full Stack)
```bash
# Build and run all services
docker-compose -f deployments/docker-compose.yml up --build

# Or build individually
docker build -f deployments/go-service.Dockerfile -t receipt-bot-go .
docker build -f deployments/python-service.Dockerfile -t receipt-bot-python .
```

## Running Tests

### Go Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/domain/recipe/...
go test ./internal/application/...

# Run with verbose output
go test -v ./...
```

### Python Tests
```bash
cd python-service
pytest
pytest --cov=src tests/
```

## Environment Variables
Required in `.env` file:
```
TELEGRAM_BOT_TOKEN=your_token
FIREBASE_PROJECT_ID=your_project
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json
GEMINI_API_KEY=your_key
PYTHON_SERVICE_URL=localhost:50051
```

## Key Directories
```
cmd/bot/                    # Application entry point
internal/
  domain/                   # Business logic (entities, value objects)
    recipe/                 # Recipe aggregate
    user/                   # User entity
  application/              # Use cases (commands, queries)
    command/               # Write operations
    query/                 # Read operations
    dto/                   # Data transfer objects
  ports/                    # Interface definitions
  adapters/                 # External implementations
    firebase/              # Firestore repository
    llm/                   # Gemini/OpenAI adapters
    telegram/              # Bot handlers
    python/                # gRPC client
  config/                   # Configuration management
python-service/             # Python microservice
  src/scrapers/            # Platform-specific scrapers
  src/video/               # Video/audio processing
proto/                      # gRPC definitions
deployments/                # Docker configs
```

## Development Workflow

1. **Before changes**: Search codebase to understand existing patterns
2. **Make changes**: Follow hexagonal architecture (ports → adapters)
3. **Test**: Run `go test ./...` after each implementation
4. **Commit**: Use conventional commits (`feat:`, `fix:`, `test:`)

## Key Learnings

### Build Optimizations
- Go builds are fast, no special optimization needed
- Python service can be slow to start (yt-dlp initialization)
- Use `go build -ldflags="-s -w"` for smaller binaries

### Testing Patterns
- Domain tests use pure unit tests (no mocks needed)
- Application tests use interface mocks
- Integration tests require Firebase emulator or test project

### Common Issues
- Firebase credentials must be valid JSON file
- Gemini API has rate limits (adjust retry logic)
- TikTok scraping may fail due to anti-bot measures

## Feature Development Quality Standards

### Testing Requirements
- Minimum 85% code coverage for new code
- All tests must pass
- Unit tests for domain logic
- Integration tests for adapters

### Git Workflow
- Use conventional commits
- Push after completing each feature
- Update .ralph/@fix_plan.md with progress

### Documentation
- Update inline comments for complex logic
- Keep this file updated with new patterns
- Document breaking changes

## Current Development Focus
See `.ralph/@fix_plan.md` for prioritized task list.

**Phase 1**: Auto-Categorization (in progress)
**Phase 2**: Ingredient Matching
**Phase 3**: Export to Notion/Obsidian
