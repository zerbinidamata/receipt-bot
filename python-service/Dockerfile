# Python service Dockerfile
FROM python:3.13-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Install poetry
RUN pip install --no-cache-dir poetry

# Copy poetry files
COPY pyproject.toml poetry.lock ./

# Configure poetry to not create virtual environment
RUN poetry config virtualenvs.create false

# Install dependencies (without dev dependencies, skip installing the project itself)
RUN poetry install --only main --no-root --no-interaction --no-ansi

# Copy application code
COPY src/ ./src/
COPY run_server.py ./

# Create non-root user
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Create temp directory for video processing
RUN mkdir -p /tmp/recipe-bot

# Expose gRPC port
EXPOSE 50051

# Run the gRPC server
CMD ["python", "run_server.py"]
