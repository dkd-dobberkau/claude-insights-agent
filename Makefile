.PHONY: build install clean test release

BINARY_NAME=claude-insights-agent
VERSION?=0.2.0
BUILD_DIR=build
INSTALL_DIR?=$(HOME)/.local/bin

# Build flags
LDFLAGS=-ldflags "-s -w"

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

# Build for all platforms
release: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/agent/
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/agent/
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/agent/
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/agent/
	@echo "Binaries built in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Lint
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build for current platform"
	@echo "  install  - Build and install to ~/.local/bin"
	@echo "  test     - Run tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  deps     - Update dependencies"
	@echo "  release  - Build for all platforms"
	@echo "  lint     - Run linter"
	@echo "  fmt      - Format code"
