#!/bin/bash

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE[0]})/..; pwd)
OUT_DIR="_output"
GO_CMD=$(which go)
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
fi\

[ -d "${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}" ] && rm -fr ${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}
[ -d "${PROJECT_ROOT}/${OUT_DIR}" ] || mkdir -pv "${PROJECT_ROOT}/${OUT_DIR}"

cd "${PROJECT_ROOT}/cmd/${BINARY_NAME}" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO_CMD} build -ldflags "-s -w" -o "${PROJECT_ROOT}/${OUT_DIR}/${BINARY_NAME}" main.go