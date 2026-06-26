# SecureVibe — Windows MSI installer

Build a Windows `.msi` for the `securevibe` CLI (part of **SecureVibe**)
using the [WiX Toolset](https://wixtoolset.org/) v4+.

## Prerequisites

1. Install the WiX Toolset: https://wixtoolset.org/docs/intro/
2. Build the CLI binary (from the repo root on a Windows machine or via
   cross-compilation):

   ```powershell
   $env:GOOS = "windows"
   $env:GOARCH = "amd64"
   go build -trimpath -ldflags "-s -w" -o securevibe.exe ./cmd/securevibe
   ```

## Build the MSI

```powershell
cd packaging\windows
wix build `
    -d BinaryPath=..\..\securevibe.exe `
    -d Version=2026.05.12.0 `
    -o build\securevibe.msi `
    securevibe.wxs
```

The resulting MSI is at `build\securevibe.msi`.

## What the installer does

- Installs `securevibe.exe` to `C:\Program Files\SecureVibe\`.
- Adds the install directory to the system `PATH`.
- Does **not** register a scheduled task; run
  `securevibe scheduler install` post-install for background updates.

## Signing (recommended)

Sign the MSI with `signtool` from the Windows SDK:

```powershell
signtool sign /f cert.pfx /p <password> /tr http://timestamp.digicert.com /td sha256 build\securevibe.msi
```
