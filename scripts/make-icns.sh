#!/bin/bash
# Convert PNG icon to macOS icns format.
# Usage: ./scripts/make-icns.sh <input.png> <output.icns>
set -euo pipefail

INPUT="${1:?Usage: make-icns.sh <input.png> <output.icns>}"
OUTPUT="${2:-AppIcon.icns}"

ICONSET=$(mktemp -d)/AppIcon.iconset
mkdir -p "$ICONSET"

for SIZE in 16 32 64 128 256 512; do
    sips -z "$SIZE" "$SIZE" "$INPUT" --out "$ICONSET/icon_${SIZE}x${SIZE}.png" >/dev/null 2>&1
done
for SIZE in 32 64 128 256 512 1024; do
    HALF=$((SIZE / 2))
    sips -z "$SIZE" "$SIZE" "$INPUT" --out "$ICONSET/icon_${HALF}x${HALF}@2x.png" >/dev/null 2>&1
done

iconutil -c icns -o "$OUTPUT" "$ICONSET"
rm -rf "$(dirname "$ICONSET")"
echo "Created: $OUTPUT"
