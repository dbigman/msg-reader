# msgReader

![alt text](doc/res/msg-reader-screenshot.png)

My collegue sends me a lot of emails in the *.msg format. It's the format that "Old Microsoft Outlook" uses if you export an email or attach it to another email.  
I couldn't even open these files on my Windows 11 machine, because of an error that stated that I need an active Microsoft 365 subscription - which is confusing (or **very** wrong), because I have one and it's active in all other MS Office apps.

Since my collegue doesn't stop sending me these <3 and my Feedback to Microsoft a few months ago didn't do anything 💤, I decided to take matters into my own hands.  
I wrote a small tool that can read these files and show them to me (about) how they are shown in email clients - with HTML and inline images and all.

I am currently writing with the Microsoft Support to get this issue fixed, but until then, I have this tool (and everyone who needs to open *.msg files and can't/wont afford a Microsoft 365 Subscription to open a fricking .msg-File).  

... also - WHY would you put a paywall in front of a file format that you created?

## Features
- Open and read *.msg and *.eml files directly in your browser or desktop application
- View HTML content and inline images
- Pin important messages
- Multiple file support with message list
- Sort messages by date
- Drag & drop support
- No server needed - everything runs in your browser or locally

## Installation

### Web Version
The web version is available at [rasalas.github.io/msg-reader/](https://rasalas.github.io/msg-reader/) and requires no installation.

### Desktop Version
Download the latest version for your operating system from the [Releases page](https://github.com/Rasalas/msg-reader/releases).

#### Windows
1. Download `msgReader-windows-amd64.zip`
2. Extract the ZIP file
3. Run `msgReader.exe`
4. (Optional) Right-click on a .msg or .eml file → Open with → Choose another app → More apps → Look for another app on this PC → Select msgReader.exe

#### macOS
1. Download `msgReader-macos-universal.zip`
2. Extract the ZIP file
3. Move `msgReader.app` to your Applications folder
4. (Optional) Right-click on a .msg or .eml file → Open With → Other... → Applications → msgReader.app → Always Open With

Note: On first launch, you might need to:
- Right-click the app and select "Open" to bypass Gatekeeper
- Go to System Settings → Privacy & Security and allow the app to run

#### Linux
1. Download `msgReader-linux-amd64.zip`
2. Extract the ZIP file
3. Make the file executable:
   ```bash
   chmod +x msgReader
   ```
4. Run the application:
   ```bash
   ./msgReader
   ```
5. (Optional) Create a desktop entry for file associations:
   ```bash
   cat > ~/.local/share/applications/msgreader.desktop << EOL
   [Desktop Entry]
   Type=Application
   Name=msgReader
   Exec=/path/to/msgReader %f
   Icon=/path/to/icon.png
   MimeType=application/vnd.ms-outlook;message/rfc822;
   Categories=Office;Email;
   EOL
   ```

### Command Line Usage
After installation, you can open files directly from the command line:

```bash
# Windows
msgReader.exe "path/to/email.msg"

# macOS
/Applications/msgReader.app/Contents/MacOS/msgReader "path/to/email.msg"

# Linux
./msgReader "path/to/email.msg"
```

Multiple files can be opened at once:
```bash
msgReader file1.msg file2.eml file3.msg
```

## Project Structure
The project is organized into several modules:
- `MessageHandler.js` - Manages message state and storage
- `UIManager.js` - Handles UI updates and rendering
- `FileHandler.js` - Manages file operations and drag & drop
- `utils.js` - Contains MSG file processing and utility functions
- `main.js` - Initializes and orchestrates the application
- `desktop/` - Contains the desktop application built with Wails and Go

## HYPER Quick Start (GitHub Pages)
1. Open [rasalas.github.io/msg-reader/](https://rasalas.github.io/msg-reader/)
2. Drag your file from your file system and drop it in the drop area.
3. Done.  

You should now see your email contents

## Quick Start (locally)
1. Clone the repository
```bash
git clone https://github.com/Rasalas/msg-reader.git
cd msg-reader
```

2. Install the dependencies
```bash
npm install
```

3. Run the application
```bash
npm start
```
A browser window should open with the application running.

## Desktop Application
The project now includes a desktop application built with Wails and Go. This allows you to run the application as a native desktop application on Windows, macOS, and Linux.

### Prerequisites
To build the desktop application, you need:
- Go 1.21 or later
- Wails CLI
- Platform-specific dependencies for Wails

### Installing Wails
Run the installation script:
```bash
cd desktop
./install-wails.sh
```

### Development
To start the development server for the desktop application:
```bash
cd desktop
./dev.sh
```

### Building
To build the desktop application:
```bash
cd desktop
./build.sh
```

The built application will be in the `desktop/build/bin` directory.

### File Associations
The desktop application can be registered as the default application for opening .msg and .eml files. This allows you to double-click on these files in your file explorer to open them directly in msgReader.

You can register file associations by clicking the "Register as Default App" button on the welcome screen of the desktop application.

## Development (Web)
1. Clone the repository
```bash
git clone https://github.com/Rasalas/msg-reader.git
cd msg-reader
```

2. Install the dependencies
```bash
npm install 
```

3. Run the application in development mode
```bash
npm run dev
```

A browser window should open with the application running. The application will automatically reload when changes are made to the source code.

## Build Process
The application uses browserify to bundle the JavaScript modules and tailwindcss for styling.

### Build Commands
- `npm run build` - Builds both JavaScript and CSS
- `npm run build:js` - Bundles JavaScript modules
- `npm run build:css` - Compiles Tailwind CSS
- `npm run watch` - Watches for changes and rebuilds
- `npm run dev` - Runs development server with live reload
- `npm run deploy` - Deploys to GitHub Pages

For the desktop application, use the scripts in the `desktop` directory:
- `./desktop/dev.sh` - Starts the Wails development server
- `./desktop/build.sh` - Builds the Wails desktop application

## Other links

[SourceForge | MsgViewer](https://sourceforge.net/projects/msgviewer/postdownload)  
Java app. Works, basically my favourite, but doesn't show inline images

[encryptomatic.com | Free Online .msg Viewer](https://www.encryptomatic.com/viewer/)  
Kinda sus. I don't really trust them, because they sell the ~same thing as a [product](https://www.encryptomatic.com/msgviewer/msgviewerpro.html). You also can't see inline images and you get an ad in the end of the email instad of in the page itself.

[GitHub | datenteiler/ReadMsgFile](https://github.com/datenteiler/ReadMsgFile)  
Seems to be ok, but it only shows the text version of the email. No inline images or HTML because it uses a command line interface.

[MS Store | MSG Viewer](https://apps.microsoft.com/detail/9nwsk3187kv3?hl=de-DE&gl=DE)  
Costst money and doesn't look promising.

## "Receipts"
account.microsoft.com account page showing an active subsciption. Next payment 26th March 2025 for 69€
![account.microsoft.com account page showing an active subsciption. Next payment 26th March 2025 for 69€](doc/res/microsoft-accounts-webpage.png)

a table showing payments of the last three years. Last payment of 69€ on 26th March 2024
![a table showing payments of the last three years. Last payment of 69€ on 26th March 2024](doc/res/abrechnungsverlauf.png)

Windows 11 Account page showing an active Microsoft 365 Single subscription
![Windows 11 Account page showing an active Microsoft 365 Single subscription](doc/res/windows-account-page.png)

An error message stating that the msg file can't be opened, because it requires an active subscription
![An error message stating that the msg file can't be opened, because it requires an active subscription](doc/res/new-outlook-error-message.png)

## Completed Features
- [x] Allow to upload multiple files at once
- [x] Drop area fills the whole screen
- [x] Show a list of all imported emails on the side
- [x] Sort by date
- [x] Show a preview of the currently selected email
- [x] Separate subject, recipients & sender, body, attachments
- [x] Pin important messages

## Future Improvements (maybe)
- [ ] Allow to download the email as a .eml file
- [ ] Add search functionality
- [ ] Add filters
  - [ ] pinned messages
  - [ ] attachments
  - [ ] sender
  - [ ] subject
- [ ] Add message categories/tags
- [ ] Add keyboard shortcuts
- [ ] Add dark mode support