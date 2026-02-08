# Build stage with Go build cache mounting
FROM golang:1.24-alpine AS builder

# Install minimal dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Pre-install swag in a separate layer for maximum caching
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

# Copy dependency files first (best cache hit rate)
COPY ucrs/go.mod ucrs/go.sum ./

# Download with cache mount for go modules
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY ucrs/ ./

# Generate docs
RUN swag init -g main.go --output ./docs

# Build with cache mounts and optimizations
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o capability_registry_service main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

COPY --from=builder /build/capability_registry_service .
COPY config.yaml.example .

RUN mkdir -p /app/keys /app/logs && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8083

CMD ["./capability_registry_service"]
