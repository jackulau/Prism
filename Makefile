.PHONY: all build dev prod clean test help

# Default target
all: help

# Development
dev: ## Start development environment
	docker-compose -f docker-compose.dev.yml up --build

dev-down: ## Stop development environment
	docker-compose -f docker-compose.dev.yml down

# Production
prod: ## Start production environment
	docker-compose up -d --build

prod-down: ## Stop production environment
	docker-compose down

prod-logs: ## View production logs
	docker-compose logs -f

# Build
build: build-backend build-frontend ## Build all components

build-backend: ## Build backend
	cd backend && go build -o prism ./cmd/server

build-frontend: ## Build frontend
	cd frontend && npm run build

build-sandbox: ## Build sandbox images
	docker build -t prism-sandbox-base ./sandbox/base
	docker build -t prism-sandbox-python ./sandbox/python
	docker build -t prism-sandbox-node ./sandbox/node
	docker build -t prism-sandbox-shell ./sandbox/shell

# Run locally
run-backend: ## Run backend locally
	cd backend && go run ./cmd/server

run-frontend: ## Run frontend locally
	cd frontend && npm run dev

# Testing
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	cd backend && go test ./...

test-frontend: ## Run frontend tests
	cd frontend && npm test

# Database
db-migrate: ## Run database migrations
	cd backend && go run ./cmd/server -migrate

# Clean
clean: ## Clean build artifacts
	rm -rf backend/prism
	rm -rf backend/tmp
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# Setup
setup: ## Initial setup
	cp .env.example .env
	cd frontend && npm install
	cd backend && go mod download

# Generate
generate-key: ## Generate encryption key
	@openssl rand -hex 32

# Help
help: ## Show this help
	@echo "Prism - Open Source Web Agent"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
