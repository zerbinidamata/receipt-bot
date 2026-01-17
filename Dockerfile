# Go service Dockerfile
# Multi-stage build for smaller image size

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/bot

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/main .

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose port (Cloud Run uses PORT env variable)
EXPOSE 8080

# Run the application
CMD ["./main"]
