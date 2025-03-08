package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                  context.Context
	filesToOpenOnStartup []string
	initialized          bool
	pendingFiles         []string
	filesMutex           sync.Mutex
	appReady             bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		pendingFiles: []string{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("App startup called, context saved")

	// Show the window immediately
	wailsRuntime.WindowShow(ctx)

	// Mark the app as initialized AND ready immediately
	a.initialized = true
	a.appReady = true
	fmt.Println("App marked as ready on startup")

	// We'll handle files from startup directly
	if len(a.filesToOpenOnStartup) > 0 {
		startupFiles := append([]string{}, a.filesToOpenOnStartup...)
		a.filesToOpenOnStartup = nil
		fmt.Println("Files to process at startup:", startupFiles)

		// CRITICAL: For immediate display of the first file, we process it immediately
		// instead of waiting for a delay
		if len(startupFiles) > 0 {
			firstFile := startupFiles[0]
			fmt.Println("Immediately processing first file:", firstFile)

			// Use direct JavaScript approach for immediate visibility
			a.executeOpenFirstFileJavaScript(firstFile)

			// Process any remaining files with a small delay
			if len(startupFiles) > 1 {
				remainingFiles := startupFiles[1:]
				go func() {
					// Short delay for remaining files
					time.Sleep(500 * time.Millisecond)
					for _, file := range remainingFiles {
						fmt.Println("Processing additional startup file:", file)
						a.executeOpenFileJavaScript(file)
					}
				}()
			}
		}
	}

	// Still listen for frontend-ready as a backup
	wailsRuntime.EventsOn(ctx, "frontend-ready", func(optionalData ...interface{}) {
		fmt.Println("Received frontend-ready event (frontend UI is fully initialized)")

		// Emit an event to tell the frontend we're ready to process files
		wailsRuntime.EventsEmit(a.ctx, "backend-ready", true)
	})
}

// handleFileOpen handles macOS file open events
func (a *App) handleFileOpen(filePath string) {
	fmt.Printf("Received macOS file open event for: %s\n", filePath)

	// Get the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Printf("Error getting absolute path for %s: %v\n", filePath, err)
		absPath = filePath
	}

	// Verify the file exists and is accessible
	if _, err := os.Stat(absPath); err != nil {
		fmt.Printf("Error accessing file %s: %v\n", absPath, err)
		return
	}

	// IMPORTANT: We're treating the app as ready once we reach here
	// This ensures files can be opened even if the frontend-ready event failed
	if !a.appReady {
		a.appReady = true
		fmt.Println("App marked as ready on file open event")
	}

	// Store the file in our cache
	a.filesMutex.Lock()
	a.pendingFiles = append(a.pendingFiles, absPath)
	a.filesMutex.Unlock()

	// Direct processing approach
	fmt.Printf("Processing file directly: %s\n", absPath)

	// Check if this is the first file message in the app
	// For the first file, we want to use the high-priority method
	a.filesMutex.Lock()
	isFirstFile := len(a.pendingFiles) <= 1
	a.filesMutex.Unlock()

	if isFirstFile {
		// Use the highest priority method for the first file
		fmt.Println("Using high-priority method for first file")
		a.executeOpenFirstFileJavaScript(absPath)
	} else {
		// Standard method for additional files
		a.executeOpenFileJavaScript(absPath)
	}

	// Also emit an event to the frontend
	wailsRuntime.EventsEmit(a.ctx, "open-file-now", absPath)
}

