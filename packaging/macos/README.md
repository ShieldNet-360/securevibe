# SecureVibe — macOS .pkg installer

Build a macOS installer package for the `securevibe` CLI (part of
**SecureVibe**) using `pkgbuild` + `productbuild` from Xcode Command
Line Tools.

## Quick Start

```bash
# Build the binary first (from the repo root):
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o securevibe ./cmd/securevibe

# Build the .pkg:
cd packaging/macos
make BINARY=../../securevibe VERSION=2026.05.12
```

The resulting `.pkg` is at `build/securevibe-2026.05.12.pkg`.

## What the installer does

- Copies `securevibe` to `/usr/local/bin/securevibe`.
- No launch daemons are installed; run `securevibe scheduler install` post-install
  if you want background updates.

## Code-signing (optional)

If you have a Developer ID Installer certificate:

```bash
productsign --sign "Developer ID Installer: Your Name" \
    build/securevibe-2026.05.12.pkg \
    build/securevibe-2026.05.12-signed.pkg
```

## Notarization (optional)

```bash
xcrun notarytool submit build/securevibe-2026.05.12-signed.pkg \
    --apple-id you@example.com \
    --team-id TEAMID \
    --password @keychain:AC_PASSWORD \
    --wait
```
