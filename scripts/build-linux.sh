#!/bin/bash
# Build VoiceCode for Linux.
# Prerequisites: sudo apt install portaudio19-dev libx11-dev xdotool
# Usage: ./scripts/build-linux.sh
set -euo pipefail

cd "$(dirname "$0")/.."

APP_NAME="voicecode"
BUILD_DIR="build"

echo "Building $APP_NAME for Linux..."

# Check dependencies
for dep in pkg-config; do
    if ! command -v "$dep" &>/dev/null; then
        echo "Error: $dep is required. Install with: sudo apt install $dep"
        exit 1
    fi
done

if ! pkg-config --exists portaudio-2.0 2>/dev/null; then
    echo "Warning: portaudio-2.0 not found. Install with: sudo apt install portaudio19-dev"
fi

mkdir -p "$BUILD_DIR"

go build -ldflags="-s -w" -o "$BUILD_DIR/$APP_NAME" ./cmd/voicecode

echo "Built: $BUILD_DIR/$APP_NAME"
ls -lh "$BUILD_DIR/$APP_NAME"
