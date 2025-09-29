# Browser Render Go - Makefile for easy operations

.PHONY: help build run test docker-build docker-run docker-stop clean

# Default target
help:
	@echo "Browser Render Go - Available commands:"
	@echo ""
	@echo "  make build         - Build the binary"
	@echo "  make run           - Run the server locally"
	@echo "  make test          - Run all tests with coverage"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run with Docker Compose"
	@echo "  make docker-stop   - Stop Docker containers"
	@echo "  make clean         - Clean build artifacts"
	@echo ""

# Build the binary
build:
	go build -o browser_render.exe ./src

# Run locally
run: build
	./browser_render.exe --server=http --debug=false

# Run tests
test:
	go test ./src/... -v -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Docker operations
docker-build:
	docker build -t browser-render-go:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Run in production mode
docker-prod:
	docker-compose -f docker-compose.prod.yml up -d

# Clean up
clean:
	rm -f browser_render.exe *.out coverage.html
	go clean -cache