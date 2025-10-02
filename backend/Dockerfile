# Multi-stage build for Go backend optimized for ARM64/Raspberry Pi

# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations for ARM64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server \
    cmd/server/main.go

# Stage 2: Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/server .

# Copy card data and assets
COPY --from=builder /build/assets ./assets

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose backend port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/api/health || exit 1

# Run the binary
CMD ["./server"]
