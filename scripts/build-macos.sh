#!/bin/bash
# Build VoiceCode.app for macOS.
# Usage: ./scripts/build-macos.sh [--dmg]
set -euo pipefail

cd "$(dirname "$0")/.."

APP_NAME="VoiceCode"
BUILD_DIR="build"
APP_DIR="$BUILD_DIR/$APP_NAME.app"
CONTENTS="$APP_DIR/Contents"
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")

echo "Building $APP_NAME $VERSION..."

# Clean
rm -rf "$APP_DIR"
mkdir -p "$CONTENTS/MacOS" "$CONTENTS/Resources"

# Build binary
echo "Compiling..."
go build -ldflags="-s -w" -o "$CONTENTS/MacOS/voicecode" ./cmd/voicecode

# Copy Info.plist
cp scripts/Info.plist "$CONTENTS/"

# Generate icns if source icon exists
if [ -f assets/icon_idle.png ]; then
    ./scripts/make-icns.sh assets/icon_idle.png "$CONTENTS/Resources/AppIcon.icns"
fi

# Sign
if [ -n "${DEVELOPER_ID:-}" ]; then
    echo "Signing with Developer ID: $DEVELOPER_ID"
    codesign --deep --force --options runtime \
        --sign "$DEVELOPER_ID" \
        --entitlements /dev/stdin "$APP_DIR" <<'ENTITLEMENTS'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.device.audio-input</key>
    <true/>
    <key>com.apple.security.automation.apple-events</key>
    <true/>
</dict>
</plist>
ENTITLEMENTS
else
    echo "Ad-hoc signing..."
    codesign --deep --force --sign - "$APP_DIR"
fi

echo "Built: $APP_DIR"
ls -lh "$CONTENTS/MacOS/voicecode"

# Create DMG if requested
if [[ "${1:-}" == "--dmg" ]]; then
    DMG_NAME="$BUILD_DIR/${APP_NAME}-${VERSION}.dmg"
    if command -v create-dmg &>/dev/null; then
        echo "Creating DMG..."
        create-dmg \
            --volname "$APP_NAME" \
            --window-size 500 300 \
            --icon "$APP_NAME.app" 150 150 \
            --app-drop-link 350 150 \
            "$DMG_NAME" "$APP_DIR"
        echo "DMG: $DMG_NAME"
    else
        echo "create-dmg not found. Creating simple DMG..."
        hdiutil create -volname "$APP_NAME" -srcfolder "$APP_DIR" \
            -ov -format UDZO "$DMG_NAME"
        echo "DMG: $DMG_NAME"
    fi
fi
