#!/bin/zsh
set -euo pipefail
if (( $# == 1 )) && [[ "$1" == --help || "$1" == -h ]]; then
  print 'Usage: ./scripts/package.sh'
  print 'Build and verify an optimized ARM64 release DMG in build/release/.'
  print 'No tests, debug build, installation or publication are performed.'
  exit 0
fi
if (( $# != 0 )); then
  print -u2 'Usage: ./scripts/package.sh (or --help)'
  exit 2
fi
cd "${0:A:h:h}"

./scripts/build.sh
VERSION=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' packaging/Info.plist)
BUILD_ROOT="$PWD/build"
WORK=$(mktemp -d "$BUILD_ROOT/.InlineExplain-package.XXXXXX")
MOUNT="$WORK/mount"
FINAL_RELEASE="$BUILD_ROOT/release"
PREVIOUS_RELEASE="$WORK/previous-release"
cleanup() {
  if mount | /usr/bin/grep -Fq " on $MOUNT ("; then hdiutil detach "$MOUNT" -force >/dev/null; fi
  if test -d "$PREVIOUS_RELEASE" && ! test -e "$FINAL_RELEASE"; then mv "$PREVIOUS_RELEASE" "$FINAL_RELEASE"; fi
  rm -rf "$WORK"
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

PAYLOAD="$WORK/payload"
OUTPUT="$WORK/release"
mkdir -p "$PAYLOAD" "$OUTPUT" "$MOUNT"
ditto "$BUILD_ROOT/InlineExplain.app" "$PAYLOAD/InlineExplain.app"
ln -s /Applications "$PAYLOAD/Applications"
DMG_NAME="InlineExplain-$VERSION-macos-arm64.dmg"
DMG="$OUTPUT/$DMG_NAME"
hdiutil create -volname InlineExplain -srcfolder "$PAYLOAD" -fs HFS+ -format UDZO "$DMG"
hdiutil verify "$DMG"
hdiutil attach -nobrowse -readonly -mountpoint "$MOUNT" "$DMG"
python3 scripts/check_identity.py --app "$MOUNT/InlineExplain.app"
if ! test -L "$MOUNT/Applications" || test "$(readlink "$MOUNT/Applications")" != /Applications; then
  print -u2 'The DMG is missing its Applications link.'
  exit 1
fi
hdiutil detach "$MOUNT"
(
  cd "$OUTPUT"
  shasum -a 256 "$DMG_NAME" > SHA256SUMS.txt
)
if test -e "$FINAL_RELEASE"; then mv "$FINAL_RELEASE" "$PREVIOUS_RELEASE"; fi
mv "$OUTPUT" "$FINAL_RELEASE"
print "Packaged: $FINAL_RELEASE/$DMG_NAME"
print "Checksums: $FINAL_RELEASE/SHA256SUMS.txt"
