.PHONY: build install clean test build-all release lint fmt deps help

BINARY_NAME=claude-insights-agent
VERSION?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DIR=build
INSTALL_DIR?=$(HOME)/.local/bin

# Build flags - inject version at build time
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

# Build for current platform
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agent/

# Install to local bin
install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Update dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

# Build for all platforms (local cross-compilation)
build-all: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/agent/
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/agent/
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/agent/
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/agent/
	@echo "Binaries built in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Create a release (interactive, triggers GitHub Actions + GoReleaser)
release:
	./scripts/release.sh

# Lint
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build for current platform"
	@echo "  install   - Build and install to ~/.local/bin"
	@echo "  test      - Run tests"
	@echo "  clean     - Remove build artifacts"
	@echo "  deps      - Update dependencies"
	@echo "  build-all - Build for all platforms (local)"
	@echo "  release   - Create release (interactive, via GoReleaser)"
	@echo "  lint      - Run linter"
	@echo "  fmt       - Format code"
