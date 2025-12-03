#!/bin/bash

# Build script - Build binaries for multiple platforms
# Usage: ./scripts/build.sh [version]

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_DIR="build"
BINARY_NAME="dudu-proxy"

# Supported platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}DuDu Proxy Build Script${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Version: ${YELLOW}${VERSION}${NC}"
echo -e "Build directory: ${YELLOW}${BUILD_DIR}${NC}"
echo ""

# Clean and create build directory
echo -e "${YELLOW}Cleaning build directory...${NC}"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# Build function
build_binary() {
    local platform=$1
    local goos=$(echo $platform | cut -d'/' -f1)
    local goarch=$(echo $platform | cut -d'/' -f2)
    
    local output_name="${BINARY_NAME}-${VERSION}-${goos}-${goarch}"
    
    # Add .exe suffix for Windows
    if [ "$goos" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    local output_path="${BUILD_DIR}/${output_name}"
    
    echo -e "${YELLOW}Building ${goos}/${goarch}...${NC}"
    
    # Build
    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -trimpath \
        -o "${output_path}" \
        main.go
    
    if [ $? -eq 0 ]; then
        # Show file info
        local size=$(ls -lh "${output_path}" | awk '{print $5}')
        echo -e "${GREEN}✓ ${output_name} (${size})${NC}"
        
        # Create ZIP archive
        echo -e "${YELLOW}  Creating ZIP archive...${NC}"
        
        # Determine simple binary name for inside ZIP
        if [ "$goos" = "windows" ]; then
            simple_name="dudu-proxy.exe"
        else
            simple_name="dudu-proxy"
        fi
        
        # Copy and rename files for ZIP
        cp "${output_path}" "${BUILD_DIR}/${simple_name}"
        cp configs/config.example.json "${BUILD_DIR}/config.json"
        
        # Create ZIP with simplified names
        (cd "${BUILD_DIR}" && zip -q "${output_name}.zip" "${simple_name}" config.json)
        
        # Clean up temporary files
        rm "${BUILD_DIR}/${simple_name}" "${BUILD_DIR}/config.json"
        
        local zip_size=$(ls -lh "${BUILD_DIR}/${output_name}.zip" | awk '{print $5}')
        echo -e "${GREEN}  ✓ ${output_name}.zip (${zip_size})${NC}"
        
        # Generate checksums
        if command -v sha256sum &> /dev/null; then
            (cd "${BUILD_DIR}" && sha256sum "${output_name}" > "${output_name}.sha256")
            (cd "${BUILD_DIR}" && sha256sum "${output_name}.zip" > "${output_name}.zip.sha256")
        elif command -v shasum &> /dev/null; then
            (cd "${BUILD_DIR}" && shasum -a 256 "${output_name}" > "${output_name}.sha256")
            (cd "${BUILD_DIR}" && shasum -a 256 "${output_name}.zip" > "${output_name}.zip.sha256")
        fi
    else
        echo -e "${RED}✗ Build failed: ${goos}/${goarch}${NC}"
        return 1
    fi
}

# Build all platforms
echo -e "${GREEN}Building all platform versions...${NC}"
echo ""

for platform in "${PLATFORMS[@]}"; do
    build_binary "$platform"
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Build complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Generate combined checksums file
echo -e "${YELLOW}Generating checksums file...${NC}"
if command -v sha256sum &> /dev/null; then
    (cd "${BUILD_DIR}" && sha256sum ${BINARY_NAME}-* | grep -v ".sha256" > checksums.txt)
elif command -v shasum &> /dev/null; then
    (cd "${BUILD_DIR}" && shasum -a 256 ${BINARY_NAME}-* | grep -v ".sha256" > checksums.txt)
fi

# Show build results
echo -e "${GREEN}Build artifacts:${NC}"
ls -lh "${BUILD_DIR}" | grep -v "^total" | awk '{printf "  %s  %s\n", $5, $9}'

echo ""
echo -e "${GREEN}Checksums:${NC}"
if [ -f "${BUILD_DIR}/checksums.txt" ]; then
    cat "${BUILD_DIR}/checksums.txt" | while read line; do
        echo "  $line"
    done
fi

echo ""
echo -e "${GREEN}All files saved to ${BUILD_DIR}/ directory${NC}"
