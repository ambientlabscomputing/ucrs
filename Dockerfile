# Build stage
FROM golang:1.24-alpine AS builder

# Install dependencies (minimal set)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Install swag as a separate layer (better caching)
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

# Copy go.mod and go.sum first (better layer caching)
COPY service/go.mod service/go.sum ./

# Download dependencies (cached if go.mod/sum unchanged)
RUN go mod download

# Copy source code
COPY service/ ./

# Generate Swagger docs
RUN swag init -g main.go --output ./docs

# Build with optimizations and parallel builds
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o capability_registry_service main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy binary from builder
COPY --from=builder /build/capability_registry_service .

# Copy config example (optional, for reference)
COPY config.yaml.example .

# Create directories for keys and logs
RUN mkdir -p /app/keys /app/logs && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Run the application
CMD ["./capability_registry_service"]
