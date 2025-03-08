// This file provides a bridge between the web frontend and the Go backend
// It will be loaded by the Wails application

// Check if we're running in a Wails context
const isWails = typeof window !== 'undefined' && typeof window.go !== 'undefined';

// Debug: Log if we're running in a Wails context
console.log('Running in Wails context:', isWails);

// Create a bridge object to expose to the frontend
window.desktopBridge = {
    // Open a file dialog and return the selected file paths
    async openFileDialog() {
        if (!isWails) return null;
        try {
            return await window.go.main.App.OpenFileDialog();
        } catch (error) {
            console.error('Error opening file dialog:', error);
            return null;
        }
    },

    // Save a file dialog and return the selected file path
    async saveFileDialog(defaultFilename) {
        if (!isWails) return null;
        try {
            return await window.go.main.App.SaveFileDialog(defaultFilename);
        } catch (error) {
            console.error('Error opening save dialog:', error);
            return null;
        }
    },

    // Read a file from the file system
    async readFile(filePath) {
        if (!isWails) return null;
        try {
            console.log('Reading file:', filePath);
            const data = await window.go.main.App.OpenFile(filePath);
            console.log('File read successfully, data length:', data.length);
            return data;
        } catch (error) {
            console.error('Error reading file:', error);
            return null;
        }
    },

    // Save a file to the file system
    async saveFile(filePath, data) {
        if (!isWails) return false;
        try {
            await window.go.main.App.SaveFile(filePath, data);
            return true;
        } catch (error) {
            console.error('Error saving file:', error);
            return false;
        }
    },

    // Check if we're running in a desktop environment
    isDesktop() {
        return isWails;
    },

    // Register the application as the default handler for .msg and .eml files
    async registerFileAssociations() {
        if (!isWails) return false;
        try {
            console.log('Registering file associations...');
            const result = await window.go.main.App.RegisterFileAssociations();
            console.log('File associations registered:', result);
            return result;
        } catch (error) {
            console.error('Error registering file associations:', error);
            return false;
        }
    },

    // Get files to open on startup
    async getFilesToOpenOnStartup() {
        if (!isWails) return [];
        try {
            console.log('Getting files to open on startup...');
            const files = await window.go.main.App.GetFilesToOpenOnStartup();
            console.log('Files to open on startup:', files);
            return files;
        } catch (error) {
            console.error('Error getting files to open on startup:', error);
            return [];
        }
    }
};

// If we're running in a Wails context, modify the file input handling
if (isWails) {
    // Wait for the DOM to be loaded
    document.addEventListener('DOMContentLoaded', () => {
        console.log('DOM loaded in Wails context');
        
        // Find all file inputs
        const fileInputs = document.querySelectorAll('input[type="file"]');
        console.log('Found file inputs:', fileInputs.length);
        
        // Add a click event listener to each file input
        fileInputs.forEach(input => {
            input.addEventListener('click', async (event) => {
                // Prevent the default file dialog
                event.preventDefault();
                
                // Open the Wails file dialog
                console.log('Opening file dialog...');
                const files = await window.desktopBridge.openFileDialog();
                console.log('Files selected from dialog:', files);
                
                // If files were selected, trigger a custom event
                if (files && files.length > 0) {
                    // Create a custom event with the selected files
                    const customEvent = new CustomEvent('wailsFileSelected', {
                        detail: { files }
                    });
                    
                    // Dispatch the event on the input element
                    console.log('Dispatching wailsFileSelected event');
                    input.dispatchEvent(customEvent);
                }
            });
        });

        // Add a menu item to register file associations
        const registerMenuItem = document.createElement('div');
        registerMenuItem.className = 'register-file-associations';
        registerMenuItem.textContent = 'Register as Default App';
        registerMenuItem.style.cursor = 'pointer';
        registerMenuItem.style.padding = '8px 16px';
        registerMenuItem.style.marginTop = '16px';
        registerMenuItem.style.backgroundColor = '#f0f9ff';
        registerMenuItem.style.color = '#0369a1';
        registerMenuItem.style.borderRadius = '4px';
        registerMenuItem.style.textAlign = 'center';
        registerMenuItem.style.fontWeight = 'bold';
        
        registerMenuItem.addEventListener('click', async () => {
            console.log('Register as Default App clicked');
            try {
                const result = await window.desktopBridge.registerFileAssociations();
                console.log('Registration result:', result);
                if (result) {
                    alert('Successfully registered as the default app for .msg and .eml files!');
                } else {
                    alert('Failed to register as the default app. Please try again with administrator privileges.');
                }
            } catch (error) {
                console.error('Error registering file associations:', error);
                alert('An error occurred while registering file associations: ' + error.message);
            }
        });
        
        // Add the menu item to the welcome screen
        const welcomeContent = document.querySelector('.welcome-content');
        if (welcomeContent) {
            console.log('Adding register menu item to welcome screen');
            welcomeContent.appendChild(registerMenuItem);
        } else {
            console.error('Welcome content element not found');
        }

        // Check for files to open on startup
        const checkFilesToOpen = async () => {
            console.log('Checking for files to open on startup');
            try {
                const files = await window.desktopBridge.getFilesToOpenOnStartup();
                console.log('Got files to open on startup:', files);
                
                if (files && files.length > 0 && window.app && window.app.fileHandler) {
                    console.log('Handling files to open on startup');
                    window.app.fileHandler.handleDesktopFiles(files);
                } else {
                    if (!files || files.length === 0) {
                        console.log('No files to open on startup');
                    } else if (!window.app) {
                        console.error('window.app is not defined');
                    } else if (!window.app.fileHandler) {
                        console.error('window.app.fileHandler is not defined');
                    }
                }
            } catch (error) {
                console.error('Error checking for files to open on startup:', error);
            }
        };
        
        // Wait for the app to be fully initialized before checking for files to open
        const waitForAppInitialization = () => {
            console.log('Waiting for app initialization...');
            if (window.app && window.app.fileHandler) {
                console.log('App is initialized, emitting frontend-ready event');
                if (isWails) {
                    window.runtime.EventsEmit('frontend-ready');
                }
                
                // Check for stored files to open
                if (window._filesToOpenWhenReady && window._filesToOpenWhenReady.length > 0) {
                    console.log('Found stored files to open:', window._filesToOpenWhenReady);
                    window.app.fileHandler.handleDesktopFiles(window._filesToOpenWhenReady);
                    window._filesToOpenWhenReady = null;
                } else {
                    // If no stored files, check for files to open on startup
                    checkFilesToOpen();
                }
            } else {
                console.log('App not yet initialized, waiting...');
                setTimeout(waitForAppInitialization, 10); // Reduced timeout significantly
            }
        };

        // Start waiting for app initialization with shorter initial delay
        setTimeout(waitForAppInitialization, 10);

        // Listen for the files-to-open event from the backend
        if (isWails) {
            console.log('Setting up files-to-open event listener');
            window.runtime.EventsOn('files-to-open', (files) => {
                console.log('Received files-to-open event:', files);
                if (files && files.length > 0) {
                    if (window.app && window.app.fileHandler) {
                        // Handle files immediately if app is ready
                        console.log('App is initialized, handling desktop files from event');
                        window.app.fileHandler.handleDesktopFiles(files);
                    } else {
                        // Store files only if app is not ready
                        console.log('App not yet initialized, storing files for later');
                        window._filesToOpenWhenReady = files;
                    }
                } else {
                    console.error('Could not handle files from files-to-open event:', {
                        filesExist: files && files.length > 0
                    });
                }
            });
        }
    });
} 