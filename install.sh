#!/bin/sh
set -e

echo "=== wit AI-Native File Context Maker Installer ==="
echo ""

# 1. Detect OS and Architecture
OS=$(uname -s)
ARCH=$(uname -m)
IS_TERMUX=0

if [ -d "/data/data/com.termux" ]; then
    IS_TERMUX=1
fi

DOWNLOAD_URL=""
DEST_DIR=""
BINARY_NAME="wit"

if [ "$IS_TERMUX" -eq 1 ]; then
    echo "Platform detected: Termux (Android)"
    DOWNLOAD_URL="https://github.com/sapirrior/wit/releases/latest/download/wit-android-arm64"
    DEST_DIR="$PREFIX/bin"
elif [ "$OS" = "Linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        echo "Platform detected: Linux (AMD64)"
        DOWNLOAD_URL="https://github.com/sapirrior/wit/releases/latest/download/wit-linux-amd64"
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        echo "Platform detected: Linux (ARM64)"
        DOWNLOAD_URL="https://github.com/sapirrior/wit/releases/latest/download/wit-linux-arm64"
    else
        echo "Unsupported Linux architecture: $ARCH"
    fi
    DEST_DIR="$HOME/.local/bin"
elif [ "$OS" = "Darwin" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        echo "Platform detected: macOS (Intel AMD64)"
        DOWNLOAD_URL="https://github.com/sapirrior/wit/releases/latest/download/wit-darwin-amd64"
    elif [ "$ARCH" = "arm64" ]; then
        echo "Platform detected: macOS (Apple Silicon ARM64)"
        DOWNLOAD_URL="https://github.com/sapirrior/wit/releases/latest/download/wit-darwin-arm64"
    fi
    DEST_DIR="$HOME/.local/bin"
else
    echo "Unknown platform: $OS ($ARCH)"
    echo "Downloading the latest Windows build to the current directory..."
    DOWNLOAD_URL="https://github.com/sapirrior/wit/releases/latest/download/wit-windows-amd64.exe"
    DEST_DIR="."
    BINARY_NAME="wit.exe"
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo "fatal: Could not determine download URL for your system."
    exit 1
fi

echo "Installation details:"
echo "  Source:      $DOWNLOAD_URL"
echo "  Destination: $DEST_DIR/$BINARY_NAME"
echo ""

# 2. Ask for Permission
printf "Do you want to proceed with the installation? (y/N): "
read -r CONFIRM
if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "Installation canceled by user."
    exit 0
fi

# 3. Create destination directory if needed
if [ "$DEST_DIR" != "." ]; then
    mkdir -p "$DEST_DIR"
fi

# 4. Download and Install
echo "Downloading binary..."
if command -v curl >/dev/null 2>&1; then
    curl -L -o "$DEST_DIR/$BINARY_NAME" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$DEST_DIR/$BINARY_NAME" "$DOWNLOAD_URL"
else
    echo "fatal: Neither curl nor wget is installed. Please install one to proceed."
    exit 1
fi

chmod +x "$DEST_DIR/$BINARY_NAME"

echo ""
echo "=== Success ==="
echo "wit has been successfully installed to: $DEST_DIR/$BINARY_NAME"

if [ "$DEST_DIR" = "$HOME/.local/bin" ]; then
    case ":$PATH:" in
        *:"$HOME/.local/bin":*) ;;
        *)
            echo ""
            echo "Warning: $HOME/.local/bin is not in your PATH."
            echo "To fix this, add the following line to your ~/.bashrc or ~/.zshrc:"
            echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
            ;;
    esac
fi
