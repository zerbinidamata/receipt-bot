# Agent Build & Development Instructions

## Project Setup

### Prerequisites
- Go 1.21+
- Python 3.11+
- Docker & Docker Compose
- Firebase project with Firestore enabled
- API keys: Telegram Bot Token, Gemini API Key, Google Cloud credentials

### Environment Variables
Create a `.env` file in project root:
```bash
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
FIREBASE_PROJECT_ID=your_firebase_project_id
FIREBASE_CREDENTIALS_PATH=/path/to/firebase-credentials.json
GEMINI_API_KEY=your_gemini_api_key
GOOGLE_CLOUD_CREDENTIALS_PATH=/path/to/gcloud-credentials.json
PYTHON_SERVICE_URL=localhost:50051
LLM_PROVIDER=gemini
TRANSCRIPTION_PROVIDER=google
```

### Initial Setup
```bash
# Install Go dependencies
go mod download

# Install Python dependencies
cd python-service
pip install -r requirements.txt
cd ..

# Generate protobuf code (if proto files changed)
make generate-proto
```

## Running the Project

### Development Mode (Local)
```bash
# Terminal 1: Start Python service
cd python-service
python run_server.py

# Terminal 2: Start Go bot
go run cmd/bot/main.go
```

### Docker Mode
```bash
# Build and start all services
docker-compose -f deployments/docker-compose.yml up --build

# Stop services
docker-compose -f deployments/docker-compose.yml down
```

## Running Tests

### Go Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain/recipe/...

# Run with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Python Tests
```bash
cd python-service
pytest

# With coverage
pytest --cov=src --cov-report=html
```

## Build Commands

### Build Go Binary
```bash
go build -o bin/receipt-bot cmd/bot/main.go
```

### Build Docker Images
```bash
# Go service
docker build -f deployments/go-service.Dockerfile -t receipt-bot-go .

# Python service
docker build -f deployments/python-service.Dockerfile -t receipt-bot-python .
```

### Generate Protobuf Code
```bash
# Requires protoc and plugins installed
protoc --go_out=. --go-grpc_out=. proto/scraper.proto
python -m grpc_tools.protoc -I. --python_out=python-service/src --grpc_python_out=python-service/src proto/scraper.proto
```

## Key Learnings & Gotchas

### Go Service
- Always run `go mod tidy` after adding/removing imports
- Firebase client requires valid credentials even for tests (use emulator or mock)
- Telegram bot token must be valid format or initialization fails
- gRPC client retries automatically on connection failure

### Python Service
- yt-dlp requires ffmpeg installed on system
- Instagram scraping may require cookies for some content
- Large video files are auto-cleaned after processing
- Google STT has 1-minute limit for synchronous calls (use async for longer)

### Docker
- Python service must be healthy before Go service starts (depends_on with healthcheck)
- Volume mount for credentials files in docker-compose
- Use host.docker.internal for localhost references on Mac/Windows

### Firebase/Firestore
- Collection names: `recipes`, `users`
- Firestore indexes may be needed for complex queries (check console for auto-suggestions)
- Timestamps stored as Firestore Timestamp type, converted to time.Time in Go

### LLM Integration
- Gemini model names: use "gemini-1.5-flash" or "gemini-1.5-pro"
- Temperature 0.1 for consistent JSON extraction
- Always validate JSON response matches expected schema
- Retry on rate limit errors with exponential backoff

## Quality Standards

### Test Coverage
- **Minimum 85%** coverage on new code
- Domain layer: aim for 95%+
- Application layer: aim for 85%+
- Adapters: test complex logic, mock external services

### Git Workflow
- Create feature branches from `main`
- Use conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
- Run tests before committing
- Keep commits atomic and focused

### Documentation
- Document public functions in Go with godoc comments
- Update README.md for user-facing changes
- Update AGENT.md for build/test changes
- Keep specs/ updated with requirement changes

### Code Review Checklist
- [ ] Tests pass locally
- [ ] No hardcoded secrets or credentials
- [ ] Error handling is explicit and informative
- [ ] Follows existing code patterns
- [ ] Domain logic has no external dependencies
- [ ] New ports/adapters follow interface pattern

## Useful Commands Reference

```bash
# Check for unused dependencies
go mod tidy

# Format Go code
go fmt ./...

# Lint Go code (if golangci-lint installed)
golangci-lint run

# Check Python formatting
black --check python-service/

# View Firestore data (requires firebase CLI)
firebase firestore:get recipes

# Test Telegram bot webhook
curl -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe"
```
