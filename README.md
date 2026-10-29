# <img src="branding/assets/mark.svg" width="36" align="left" alt=""> EvenCord Installer

Install, update, repair or uninstall [EvenCord](https://github.com/alrimii/evencord) on Discord Stable for Windows.

[Website](https://alrimii.github.io/evencord/) · [Support](https://github.com/alrimii/evencord-installer/issues) · [Releases](https://github.com/alrimii/evencord-installer/releases)

## Download

**Windows x64 · Public beta**

| Version | Download |
| --- | --- |
| GUI | [Download EvenCord Installer](https://github.com/alrimii/evencord-installer/releases/download/installer-v0.1.0-beta.6/EvenCordInstaller-x64.exe) |
| CLI | [Download CLI](https://github.com/alrimii/evencord-installer/releases/download/installer-v0.1.0-beta.6/EvenCordInstallerCli-x64.exe) |

## Usage

1. Open the installer. Discord Stable is detected automatically.
2. Save unsaved clips and fully close Discord, including its tray icon.
3. Select **Install EvenCord**, **Repair** or **Uninstall**.

Uninstall restores Discord's original loader and keeps your settings. If an operation is interrupted, refresh the installation status and use **Recover** when offered. Other client modifications must be removed with their own installer first.

The beta supports Windows 10/11 x64 with reviewed Stable **1.0.9266**, PTB **1.0.1229** and Canary **1.0.1285** archives. Each channel has a separate EvenCord profile. Beta.1 and beta.2 require a one-time upgrade with the new installer. From beta.3, EvenCord checks for signed updates automatically; control background downloads in EvenCord > Updater. Restart only after a verified update is ready. The executable is not Windows Authenticode signed and may show a SmartScreen warning.

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
