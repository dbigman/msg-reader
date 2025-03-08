// This is a modified version of main.js for the desktop application
// It ensures proper initialization and file handling

// Make classes available globally
if (typeof window !== 'undefined') {
    // Wait for the bundle to be loaded
    window.addEventListener('load', () => {
        console.log('Window loaded, checking for required modules');
        
        // Check if modules are available from the bundle
        if (!window.MessageHandler || !window.UIManager || !window.FileHandler || !window.extractMsg || !window.extractEml) {
            console.error('Required modules not found in window object. Available globals:', Object.keys(window));
            return;
        }
        
        // Initialize the app
        initializeApp();
    });
}

function initializeApp() {
    console.log('Initializing app with required modules');
    console.log('MessageHandler available:', !!window.MessageHandler);
    console.log('UIManager available:', !!window.UIManager);
    console.log('FileHandler available:', !!window.FileHandler);
    console.log('extractMsg available:', !!window.extractMsg);
    console.log('extractEml available:', !!window.extractEml);
    
    class App {
        constructor() {
            console.log('App constructor called');
            
            // Initialize core components
            this.messageHandler = new window.MessageHandler();
            this.uiManager = new window.UIManager(this.messageHandler);
            this.fileHandler = new window.FileHandler(this.messageHandler, this.uiManager);
            
            // Setup event listeners first
            this.setupEventListeners();
            
            // IMPORTANT: Send multiple ready signals to ensure the backend receives it
            this.signalReadyToBackend();
        }
        
        // Send ready signal multiple ways to ensure it's received
        signalReadyToBackend() {
            // First attempt - immediate
            this.sendReadySignal();
            
            // Second attempt - after requestAnimationFrame
            requestAnimationFrame(() => {
                this.sendReadySignal();
                
                // Third attempt - after a short delay
                setTimeout(() => {
                    this.sendReadySignal();
                }, 500);
            });
        }
        
        // Helper to send the ready signal
        sendReadySignal() {
            if (window.runtime) {
                console.log('Sending frontend-ready signal to backend');
                window.runtime.EventsEmit('frontend-ready');
            } else {
                console.error('Runtime not available, cannot signal readiness');
            }
        }
        
        setupEventListeners() {
            if (!window.runtime) {
                console.error('Wails runtime not available for event setup');
                return;
            }
            
            console.log('Setting up Wails runtime event listeners');
            
            // Listen for backend-ready event - backend has finished processing our frontend-ready signal
            window.runtime.EventsOn('backend-ready', async (isReady) => {
                console.log('Received backend-ready event, checking for pending files');
                
                try {
                    // Actively fetch any pending files from the backend
                    const pendingFiles = await window.go.main.App.GetPendingFiles();
                    console.log('Retrieved pending files from backend:', pendingFiles);
                    
                    if (pendingFiles && pendingFiles.length > 0) {
                        console.log('Processing pending files:', pendingFiles);
                        this.fileHandler.handleDesktopFiles(pendingFiles);
                    } else {
                        console.log('No pending files to process');
                    }
                } catch (error) {
                    console.error('Error retrieving pending files:', error);
                }
            });
            
            // Listen for new-files-available event - files have been added while app is running
            window.runtime.EventsOn('new-files-available', async () => {
                console.log('Received new-files-available event, fetching new files');
                
                try {
                    // Fetch the new files from the backend
                    const newFiles = await window.go.main.App.GetPendingFiles();
                    console.log('Retrieved new files from backend:', newFiles);
                    
                    if (newFiles && newFiles.length > 0) {
                        console.log('Processing new files:', newFiles);
                        this.fileHandler.handleDesktopFiles(newFiles);
                    } else {
                        console.log('No new files to process');
                    }
                } catch (error) {
                    console.error('Error retrieving new files:', error);
                }
            });
            
            // Direct file open event - most direct way to open a file
            window.runtime.EventsOn('open-file-now', (filePath) => {
                console.log('Received direct open-file-now event for:', filePath);
                
                if (this.fileHandler) {
                    console.log('Processing file directly:', filePath);
                    this.fileHandler.handleDesktopFiles([filePath]);
                } else {
                    console.error('FileHandler not initialized, cannot process file:', filePath);
                }
            });
        }

        showMessage(index) {
            const messages = this.messageHandler.getMessages();
            if (messages[index]) {
                this.uiManager.showMessage(messages[index]);
            }
        }

        togglePin(index) {
            const message = this.messageHandler.togglePin(index);
            this.uiManager.updateMessageList();
            this.uiManager.showMessage(message);
        }

        deleteMessage(index) {
            const nextMessage = this.messageHandler.deleteMessage(index);
            this.uiManager.updateMessageList();
            
            if (nextMessage) {
                this.uiManager.showMessage(nextMessage);
            } else {
                this.uiManager.showWelcomeScreen();
            }
        }
    }

    // Initialize the app when the DOM is loaded
    if (typeof window !== 'undefined') {
        window.App = App;
        console.log('Creating new App instance');
        window.app = new App();
    }
}

// Export the App class if we're in a module environment
if (typeof module !== 'undefined' && module.exports) {
    module.exports = App;
} 