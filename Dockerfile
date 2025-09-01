# Multi-stage build for minimal final image

# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application (without CGO for portability)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=docker" \
    -a -installsuffix cgo \
    -o fakesmtp \
    ./cmd/server

# Stage 2: Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget bash

# Create non-root user
RUN addgroup -g 1000 fakesmtp && \
    adduser -u 1000 -G fakesmtp -s /bin/sh -D fakesmtp

# Create directories
RUN mkdir -p /app/data /app/logs && \
    chown -R fakesmtp:fakesmtp /app

# Copy binary and scripts
COPY --from=builder /app/fakesmtp /usr/local/bin/fakesmtp
COPY scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Make scripts executable
RUN chmod +x /usr/local/bin/fakesmtp /usr/local/bin/docker-entrypoint.sh

# Switch to non-root user
USER fakesmtp

WORKDIR /app

# Expose ports
EXPOSE 2525 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/stats || exit 1

# Set default environment variables
ENV SMTP_PORT=2525
ENV HTTP_PORT=8080
ENV DB_PATH=/app/data/emails.db
ENV LOG_PATH=/app/logs/fakesmtp.log
ENV CLEAR_ON_SHUTDOWN=true

# Volumes for persistent data
VOLUME ["/app/data", "/app/logs"]

# Use entrypoint script
ENTRYPOINT ["docker-entrypoint.sh"]