#!/bin/bash

# Exit on error
set -e

# Create frontend directory if it doesn't exist
mkdir -p frontend

# Copy the web assets to the frontend directory
cp -r ../dist frontend/
cp ../index.html frontend/
cp -r ../res frontend/

# Build the Wails application
wails build 