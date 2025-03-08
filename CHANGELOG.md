# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-03-06

### Added
- Desktop application support using Wails and Go
- File associations for .msg and .eml files
- Command-line support for opening files
- macOS native file handling support
- Automated builds for Windows, macOS, and Linux
- GitHub Actions workflow for automated releases

### Changed
- Restructured project to support both web and desktop versions
- Improved file handling with better error messages
- Enhanced drag and drop functionality

### Fixed
- File path handling on different operating systems
- Special character handling in file names
- macOS-specific command line argument handling

## [0.1.0] - 2024-01-24

### Added
- Initial web application release
- Support for .msg and .eml file formats
- HTML content and inline image display
- Message pinning functionality
- Multiple file support with message list
- Date-based message sorting
- Drag and drop support
- Basic message viewer interface
- Message list with sorting capabilities
- Pin/unpin functionality for important messages
- File upload through button or drag & drop
- HTML email rendering
- Inline image support
- Attachment handling

[1.0.0]: https://github.com/Rasalas/msg-reader/releases/tag/v1.0.0
[0.1.0]: https://github.com/Rasalas/msg-reader/releases/tag/v0.1.0 