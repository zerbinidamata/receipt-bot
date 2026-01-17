#!/bin/bash
set -e

echo "🤖 Starting Recipe Bot..."
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found"
    echo "Please copy .env.example to .env and fill in your credentials:"
    echo "  cp .env.example .env"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    echo "Please install Go 1.23 or higher: https://golang.org/dl/"
    exit 1
fi

# Check if Python service is running
echo "🔍 Checking if Python service is running..."
if ! nc -z localhost 50051 2>/dev/null; then
    echo "⚠️  Warning: Python service doesn't seem to be running on localhost:50051"
    echo "Please start the Python service first:"
    echo "  cd python-service"
    echo "  poetry run python run_server.py"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo "📦 Installing Go dependencies..."
go mod download

echo "🚀 Starting Telegram bot..."
echo ""
go run cmd/bot/main.go
