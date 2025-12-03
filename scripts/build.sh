#!/bin/bash

# 构建脚本 - 用于本地构建多平台版本
# 使用方法: ./scripts/build.sh [version]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取版本号
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_DIR="build"
BINARY_NAME="dudu-proxy"

# 支持的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}DuDu Proxy 构建脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "版本: ${YELLOW}${VERSION}${NC}"
echo -e "构建目录: ${YELLOW}${BUILD_DIR}${NC}"
echo ""

# 清理并创建构建目录
echo -e "${YELLOW}清理构建目录...${NC}"
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# 构建函数
build_binary() {
    local platform=$1
    local goos=$(echo $platform | cut -d'/' -f1)
    local goarch=$(echo $platform | cut -d'/' -f2)
    
    local output_name="${BINARY_NAME}-${VERSION}-${goos}-${goarch}"
    
    # Windows 需要 .exe 后缀
    if [ "$goos" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    local output_path="${BUILD_DIR}/${output_name}"
    
    echo -e "${YELLOW}构建 ${goos}/${goarch}...${NC}"
    
    # 构建
    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -trimpath \
        -o "${output_path}" \
        main.go
    
    if [ $? -eq 0 ]; then
        # 显示文件信息
        local size=$(ls -lh "${output_path}" | awk '{print $5}')
        echo -e "${GREEN}✓ ${output_name} (${size})${NC}"
        
        # 生成校验和
        if command -v sha256sum &> /dev/null; then
            (cd "${BUILD_DIR}" && sha256sum "${output_name}" > "${output_name}.sha256")
        elif command -v shasum &> /dev/null; then
            (cd "${BUILD_DIR}" && shasum -a 256 "${output_name}" > "${output_name}.sha256")
        fi
    else
        echo -e "${RED}✗ 构建失败: ${goos}/${goarch}${NC}"
        return 1
    fi
}

# 构建所有平台
echo -e "${GREEN}开始构建所有平台版本...${NC}"
echo ""

for platform in "${PLATFORMS[@]}"; do
    build_binary "$platform"
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}构建完成!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 生成总校验和文件
echo -e "${YELLOW}生成校验和文件...${NC}"
if command -v sha256sum &> /dev/null; then
    (cd "${BUILD_DIR}" && sha256sum ${BINARY_NAME}-* | grep -v ".sha256" > checksums.txt)
elif command -v shasum &> /dev/null; then
    (cd "${BUILD_DIR}" && shasum -a 256 ${BINARY_NAME}-* | grep -v ".sha256" > checksums.txt)
fi

# 显示构建结果
echo -e "${GREEN}构建产物:${NC}"
ls -lh "${BUILD_DIR}" | grep -v "^total" | awk '{printf "  %s  %s\n", $5, $9}'

echo ""
echo -e "${GREEN}校验和:${NC}"
if [ -f "${BUILD_DIR}/checksums.txt" ]; then
    cat "${BUILD_DIR}/checksums.txt" | while read line; do
        echo "  $line"
    done
fi

echo ""
echo -e "${GREEN}所有文件已保存到 ${BUILD_DIR}/ 目录${NC}"
