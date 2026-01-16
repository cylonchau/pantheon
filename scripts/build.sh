#!/bin/bash

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/..; pwd)
set -e

OUT_DIR="target"
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

# 检查模块是否存在
if [ ! -d "${PROJECT_ROOT}/cmd/${BINARY_NAME}" ]; then
    echo "Error: Invalid module name '${BINARY_NAME}'."
    echo "Available modules are: $(ls ${PROJECT_ROOT}/cmd)"
    exit 119
fi

[ -d "${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}" ] && rm -fr ${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}
[ -d "${PROJECT_ROOT}/${OUT_DIR}" ] || mkdir -p "${PROJECT_ROOT}/${OUT_DIR}"

# 获取版本号
VERSION=$(git describe --tags --always 2>/dev/null || echo "v0.0.0-dev")

# 编译 (使用本地平台默认设置，方便本地测试；发布时可通过环境变量指定)
TARGET_OS=${GOOS:-$(go env GOOS)}
TARGET_ARCH=${GOARCH:-$(go env GOARCH)}

cd "${PROJECT_ROOT}/cmd/${BINARY_NAME}" && \
    CGO_ENABLED=0 GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} ${GO_CMD} build \
    -ldflags "-s -w -X 'github.com/cylonchau/pantheon/pkg/version.Version=${VERSION}'" \
    -o "${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}" main.go

if command -v upx >/dev/null 2>&1; then
    if [[ "$TARGET_OS" == "darwin" ]]; then
       echo "Skipping UPX compression for macOS to ensure stability."
    else
       upx -1 "${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}"
    fi
fi

echo "Done building ${BINARY_NAME}."