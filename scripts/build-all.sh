#!/bin/bash
#
# Multi-platform build script for MCP Manager
# Usage: ./scripts/build-all.sh [windows] [macos] [linux]
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR="build/releases/${VERSION}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

echo -e "${GREEN}=== MCP Manager Multi-Platform Build ===${NC}"
echo "Version: ${VERSION}"
echo "Build Directory: ${BUILD_DIR}"
echo ""

# Clean previous builds
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# Function to build for a platform
build_platform() {
    local platform=$1
    local arch=$2
    local output_name=$3

    echo -e "${YELLOW}Building for ${platform}/${arch}...${NC}"

    if wails build -clean -platform "${platform}/${arch}" -ldflags "-s -w" -o "${output_name}"; then
        echo -e "${GREEN}✓ Build successful for ${platform}/${arch}${NC}"
        return 0
    else
        echo -e "${RED}✗ Build failed for ${platform}/${arch}${NC}"
        return 1
    fi
}

# Function to create archive
create_archive() {
    local platform=$1
    local file=$2
    local archive=$3

    cd build/bin
    if [ -f "${file}" ] || [ -d "${file}" ]; then
        if [[ "${platform}" == "windows" ]]; then
            zip -r "../../${BUILD_DIR}/${archive}.zip" "${file}"
        else
            tar -czf "../../${BUILD_DIR}/${archive}.tar.gz" "${file}"
        fi
        echo -e "${GREEN}✓ Created archive: ${archive}${NC}"
    else
        echo -e "${RED}✗ File not found: ${file}${NC}"
    fi
    cd ../..
}

# Parse arguments (default: build all)
PLATFORMS=("$@")
if [ ${#PLATFORMS[@]} -eq 0 ]; then
    PLATFORMS=("windows" "macos" "linux")
fi

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    case "${platform}" in
        windows|win)
            if build_platform "windows" "amd64" "mcpmanager.exe"; then
                create_archive "windows" "mcpmanager.exe" "mcpmanager-${VERSION}-windows-amd64"
            fi
            ;;

        macos|darwin|mac)
            if build_platform "darwin" "universal" "mcpmanager.app"; then
                create_archive "macos" "mcpmanager.app" "mcpmanager-${VERSION}-macos-universal"
            fi
            ;;

        linux)
            if build_platform "linux" "amd64" "mcpmanager"; then
                create_archive "linux" "mcpmanager" "mcpmanager-${VERSION}-linux-amd64"
            fi
            ;;

        *)
            echo -e "${RED}Unknown platform: ${platform}${NC}"
            echo "Supported platforms: windows, macos, linux"
            exit 1
            ;;
    esac
done

# Generate checksums
if [ -d "${BUILD_DIR}" ] && [ "$(ls -A ${BUILD_DIR})" ]; then
    echo ""
    echo -e "${YELLOW}Generating checksums...${NC}"
    cd "${BUILD_DIR}"
    sha256sum * > SHA256SUMS
    echo -e "${GREEN}✓ Checksums generated${NC}"
    cd -
fi

# Summary
echo ""
echo -e "${GREEN}=== Build Complete ===${NC}"
echo "Artifacts location: ${BUILD_DIR}"
echo ""
ls -lh "${BUILD_DIR}"

# Display checksums
if [ -f "${BUILD_DIR}/SHA256SUMS" ]; then
    echo ""
    echo -e "${YELLOW}SHA256 Checksums:${NC}"
    cat "${BUILD_DIR}/SHA256SUMS"
fi

echo ""
echo -e "${GREEN}✓ All builds completed successfully!${NC}"
