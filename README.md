# EvenCord Installer

Phase 3 signed-release and recovery infrastructure is implemented. Live distribution remains disabled. See [PHASE3.md](PHASE3.md) for validated behavior, protocol details and remaining activation requirements.

![EvenCord](branding/assets/banner.svg)

The development foundation for an independent Windows x64 installer for [EvenCord](https://github.com/alrimii/evencord), based on [Equilotl](https://github.com/Equicord/Equilotl) and [Vencord Installer](https://github.com/Vencord/Installer).

**No public installer downloads are available.** The GUI is a non-operational development preview. CLI metadata is available, but installation, removal, repair, OpenAsar changes, downloads and self-update are disabled in the backend. Do not enable these capabilities before Phase 3.

## Development

The current source requires Go 1.26.2. The GUI additionally requires the inherited Windows CGO toolchain. Windows resources use the existing go-winres tool. Keep dependency/tool caches in this checkout when working alongside other development environments.

    go build -tags cli -o build/EvenCordInstallerCli-x64.exe .
    go run ./cmd/brand-resources
    go-winres make --in build/winres.json --arch amd64
    go build -tags "static gui" -ldflags "-H=windowsgui" -o build/EvenCordInstaller-x64.exe .

For CLI resource metadata, pass -cli to the resource generator before embedding resources and building the CLI. These commands compile source; they do not install the mod. Do not execute installation or repair flows during Phase 2.

The Go module remains equilotl for upstream compatibility. Component versioning is independent of the shared branding contract. GitHub Actions and inherited publishing jobs remain disabled.

## Identity and safety

See [branding/README.md](branding/README.md), [FOUNDATION.md](FOUNDATION.md), and [PHASE2.md](PHASE2.md). The intended payload is evencord.asar in the EvenCord data directory; it is not downloaded or written by this phase.

Existing _app.asar files indicate a modified Discord installation, not EvenCord ownership. Loader recognition, safe switching, recovery, verified downloads and signing require Phase 3 work.

## Support and attribution

Support information: https://github.com/alrimii/evencord#support. No separate support server, domain or cloud service is operated by this project. Do not seek support for this fork from upstream maintainers.

All upstream history, source headers and [GPL license](LICENSE) are retained. Copyright remains with Vendicated, Vencord contributors, Equicord contributors and the respective original authors; new branding contributions are credited to EvenCord contributors. The original artwork is GPL-3.0-or-later.

EvenCord is independent and not affiliated with Discord Inc., Equicord or Vencord. The name remains subject to trademark clearance before public distribution.
