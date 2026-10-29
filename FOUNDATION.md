# EvenCord foundation status

Phase 2 branding and identity are implemented. See [PHASE2.md](PHASE2.md). The following Phase 1 record is historical; its references to unchanged runtime branding describe the initial foundation, not the current source.

# Phase 1 installer foundation

This is [alrimii/evencord-installer](https://github.com/alrimii/evencord-installer), a history-preserving fork of [Equicord/Equilotl](https://github.com/Equicord/Equilotl), itself derived from [Vencord/Installer](https://github.com/Vencord/Installer). It is not an official upstream release. The repository name is temporary, and existing licenses, copyright notices, and contributor history are retained.

## Repository relationships

- Core fork: [alrimii/evencord](https://github.com/alrimii/evencord).
- Installer checkout: `C:\Users\MainRec\Desktop\DiscordModFoundation\installer`.
- Core checkout: `C:\Users\MainRec\Desktop\DiscordModFoundation\core`.
- Installer `origin`: `https://github.com/alrimii/evencord-installer.git`.
- Installer `upstream`: `https://github.com/Equicord/Equilotl.git`.
- Installer starting commit: `4f204a1b8e3e6e1c79f5f4a09591465fc4a97223`.

The checkout has its own `.git` directory. `main` tracks `origin/main`; local pulls are fast-forward only and pushes default to `origin`. The existing Equicord development repository and installed Discord client are outside this project and must remain untouched.

## Current state

GitHub Actions is disabled at repository level. Every job in `.github/workflows/release.yml` and the job in `.github/workflows/winget-submission.yml` also has an unconditional false guard on `main`. The build and signing infrastructure remains present for later review, but no secrets, signing credentials, release binaries, or package submissions were added.

Inherited branches and tags retain upstream history. Before any CI/CD activation, review all branches and tags that could trigger workflows. Do not remove guards or enable Actions without explicit owner approval.

Installer code is unchanged. Its download and self-update destinations still refer to upstream Equicord and Equilotl. The existing install scripts and README download links are upstream instructions, not installers for this project. Do not build, run, or distribute this installer as a branded release during Phase 1.

The planned initial target is existing Discord Stable on Windows x64. No standalone application, Reclip integration, or customized Quoter integration is included.

## Later synchronization and release policy

Equilotl is the direct upstream for the installer; Equicord is the sole direct upstream for the companion core. Future updates require review branches, explicit upstream SHAs, conflict reports, preserved merge ancestry, and human approval. No automatic synchronization is configured. Do not squash upstream synchronization merges or force-push public main.

Before distribution, separately review loader ownership and directory/archive compatibility, repair error propagation, verified downloads, self-update behavior, data isolation, and release destinations. Never reuse upstream release binaries as this project's own builds. Installation/update endpoints and repository identities must agree with the companion core before any public release. See the [core foundation plan](https://github.com/alrimii/evencord/blob/main/FOUNDATION.md).

The next step is approval of the final project name and branding configuration. Installer redesign, updater changes, signing, release publishing, and enabling CI/CD remain outside this phase.

Phase 1 was prepared with OpenAI Codex assistance. Verification is limited to source/configuration inspection and GitHub/Git metadata; no build, test, installation, or runtime compatibility validation was performed.
