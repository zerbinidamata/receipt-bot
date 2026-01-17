#!/bin/bash

# Docker logs script for Receipt Bot

cd "$(dirname "$0")/.."

# Check if a service name was provided
if [ -z "$1" ]; then
    echo "📝 Showing logs for all services (use Ctrl+C to stop)..."
    docker-compose -f deployments/docker-compose.yml logs -f
else
    echo "📝 Showing logs for $1 (use Ctrl+C to stop)..."
    docker-compose -f deployments/docker-compose.yml logs -f "$1"
fi
