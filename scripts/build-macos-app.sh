#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APP_NAME="YAML Struct Editor"
BUNDLE_DIR="$ROOT_DIR/dist/${APP_NAME}.app"
CONTENTS_DIR="$BUNDLE_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"

rm -rf "$BUNDLE_DIR"
mkdir -p "$MACOS_DIR"

(cd "$ROOT_DIR/frontend" && npm run build)
(cd "$ROOT_DIR" && go build -mod=vendor -tags production -trimpath -buildvcs=false -ldflags="-w -s" -o "$MACOS_DIR/yaml-struct-editor" ./cmd/yaml-struct-editor)

cat > "$CONTENTS_DIR/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>ja</string>
	<key>CFBundleExecutable</key>
	<string>yaml-struct-editor</string>
	<key>CFBundleIdentifier</key>
	<string>com.mryskyj.yaml-struct-editor</string>
	<key>CFBundleName</key>
	<string>YAML Struct Editor</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>0.1.0</string>
	<key>CFBundleVersion</key>
	<string>0.1.0</string>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
PLIST

echo "Built: $BUNDLE_DIR"
echo "Open with: open \"$BUNDLE_DIR\""
