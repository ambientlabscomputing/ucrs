# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache --no-scripts git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . ./

# Generate swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g main.go --output ./docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o capability_registry_service main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS (--no-scripts avoids trigger issues in ARM64 QEMU)
RUN apk --no-cache add --no-scripts ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/capability_registry_service .

# Create a non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose the application port
EXPOSE 8083

# Run the application
CMD ["./capability_registry_service"]
