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
        
        // Sanity check - early return if there are no files
        if (!filePaths || (Array.isArray(filePaths) && filePaths.length === 0)) {
            console.error('No files provided to handleDesktopFiles');
            return;
        }
        
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
            console.error('No valid files to process after filtering');
            return;
        }
        
        console.log('Processing', paths.length, 'files:', paths);
        
        try {
            // Always show app container before processing files
            console.log('Showing app container');
            this.uiManager.showAppContainer();
            
            // Process each file one by one
            let lastProcessedMessage = null;
            
            for (const path of paths) {
                try {
                    console.log('Processing file:', path);
                    
                    // Get the file name from the path
                    const fileName = path.split(/[/\\]/).pop();
                    console.log('File name:', fileName);
                    
                    // Get the extension
                    const extension = fileName.toLowerCase().split('.').pop();
                    console.log('File extension:', extension);
                    
                    // Process differently based on extension
                    if (extension === 'msg' || extension === 'eml') {
                        // Read the file
                        console.log('Reading file data from:', path);
                        let fileData = await this.readFile(path);
                        
                        if (!fileData || fileData.length === 0) {
                            console.error('File data is empty or could not be read');
                            continue; // Skip to next file
                        }
                        
                        console.log('File data loaded, size:', fileData.length);
                        
                        // Extract message info
                        let msgInfo;
                        if (extension === 'msg') {
                            console.log('Extracting MSG file...');
                            msgInfo = window.extractMsg(fileData);
                        } else {
                            console.log('Extracting EML file...');
                            msgInfo = window.extractEml(fileData);
                        }
                        
                        if (!msgInfo) {
                            console.error('Failed to extract message info from file:', fileName);
                            continue; // Skip to next file
                        }
                        
                        console.log('Message extracted successfully');
                        
                        // Add to message handler
                        const message = this.messageHandler.addMessage(msgInfo, fileName);
                        console.log('Message added to handler');
                        
                        // Update the message list in UI
                        this.uiManager.updateMessageList();
                        
                        // Store the last processed message
                        lastProcessedMessage = message;
                    } else {
                        console.warn('Unsupported file extension:', extension);
                    }
                } catch (error) {
                    console.error('Error processing file:', path, error);
                    // Continue with next file
                }
            }
            
            // Display the message
            if (lastProcessedMessage) {
                console.log('Showing last processed message');
                this.uiManager.showMessage(lastProcessedMessage);
            } else {
                // Fallback: try to show any available message
                const messages = this.messageHandler.getMessages();
                if (messages.length > 0) {
                    console.log('Showing first message from message list');
                    this.uiManager.showMessage(messages[0]);
                } else {
                    console.log('No messages to display');
                }
            }
        } catch (error) {
            console.error('Error in file processing workflow:', error);
        }
    }
    
    // Helper to read a file from the backend
    async readFile(filePath) {
        try {
            console.log('Reading file:', filePath);
            
            if (!window.go || !window.go.main || !window.go.main.App) {
                throw new Error('Go backend not available');
            }
            
            // Try OpenFile first
            if (typeof window.go.main.App.OpenFile === 'function') {
                return await window.go.main.App.OpenFile(filePath);
            }
            
            // Fall back to ReadFile if available
            if (typeof window.go.main.App.ReadFile === 'function') {
                return await window.go.main.App.ReadFile(filePath);
            }
            
            throw new Error('No suitable file reading method available');
        } catch (error) {
            console.error('Error reading file:', filePath, error);
            throw error;
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
                
                // Read the file using the appropriate method
                let fileBuffer;
                try {
                    console.log('Reading file:', filePath);
                    // Check which method is available (for compatibility)
                    if (window.go && window.go.main && window.go.main.App && window.go.main.App.OpenFile) {
                        fileBuffer = await window.go.main.App.OpenFile(filePath);
                    } else if (window.go && window.go.main && window.go.main.App && window.go.main.App.ReadFile) {
                        fileBuffer = await window.go.main.App.ReadFile(filePath);
                    } else {
                        throw new Error('No method available to read files from Go backend');
                    }
                    
                    if (!fileBuffer || fileBuffer.length === 0) {
                        throw new Error('Received empty file buffer');
                    }
                    
                    console.log('File read successfully:', fileName, 'Buffer size:', fileBuffer.length);
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
                    
                    console.log('Successfully extracted message info from file:', fileName);
                    
                    // Add message to the handler
                    const message = this.messageHandler.addMessage(msgInfo, fileName);
                    
                    // Update message list
                    this.uiManager.updateMessageList();
                    
                    console.log('Message added and UI updated for file:', fileName);
                    
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