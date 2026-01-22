#!/bin/bash
set -e

# Claude Insights Agent Installer
# Downloads and installs the latest release

VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
GITHUB_REPO="dkd/claude-insights-agent"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    case "$OS" in
        darwin)
            PLATFORM="darwin-$ARCH"
            ;;
        linux)
            PLATFORM="linux-$ARCH"
            ;;
        *)
            error "Unsupported OS: $OS"
            ;;
    esac

    info "Detected platform: $PLATFORM"
}

# Get latest version if not specified
get_version() {
    if [ "$VERSION" = "latest" ]; then
        info "Fetching latest version..."
        VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v?([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            error "Could not determine latest version"
        fi
    fi
    info "Installing version: $VERSION"
}

# Download binary
download() {
    DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/v$VERSION/claude-insights-agent-$PLATFORM"

    info "Downloading from $DOWNLOAD_URL..."

    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/claude-insights-agent"

    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE" || error "Download failed"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE" || error "Download failed"
    else
        error "curl or wget required"
    fi

    chmod +x "$TMP_FILE"
}

# Install binary
install_binary() {
    info "Installing to $INSTALL_DIR..."

    mkdir -p "$INSTALL_DIR"
    mv "$TMP_FILE" "$INSTALL_DIR/claude-insights-agent"

    # Clean up
    rm -rf "$TMP_DIR"

    info "Installed successfully!"
}

# Check PATH
check_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "Add $INSTALL_DIR to your PATH:"
        echo ""
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        echo "Add this to your ~/.bashrc or ~/.zshrc"
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo "============================================"
    echo "  Claude Insights Agent installed!"
    echo "============================================"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Initialize configuration:"
    echo "     claude-insights-agent init"
    echo ""
    echo "  2. Start the agent:"
    echo "     claude-insights-agent run"
    echo ""
    echo "  Or run a one-time sync:"
    echo "     claude-insights-agent sync"
    echo ""
    echo "For more info:"
    echo "     claude-insights-agent help"
    echo ""
}

# Build from source (for development)
build_from_source() {
    info "Building from source..."

    if ! command -v go &> /dev/null; then
        error "Go is required to build from source"
    fi

    go build -o "$INSTALL_DIR/claude-insights-agent" ./cmd/agent/
    info "Built and installed to $INSTALL_DIR/claude-insights-agent"
}

# Main
main() {
    echo ""
    echo "Claude Insights Agent Installer"
    echo "================================"
    echo ""

    # Check if running from source directory
    if [ -f "go.mod" ] && [ -f "cmd/agent/main.go" ]; then
        info "Source directory detected, building from source..."
        build_from_source
    else
        detect_platform
        get_version
        download
        install_binary
    fi

    check_path
    print_next_steps
}

main "$@"
