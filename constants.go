/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"equilotl/branding"
	"equilotl/buildinfo"
	"errors"
	"image/color"
)

var ReleaseUrl = "https://api.github.com/repos/" + branding.Project.Repositories.Core.Slug + "/releases/latest"

const ReleaseUrlFallback = ""

var InstallerReleaseUrl = "https://api.github.com/repos/" + branding.Project.Repositories.Installer.Slug + "/releases/latest"

const InstallerReleaseUrlFallback = ""

var UserAgent = branding.Project.InstallerName + "/" + buildinfo.InstallerGitHash + " (https://github.com/" + branding.Project.Repositories.Installer.Slug + ")"

var SupportUrl = branding.Project.Links.Support
var ErrDistributionDisabled = errors.New("EvenCord distribution is disabled in this build.")

var (
	DiscordGreen        = color.RGBA{128, 82, 99, 0xff}
	DiscordGreenHovered = color.RGBA{157, 105, 125, 0xff}
	DiscordRed          = color.RGBA{210, 45, 57, 0xff}
	DiscordRedHovered   = color.RGBA{169, 35, 46, 0xff}
	DiscordBlue         = color.RGBA{52, 53, 59, 0xff}
	DiscordBlueHovered  = color.RGBA{75, 65, 73, 0xff}
	DiscordYellow       = color.RGBA{0xfe, 0xe7, 0x5c, 0xff}
)

var LinuxDiscordNames = []string{
	"Discord",
	"DiscordPTB",
	"DiscordCanary",
	"DiscordDevelopment",
	"discord",
	"discordptb",
	"discordcanary",
	"discorddevelopment",
	"discord-ptb",
	"discord-canary",
	"discord-development",
	// Flatpak
	"com.discordapp.Discord",
	"com.discordapp.DiscordPTB",
	"com.discordapp.DiscordCanary",
	"com.discordapp.DiscordDevelopment",
}
