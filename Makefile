# Makefile for ygo - Y.js CRDT library for Go

.PHONY: all build build-mock test test-mock clean setup-yffi deps

# Default target
all: build

# Setup yffi dependencies
setup-yffi:
	@echo "🦀 Setting up official yffi..."
	./scripts/setup-yffi.sh

# Install Go dependencies
deps:
	@echo "📦 Installing Go dependencies..."
	go mod download
	go mod tidy

# Build with yffi support (size-optimized)
build: setup-yffi
	@echo "🔧 Building ygo with yffi support (size-optimized)..."
	go build -ldflags="-s -w" -v ./...

# Build mock version (no Rust dependencies)
build-mock:
	@echo "🔧 Building ygo with mock implementation..."
	go build -tags mock -v ./...

# Run tests with yffi
test: setup-yffi
	@echo "🧪 Running tests with yffi..."
	go test -v ./...

# Run tests with mock implementation
test-mock:
	@echo "🧪 Running tests with mock implementation..."
	go test -tags mock -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf lib/
	rm -rf y-crdt/
	go clean -cache
	go clean -testcache

# Install as Go module dependency
install:
	@echo "📦 Installing ygo module..."
	go install ./...

# Run benchmarks
bench: setup-yffi
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run benchmarks with mock
bench-mock:
	@echo "⚡ Running benchmarks with mock..."
	go test -tags mock -bench=. -benchmem ./...

# Check if yffi is working
check-yffi:
	@echo "🔍 Checking yffi installation..."
	@if [ -f "lib/libyrs.dylib" ] || [ -f "lib/libyrs.so" ]; then \
		echo "✅ yffi library found"; \
		ls -la lib/; \
	else \
		echo "❌ yffi library not found - run 'make setup-yffi'"; \
		exit 1; \
	fi

# Development helpers
dev-test: test-mock
	@echo "🚀 Development testing complete"

dev-build: build-mock
	@echo "🚀 Development build complete"

# Show available make targets
help:
	@echo "Available targets:"
	@echo "  all          - Default build (with yffi)"
	@echo "  build        - Build with yffi support"
	@echo "  build-mock   - Build with mock implementation only"
	@echo "  test         - Run tests with yffi"
	@echo "  test-mock    - Run tests with mock implementation"
	@echo "  setup-yffi   - Setup official yffi from y-crdt"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install Go dependencies"
	@echo "  install      - Install as Go module"
	@echo "  bench        - Run benchmarks with yffi"
	@echo "  bench-mock   - Run benchmarks with mock"
	@echo "  check-yffi   - Check yffi installation"
	@echo "  dev-test     - Quick development test (mock)"
	@echo "  dev-build    - Quick development build (mock)"
	@echo "  help         - Show this help"