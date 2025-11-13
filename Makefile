# Terraforming Mars - Unified Development Makefile
# Run from project root directory

.PHONY: help run frontend backend kill lint typecheck test test-backend test-frontend test-verbose test-coverage clean build format format-backend format-frontend format-json install-cli generate parse-cards

# Default target - show help
help:
	@echo "🚀 Terraforming Mars Development Commands"
	@echo ""
	@echo "🎯 Main Commands:"
	@echo "  make run          - Run both frontend and backend servers"
	@echo "  make frontend     - Run frontend development server (port 3000)"
	@echo "  make backend      - Run backend development server with auto-restart (port 3001)"
	@echo "  make kill         - Kill all frontend and backend development processes"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test         - Run all tests (backend + frontend)"
	@echo "  make test-backend - Run backend tests only"
	@echo "  make test-verbose - Run backend tests with verbose output"
	@echo "  make test-coverage- Run backend tests with coverage report"
	@echo ""
	@echo "🔧 Code Quality:"
	@echo "  make lint         - Run all linters (backend + frontend)"
	@echo "  make typecheck    - Run TypeScript type checking"
	@echo "  make format       - Format all code (Go + TypeScript + JSON)"
	@echo "  make generate     - Generate TypeScript types from Go structs"
	@echo "  make parse-cards  - Parse card data from CSV to JSON"
	@echo ""
	@echo "🏗️  Build & Deploy:"
	@echo "  make build        - Build production binaries"
	@echo "  make clean        - Clean build artifacts"
	@echo ""

# Main development commands
run:
	@echo "🚀 Starting both servers..."
	@echo "Frontend: http://localhost:3000"
	@echo "Backend: http://localhost:3001 (with auto-reload)"
	cd frontend && npm start & cd backend && air

frontend:
	@echo "🎨 Starting frontend development server..."
	cd frontend && npm start

backend:
	@echo "🔄 Starting backend development server with auto-restart..."
	@echo "   Watching for changes in backend/ directory"
	cd backend && air

kill:
	@echo "🛑 Killing all development servers..."
	./kill-servers.sh

# Testing commands
test: test-backend

test-backend:
	@echo "🧪 Running backend tests..."
	cd backend && go test ./test/...

test-frontend:
	@echo "🧪 Running frontend tests..."
	@echo "⚠️  No test script found in frontend package.json"
	@echo "ℹ️  Running linter instead..."
	cd frontend && npm run lint

test-verbose:
	@echo "🧪 Running backend tests (verbose)..."
	cd backend && go test -v ./test/...

test-coverage:
	@echo "🧪 Running backend tests with coverage..."
	cd backend && go test -v -coverprofile=coverage.out -coverpkg=./internal/... ./test/...
	@cd backend && if [ -s coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html && \
		echo "📊 Coverage report generated: backend/coverage.html"; \
	else \
		echo "⚠️ No coverage data generated - skipping HTML report"; \
	fi
	@echo "✅ Test coverage completed"

# Quick test commands for development
test-quick:
	@echo "⚡ Running quick test suite..."
	@cd backend && go test ./test/service/... && echo "✅ Service tests passed" || echo "❌ Service tests failed"
	@cd backend && go test ./test/delivery/websocket/hub_test.go && echo "✅ Hub tests passed" || echo "❌ Hub tests failed"
	@cd backend && go test ./test/delivery/websocket/message_test.go && echo "✅ Message tests passed" || echo "❌ Message tests failed"
	@cd backend && go test ./test/delivery/websocket/client_test.go && echo "✅ Client tests passed" || echo "❌ Client tests failed"

# Code quality commands
lint: lint-backend lint-frontend typecheck

typecheck:
	@echo "🔍 Running TypeScript type checking..."
	cd frontend && npm run typecheck
	@echo "✅ Type checking complete"

lint-backend:
	@echo "🔍 Running backend linting (Go fmt)..."
	cd backend && go fmt ./...
	@echo "✅ Backend formatting complete"

lint-frontend:
	@echo "🔍 Running frontend linting (oxlint)..."
	cd frontend && npm run lint
	@echo "✅ Frontend linting complete"

format: format-backend format-frontend format-json

format-backend:
	@echo "🎨 Formatting backend Go code..."
	cd backend && find . -name "*.go" -exec gofmt -s -w {} \;
	@echo "✅ Backend formatting complete"

format-frontend:
	@echo "🎨 Formatting frontend TypeScript code..."
	cd frontend && npm run format:write
	@echo "✅ Frontend formatting complete"

format-json:
	@echo "🎨 Formatting all JSON files..."
	cd frontend && npx prettier --write "../**/*.json"
	@echo "✅ JSON formatting complete"

# Build and deployment
build: build-backend build-frontend

build-backend:
	@echo "🏗️  Building backend binary..."
	cd backend && go build -o bin/server cmd/server/main.go
	@echo "✅ Backend binary: backend/bin/server"

build-frontend:
	@echo "🏗️  Building frontend for production..."
	cd frontend && npm run build
	@echo "✅ Frontend build: frontend/dist/"

# Cleanup
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd backend && rm -f bin/server bin/tm coverage.out coverage.html
	cd frontend && rm -rf dist build
	cd backend && go clean
	@echo "✅ Cleanup complete"

# Development helpers
dev-setup:
	@echo "🔧 Setting up development environment..."
	cd backend && go mod tidy
	cd frontend && npm install
	@echo "✅ Development setup complete"

# Type generation
generate:
	@echo "🔄 Generating TypeScript types from Go structs..."
	cd backend && tygo generate
	@echo "✅ TypeScript types generated"

# Card data parsing
parse-cards:
	@echo "🃏 Parsing card data from CSV files..."
	cd backend && go run tools/parse_cards.go assets/terraforming_mars_cards.json
	@echo "✅ Card data parsed to backend/assets/terraforming_mars_cards.json"

# Watch for changes (requires entr: apt install entr)
test-watch:
	@echo "👀 Watching for Go file changes and running tests..."
	cd backend && find . -name "*.go" | entr -c make test-quick