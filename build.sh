#!/bin/bash

# Build script for the tr CLI tool

set -e

echo "Building TR - English-Spanish Translation CLI Tool"

# Clean previous builds
echo "Cleaning previous builds..."
rm -f tr tr.exe tr-linux tr-macos tr-windows.exe

# Get dependencies
echo "Downloading dependencies..."
go mod download

# Build for current platform
echo "Building for current platform..."
go build -o tr ./cmd/tr

# Build for multiple platforms
echo "Building for multiple platforms..."

# Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o tr-windows.exe ./cmd/tr

# macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o tr-macos ./cmd/tr

# Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o tr-linux ./cmd/tr

echo "Build complete!"
echo "Executables created:"
echo "  - tr (current platform)"
echo "  - tr-windows.exe (Windows)"
echo "  - tr-macos (macOS)"
echo "  - tr-linux (Linux)"

# Test the current platform build
echo ""
echo "Testing the build..."
./tr --version
echo "Build test successful!"
