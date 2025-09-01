#!/bin/bash
# FakeSMTP Installation Script

set -e

# Configuration
BINARY_NAME="fakesmtp"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="fakesmtp"
USER="fakesmtp"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) log_error "Unsupported architecture: $ARCH" && exit 1 ;;
    esac
    
    log_info "Detected platform: $OS-$ARCH"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/your-username/fakesmtp/releases/latest/download/fakesmtp-$OS-$ARCH"
    
    log_info "Downloading $BINARY_NAME..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$BINARY_NAME" "$DOWNLOAD_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$BINARY_NAME" "$DOWNLOAD_URL"
    else
        log_error "curl or wget is required"
        exit 1
    fi
    
    chmod +x "$BINARY_NAME"
}

# Install binary
install_binary() {
    log_info "Installing $BINARY_NAME to $INSTALL_DIR..."
    
    if [[ $EUID -ne 0 ]]; then
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    else
        mv "$BINARY_NAME" "$INSTALL_DIR/"
    fi
    
    log_info "Installation completed!"
}

# Create systemd service (Linux)
create_systemd_service() {
    if [[ "$OS" != "linux" ]]; then
        return
    fi
    
    log_info "Creating systemd service..."
    
    cat > /tmp/fakesmtp.service << EOF
[Unit]
Description=FakeSMTP Email Testing Server
After=network.target

[Service]
Type=simple
User=$USER
Group=$USER
ExecStart=$INSTALL_DIR/$BINARY_NAME --daemon --smtp-port=2525 --http-port=8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    if [[ $EUID -ne 0 ]]; then
        sudo mv /tmp/fakesmtp.service /etc/systemd/system/
        sudo systemctl daemon-reload
        sudo systemctl enable fakesmtp
        log_info "Systemd service created. Use 'sudo systemctl start fakesmtp' to start"
    else
        mv /tmp/fakesmtp.service /etc/systemd/system/
        systemctl daemon-reload
        systemctl enable fakesmtp
        log_info "Systemd service created. Use 'systemctl start fakesmtp' to start"
    fi
}

# Create user for service
create_user() {
    if [[ "$OS" != "linux" ]]; then
        return
    fi
    
    if ! id "$USER" >/dev/null 2>&1; then
        log_info "Creating user $USER..."
        if [[ $EUID -ne 0 ]]; then
            sudo useradd -r -s /bin/false -d /nonexistent $USER
        else
            useradd -r -s /bin/false -d /nonexistent $USER
        fi
    fi
}

# Main installation
main() {
    log_info "Starting FakeSMTP installation..."
    
    detect_platform
    download_binary
    install_binary
    
    if [[ "$1" == "--service" ]]; then
        create_user
        create_systemd_service
    fi
    
    log_info "FakeSMTP installed successfully!"
    log_info "Run 'fakesmtp --help' to see available options"
    
    if [[ "$1" == "--service" ]]; then
        log_info "Service installed. Start with: sudo systemctl start fakesmtp"
    fi
}

# Run main function
main "$@"