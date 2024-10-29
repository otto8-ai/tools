#!/bin/bash

set -e

REPO="gptscript-ai/knowledge"
INSTALL_DIR="/usr/local/bin"

# Function to determine the OS and architecture
get_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        arm64)
            ARCH="arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    echo "${OS}-${ARCH}"
}

# Function to download the latest release
download_latest_release() {
    OS_ARCH=$1
    LATEST_RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"

    DOWNLOAD_URL=$(curl -s $LATEST_RELEASE_URL | grep "browser_download_url.*$OS_ARCH" | cut -d '"' -f 4)

    if [[ -z "$DOWNLOAD_URL" ]]; then
        echo "No binary found for $OS_ARCH"
        exit 1
    fi

    TEMP_DIR=$(mktemp -d)
    TEMP_FILE="$TEMP_DIR/knowledge"

    echo "Downloading $DOWNLOAD_URL..."
    curl -sL "$DOWNLOAD_URL" -o "$TEMP_FILE"

    chmod +x $TEMP_FILE

    mv $TEMP_FILE "$INSTALL_DIR/knowledge"

    rm -rf "$TEMP_DIR"

    echo "Installed knowledge to $INSTALL_DIR/knowledge"
}

# Ensure the script is run as root
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root. Please run with sudo."
    exit 1
fi

OS_ARCH=$(get_os_arch)
download_latest_release "$OS_ARCH"
