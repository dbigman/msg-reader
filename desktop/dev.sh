#!/bin/bash

# Exit on error
set -e

# Create frontend directory if it doesn't exist
mkdir -p frontend

# Create symbolic links to the web assets
ln -sf ../../dist frontend/dist
ln -sf ../../index.html frontend/index.html
ln -sf ../../res frontend/res

# Make sure the FileHandler.js is in the frontend directory
cp -f FileHandler.js frontend/ || echo "FileHandler.js not found in current directory"

# Make sure the FileHandler.js and wails-bridge.js are in the dist directory
mkdir -p frontend/dist
cp -f frontend/FileHandler.js frontend/dist/ || echo "FileHandler.js not found in frontend directory"
cp -f frontend/wails-bridge.js frontend/dist/ || echo "wails-bridge.js not found in frontend directory"

# Start the Wails development server
wails dev 