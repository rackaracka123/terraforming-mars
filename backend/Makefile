# Backend Testing Makefile
# Run from backend/ directory

.PHONY: test test-verbose test-coverage test-quick test-service test-integration test-websocket format lint clean

# Main test command
test:
	@echo "🧪 Running all backend tests..."
	go test ./test/...

test-backend:
	@echo "🧪 Running all backend tests..."
	go test ./test/...

# Test with verbose output
test-verbose:
	@echo "🧪 Running backend tests (verbose)..."
	go test -v ./test/...

# Test with coverage report
test-coverage:
	@echo "🧪 Running backend tests with coverage..."
	go test -v -coverprofile=coverage.out ./test/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Quick test suite for development iteration
test-quick:
	@echo "⚡ Running quick test suite..."
	@go test ./test/service/... && echo "✅ Service tests passed" || echo "❌ Service tests failed"
	@go test ./test/delivery/websocket/hub_test.go && echo "✅ Hub tests passed" || echo "❌ Hub tests failed"
	@go test ./test/delivery/websocket/message_test.go && echo "✅ Message tests passed" || echo "❌ Message tests failed"
	@go test ./test/delivery/websocket/client_test.go && echo "✅ Client tests passed" || echo "❌ Client tests failed"

# Test specific categories
test-service:
	@echo "🔧 Running service layer tests..."
	go test ./test/service/...

test-integration:
	@echo "🔗 Running integration tests..."
	go test ./test/integration/...

test-websocket:
	@echo "🌐 Running WebSocket tests..."
	go test ./test/delivery/websocket/...

# Code quality
format:
	@echo "🎨 Formatting Go code..."
	gofmt -s -w .
	@echo "✅ Backend formatting complete"

lint:
	@echo "🔍 Running Go fmt..."
	go fmt ./...
	@echo "✅ Backend linting complete"

# Cleanup
clean:
	@echo "🧹 Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	go clean
	@echo "✅ Cleanup complete"

# Run specific test files (usage: make test-file FILE=test/service/game_service_test.go)
test-file:
	@echo "🎯 Running specific test file: $(FILE)"
	go test ./$(FILE)

# Run tests with race detection
test-race:
	@echo "🏁 Running tests with race detection..."
	go test -race ./test/...

# Watch for changes and run tests (requires entr: apt install entr)
test-watch:
	@echo "👀 Watching for Go file changes and running tests..."
	find . -name "*.go" | entr -c make test-quick