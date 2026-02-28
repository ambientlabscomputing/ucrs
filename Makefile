.PHONY: run build test docs clean help generate-keys

## help: Generate and display this message
help:
	@echo "UCRS - Underleaf Capability Registry Service"
	@echo ""
	@echo "Available targets:"
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = "## |: "}; {printf "  \033[36m%-20s\033[0m %s\n", $$2, $$3}'

## generate-keys: Generate Ed25519 signing keys (only needs to be run once)
generate-keys:
	@echo "Generating Ed25519 signing keys..."
	@go run scripts/generate_keys.go
	@echo "✓ Keys generated in keys/ directory"
	@echo "⚠️  Keep these keys secure! They are used to sign registry snapshots"

## run: Run the service
run:
	export CONFIG_PATH=${PWD}/config.yaml && \
	go run main.go

## build: Build the service binary
build:
	go build -o capability_registry_service main.go
	@echo "Binary created: capability_registry_service"

## test: Run tests
test:
	go test ./... -v

## docs: Generate Swagger documentation
docs:
	swag init -g main.go --output ./docs
	@echo "Swagger docs generated in docs/"

## clean: Clean build artifacts
clean:
	rm -f capability_registry_service
	rm -rf docs/
	@echo "Cleaned build artifacts"

## deps: Install dependencies
deps:
	go mod download
	go mod tidy

## docker-up: Run with Docker Compose
docker-up:
	docker-compose up -d

## docker-down: Stop Docker Compose
docker-down:
	docker-compose down

## docker-logs: View logs
docker-logs:
	docker-compose logs -f capability_registry

## docker-build: Build from Dockerfile
docker-build:
	docker build -t ambientlabsjose/ucrs:develop .

## docker-publish: Push new dev image
docker-publish: docker-build
	docker push ambientlabsjose/ucrs:develop
