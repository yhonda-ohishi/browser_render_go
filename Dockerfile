# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (skip go mod tidy as imports are internal)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o browser_render ./src

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    chromium \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ttf-freefont \
    sqlite

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/browser_render .

# Create data directory
RUN mkdir -p /app/data /app/logs

# Set Chrome path for Rod
ENV ROD_BROWSER=/usr/bin/chromium-browser

# Expose ports
EXPOSE 50051 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./browser_render", "--headless=true"]