// executeOpenFileJavaScript executes JavaScript directly to open a file
func (a *App) executeOpenFileJavaScript(filePath string) {
	// Read the file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s for direct execution: %v\n", filePath, err)
		return
	}

	// Get the file extension
	extension := filepath.Ext(filePath)
	if extension != "" {
		extension = extension[1:] // Remove the leading dot
	}
	extension = strings.ToLower(extension)

	// Get the filename
	fileName := filepath.Base(filePath)

	fmt.Printf("Executing direct JavaScript to open file: %s\n", fileName)

	// Base64 encode the file data for safe JavaScript transfer
	base64Data := base64.StdEncoding.EncodeToString(data)

	// Create JavaScript to handle the file
	js := fmt.Sprintf(`
		console.log("Direct JavaScript execution to open file: %s");
		
		function processFile() {
			try {
				// Check if the app is ready
				if (!window.app || !window.app.fileHandler || !window.app.uiManager) {
					console.log("App not ready, will retry in 200ms");
					setTimeout(processFile, 200);
					return;
				}
				
				// Convert base64 data to ArrayBuffer
				const binaryString = window.atob("%s");
				const bytes = new Uint8Array(binaryString.length);
				for (let i = 0; i < binaryString.length; i++) {
					bytes[i] = binaryString.charCodeAt(i);
				}
				const buffer = bytes.buffer;
				
				// Process the file as if it was dropped into the app
				const extension = "%s";
				const fileName = "%s";
				
				console.log("Processing file:", fileName);
				
				// Make sure UI is visible
				window.app.uiManager.showAppContainer();
				
				// Extract the message info
				let msgInfo;
				if (extension === 'msg' && window.extractMsg) {
					console.log("Extracting MSG file...");
					msgInfo = window.extractMsg(buffer);
				} else if (extension === 'eml' && window.extractEml) {
					console.log("Extracting EML file...");
					msgInfo = window.extractEml(buffer);
				} else {
					console.error("Unsupported file extension:", extension);
					return;
				}
				
				if (!msgInfo) {
					console.error("Failed to extract message info");
					return;
				}
				
				console.log("Message extracted successfully");
				
				// Add the message
				const message = window.app.messageHandler.addMessage(msgInfo, fileName);
				
				// Update the message list
				window.app.uiManager.updateMessageList();
				
				// If only 1 message exists, show it
				const messages = window.app.messageHandler.getMessages();
				if (messages.length === 1) {
					// This is the first message, show it
					window.app.uiManager.showMessage(message);
				}
				
				console.log("File processed successfully via direct JavaScript execution");
			} catch (error) {
				console.error("Error in direct file processing:", error);
			}
		}
		
		// Start processing
		processFile();
	`, fileName, base64Data, extension, fileName)

	// Execute the JavaScript
	wailsRuntime.WindowExecJS(a.ctx, js)
}

// executeOpenFirstFileJavaScript executes JavaScript specifically optimized for opening the first file
// with highest priority and immediate visibility
func (a *App) executeOpenFirstFileJavaScript(filePath string) {
	// Read the file data
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s for direct execution: %v\n", filePath, err)
		return
	}

	// Get the file extension
	extension := filepath.Ext(filePath)
	if extension != "" {
		extension = extension[1:] // Remove the leading dot
	}
	extension = strings.ToLower(extension)

	// Get the filename
	fileName := filepath.Base(filePath)

	fmt.Printf("Opening FIRST FILE directly via JavaScript: %s\n", fileName)

	// Base64 encode the file data for safe JavaScript transfer
	base64Data := base64.StdEncoding.EncodeToString(data)

	// Create JavaScript to handle the file with high priority
	js := fmt.Sprintf(`
		console.log("PRIORITY: Opening first file directly: %s");
		
		// Function to process the file with high priority
		function processFirstFile() {
			try {
				// Check if the app is ready
				if (!window.app || !window.app.fileHandler || !window.app.uiManager) {
					console.log("App not fully ready yet, will retry in 100ms");
					setTimeout(processFirstFile, 100);
					return;
				}
				
				// Convert base64 data to ArrayBuffer
				const binaryString = window.atob("%s");
				const bytes = new Uint8Array(binaryString.length);
				for (let i = 0; i < binaryString.length; i++) {
					bytes[i] = binaryString.charCodeAt(i);
				}
				const buffer = bytes.buffer;
				
				// Process the file as if it was dropped into the app
				const extension = "%s";
				const fileName = "%s";
				
				console.log("Processing primary file:", fileName);
				
				// First ensure UI is visible
				window.app.uiManager.showAppContainer();
				
				// Extract the message info
				let msgInfo;
				if (extension === 'msg' && window.extractMsg) {
					console.log("Extracting MSG file...");
					msgInfo = window.extractMsg(buffer);
				} else if (extension === 'eml' && window.extractEml) {
					console.log("Extracting EML file...");
					msgInfo = window.extractEml(buffer);
				} else {
					console.error("Unsupported file extension:", extension);
					return;
				}
				
				if (!msgInfo) {
					console.error("Failed to extract message info");
					return;
				}
				
				console.log("Message extracted successfully");
				
				// Add the message
				const message = window.app.messageHandler.addMessage(msgInfo, fileName);
				
				// Update the message list
				window.app.uiManager.updateMessageList();
				
				// Show the message IMMEDIATELY
				window.app.uiManager.showMessage(message);
				
				console.log("PRIORITY FILE processed and displayed successfully");
			} catch (error) {
				console.error("Error in first file processing:", error);
			}
		}
		
		// Start processing immediately
		processFirstFile();
	`, fileName, base64Data, extension, fileName)

	// Execute the JavaScript
	wailsRuntime.WindowExecJS(a.ctx, js)
}

