#!/bin/bash

# Exit on error
set -e

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Install Wails CLI
echo "Installing Wails CLI..."
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Check if Wails was installed successfully
if ! command -v wails &> /dev/null; then
    echo "Failed to install Wails CLI. Please check your Go installation and try again."
    exit 1
fi

echo "Wails CLI installed successfully!"
echo "You can now run './dev.sh' to start the development server or './build.sh' to build the desktop application." 