.PHONY: dev dev-backend dev-frontend build clean frontend backend install run test

# Version info (can be overridden)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Air for hot reload..."
	@go install github.com/air-verse/air@latest
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install
	@echo "Done! Run 'make dev' to start development servers."

# Development (run both backend and frontend with hot reload)
dev:
	@echo "Starting development servers..."
	@echo "Backend: http://localhost:3000 (API)"
	@echo "Frontend: http://localhost:5173 (Vite dev server)"
	@echo ""
	@make -j2 dev-backend dev-frontend

dev-backend:
	@air -c .air.toml

dev-frontend:
	@cd frontend && npm run dev

# Build for production
build: frontend backend
	@echo ""
	@echo "Build complete: ./bin/central-logs"
	@echo "Run with: ./bin/central-logs"

frontend:
	@echo "Building frontend..."
	@cd frontend && npm ci && npm run build
	@rm -rf web/dist
	@cp -r frontend/dist web/dist
	@echo "Frontend built and copied to web/dist"

backend:
	@echo "Building backend..."
	@echo "Version: $(VERSION) | Commit: $(GIT_COMMIT) | Build: $(BUILD_TIME)"
	@CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o bin/central-logs ./cmd/server
	@echo "Backend built: bin/central-logs"

# Cross compile for Linux (AMD64)
build-linux: frontend
	@echo "Building for Linux (amd64)..."
	@echo "Version: $(VERSION) | Commit: $(GIT_COMMIT) | Build: $(BUILD_TIME)"
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc go build -ldflags "$(LDFLAGS)" -o bin/central-logs-linux-amd64 ./cmd/server
	@echo "Built: bin/central-logs-linux-amd64"

# Run production binary
run:
	@./bin/central-logs

# Run tests
test:
	@go test -v ./...

# Clean build artifacts
clean:
	@rm -rf bin/ web/dist/ frontend/dist/ tmp/ build-errors.log
	@echo "Cleaned build artifacts"

# Clean everything including node_modules
clean-all: clean
	@rm -rf frontend/node_modules/
	@echo "Cleaned all including node_modules"

# Generate default config
config:
	@go run ./cmd/server -generate-config

# Database operations
db-reset:
	@rm -f data/central-logs.db
	@echo "Database deleted. Will be recreated on next run."

# Docker commands
docker-up:
	@docker-compose up -d
	@echo "Redis started"

docker-down:
	@docker-compose down
	@echo "Docker services stopped"

# Help
help:
	@echo "Central Logs - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make install      - Install all dependencies"
	@echo "  make dev          - Start dev servers (Air + Vite)"
	@echo "  make dev-backend  - Start only backend with Air"
	@echo "  make dev-frontend - Start only frontend with Vite"
	@echo ""
	@echo "Build:"
	@echo "  make build        - Build production binary"
	@echo "  make build-linux  - Cross-compile for Linux"
	@echo "  make frontend     - Build frontend only"
	@echo "  make backend      - Build backend only"
	@echo ""
	@echo "Run:"
	@echo "  make run          - Run production binary"
	@echo "  make test         - Run tests"
	@echo ""
	@echo "Clean:"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make clean-all    - Clean everything"
	@echo "  make db-reset     - Delete database"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up    - Start Redis"
	@echo "  make docker-down  - Stop Docker services"