// GetPendingFiles returns any files waiting to be processed
func (a *App) GetPendingFiles() []string {
	a.filesMutex.Lock()
	defer a.filesMutex.Unlock()

	// Combine all pending files
	allPendingFiles := append([]string{}, a.filesToOpenOnStartup...)
	allPendingFiles = append(allPendingFiles, a.pendingFiles...)

	// Clear the caches after returning them
	a.filesToOpenOnStartup = []string{}
	a.pendingFiles = []string{}

	fmt.Printf("GetPendingFiles returning %d files: %v\n", len(allPendingFiles), allPendingFiles)
	return allPendingFiles
}

// OpenFile allows the frontend to request opening a file
func (a *App) OpenFile(filePath string) ([]byte, error) {
	fmt.Println("OpenFile called with path:", filePath)

	// Verify file exists before trying to read
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", filePath)
			return nil, fmt.Errorf("file does not exist: %w", err)
		}
		fmt.Printf("Error checking file: %s - %v\n", filePath, err)
		return nil, fmt.Errorf("error checking file: %w", err)
	}

	// File exists, read it
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %s - %v\n", filePath, err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("File read successfully: %s, size: %d bytes\n", filePath, len(data))
	return data, nil
}

// ReadFile is an alias for OpenFile for better naming compatibility
func (a *App) ReadFile(filePath string) ([]byte, error) {
	fmt.Println("ReadFile called (alias for OpenFile) with path:", filePath)
	return a.OpenFile(filePath)
}

// ReadFileHeader reads just the first n bytes of a file to verify it exists and is readable
func (a *App) ReadFileHeader(filePath string, bytesToRead int) ([]byte, error) {
	fmt.Printf("ReadFileHeader called for file: %s, reading %d bytes\n", filePath, bytesToRead)

	// Verify file exists before trying to read
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", filePath)
			return nil, fmt.Errorf("file does not exist: %w", err)
		}
		fmt.Printf("Error checking file: %s - %v\n", filePath, err)
		return nil, fmt.Errorf("error checking file: %w", err)
	}

	// Check file size and adjust bytes to read if necessary
	fileSize := fileInfo.Size()
	if fileSize == 0 {
		fmt.Printf("File is empty: %s\n", filePath)
		return []byte{}, nil
	}

	if bytesToRead <= 0 {
		bytesToRead = 10 // Default to reading 10 bytes
	}

	if int64(bytesToRead) > fileSize {
		bytesToRead = int(fileSize)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %s - %v\n", filePath, err)
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes
	header := make([]byte, bytesToRead)
	n, err := file.Read(header)
	if err != nil {
		fmt.Printf("Error reading file header: %s - %v\n", filePath, err)
		return nil, fmt.Errorf("error reading file header: %w", err)
	}

	// Return only the bytes that were actually read
	fmt.Printf("Successfully read %d bytes from file: %s\n", n, filePath)
	return header[:n], nil
}

