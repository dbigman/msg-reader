// This is a modified version of main.js for the desktop application
// It ensures proper initialization and file handling

// Make classes available globally
if (typeof window !== 'undefined') {
    window.MessageHandler = require('./dist/MessageHandler');
    window.UIManager = require('./dist/UIManager');
    window.FileHandler = require('./dist/FileHandler');
    window.extractMsg = require('./dist/utils').extractMsg;
    window.extractEml = require('./dist/utils').extractEml;
}

class App {
    constructor() {
        console.log('App constructor called');
        this.messageHandler = new window.MessageHandler();
        this.uiManager = new window.UIManager(this.messageHandler);
        this.fileHandler = new window.FileHandler(this.messageHandler, this.uiManager);
        
        // Check for stored files to open
        if (window._filesToOpenWhenReady && window._filesToOpenWhenReady.length > 0) {
            console.log('Found stored files to open in App constructor:', window._filesToOpenWhenReady);
            setTimeout(() => {
                this.fileHandler.handleDesktopFiles(window._filesToOpenWhenReady);
                // Clear the stored files
                window._filesToOpenWhenReady = null;
            }, 500);
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
    document.addEventListener('DOMContentLoaded', () => {
        console.log('DOM content loaded, initializing app');
        window.app = new App();
        
        // Signal that the frontend is ready
        if (typeof window.runtime !== 'undefined') {
            console.log('Emitting frontend-ready event');
            window.runtime.EventsEmit('frontend-ready');
        }
    });
}

module.exports = App; 