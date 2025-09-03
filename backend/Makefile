# Terraforming Mars Backend Makefile

.PHONY: test test-verbose test-service test-websocket test-coverage clean build run help

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run all tests with verbose output"
	@echo "  test-service  - Run only service layer tests"
	@echo "  test-websocket- Run only websocket tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  build         - Build the server binary"
	@echo "  run           - Run the development server"
	@echo "  clean         - Clean build artifacts"

# Run all tests
test:
	@echo "Running all backend tests..."
	go test ./test/...

# Run all tests with verbose output
test-verbose:
	@echo "Running all backend tests (verbose)..."
	go test -v ./test/...

# Run only service tests
test-service:
	@echo "Running service layer tests..."
	go test -v ./test/service/...

# Run only websocket tests
test-websocket:
	@echo "Running websocket tests..."
	go test -v ./test/delivery/websocket/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage report..."
	go test -v -coverprofile=coverage.out ./test/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the server binary
build:
	@echo "Building server binary..."
	go build -o bin/server cmd/server/main.go

# Run the development server
run:
	@echo "Starting development server..."
	go run cmd/server/main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f bin/server
	rm -f coverage.out
	rm -f coverage.html

# Test individual components
test-game-service:
	@echo "Running GameService tests..."
	go test -v ./test/service/ -run TestGameService

test-hub:
	@echo "Running WebSocket Hub tests..."
	go test -v ./test/delivery/websocket/ -run TestHub

test-client:
	@echo "Running WebSocket Client tests..."
	go test -v ./test/delivery/websocket/ -run TestClient

test-messages:
	@echo "Running WebSocket Message tests..."
	go test -v ./test/delivery/websocket/ -run TestMessage

# Quick test - run tests without verbose output for faster feedback
test-quick:
	@echo "Running quick test suite..."
	@go test ./test/service/... && echo "✅ Service tests passed" || echo "❌ Service tests failed"
	@go test ./test/delivery/websocket/hub_test.go && echo "✅ Hub tests passed" || echo "❌ Hub tests failed"
	@go test ./test/delivery/websocket/message_test.go && echo "✅ Message tests passed" || echo "❌ Message tests failed"
	@go test ./test/delivery/websocket/client_test.go && echo "✅ Client tests passed" || echo "❌ Client tests failed"

# Run tests continuously on file changes (requires entr: apt install entr)
test-watch:
	@echo "Watching for changes and running tests..."
	find . -name "*.go" | entr -c make test-quick