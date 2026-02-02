.PHONY: run build test docs clean help

# Default target
help:
	@echo "Underleaf Capability Registry Service - Make targets:"
	@echo "  make run      - Run the service locally"
	@echo "  make build    - Build the binary"
	@echo "  make test     - Run tests"
	@echo "  make docs     - Generate Swagger documentation"
	@echo "  make clean    - Clean build artifacts"

# Run the service
run:
	export CONFIG_PATH=${PWD}/config.yaml && \
	cd service && go run main.go

# Build the service binary
build:
	cd service && go build -o capability_registry_service main.go
	@echo "Binary created: service/capability_registry_service"

# Run tests
test:
	cd service && go test ./... -v

# Generate Swagger documentation
docs:
	cd service && swag init -g main.go --output ./docs
	@echo "Swagger docs generated in service/docs/"

# Clean build artifacts
clean:
	rm -f service/capability_registry_service
	rm -rf service/docs/
	@echo "Cleaned build artifacts"

# Install dependencies
deps:
	cd service && go mod download
	cd service && go mod tidy

# Run with Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f capability_registry

# Generate Ed25519 signing keys
generate-keys:
	@mkdir -p keys
	@echo "Generating Ed25519 signing key pair..."
	@cd service && go run scripts/generate_keys.go
	@echo "Keys generated in keys/ directory"
