#!/bin/bash

# Exit on error
set -e

# Create frontend directory if it doesn't exist
mkdir -p frontend

# Copy the web assets to the frontend directory
cp -r ../dist frontend/
cp ../index.html frontend/
cp -r ../res frontend/

# Make sure the FileHandler.js and wails-bridge.js are in the frontend directory
# First check if they exist in the current directory
if [ -f "FileHandler.js" ]; then
    cp -f FileHandler.js frontend/
else
    echo "FileHandler.js not found in current directory, using existing file in frontend directory"
fi

if [ -f "wails-bridge.js" ]; then
    cp -f wails-bridge.js frontend/
else
    echo "wails-bridge.js not found in current directory, using existing file in frontend directory"
fi

# Copy our custom main.js file
if [ -f "main.js" ]; then
    cp -f main.js frontend/
else
    echo "main.js not found in current directory, using existing file in frontend directory"
fi

# Make sure the FileHandler.js, wails-bridge.js, and main.js are in the dist directory
mkdir -p frontend/dist
if [ -f "frontend/FileHandler.js" ]; then
    cp -f frontend/FileHandler.js frontend/dist/
else
    echo "FileHandler.js not found in frontend directory"
fi

if [ -f "frontend/wails-bridge.js" ]; then
    cp -f frontend/wails-bridge.js frontend/dist/
else
    echo "wails-bridge.js not found in frontend directory"
fi

if [ -f "frontend/main.js" ]; then
    cp -f frontend/main.js frontend/dist/
else
    echo "main.js not found in frontend directory"
fi

# Build the Wails application with developer tools enabled
wails build -devtools -debug