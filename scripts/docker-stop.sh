#!/bin/bash

# Docker stop script for Receipt Bot

set -e

echo "🛑 Stopping Receipt Bot services..."

cd "$(dirname "$0")/.."

docker-compose -f deployments/docker-compose.yml down

echo "✅ Services stopped successfully!"