// SaveFile allows the frontend to save a file
func (a *App) SaveFile(filePath string, data []byte) error {
	fmt.Println("SaveFile called with path:", filePath)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Println("Error creating directory:", err)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Println("Error writing file:", err)
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println("File saved successfully")
	return nil
}

// OpenFileDialog opens a file dialog and returns the selected file paths
func (a *App) OpenFileDialog() ([]string, error) {
	fmt.Println("OpenFileDialog called")
	files, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select MSG or EML files",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "Email Files (*.msg, *.eml)",
				Pattern:     "*.msg;*.eml",
			},
			{
				DisplayName: "MSG Files (*.msg)",
				Pattern:     "*.msg",
			},
			{
				DisplayName: "EML Files (*.eml)",
				Pattern:     "*.eml",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		fmt.Println("Error opening file dialog:", err)
	} else {
		fmt.Println("Files selected from dialog:", files)
	}

	return files, err
}

// SaveFileDialog opens a save file dialog and returns the selected file path
func (a *App) SaveFileDialog(defaultFilename string) (string, error) {
	fmt.Println("SaveFileDialog called with default filename:", defaultFilename)
	file, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Save File",
		DefaultFilename: defaultFilename,
	})

	if err != nil {
		fmt.Println("Error opening save dialog:", err)
	} else {
		fmt.Println("File selected for saving:", file)
	}

	return file, err
}

