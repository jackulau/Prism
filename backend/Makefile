# Backend Makefile

.PHONY: all build run dev test clean typegen

# Default target
all: build

# Build the main server binary
build:
	go build -o bin/server ./cmd/server

# Run the server
run: build
	./bin/server

# Run with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf tmp/

# Generate TypeScript types from Go structs
typegen:
	@echo "Generating TypeScript types..."
	go run cmd/typegen/main.go ../frontend/src/types
	@echo "TypeScript types generated successfully"

# Install development dependencies
deps:
	go mod download
	go install github.com/cosmtrek/air@latest

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Help
help:
	@echo "Available targets:"
	@echo "  make build   - Build the server binary"
	@echo "  make run     - Build and run the server"
	@echo "  make dev     - Run with hot reload (air)"
	@echo "  make test    - Run tests"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make typegen - Generate TypeScript types from Go structs"
	@echo "  make deps    - Install development dependencies"
	@echo "  make lint    - Run linter"
	@echo "  make fmt     - Format code"
