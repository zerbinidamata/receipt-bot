#!/bin/bash

# Docker startup script for Receipt Bot
# This script starts both services using Docker Compose

set -e

echo "🚀 Starting Receipt Bot with Docker..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found"
    echo "📝 Please copy .env.example to .env and configure your credentials"
    exit 1
fi

# Source .env file to check required variables
set -a
source .env
set +a

# Check required environment variables
REQUIRED_VARS=(
    "TELEGRAM_BOT_TOKEN"
    "GEMINI_API_KEY"
    "FIREBASE_PROJECT_ID"
    "FIREBASE_CREDENTIALS_PATH"
    "GOOGLE_CLOUD_CREDENTIALS_PATH"
)

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Error: $var is not set in .env"
        exit 1
    fi
done

# Check if credential files exist
if [ ! -f "$FIREBASE_CREDENTIALS_PATH" ]; then
    echo "❌ Error: Firebase credentials file not found at $FIREBASE_CREDENTIALS_PATH"
    exit 1
fi

if [ ! -f "$GOOGLE_CLOUD_CREDENTIALS_PATH" ]; then
    echo "❌ Error: Google Cloud credentials file not found at $GOOGLE_CLOUD_CREDENTIALS_PATH"
    exit 1
fi

# Create temp directory
mkdir -p /tmp/recipe-bot

echo "✅ Prerequisites check passed"
echo ""
echo "📦 Building and starting services..."

# Navigate to deployments directory
cd "$(dirname "$0")/.."

# Build and start services
docker-compose -f deployments/docker-compose.yml up --build -d

echo ""
echo "✅ Services started successfully!"
echo ""
echo "📊 Service status:"
docker-compose -f deployments/docker-compose.yml ps
echo ""
echo "📝 View logs with:"
echo "   docker-compose -f deployments/docker-compose.yml logs -f"
echo ""
echo "🛑 Stop services with:"
echo "   docker-compose -f deployments/docker-compose.yml down"
