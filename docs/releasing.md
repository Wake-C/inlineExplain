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

Builds require the fixed code-signing certificate, named `InlineExplain Release`
by default. Set `INLINEEXPLAIN_SIGN_IDENTITY` to its SHA-1 fingerprint to select
it explicitly. Ad-hoc signing is rejected. Packages are not notarized.

## One-time signing setup

Use the same certificate and private key for local builds and GitHub releases.
Do not generate a new identity for each build or CI runner.

1. Open Keychain Access → Certificate Assistant → Create a Certificate.
   Name it `InlineExplain Release`, choose **Self Signed Root** and **Code
   Signing**. Choose a suitable validity period under the default overrides;
   keep the identity in the login Keychain. Do not enable system-wide trust.
2. Under **My Certificates**, export this identity (certificate and private
   key) as a password-protected `.p12`, using the interactive export dialog.
   Keep any recovery copy encrypted and outside the repository. Never commit
   the private key or export password.
3. Configure repository Actions secrets `MACOS_SIGNING_P12_BASE64` (base64 of
   the encrypted export) and `MACOS_SIGNING_P12_PASSWORD` (its export password).
   Use secure input, not secret values in shell arguments or logs. Delete the
   temporary export after upload unless it is an intentional protected backup.
4. Set repository Actions variable `MACOS_SIGNING_IDENTITY` to the certificate's
   40-character SHA-1 fingerprint. `security find-identity -p codesigning`
   lists identities; do not use `-v`, which can omit self-signed identities.
5. Use that fingerprint as `INLINEEXPLAIN_SIGN_IDENTITY` locally too.

The workflow refuses missing credentials, imports the identity into a temporary
Keychain using a pinned action, signs both the app and OCR helper, checks their
certificate fingerprints, and removes the Keychain even after failure. Only
trusted maintainers should be able to change or run the release workflow.

A stable certificate is intended to keep macOS TCC permission identity stable;
it does not establish Gatekeeper trust. Switching from an ad-hoc build may
require another Screen Recording grant. Verify retention by installing build A,
granting permission, then installing a changed build B signed with the same
certificate and checking capture without reauthorization. Signature checks
alone do not establish permission retention. Certificate replacement or loss
can require users to grant permission again. Never publish a replacement
identity as if it were the old one.

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
