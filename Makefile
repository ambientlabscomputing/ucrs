.PHONY: run build test docs clean help generate-keys

# Default target
help:
	@echo "Underleaf Capability Registry Service - Make targets:"
	@echo "  make generate-keys  - Generate Ed25519 signing keys (run once)"
	@echo "  make run            - Run the service locally"
	@echo "  make build          - Build the binary"
	@echo "  make test           - Run tests"
	@echo "  make docs           - Generate Swagger documentation"
	@echo "  make clean          - Clean build artifacts"

# Generate Ed25519 signing keys (only needs to be run once)
generate-keys:
	@echo "Generating Ed25519 signing keys..."
	@cd ucrs && go run scripts/generate_keys.go
	@echo "✓ Keys generated in keys/ directory"
	@echo "⚠️  Keep these keys secure! They are used to sign registry snapshots"

# Run the service
run:
	export CONFIG_PATH=${PWD}/config.yaml && \
	cd ucrs && go run main.go

# Build the service binary
build:
	cd ucrs && go build -o capability_registry_service main.go
	@echo "Binary created: ucrs/capability_registry_service"

# Run tests
test:
	cd ucrs && go test ./... -v

# Generate Swagger documentation
docs:
	cd ucrs && swag init -g main.go --output ./docs
	@echo "Swagger docs generated in ucrs/docs/"

# Clean build artifacts
clean:
	rm -f ucrs/capability_registry_service
	rm -rf ucrs/docs/
	@echo "Cleaned build artifacts"

# Install dependencies
deps:
	cd ucrs && go mod download
	cd ucrs && go mod tidy

# Run with Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f capability_registry
