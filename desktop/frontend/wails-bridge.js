// This file provides a bridge between the web frontend and the Go backend
// It will be loaded by the Wails application

// Check if we're running in a Wails context
const isWails = typeof window.go !== 'undefined';

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
            const data = await window.go.main.App.OpenFile(filePath);
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
    }
};

// If we're running in a Wails context, modify the file input handling
if (isWails) {
    // Wait for the DOM to be loaded
    document.addEventListener('DOMContentLoaded', () => {
        // Find all file inputs
        const fileInputs = document.querySelectorAll('input[type="file"]');
        
        // Add a click event listener to each file input
        fileInputs.forEach(input => {
            input.addEventListener('click', async (event) => {
                // Prevent the default file dialog
                event.preventDefault();
                
                // Open the Wails file dialog
                const files = await window.desktopBridge.openFileDialog();
                
                // If files were selected, trigger a custom event
                if (files && files.length > 0) {
                    // Create a custom event with the selected files
                    const customEvent = new CustomEvent('wailsFileSelected', {
                        detail: { files }
                    });
                    
                    // Dispatch the event on the input element
                    input.dispatchEvent(customEvent);
                }
            });
        });
    });
} 