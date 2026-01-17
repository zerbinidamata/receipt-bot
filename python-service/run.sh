#!/bin/bash
set -e

echo "🐍 Starting Python Service with Poetry..."
echo ""

# Check if Poetry is installed
if ! command -v poetry &> /dev/null; then
    echo "❌ Error: Poetry is not installed"
    echo "Please install Poetry: https://python-poetry.org/docs/#installation"
    echo "  curl -sSL https://install.python-poetry.org | python3 -"
    exit 1
fi

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  Warning: .env file not found"
    echo "Creating .env from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ Created .env file. Please edit it with your credentials."
        echo "  nano .env"
        exit 1
    else
        echo "❌ Error: .env.example not found"
        exit 1
    fi
fi

# Check if dependencies are installed
if [ ! -f "poetry.lock" ]; then
    echo "📦 Installing dependencies with Poetry..."
    poetry install
else
    echo "✅ Dependencies already installed"
fi

# Check if proto files exist, generate if not
if [ ! -f "src/scraper_pb2.py" ] || [ ! -f "src/scraper_pb2_grpc.py" ]; then
    echo "📝 Generating protobuf files..."
    make generate
fi

echo "🚀 Starting gRPC server..."
echo ""
poetry run python run_server.py
