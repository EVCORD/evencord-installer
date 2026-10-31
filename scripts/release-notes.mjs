/*
 * Vencord, a Discord client mod
 * Copyright (c) 2026 Vendicated and contributors
 * SPDX-License-Identifier: GPL-3.0-or-later
 */
const version = process.argv[2];
const base = `https://github.com/alrimii/evencord-installer/releases/download/installer-v${version}/`;
console.log(`# EvenCord ${version}

## Download

| Architecture | Windows | macOS | Linux |
| --- | --- | --- | --- |
| Intel / AMD (x64) | [EXE](${base}EvenCord.exe) | [DMG](${base}EvenCord.dmg) | [Download](${base}EvenCord) |
| ARM64 | [EXE](${base}EvenCord-arm64.exe) | [DMG](${base}EvenCord-arm64.dmg) | [Download](${base}EvenCord-arm64) |
| Universal | | [DMG](${base}EvenCord.dmg) | |

<details>
<summary>Command line and Linux variants</summary>

| Architecture | Windows CLI | macOS CLI | Linux CLI |
| --- | --- | --- | --- |
| x64 | [Download](${base}EvenCord-cli.exe) | [Download](${base}EvenCord-cli-x64) | [Download](${base}EvenCord-cli-linux) |
| ARM64 | [Download](${base}EvenCord-cli-arm64.exe) | [Download](${base}EvenCord-cli-arm64) | [Download](${base}EvenCord-cli-linux-arm64) |

Linux: [X11](${base}EvenCord-x11) ? [Wayland](${base}EvenCord-wayland) ? [X11 ARM64](${base}EvenCord-x11-arm64) ? [Wayland ARM64](${base}EvenCord-wayland-arm64).

</details>

Close Discord, open EvenCord and choose Install or Repair. Your settings are kept.
`);
