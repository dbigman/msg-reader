package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                 context.Context
	filesToOpenOnStartup []string
	initialized         bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("App startup called, context saved")
	
	// If there are files to open on startup, notify the frontend
	if len(a.filesToOpenOnStartup) > 0 {
		fmt.Println("Files to open on startup:", a.filesToOpenOnStartup)
		
		// Emit the event immediately as well (in case the frontend is already ready)
		fmt.Println("Emitting initial files-to-open event with:", a.filesToOpenOnStartup)
		wailsRuntime.EventsEmit(ctx, "files-to-open", a.filesToOpenOnStartup)
		
		// Also wait for the frontend to signal it's ready
		fmt.Println("Setting up frontend-ready event listener")
		wailsRuntime.EventsOn(ctx, "frontend-ready", func(optionalData ...interface{}) {
			fmt.Println("Received frontend-ready event, emitting files-to-open event again with:", a.filesToOpenOnStartup)
			// Emit an event to the frontend with the files to open
			wailsRuntime.EventsEmit(ctx, "files-to-open", a.filesToOpenOnStartup)
			
			// Also emit a direct event to force file opening
			fmt.Println("Emitting force-open-files event with:", a.filesToOpenOnStartup)
			wailsRuntime.EventsEmit(ctx, "force-open-files", a.filesToOpenOnStartup)
			
			// Also try to directly open the first file
			if len(a.filesToOpenOnStartup) > 0 {
				fmt.Println("Directly opening first file:", a.filesToOpenOnStartup[0])
				a.DirectOpenFile(a.filesToOpenOnStartup[0])
			}
		})
		
		// Set up a DOM ready event listener to handle the case where the frontend is ready but hasn't emitted the frontend-ready event
		fmt.Println("Setting up dom-ready event listener")
		wailsRuntime.EventsOn(ctx, "dom-ready", func(optionalData ...interface{}) {
			fmt.Println("Received dom-ready event, checking if file has been opened")
			
			// Check if the file has been opened already
			if len(a.filesToOpenOnStartup) > 0 {
				fmt.Println("Opening file after DOM ready:", a.filesToOpenOnStartup[0])
				// Try to open the file directly
				a.DirectOpenFile(a.filesToOpenOnStartup[0])
			}
		})
		
		// Also execute JavaScript directly to handle the file
		// This is a more direct approach that doesn't rely on events
		go func() {
			// Wait a short time for the window to be created
			time.Sleep(500 * time.Millisecond)
			
			if len(a.filesToOpenOnStartup) > 0 {
				fmt.Println("Executing JavaScript to handle file directly:", a.filesToOpenOnStartup[0])
				
				// Read the file
				data, err := os.ReadFile(a.filesToOpenOnStartup[0])
				if err != nil {
					fmt.Println("Error reading file:", err)
					return
				}
				
				// Get the file name from the path
				fileName := filepath.Base(a.filesToOpenOnStartup[0])
				fmt.Println("File name extracted from path:", fileName)
				
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
				wailsRuntime.WindowExecJS(ctx, js)
			}
		}()
		
		// Mark as initialized
		a.initialized = true
		fmt.Println("App marked as initialized (with files to open)")
	} else {
		// Mark as initialized immediately if no files to open
		a.initialized = true
		fmt.Println("App marked as initialized (no files to open)")
	}
}

// OpenFile allows the frontend to request opening a file
func (a *App) OpenFile(filePath string) ([]byte, error) {
	fmt.Println("OpenFile called with path:", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	fmt.Println("File read successfully, size:", len(data))
	return data, nil
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