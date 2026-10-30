# <img src="branding/assets/mark.svg" width="36" align="left" alt=""> EvenCord Installer

Install, update, repair or uninstall [EvenCord](https://github.com/alrimii/evencord) on Discord Stable, PTB or Canary for Windows.

[Website](https://alrimii.github.io/evencord/) · [Support](https://github.com/alrimii/evencord-installer/issues) · [Releases](https://github.com/alrimii/evencord-installer/releases)

## Download

**Windows x64 · Stable release**

| Version | Download |
| --- | --- |
| GUI | [Download EvenCord Installer](https://github.com/alrimii/evencord-installer/releases/download/installer-v1.0.0/EvenCordInstaller-x64.exe) |
| CLI | [Download CLI](https://github.com/alrimii/evencord-installer/releases/download/installer-v1.0.0/EvenCordInstallerCli-x64.exe) |

## Usage

1. Open the installer. Select Discord Stable, PTB or Canary. Installed channels are detected automatically.
2. Save unsaved clips and fully close Discord, including its tray icon.
3. Select **Install EvenCord**, **Repair** or **Uninstall**.

Uninstall restores Discord's original loader and keeps your settings. If an operation is interrupted, refresh the installation status and use **Recover** when offered. Other client modifications must be removed with their own installer first.

Windows 10/11 x64 supports reviewed Stable **1.0.9266**, PTB **1.0.1229** and Canary **1.0.1285/1.0.1286** archives. Each channel has a separate EvenCord profile. The installer retrieves signed EvenCord stable releases, preserving settings and recovery. Existing update-enabled beta clients can migrate through the signed transition update. Very old update-disabled builds require this installer. Windows executables are not Authenticode signed and may show a SmartScreen warning.

Downloads are verified using pinned Ed25519 signatures, artifact sizes and SHA-256. [Verification and recovery details](docs/RELEASE-SECURITY.md)

The installer lets you select Stable, PTB or Canary and shows detection status. Installation, repair and uninstall are available for reviewed archives in all three channels, with independent plugin settings and recovery.

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

The installer follows upstream close-before-patching behavior and automatically closes the selected Discord installation when an operation starts. Other channels stay open. If an older failed attempt shows Recovery available, choose Recover before Install.
