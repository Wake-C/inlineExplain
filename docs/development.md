# Development

Use macOS 15 or later with Xcode Command Line Tools, Python 3 and the Go
version declared in [`go.mod`](../go.mod). Automated frontend checks also use
Node.js 18 or later. The app targets Apple Silicon. The frontend has no package
install or build step.

```sh
xcode-select --install
./scripts/build.sh
open build/InlineExplain.app
```

If Go is not on `PATH`, set `INLINEEXPLAIN_GO` to its executable path. The build
uses optimized Go and Swift binaries, generates the macOS icon from
[`assets/icon.png`](../assets/icon.png), signs the app and checks its bundle.
[`packaging/Info.plist`](../packaging/Info.plist) owns the release version and
build number; the build injects that version into the app and Settings.
A raw `go run .` uses the development version `dev`.

For normal use, copy the complete app to Applications before granting Screen
Recording access. Keep the app's bundle and helper identifiers stable when
changing packaging.

## Automated checks

```sh
./scripts/test.sh
```

The script runs frontend Settings checks, build and package failure checks,
Go race tests, Swift context tests and a real Apple Vision OCR fixture. It
writes temporary artifacts to `~/Documents/tmp/inlineExplain`; set
`INLINEEXPLAIN_TEST_DIR` to use another task directory.

Live model checks are optional and use an explicitly authorized JSON profile
in macOS Keychain. Set `INLINEEXPLAIN_LIVE_KEYCHAIN_SERVICE` and
`INLINEEXPLAIN_LIVE_KEYCHAIN_ACCOUNT` to select it. The entry must contain
`endpoint` and `apiKey`; `model` is optional. Tests use synthetic passages and
do not log the credential or provider response body. Never commit a provider
profile or use another account's credential without authorization.

Automated checks do not prove native focus, permissions or full-screen
behavior. The affected installed-app journeys and limits are in
[Architecture](architecture.md#runtime-and-verification-boundaries).

## Popup preview

The README illustration uses the production popup renderer with synthetic
data. Serve the repository root:

```sh
python3 -m http.server 8765 --bind 127.0.0.1
```

Open `http://127.0.0.1:8765/docs/popup-preview.html`. At 1280 × 800, wait for
`data-preview-ready="true"` on the document element before capturing
`docs/assets/popup-preview.png`. The preview supports Option+S, Escape and
clicking the example word. It does not capture the screen or call a provider.

For DMG generation and the manual GitHub workflow, see [Releasing](releasing.md).
