# msgReader Desktop Application

This is the desktop version of the msgReader web application, built using Wails and Go.

## Prerequisites

To build the desktop application, you need to have the following installed:

- Go 1.21 or later
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Platform-specific dependencies for Wails (see [Wails Installation Guide](https://wails.io/docs/gettingstarted/installation))

## Installation

Run the installation script to install the Wails CLI:

```bash
./install-wails.sh
```

## Development

To start the development server:

```bash
./dev.sh
```

This will:
1. Create symbolic links to the web assets
2. Start the Wails development server

## Building

To build the desktop application:

```bash
./build.sh
```

This will:
1. Copy the web assets to the frontend directory
2. Build the Wails application

The built application will be in the `build/bin` directory.

## Architecture

The desktop application uses:
- Go for the backend (file system operations, native dialogs)
- The existing web frontend (HTML, CSS, JavaScript)

The Go backend exposes methods to the frontend for:
- Opening files
- Saving files
- Opening file dialogs
- Saving file dialogs

## Advantages over Electron

This Wails-based desktop application has several advantages over an Electron-based solution:

1. **Smaller binary size** - Wails produces much smaller binaries (typically 10-20MB) compared to Electron (100MB+)
2. **Better performance** - Wails uses the system's native webview, resulting in better performance
3. **Native look and feel** - The application integrates better with the operating system
4. **Lower memory usage** - Wails applications typically use less memory than Electron applications

## Building for Different Platforms

### Windows
```bash
wails build -platform windows/amd64
```

### macOS
```bash
wails build -platform darwin/universal
```

### Linux
```bash
wails build -platform linux/amd64
``` 