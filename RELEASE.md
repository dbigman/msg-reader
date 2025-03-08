# Release

This project uses GitHub Actions to automatically build and release the application for multiple platforms (macOS, Windows, and Linux).

### Release Process

1. **Prepare your changes**
   - Make sure all your changes are committed
   - Update the version number in relevant files
   - Update the CHANGELOG.md if it exists

2. **Create and push a new tag**
   ```bash
   # Create a new tag
   git tag v1.0.0  # Replace with your version number

   # Push the tag to GitHub
   git push origin v1.0.0
   ```

3. **Automated Build Process**
   Once you push a tag, GitHub Actions will automatically:
   - Build the application for all supported platforms:
     - macOS (Universal Binary)
     - Windows (64-bit)
     - Linux (64-bit)
   - Package each build into a ZIP file
   - Create a draft release
   - Attach the built packages to the release

4. **Review and Publish**
   - Go to the GitHub repository's "Releases" page
   - Find the draft release created by the workflow
   - Review the automatically generated release notes
   - Make any necessary adjustments to the release description
   - Click "Publish release" when ready

### Build Artifacts

The following artifacts will be available in each release:
- `msgReader-macos-universal.zip` - macOS application bundle
- `msgReader-windows-amd64.zip` - Windows executable
- `msgReader-linux-amd64.zip` - Linux binary

### Troubleshooting

If the automated build fails:
1. Check the GitHub Actions logs for error messages
2. Ensure all dependencies are properly specified
3. Verify that the version tag follows the format `v*.*.*`
4. Make sure all required files are committed

### Development Builds

For development and testing, you can build the application locally using:
```bash
cd desktop
./build.sh  # On macOS/Linux
build.bat   # On Windows
```