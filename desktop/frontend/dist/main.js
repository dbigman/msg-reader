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
            this.messageHandler = new window.MessageHandler();
            this.uiManager = new window.UIManager(this.messageHandler);
            this.fileHandler = new window.FileHandler(this.messageHandler, this.uiManager);
            
            // Set up event listeners for file opening
            if (typeof window.runtime !== 'undefined') {
                console.log('Setting up Wails runtime event listeners');
                
                // Listen for files-to-open event
                window.runtime.EventsOn('files-to-open', (filePaths) => {
                    console.log('Received files-to-open event:', filePaths);
                    if (this.fileHandler) {
                        this.fileHandler.handleDesktopFiles(filePaths);
                    } else {
                        console.error('FileHandler not initialized when receiving files-to-open event');
                        // Store the files to open later
                        window._filesToOpenWhenReady = filePaths;
                    }
                });
                
                // Listen for force-open-files event
                window.runtime.EventsOn('force-open-files', (filePaths) => {
                    console.log('Received force-open-files event:', filePaths);
                    if (this.fileHandler) {
                        this.fileHandler.handleDesktopFiles(filePaths);
                    }
                });
            }
            
            // Check for stored files to open
            if (window._filesToOpenWhenReady && window._filesToOpenWhenReady.length > 0) {
                console.log('Found stored files to open in App constructor:', window._filesToOpenWhenReady);
                setTimeout(() => {
                    this.fileHandler.handleDesktopFiles(window._filesToOpenWhenReady);
                    // Clear the stored files
                    window._filesToOpenWhenReady = null;
                }, 500);
            }
            
            // Signal that the frontend is ready
            if (typeof window.runtime !== 'undefined') {
                console.log('Emitting frontend-ready event');
                window.runtime.EventsEmit('frontend-ready');
            }
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