# <img src="branding/assets/mark.svg" width="36" align="left" alt=""> EvenCord Installer

Install, update, repair or uninstall [EvenCord](https://github.com/alrimii/evencord) on Discord Stable for Windows.

[Website](https://alrimii.github.io/evencord/) · [Support](https://github.com/alrimii/evencord-installer/issues) · [Releases](https://github.com/alrimii/evencord-installer/releases)

## Download

**Windows x64 · Public beta**

| Version | Download |
| --- | --- |
| GUI | [Download EvenCord Installer](https://github.com/alrimii/evencord-installer/releases/download/installer-v0.1.0-beta.2/EvenCordInstaller-x64.exe) |
| CLI | [Download CLI](https://github.com/alrimii/evencord-installer/releases/download/installer-v0.1.0-beta.2/EvenCordInstallerCli-x64.exe) |

## Usage

1. Open the installer. Discord Stable is detected automatically.
2. Fully close Discord, including its tray icon.
3. Select **Install EvenCord**, **Repair** or **Uninstall**.

Uninstall restores Discord's original loader and keeps your settings. If an operation is interrupted, refresh the installation status and use **Recover** when offered. Other client modifications must be removed with their own installer first.

The beta currently supports Windows 10/11 x64 and the reviewed Discord Stable **1.0.9266** archive. Beta.1 and beta.2 require a one-time upgrade with the new installer. From beta.3, EvenCord checks for signed updates automatically; opt into background downloads in EvenCord > Updater. Restart only after a verified update is ready. The executable is not Windows Authenticode signed and may show a SmartScreen warning.

Downloads are verified using pinned Ed25519 signatures, artifact sizes and SHA-256. [Verification and recovery details](docs/RELEASE-SECURITY.md)

## Building from source

Use the Go version in `go.mod`, `go-winres`, and a Windows x64 C/C++ compiler compatible with giu/GLFW for GUI builds.

```powershell
go run ./cmd/brand-resources
go-winres make --in build/winres.json --arch amd64
go build -tags 'static gui' -trimpath -ldflags '-H=windowsgui' -o build/EvenCordInstaller-x64.exe .
```

For the CLI, generate resources with `go run ./cmd/brand-resources -cli`, then build with `go build -tags cli -trimpath -o build/EvenCordInstallerCli-x64.exe .`.

## Credits

Based on [Equilotl](https://github.com/Equicord/Equilotl) and the [Vencord Installer](https://github.com/Vencord/Installer). Original authors, copyright notices, Git history and [GPL license](LICENSE) are preserved.

This is the EvenCord installer, not an official Discord, Equicord or Vencord product.
