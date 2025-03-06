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

## File Associations

The desktop application can be registered as the default application for opening .msg and .eml files. This allows you to double-click on these files in your file explorer to open them directly in msgReader.

### Registering File Associations

There are two ways to register file associations:

1. **From within the application**: Click the "Register as Default App" button on the welcome screen.

2. **Manually**:
   - **Windows**: Right-click on a .msg or .eml file, select "Open with" > "Choose another app", select msgReader, and check "Always use this app to open .msg/.eml files".
   - **macOS**: Right-click on a .msg or .eml file, select "Get Info", change the "Open with" option to msgReader, and click "Change All".
   - **Linux**: The application will create the necessary desktop entry and MIME type files when you click the "Register as Default App" button.

### Opening Files from Command Line

You can also open files directly from the command line:

```bash
./msgReader path/to/file.msg
```

## Architecture

The desktop application uses:
- Go for the backend (file system operations, native dialogs)
- The existing web frontend (HTML, CSS, JavaScript)

The Go backend exposes methods to the frontend for:
- Opening files
- Saving files
- Opening file dialogs
- Saving file dialogs
- Registering file associations

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