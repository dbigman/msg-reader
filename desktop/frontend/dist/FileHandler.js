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
    handleDesktopFiles(filePaths) {
        console.log('handleDesktopFiles called with paths:', filePaths);
        if (!Array.isArray(filePaths)) {
            console.error('filePaths is not an array:', filePaths);
            
            // Try to handle it as a single string
            if (typeof filePaths === 'string') {
                console.log('Trying to handle as a single string path');
                this.handleDesktopFile(filePaths);
                return;
            }
            
            return;
        }
        
        if (filePaths.length === 0) {
            console.error('No file paths provided to handleDesktopFiles');
            return;
        }
        
        // Process files one by one with a small delay between them
        const processFiles = (index) => {
            if (index >= filePaths.length) {
                console.log('Finished processing all files');
                return;
            }
            
            const filePath = filePaths[index];
            if (!filePath) {
                console.error('Empty file path in filePaths array');
                processFiles(index + 1);
                return;
            }
            
            // Normalize the file path (handle spaces and special characters)
            const normalizedPath = decodeURIComponent(filePath);
            console.log('Normalized file path:', normalizedPath);
            
            const extension = normalizedPath.toLowerCase().split('.').pop();
            console.log('Processing file path:', normalizedPath, 'with extension:', extension);
            if (extension === 'msg' || extension === 'eml') {
                // Process the file
                this.handleDesktopFile(normalizedPath)
                    .then(() => {
                        console.log('Successfully processed file:', normalizedPath);
                        // Process the next file after a small delay
                        setTimeout(() => processFiles(index + 1), 100);
                    })
                    .catch(error => {
                        console.error('Error processing file:', normalizedPath, error);
                        // Continue with the next file even if there was an error
                        setTimeout(() => processFiles(index + 1), 100);
                    });
            } else {
                console.warn('Skipping file with unsupported extension:', extension);
                // Continue with the next file
                setTimeout(() => processFiles(index + 1), 100);
            }
        };
        
        // Start processing files
        processFiles(0);
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
                const fileName = filePath.split(/[\\/]/).pop();
                console.log('File name extracted from path:', fileName);
                
                // Read the file using the desktop bridge
                console.log('Reading file using desktop bridge:', filePath);
                let fileBuffer;
                try {
                    console.log('Calling window.desktopBridge.readFile with path:', filePath);
                    fileBuffer = await window.desktopBridge.readFile(filePath);
                    console.log('readFile returned successfully, buffer exists:', !!fileBuffer);
                    if (fileBuffer) {
                        console.log('Buffer length:', fileBuffer.length);
                    }
                } catch (readError) {
                    console.error('Error reading file with original path:', readError);
                    
                    // Try with URI decoded path
                    try {
                        const decodedPath = decodeURIComponent(filePath);
                        console.log('Trying with decoded path:', decodedPath);
                        fileBuffer = await window.desktopBridge.readFile(decodedPath);
                        console.log('readFile with decoded path returned successfully, buffer exists:', !!fileBuffer);
                        if (fileBuffer) {
                            console.log('Buffer length:', fileBuffer.length);
                        }
                    } catch (decodedError) {
                        console.error('Error reading file with decoded path:', decodedError);
                        reject(decodedError);
                        return;
                    }
                }
                
                if (!fileBuffer) {
                    const error = new Error('Failed to read file: ' + filePath);
                    console.error(error);
                    reject(error);
                    return;
                }
                
                console.log('File read successfully:', filePath, 'buffer length:', fileBuffer.length);
                
                const extension = fileName.toLowerCase().split('.').pop();
                console.log('File extension:', extension);
                
                let msgInfo;
                try {
                    if (extension === 'msg') {
                        console.log('Extracting MSG file:', fileName);
                        console.log('window.extractMsg exists:', !!window.extractMsg);
                        msgInfo = window.extractMsg(fileBuffer);
                        console.log('MSG extraction complete, msgInfo exists:', !!msgInfo);
                    } else if (extension === 'eml') {
                        console.log('Extracting EML file:', fileName);
                        console.log('window.extractEml exists:', !!window.extractEml);
                        msgInfo = window.extractEml(fileBuffer);
                        console.log('EML extraction complete, msgInfo exists:', !!msgInfo);
                    }
                    
                    console.log('File extracted successfully:', fileName);
                    console.log('Message info:', msgInfo ? 'valid' : 'invalid');
                    
                    if (!msgInfo) {
                        const error = new Error('Failed to extract message info from file: ' + fileName);
                        console.error(error);
                        reject(error);
                        return;
                    }
                    
                    console.log('Adding message to message handler');
                    console.log('this.messageHandler exists:', !!this.messageHandler);
                    const message = this.messageHandler.addMessage(msgInfo, fileName);
                    console.log('Message added successfully');
                    
                    // Hide welcome screen and show app
                    console.log('Showing app container');
                    console.log('this.uiManager exists:', !!this.uiManager);
                    this.uiManager.showAppContainer();
                    console.log('App container shown');

                    // Update message list
                    console.log('Updating message list');
                    this.uiManager.updateMessageList();
                    console.log('Message list updated');
                    
                    // Show first message if it's the only one
                    console.log('Checking if this is the only message');
                    console.log('Messages count:', this.messageHandler.getMessages().length);
                    if (this.messageHandler.getMessages().length === 1) {
                        console.log('Showing the message as it is the only one');
                        this.uiManager.showMessage(message);
                        console.log('Message shown');
                    } else {
                        console.log('Not showing the message as there are multiple messages');
                    }
                    
                    resolve();
                    console.log('handleDesktopFile completed successfully');
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