# syntax=docker/dockerfile:1.7-labs
# Multi-stage Dockerfile for CostScope

# Multi-stage Dockerfile for CostScope

# Stage 1: Build stage
FROM golang:1.26-alpine AS builder

# Build arguments (can be passed during docker build)
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown
ARG GOVERSION=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make bash

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with BuildKit cache mounts
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download

# Copy source code
COPY . .

# Build the application (stamped via build args)
# Use ldflags to stamp version metadata when provided
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo \
    -ldflags "-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE} -X main.GoVersion=${GOVERSION} -extldflags '-Wl,-dead_strip -Wl,-x'" \
    -trimpath \
    -o costscope .

# Stage 2: Final stage
FROM alpine:3.23

# Metadata arguments (propagated as labels)
ARG VERSION=dev
ARG COMMIT=none

# Install minimal runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S costscope && \
  adduser -u 1001 -S costscope -G costscope

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/costscope .

# Copy configuration directory if present
COPY --from=builder /app/configs ./configs

# Change ownership
RUN chown -R costscope:costscope /app

# Switch to non-root user
USER costscope

# Expose port
EXPOSE 8080

# OCI labels with sensible defaults; docker build should pass --label or --build-arg to set real values
LABEL org.opencontainers.image.source="https://github.com/your/repo"
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.revision=${COMMIT}

# Health check (uses busybox wget provided by base image)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=5 \
  CMD wget -q -O - http://127.0.0.1:8080/health/ready >/dev/null 2>&1 || exit 1

# Run the enterprise API by default for production containers
CMD ["./costscope", "enterprise", "--host", "0.0.0.0", "--port", "8080"]
