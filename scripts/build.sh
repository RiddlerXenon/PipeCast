#!/bin/bash

set -e  # Exit the script when any command fails
set -u  # Exit the script when using undeclared variables
set -o pipefail  # Exit the script when any command in a pipeline fails

# Version information for Go build
VERSION="1.0.0"
BUILD_TIME=$(date +%Y-%m-%d_%H:%M:%S)
COMMIT=$(git rev-parse HEAD)

# Paths
DM_PATH="internal/c_modules"
OUTPUT_DIR="$DM_PATH/obj"
LIB_DIR="$DM_PATH/lib"
BIN_DIR="bin"
GO_MAIN="cmd/pipecast/main.go"

# pkg-config flags
PW_FLAGS=$(pkg-config --cflags --libs libpipewire-0.3)
JSONC_FLAGS=$(pkg-config --cflags --libs json-c)

# Ensure necessary directories exist
mkdir -p "$OUTPUT_DIR" "$LIB_DIR" "$BIN_DIR"

echo ">>> Compiling C library..."
gcc -c "$DM_PATH/source/device_monitor.c" -o "$OUTPUT_DIR/dm.o" $PW_FLAGS $JSONC_FLAGS
ar rcs "$LIB_DIR/libdm.a" "$OUTPUT_DIR/dm.o"

echo ">>> Building Go application..."
go build -ldflags "-X main.version=$VERSION -X main.buildTime=$BUILD_TIME -X main.commit=$COMMIT" -o "$BIN_DIR/pipecast" "$GO_MAIN"

echo ">>> Build completed successfully!"

mkdir -p $HOME/.cache/pipecast