// RegisterFileAssociations registers the application as the default handler for .msg and .eml files
func (a *App) RegisterFileAssociations() (bool, error) {
	fmt.Println("RegisterFileAssociations called")
	switch runtime.GOOS {
	case "windows":
		return a.registerFileAssociationsWindows()
	case "darwin":
		return a.registerFileAssociationsMacOS()
	case "linux":
		return a.registerFileAssociationsLinux()
	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// registerFileAssociationsWindows registers file associations on Windows
func (a *App) registerFileAssociationsWindows() (bool, error) {
	fmt.Println("Registering file associations on Windows")
	// Get the path to the executable
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Println("Executable path:", exePath)

	// Register .msg file association
	msgCmd := exec.Command("reg", "add", "HKCU\\Software\\Classes\\.msg", "/ve", "/t", "REG_SZ", "/d", "msgReader.MSG", "/f")
	if err := msgCmd.Run(); err != nil {
		fmt.Println("Error registering .msg file association:", err)
		return false, fmt.Errorf("failed to register .msg file association: %w", err)
	}

	// Register .eml file association
	emlCmd := exec.Command("reg", "add", "HKCU\\Software\\Classes\\.eml", "/ve", "/t", "REG_SZ", "/d", "msgReader.EML", "/f")
	if err := emlCmd.Run(); err != nil {
		fmt.Println("Error registering .eml file association:", err)
		return false, fmt.Errorf("failed to register .eml file association: %w", err)
	}

	// Create program entry for .msg
	msgProgCmd := exec.Command("reg", "add", "HKCU\\Software\\Classes\\msgReader.MSG", "/ve", "/t", "REG_SZ", "/d", "MSG Email File", "/f")
	if err := msgProgCmd.Run(); err != nil {
		fmt.Println("Error creating program entry for .msg:", err)
		return false, fmt.Errorf("failed to create program entry for .msg: %w", err)
	}

	// Create program entry for .eml
	emlProgCmd := exec.Command("reg", "add", "HKCU\\Software\\Classes\\msgReader.EML", "/ve", "/t", "REG_SZ", "/d", "EML Email File", "/f")
	if err := emlProgCmd.Run(); err != nil {
		fmt.Println("Error creating program entry for .eml:", err)
		return false, fmt.Errorf("failed to create program entry for .eml: %w", err)
	}

	// Create command for .msg
	msgOpenCmd := exec.Command("reg", "add", "HKCU\\Software\\Classes\\msgReader.MSG\\shell\\open\\command", "/ve", "/t", "REG_SZ", "/d", fmt.Sprintf("\"%s\" \"%%1\"", exePath), "/f")
	if err := msgOpenCmd.Run(); err != nil {
		fmt.Println("Error creating command for .msg:", err)
		return false, fmt.Errorf("failed to create command for .msg: %w", err)
	}

	// Create command for .eml
	emlOpenCmd := exec.Command("reg", "add", "HKCU\\Software\\Classes\\msgReader.EML\\shell\\open\\command", "/ve", "/t", "REG_SZ", "/d", fmt.Sprintf("\"%s\" \"%%1\"", exePath), "/f")
	if err := emlOpenCmd.Run(); err != nil {
		fmt.Println("Error creating command for .eml:", err)
		return false, fmt.Errorf("failed to create command for .eml: %w", err)
	}

	fmt.Println("File associations registered successfully on Windows")
	return true, nil
}

// registerFileAssociationsMacOS registers file associations on macOS
func (a *App) registerFileAssociationsMacOS() (bool, error) {
	fmt.Println("Registering file associations on macOS")
	// On macOS, file associations are defined in the Info.plist file
	// This is handled during the build process in the wails.json file
	// We'll just show a message to the user
	wailsRuntime.MessageDialog(a.ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.InfoDialog,
		Title:   "File Associations",
		Message: "On macOS, file associations are handled by the system. Please right-click on a .msg or .eml file, select 'Get Info', change the 'Open with' option to msgReader, and click 'Change All'.",
	})

	fmt.Println("Showed macOS file association dialog")
	return true, nil
}

// registerFileAssociationsLinux registers file associations on Linux
func (a *App) registerFileAssociationsLinux() (bool, error) {
	fmt.Println("Registering file associations on Linux")
	// Get the path to the executable
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	fmt.Println("Executable path:", exePath)

	// Create desktop entry file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return false, fmt.Errorf("failed to get user home directory: %w", err)
	}

	fmt.Println("User home directory:", homeDir)

	// Create applications directory if it doesn't exist
	applicationsDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(applicationsDir, 0755); err != nil {
		fmt.Println("Error creating applications directory:", err)
		return false, fmt.Errorf("failed to create applications directory: %w", err)
	}

	fmt.Println("Applications directory:", applicationsDir)

	// Create desktop entry file
	desktopEntryPath := filepath.Join(applicationsDir, "msgreader.desktop")
	desktopEntry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=msgReader
Exec=%s %%f
Icon=mail-message
MimeType=application/vnd.ms-outlook;message/rfc822;
Categories=Office;Email;
Comment=MSG and EML file viewer
Terminal=false
`, exePath)

	fmt.Println("Writing desktop entry to:", desktopEntryPath)
	if err := os.WriteFile(desktopEntryPath, []byte(desktopEntry), 0644); err != nil {
		fmt.Println("Error creating desktop entry file:", err)
		return false, fmt.Errorf("failed to create desktop entry file: %w", err)
	}

	// Update desktop database
	updateCmd := exec.Command("update-desktop-database", applicationsDir)
	if err := updateCmd.Run(); err != nil {
		// Not critical, just log the error
		fmt.Printf("Warning: failed to update desktop database: %v\n", err)
	}

	// Create MIME types directory if it doesn't exist
	mimeDir := filepath.Join(homeDir, ".local", "share", "mime", "packages")
	if err := os.MkdirAll(mimeDir, 0755); err != nil {
		fmt.Println("Error creating MIME types directory:", err)
		return false, fmt.Errorf("failed to create MIME types directory: %w", err)
	}

	fmt.Println("MIME types directory:", mimeDir)

	// Create MIME types file
	mimeFilePath := filepath.Join(mimeDir, "msgreader.xml")
	mimeFile := `<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="application/vnd.ms-outlook">
    <comment>Microsoft Outlook Message</comment>
    <glob pattern="*.msg"/>
  </mime-type>
  <mime-type type="message/rfc822">
    <comment>Email Message</comment>
    <glob pattern="*.eml"/>
  </mime-type>
</mime-info>
`

	fmt.Println("Writing MIME types file to:", mimeFilePath)
	if err := os.WriteFile(mimeFilePath, []byte(mimeFile), 0644); err != nil {
		fmt.Println("Error creating MIME types file:", err)
		return false, fmt.Errorf("failed to create MIME types file: %w", err)
	}

	// Update MIME database
	updateMimeCmd := exec.Command("update-mime-database", filepath.Join(homeDir, ".local", "share", "mime"))
	if err := updateMimeCmd.Run(); err != nil {
		// Not critical, just log the error
		fmt.Printf("Warning: failed to update MIME database: %v\n", err)
	}

	fmt.Println("File associations registered successfully on Linux")
	return true, nil
}

// GetFilesToOpenOnStartup returns the files to open on startup
func (a *App) GetFilesToOpenOnStartup() []string {
	fmt.Println("GetFilesToOpenOnStartup called, returning:", a.filesToOpenOnStartup)
	return a.filesToOpenOnStartup
}

// IsInitialized returns whether the app has been initialized
func (a *App) IsInitialized() bool {
	return a.initialized
}

// DirectOpenFile directly opens a file and emits its content to the frontend
func (a *App) DirectOpenFile(filePath string) {
	fmt.Println("DirectOpenFile called with path:", filePath)

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error checking file:", err)
		// Try to resolve the path
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			fmt.Println("Error getting absolute path:", err)
		} else {
			fmt.Println("Absolute path:", absPath)
			filePath = absPath
		}
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Get the file name from the path
	fileName := filepath.Base(filePath)
	fmt.Println("File name extracted from path:", fileName)
	fmt.Println("File data size:", len(data))

	// Convert the data to a base64 string for embedding in JavaScript
	base64Data := base64.StdEncoding.EncodeToString(data)
	fmt.Println("Base64 data length:", len(base64Data))

	// Create a JavaScript function to process the file
	js := fmt.Sprintf(`
		console.log("Processing file directly in JavaScript");
		
		// Function to convert base64 to ArrayBuffer
		function base64ToArrayBuffer(base64) {
			var binary_string = window.atob(base64);
			var len = binary_string.length;
			var bytes = new Uint8Array(len);
			for (var i = 0; i < len; i++) {
				bytes[i] = binary_string.charCodeAt(i);
			}
			return bytes.buffer;
		}
		
		// Convert the base64 data to an ArrayBuffer
		var fileData = base64ToArrayBuffer("%s");
		console.log("File data converted to ArrayBuffer, length:", fileData.byteLength);
		
		// Get the file extension
		var fileName = "%s";
		var extension = fileName.toLowerCase().split('.').pop();
		console.log("File extension:", extension);
		
		// Function to process the file when the app is ready
		function processFileWhenReady() {
			console.log("Checking if app is ready to process file");
			if (window.app && window.app.fileHandler) {
				console.log("App is ready, processing file");
				
				try {
					// Extract the message info
					var msgInfo;
					if (extension === 'msg' && window.extractMsg) {
						console.log("Extracting MSG file");
						msgInfo = window.extractMsg(fileData);
					} else if (extension === 'eml' && window.extractEml) {
						console.log("Extracting EML file");
						msgInfo = window.extractEml(fileData);
					} else {
						console.error("Unsupported file extension or extraction function not available");
						return;
					}
					
					if (!msgInfo) {
						console.error("Failed to extract message info");
						return;
					}
					
					console.log("Message extracted successfully");
					
					// Add the message to the message handler
					var message = window.app.messageHandler.addMessage(msgInfo, fileName);
					
					// Show the app container
					window.app.uiManager.showAppContainer();
					
					// Update the message list
					window.app.uiManager.updateMessageList();
					
					// Show the message
					window.app.uiManager.showMessage(message);
					
					console.log("Message displayed successfully");
				} catch (error) {
					console.error("Error processing file:", error);
				}
			} else {
				console.log("App not ready yet, waiting...");
				setTimeout(processFileWhenReady, 500);
			}
		}
		
		// Start processing the file
		processFileWhenReady();
	`, base64Data, fileName)

	// Execute the JavaScript
	fmt.Println("Executing JavaScript to process file")
	wailsRuntime.WindowExecJS(a.ctx, js)
}
