#!/bin/bash

# Build script for Audio Track Switcher
# This script builds the Go backend and prepares it for Tauri

echo "Building Go backend..."

cd go-backend

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o ../src-tauri/audio-track-backend.exe main.go

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o ../src-tauri/audio-track-backend main.go

# Build for macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o ../src-tauri/audio-track-backend-darwin main.go

cd ..

echo "Go backend build complete!"
echo "Binaries created in src-tauri directory"
