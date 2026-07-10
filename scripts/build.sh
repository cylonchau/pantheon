#!/bin/bash

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/..; pwd)
set -e

OUT_DIR="target/bin"
GO_CMD=$(command -v go || echo "")

if [ -z "$GO_CMD" ]; then
    echo "Error: 'go' command not found."
    exit 1
fi

BINARY_NAME=$1

# 确认传入的模块名
if [ -z "$BINARY_NAME" ]; then
    echo "Error: No module name specified."
    echo "Usage: $0 <module_name>"
    exit 1
fi

# 确定编译路径和二进制名称
if [ -d "${PROJECT_ROOT}/cmd/${BINARY_NAME}" ]; then
    SRC_PATH="${PROJECT_ROOT}/cmd/${BINARY_NAME}"
else
    echo "Error: Could not find source for module '${BINARY_NAME}'."
    echo "Available modules: $(ls ${PROJECT_ROOT}/cmd)"
    exit 119
fi

[ -d "${PROJECT_ROOT}/${OUT_DIR}" ] || mkdir -p "${PROJECT_ROOT}/${OUT_DIR}"

# 获取版本号
VERSION=$(git describe --tags --always 2>/dev/null || echo "v0.0.0-dev")

# 编译 (使用本地平台默认设置，方便本地测试；发布时可通过环境变量指定)
TARGET_OS=${GOOS:-$(go env GOHOSTOS)}
TARGET_ARCH=${GOARCH:-$(go env GOHOSTARCH)}
HOST_OS=$(go env GOHOSTOS)
HOST_ARCH=$(go env GOHOSTARCH)

# 根据平台决定扩展名
EXTENSION=""
if [ "$TARGET_OS" == "windows" ]; then
    EXTENSION=".exe"
fi

BINARY_PATH="${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}"
# 如果是跨平台编译，在文件名中加入 OS 和 ARCH
if [ "$TARGET_OS" != "$HOST_OS" ] || [ "$TARGET_ARCH" != "$HOST_ARCH" ]; then
    BINARY_PATH="${BINARY_PATH}-${TARGET_OS}-${TARGET_ARCH}${EXTENSION}"
else
    BINARY_PATH="${BINARY_PATH}${EXTENSION}"
fi

# 编译
cd "${PROJECT_ROOT}" && \
    CGO_ENABLED=0 GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} ${GO_CMD} build \
    -ldflags "-s -w -X 'github.com/cylonchau/pantheon/pkg/version.Version=${VERSION}'" \
    -o "${BINARY_PATH}" "${SRC_PATH}"

if command -v upx >/dev/null 2>&1; then
    if [[ "$TARGET_OS" == "darwin" ]]; then
       echo "Skipping UPX compression for macOS to ensure stability."
    else
       upx -1 "${BINARY_PATH}" 2>/dev/null || true
    fi
fi

echo "Done building ${BINARY_NAME} for ${TARGET_OS}/${TARGET_ARCH}."