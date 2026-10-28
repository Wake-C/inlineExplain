# Releasing

[`scripts/package.sh`](../scripts/package.sh) is the release entry point for
both local use and GitHub Actions. It builds optimized ARM64 binaries, exports
the icon, signs the bundle, creates a compressed HFS+ DMG, mounts it read-only and
checks the app, version, architecture, icon, licenses and Applications link.
It replaces the previous `build/release` directory only after all package
checks pass. A failure keeps the previous valid package.

## Local package

Use the prerequisites in [Development](development.md), then run from the
repository root:

```sh
./scripts/package.sh
```

The executable also works when called by absolute path from another working
directory. Output is under the ignored `build/release/` directory:

- `InlineExplain-<version>-macos-arm64.dmg`
- `SHA256SUMS.txt`

The command builds only the release package. It does not run tests or create a
debug build. Run the relevant checks before packaging or publishing; a valid
DMG does not replace installed-app acceptance.

The default signature is ad-hoc, and packages are not notarized. For local
certificate signing, set `INLINEEXPLAIN_SIGN_IDENTITY` to a valid identity in
your Keychain. This does not add notarization. The GitHub workflow uses the
ad-hoc default and needs no signing secret.

## GitHub release

1. Update `CFBundleShortVersionString` and increment `CFBundleVersion` in
   [`packaging/Info.plist`](../packaging/Info.plist). Do not add version copies
   to Go, HTML or workflow files.
2. Verify the intended changes, commit and push them to `main`.
3. Open **Actions → Package release → Run workflow**, using `main`.

The workflow is manual only. Pushes, pull requests and schedules do not start
packaging, debug builds or test jobs. It calls the same local package command,
creates a draft release tagged `v<version>` at the selected source commit,
uploads the DMG and checksums, downloads them and checks their integrity, then
publishes the release. Third-party Actions are pinned to commit IDs; only the
release job receives `contents: write` through GitHub's temporary token.

An existing tag is not overwritten. Use a new version for another release. If
upload or verification fails after draft creation, inspect that draft and its
assets before taking further action; do not rerun publication blindly.

After publishing, download the release assets and check them:

```sh
shasum -a 256 -c SHA256SUMS.txt
hdiutil verify InlineExplain-<version>-macos-arm64.dmg
```

Also open the DMG and check the complete app from its installed location. Keep
artifact integrity checks, automated tests and native acceptance results
separate when reporting a release.
