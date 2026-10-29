# Private local installer

`cmd/local-preview` is a separate offline CLI using the existing signed release verifier, loader ownership checks, install transactions and recovery engine. Public CLI/GUI distribution gates are unchanged. It has no network download path and refuses to operate without a compiled public-trust fingerprint.

Build on Windows x64:

```powershell
$fingerprint = (Get-FileHash <bundle>/trust.json -Algorithm SHA256).Hash.ToLower()
go build -ldflags "-X main.trustSHA256=$fingerprint" -o build/EvenCordLocalInstaller-x64.exe ./cmd/local-preview
```

Run with explicit paths and `--action inspect|install|uninstall|recover --bundle <bundle> --resources <Discord-Stable-resources>`.

The local key must have a `local-preview-` identifier, the manifest must pass the standard Ed25519 and repository identity checks, the payload must match its signed size/digest, and the original Discord archive must match the signed allowlist. Data is stored only in `%APPDATA%/EvenCord`. Preserve the local bundle and installer for repair/uninstallation; neither contains the private signing key.

This command never removes a foreign modification. Migration from Equicord requires separate, explicit review of the existing loader and original archive, followed by restoring that original while Discord is stopped. Keep the previous shim outside Discord's active resources for recovery. If installation fails, use recover and then uninstall any verified EvenCord loader to restore original Discord. Never guess ownership or delete an unrecognized archive.

This is not a public distribution channel, automatic updater, or proof of runtime compatibility. Production keys remain empty and all public capability flags remain false.
