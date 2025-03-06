#!/bin/bash

# Exit on error
set -e

# Create frontend directory if it doesn't exist
mkdir -p frontend

# Create symbolic links to the web assets
ln -sf ../../dist frontend/dist
ln -sf ../../index.html frontend/index.html
ln -sf ../../res frontend/res

# Start the Wails development server
wails dev 