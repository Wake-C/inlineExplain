#!/bin/zsh
set -euo pipefail
cd "${0:A:h:h}"
INLINEEXPLAIN_GO=${INLINEEXPLAIN_GO:-go}
if ! command -v "$INLINEEXPLAIN_GO" >/dev/null 2>&1; then
  print -u2 'Install the Go version in go.mod, or set INLINEEXPLAIN_GO to its executable path.'
  exit 1
fi
python3 scripts/check_identity.py
APP_VERSION=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' packaging/Info.plist)
BUILD_ROOT="$PWD/build"
mkdir -p "$BUILD_ROOT"
WORK=$(mktemp -d "$BUILD_ROOT/.InlineExplain-build.XXXXXX")
FINAL_APP="$BUILD_ROOT/InlineExplain.app"
PREVIOUS_APP="$WORK/previous.app"
cleanup() {
  if test -d "$PREVIOUS_APP" && ! test -e "$FINAL_APP"; then mv "$PREVIOUS_APP" "$FINAL_APP"; fi
  rm -rf "$WORK"
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
APP="$WORK/InlineExplain.app"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 "$INLINEEXPLAIN_GO" build -trimpath \
  -ldflags "-s -w -X main.appVersion=$APP_VERSION" -o "$APP/Contents/MacOS/inlineexplain" .
xcrun swiftc -O -swift-version 5 -target arm64-apple-macos15.0 \
  native/ContextWindow.swift native/Capture.swift native/Helper.swift \
  -o "$APP/Contents/MacOS/InlineExplainOCR"
cp packaging/Info.plist "$APP/Contents/Info.plist"
ICONSET="$WORK/InlineExplain.iconset"
mkdir -p "$ICONSET"
for size in 16 32 128 256 512; do
  sips -z "$size" "$size" assets/icon.png --out "$ICONSET/icon_${size}x${size}.png" >/dev/null
  retina=$((size * 2))
  sips -z "$retina" "$retina" assets/icon.png --out "$ICONSET/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil --convert icns "$ICONSET" --output "$APP/Contents/Resources/InlineExplain.icns"
for locale in zh-Hans en; do
  mkdir -p "$APP/Contents/Resources/$locale.lproj"
  cp "packaging/$locale.lproj/InfoPlist.strings" "$APP/Contents/Resources/$locale.lproj/InfoPlist.strings"
done
cp NOTICE "$APP/Contents/Resources/NOTICE"
if test -f LICENSE; then cp LICENSE "$APP/Contents/Resources/LICENSE"; fi
mkdir -p "$APP/Contents/Resources/ThirdParty"
"$INLINEEXPLAIN_GO" list -m -f '{{if and (not .Main) .Dir}}{{.Dir}}{{end}}' all | while IFS= read -r module_dir; do
  for notice in "$module_dir/LICENSE" "$module_dir/COPYING" "$module_dir/NOTICE"; do
    if test -f "$notice"; then cp -f "$notice" "$APP/Contents/Resources/ThirdParty/${module_dir:t}-${notice:t}"; fi
  done
done
SIGN_IDENTITY=${INLINEEXPLAIN_SIGN_IDENTITY:--}
codesign --force --sign "$SIGN_IDENTITY" --identifier com.inlineexplain.desktop.ocr "$APP/Contents/MacOS/InlineExplainOCR"
codesign --force --sign "$SIGN_IDENTITY" --identifier com.inlineexplain.desktop "$APP"
codesign --verify --deep --strict "$APP"
python3 scripts/check_identity.py --app "$APP"
if test -e "$FINAL_APP"; then mv "$FINAL_APP" "$PREVIOUS_APP"; fi
mv "$APP" "$FINAL_APP"
print "Built: $FINAL_APP"
