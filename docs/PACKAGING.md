# MCP Manager - Production Packaging Guide

## Overview

This document describes how to build and package MCP Manager for production distribution across Windows, macOS, and Linux platforms.

## Prerequisites

### All Platforms
- Go 1.21 or later
- Node.js 18 or later
- Wails CLI v2.x: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Windows
- Windows 10/11 or Windows Server 2019+
- WebView2 Runtime (automatically installed on Windows 11)
- MSYS2/MinGW-w64 (for CGO)

### macOS
- macOS 10.15 (Catalina) or later
- Xcode Command Line Tools
- Apple Developer account (for code signing)

### Linux
- GTK+3
- WebKitGTK
- Build essentials: `sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev`

## Quick Build

### Development Build
```bash
# Build for current platform only
wails build
```

### Production Build
```bash
# Clean build with optimizations
wails build -clean -production
```

## Platform-Specific Builds

### Windows

```bash
# Build Windows executable
wails build -platform windows/amd64 -clean

# Output: build/bin/mcpmanager.exe
```

**Windows Installer** (Optional):
- Use Inno Setup or NSIS to create installer
- Include WebView2 bootstrapper
- Example Inno Setup script: `scripts/windows-installer.iss`

### macOS

```bash
# Build macOS application bundle
wails build -platform darwin/universal -clean

# Output: build/bin/mcpmanager.app
```

**Code Signing**:
```bash
# Sign the application
codesign --deep --force --verify --verbose --sign "Developer ID Application: Your Name" build/bin/mcpmanager.app

# Create DMG
hdiutil create -volname "MCP Manager" -srcfolder build/bin/mcpmanager.app -ov -format UDZO mcpmanager-macos.dmg
```

**Notarization** (required for distribution):
```bash
# Submit for notarization
xcrun notarytool submit mcpmanager-macos.dmg --apple-id your@email.com --team-id TEAMID --password app-specific-password

# Staple the ticket
xcrun stapler staple mcpmanager-macos.dmg
```

### Linux

```bash
# Build Linux binary
wails build -platform linux/amd64 -clean

# Output: build/bin/mcpmanager
```

**Package Formats**:

**DEB (Debian/Ubuntu)**:
```bash
# See scripts/build-deb.sh
./scripts/build-deb.sh
```

**RPM (Fedora/RHEL)**:
```bash
# See scripts/build-rpm.sh
./scripts/build-rpm.sh
```

**AppImage**:
```bash
# See scripts/build-appimage.sh
./scripts/build-appimage.sh
```

## Multi-Platform Build Script

Use the provided build script for automated multi-platform builds:

```bash
# Build for all platforms
./scripts/build-all.sh

# Build for specific platforms
./scripts/build-all.sh windows macos linux
```

## Release Artifacts

After building, create release artifacts:

```bash
# Windows
cd build/bin
zip mcpmanager-windows-amd64.zip mcpmanager.exe

# macOS (after creating DMG)
# mcpmanager-macos-universal.dmg

# Linux
cd build/bin
tar -czf mcpmanager-linux-amd64.tar.gz mcpmanager
```

## Build Configuration

### wails.json

Key configuration options:

```json
{
  "outputfilename": "mcpmanager",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "wailsjsdir": "./frontend",
  "assetdir": "./frontend/dist",
  "reloaddirs": "frontend/src"
}
```

### Build Flags

- `-clean`: Remove build directory before building
- `-production`: Enable production mode (minification, optimization)
- `-upx`: Compress binary with UPX (reduces size by ~50%)
- `-ldflags`: Pass flags to Go linker (e.g., strip debug info)

Example with optimization:
```bash
wails build -clean -production -ldflags "-s -w" -upx
```

## Size Optimization

1. **Strip Debug Symbols**:
   ```bash
   wails build -ldflags "-s -w"
   ```

2. **UPX Compression**:
   ```bash
   wails build -upx
   ```

3. **Frontend Optimization**:
   - Ensure `npm run build` includes minification
   - Remove unused dependencies
   - Use tree-shaking

## Continuous Deployment

The CI/CD pipeline (`.github/workflows/ci.yml`) automatically builds for all platforms on:
- Push to `main` branch
- Release tags (`v*`)

Artifacts are uploaded and available for download from the Actions tab.

## Version Management

Update version in:
1. `wails.json` - `info.productVersion`
2. `frontend/package.json` - `version`
3. Git tag: `git tag v0.1.0`

## Checksums

Generate checksums for release artifacts:

```bash
# SHA256 checksums
sha256sum mcpmanager-* > SHA256SUMS

# Verify
sha256sum -c SHA256SUMS
```

## Distribution

### GitHub Releases
1. Create a new release on GitHub
2. Upload all platform artifacts
3. Include SHA256SUMS file
4. Add release notes

### Alternative Distribution
- Homebrew (macOS): Create Formula
- Chocolatey (Windows): Create package
- Snap Store (Linux): Create snap
- Flathub (Linux): Create flatpak

## Troubleshooting

### Build Fails

**Issue**: Missing dependencies
```bash
# Windows: Install MSYS2/MinGW
# macOS: xcode-select --install
# Linux: apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev
```

**Issue**: Frontend build fails
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Binary Too Large

- Use `-upx` flag
- Strip debug symbols with `-ldflags "-s -w"`
- Optimize frontend assets

### Permissions Issues (macOS/Linux)

```bash
chmod +x build/bin/mcpmanager
```

## Security Considerations

1. **Code Signing**: Always sign binaries for distribution
2. **Checksums**: Provide SHA256 checksums for verification
3. **Secure Distribution**: Use HTTPS for downloads
4. **Updates**: Implement secure update mechanism

## Performance Targets

Per project requirements:
- **Startup Time**: < 2 seconds (FR-037)
- **Idle Memory**: < 100 MB (FR-039)
- **Binary Size**: < 50 MB (optimized with UPX)

Verify with performance benchmarks:
```bash
go test ./tests/performance -v
```

## Support

For build issues, see:
- Wails Documentation: https://wails.io
- Project Issues: GitHub Issues
- Build logs: `.github/workflows/ci.yml` runs
