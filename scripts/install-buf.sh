#!/bin/bash

# Buf installation script for different platforms

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)
        echo "Installing buf for Linux..."
        curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64" -o /usr/local/bin/buf
        chmod +x /usr/local/bin/buf
        ;;
    Darwin*)
        echo "Installing buf for macOS..."
        if command -v brew &> /dev/null; then
            brew install bufbuild/buf/buf
        else
            curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Darwin-x86_64" -o /usr/local/bin/buf
            chmod +x /usr/local/bin/buf
        fi
        ;;
    MINGW* | MSYS* | CYGWIN*)
        echo "Installing buf for Windows..."
        curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Windows-x86_64.exe" -o buf.exe
        echo "Please move buf.exe to a directory in your PATH"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "Buf installation completed!"
buf --version