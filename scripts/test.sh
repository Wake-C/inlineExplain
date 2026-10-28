#!/bin/zsh
set -euo pipefail
cd "${0:A:h:h}"
INLINEEXPLAIN_GO=${INLINEEXPLAIN_GO:-go}
INLINEEXPLAIN_TEST_DIR=${INLINEEXPLAIN_TEST_DIR:-$HOME/Documents/tmp/inlineExplain}
mkdir -p "$INLINEEXPLAIN_TEST_DIR/native-build"
python3 scripts/check_identity.py
node scripts/test_frontend.cjs
python3 scripts/test_build.py
"$INLINEEXPLAIN_GO" test -race ./...
xcrun swiftc -O -swift-version 5 native/ContextWindow.swift native/ContextWindowTests.swift \
  -o "$INLINEEXPLAIN_TEST_DIR/native-build/context-tests"
"$INLINEEXPLAIN_TEST_DIR/native-build/context-tests"
xcrun swiftc -O -swift-version 5 -target arm64-apple-macos15.0 \
  native/ContextWindow.swift native/Capture.swift native/OCRTests.swift \
  -o "$INLINEEXPLAIN_TEST_DIR/native-build/ocr-tests"
"$INLINEEXPLAIN_TEST_DIR/native-build/ocr-tests"
if test -n "${INLINEEXPLAIN_LIVE_KEYCHAIN_SERVICE:-}" && test -n "${INLINEEXPLAIN_LIVE_KEYCHAIN_ACCOUNT:-}"; then
  "$INLINEEXPLAIN_GO" test -count=1 -tags live ./... -v
else
  print 'NOT RUN: live provider checks. Select an authorized JSON Keychain profile with INLINEEXPLAIN_LIVE_KEYCHAIN_SERVICE and INLINEEXPLAIN_LIVE_KEYCHAIN_ACCOUNT.'
fi
print "Test artifacts: $INLINEEXPLAIN_TEST_DIR"
