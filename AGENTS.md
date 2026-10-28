# InlineExplain

## Product taste

Help readers understand an unfamiliar English word in its visible sentence,
then return to reading. Keep help brief, close to the word and easy to dismiss.
Preserve the source application's focus, full-screen Space and playback intent.
Keep provider setup and other occasional controls outside the reading card.

## Boundaries

- This repository owns the macOS app. Peekturn is a separate Android product;
  shared prompt provenance does not make its workspace part of this project.
- Keep screen capture and OCR local and on demand. Do not save or log screenshots,
  reading text, model answers or credentials.
- Store API keys only in macOS Keychain. Preserve provider-origin isolation;
  never guess or reuse ambiguous legacy credentials.
- Keep public defaults independent of company services and accounts.
- Keep task reports and one-off validation artifacts out of the repository.
  Retain useful regression tests beside the code or in `scripts/`.
- Commit, push, tag and publish only on explicit user request.

## Read when relevant

- [Design](DESIGN.md): reading-card and Settings presentation.
- [Architecture](docs/architecture.md): capture, context, provider and native
  integration boundaries; verification limits.
- [Third-party notices](NOTICE): reused code and prompt provenance.
- [Development](docs/development.md): build, automated checks and popup preview.
- [Releasing](docs/releasing.md): canonical version, package output and manual
  GitHub release workflow.
- [`scripts/test.sh`](scripts/test.sh): current automated checks.
