# Python service Dockerfile
FROM python:3.14

# Install system dependencies and build tools
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    build-essential \
    gcc \
    g++ \
    libxml2-dev \
    libxslt1-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Poetry
RUN pip install --no-cache-dir poetry

# Set working directory
WORKDIR /app

# Copy Poetry configuration files first for better caching
COPY python-service/pyproject.toml python-service/poetry.lock* ./

# Configure Poetry and install dependencies (without dev dependencies, skip root package)
RUN poetry config virtualenvs.create false \
    && poetry install --no-interaction --no-ansi --no-root --without dev

# Copy Python service code
COPY python-service/ .

# Install the project itself and create temp directory
RUN poetry install --no-interaction --no-ansi --only-root \
    && mkdir -p /tmp/recipe-bot

# Remove build dependencies to reduce image size (keep runtime libraries)
RUN apt-get purge -y --auto-remove \
    build-essential \
    gcc \
    g++ \
    libxml2-dev \
    libxslt1-dev \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV GRPC_PORT=50051
ENV LOG_LEVEL=INFO
ENV TEMP_DIR=/tmp/recipe-bot

# Expose gRPC port
EXPOSE 50051

# Run the server
CMD ["python", "run_server.py"]
