// This is a modified version of FileHandler.js for the desktop application
// It adds support for handling file paths from the Go backend

class FileHandler {
    constructor(messageHandler, uiManager) {
        console.log('FileHandler constructor called');
        this.messageHandler = messageHandler;
        this.uiManager = uiManager;
        this.setupEventListeners();
    }

    setupEventListeners() {
        console.log('Setting up FileHandler event listeners');
        document.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.uiManager.showDropOverlay();
        });

        document.addEventListener('dragleave', (e) => {
            e.preventDefault();
            if (e.clientX <= 0 || e.clientY <= 0 || 
                e.clientX >= window.innerWidth || e.clientY >= window.innerHeight) {
                this.uiManager.hideDropOverlay();
            }
        });

        document.addEventListener('drop', (e) => {
            e.preventDefault();
            this.uiManager.hideDropOverlay();
            this.handleFiles(e.dataTransfer.files);
        });

        // File input handlers
        const fileInput = document.getElementById('fileInput');
        if (fileInput) {
            console.log('Setting up fileInput event listener');
            fileInput.addEventListener('change', (e) => {
                this.handleFiles(e.target.files);
            });
        } else {
            console.error('fileInput element not found');
        }

        const fileInputInApp = document.getElementById('fileInputInApp');
        if (fileInputInApp) {
            console.log('Setting up fileInputInApp event listener');
            fileInputInApp.addEventListener('change', (e) => {
                this.handleFiles(e.target.files);
            });
        } else {
            console.error('fileInputInApp element not found');
        }
    }

    // Handle browser File objects
    handleFiles(files) {
        console.log('handleFiles called with', files ? files.length : 0, 'files');
        if (!files || files.length === 0) {
            console.error('No files provided to handleFiles');
            return;
        }
        
        Array.from(files).forEach(file => {
            const extension = file.name.toLowerCase().split('.').pop();
            console.log('Processing file:', file.name, 'with extension:', extension);
            if (extension === 'msg' || extension === 'eml') {
                this.handleFile(file);
            } else {
                console.warn('Skipping file with unsupported extension:', extension);
            }
        });
    }

    // Handle file paths from the desktop application
    async handleDesktopFiles(filePaths) {
        console.log('handleDesktopFiles called with paths:', filePaths);
        
        // Normalize input to array
        let paths = filePaths;
        if (typeof filePaths === 'string') {
            console.log('Converting single string path to array');
            paths = [filePaths];
        } else if (!Array.isArray(filePaths)) {
            console.error('Invalid input to handleDesktopFiles:', filePaths);
            return;
        }
        
        // Filter out any invalid paths
        paths = paths.filter(path => {
            if (!path) {
                console.warn('Skipping empty path');
                return false;
            }
            const extension = path.toLowerCase().split('.').pop();
            if (extension !== 'msg' && extension !== 'eml') {
                console.warn('Skipping file with unsupported extension:', path);
                return false;
            }
            return true;
        });
        
        if (paths.length === 0) {
            console.error('No valid files to process');
            return;
        }
        
        console.log('Processing', paths.length, 'files:', paths);
        
        // Process all files
        try {
            // Hide welcome screen before processing files
            this.uiManager.showAppContainer();
            
            // Process all files and collect their promises
            const promises = paths.map(path => this.handleDesktopFile(path));
            
            // Wait for all files to be processed
            await Promise.all(promises);
            
            // After all files are processed, show the last file
            const messages = this.messageHandler.getMessages();
            if (messages.length > 0) {
                // Show the most recently added message (first in the list since we unshift)
                this.uiManager.showMessage(messages[0]);
            }
            
            console.log('All files processed successfully');
        } catch (error) {
            console.error('Error processing files:', error);
        }
    }

    // Handle a browser File object
    handleFile(file) {
        console.log('handleFile called with file:', file.name);
        const reader = new FileReader();
        reader.onload = (e) => {
            console.log('File read complete:', file.name);
            const fileBuffer = e.target.result;
            const extension = file.name.toLowerCase().split('.').pop();
            
            let msgInfo;
            try {
                if (extension === 'msg') {
                    console.log('Extracting MSG file:', file.name);
                    msgInfo = window.extractMsg(fileBuffer);
                } else if (extension === 'eml') {
                    console.log('Extracting EML file:', file.name);
                    msgInfo = window.extractEml(fileBuffer);
                }
                
                console.log('File extracted successfully:', file.name);
                console.log('Message info:', msgInfo ? 'valid' : 'invalid');
                
                if (!msgInfo) {
                    console.error('Failed to extract message info from file:', file.name);
                    return;
                }
                
                const message = this.messageHandler.addMessage(msgInfo, file.name);
                
                // Hide welcome screen and show app
                this.uiManager.showAppContainer();

                // Update message list
                this.uiManager.updateMessageList();
                
                // Show first message if it's the only one
                if (this.messageHandler.getMessages().length === 1) {
                    this.uiManager.showMessage(message);
                }
            } catch (error) {
                console.error('Error processing file:', file.name, error);
            }
        };
        
        reader.onerror = (error) => {
            console.error('Error reading file:', file.name, error);
        };
        
        reader.readAsArrayBuffer(file);
    }

    // Handle a file path from the desktop application
    async handleDesktopFile(filePath) {
        console.log('handleDesktopFile called with path:', filePath);
        
        return new Promise(async (resolve, reject) => {
            try {
                // Get the file name from the path
                const fileName = filePath.split(/[/\\]/).pop();
                console.log('Processing file:', fileName);
                
                // Get the file extension
                const extension = fileName.toLowerCase().split('.').pop();
                if (extension !== 'msg' && extension !== 'eml') {
                    const error = new Error('Unsupported file type: ' + extension);
                    console.error(error);
                    reject(error);
                    return;
                }
                
                // Read the file
                let fileBuffer;
                try {
                    console.log('Reading file:', filePath);
                    fileBuffer = await window.go.main.App.ReadFile(filePath);
                    console.log('File read successfully:', fileName);
                } catch (error) {
                    console.error('Error reading file:', filePath, error);
                    reject(error);
                    return;
                }
                
                let msgInfo;
                try {
                    if (extension === 'msg') {
                        console.log('Extracting MSG file:', fileName);
                        msgInfo = window.extractMsg(fileBuffer);
                    } else if (extension === 'eml') {
                        console.log('Extracting EML file:', fileName);
                        msgInfo = window.extractEml(fileBuffer);
                    }
                    
                    if (!msgInfo) {
                        const error = new Error('Failed to extract message info from file: ' + fileName);
                        console.error(error);
                        reject(error);
                        return;
                    }
                    
                    // Add message to the handler
                    const message = this.messageHandler.addMessage(msgInfo, fileName);
                    
                    // Update message list
                    this.uiManager.updateMessageList();
                    
                    resolve(message);
                } catch (error) {
                    console.error('Error extracting file:', fileName, error);
                    reject(error);
                }
            } catch (error) {
                console.error('Error handling desktop file:', filePath, error);
                reject(error);
            }
        });
    }
}

// Export the FileHandler class
module.exports = FileHandler; 