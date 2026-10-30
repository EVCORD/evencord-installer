//go:build !cli

// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"equilotl/branding"
	"equilotl/buildinfo"
	"equilotl/internal/install"
	"equilotl/internal/release"
	"errors"
	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func secureGUI() {
	g.SetDefaultFontSize(16)
	win = g.NewMasterWindow(branding.Project.InstallerName, 900, 780, 0)
	win.SetSizeLimits(780, 740, g.DontCare, g.DontCare)
	applyBrandTheme()
	icon, _, iconErr := image.Decode(bytes.NewReader(iconBytes))
	if iconErr == nil {
		win.SetIcon([]image.Image{icon})
	}
	var mu sync.Mutex
	state, detail, ownership, busy := "Looking for Discord", "", "", true
	channels := []struct{ name, branch, directory string }{
		{"Discord Stable", "stable", "Discord"},
		{"Discord PTB", "ptb", "DiscordPTB"},
		{"Discord Canary", "canary", "DiscordCanary"},
	}
	selectedChannel := 0
	detected := [3]bool{}
	var chosen *DiscordInstall
	set := func(s, d string) { mu.Lock(); state, detail = s, d; mu.Unlock(); g.Update() }
	discover := func() {
		mu.Lock()
		index := selectedChannel
		mu.Unlock()
		var installs [3]*DiscordInstall
		var available [3]bool
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			for i, channel := range channels {
				installs[i] = ParseDiscord(filepath.Join(local, channel.directory), channel.branch)
				available[i] = installs[i] != nil
			}
		}
		target := installs[index]
		status, message := "Discord not found", "Install the selected Discord channel, then select Refresh."
		if target != nil {
			status = target.ownership()
			switch status {
			case "Not installed":
				message = "Ready to install EvenCord. Close Discord before continuing."
			case "Installed":
				message = "EvenCord is installed. You can update, repair or uninstall it."
			case "Repair available":
				message = "The EvenCord payload needs repair. Close Discord and select Repair."
			case "Recovery available":
				message = "An earlier operation was interrupted. Select Recover before continuing."
			default:
				message = "This installation cannot be changed safely. Remove other modifications with their own installer, or contact support."
			}
		}
		mu.Lock()
		chosen, ownership, state, detail, busy = target, status, status, message, false
		detected = available
		mu.Unlock()
		g.Update()
	}
	chooseChannel := func(index int) {
		mu.Lock()
		if busy {
			mu.Unlock()
			return
		}
		selectedChannel, chosen, busy = index, nil, true
		state, detail, ownership = "Looking for Discord", "", ""
		mu.Unlock()
		go discover()
	}
	launch := func(operation string) {
		mu.Lock()
		if busy {
			mu.Unlock()
			return
		}
		target := chosen
		busy = true
		mu.Unlock()
		go func() {
			defer func() { mu.Lock(); busy = false; mu.Unlock(); g.Update() }()
			var err error
			if target == nil {
				err = errors.New("select a supported Discord installation first")
			} else if err = target.closeForOperation(); err != nil {
				// The selected channel could not be closed.
			} else if err = requireLive(); err == nil {
				switch operation {
				case "Uninstall":
					set("Uninstalling EvenCord", "Restoring the original Discord loader. Your settings will be kept.")
					err = target.unpatch()
				case "Recover":
					set("Recovering installation", "Restoring a consistent state from the recovery records.")
					err = target.recover()
				default:
					set("Checking for updates", "Finding the latest signed EvenCord release.")
					approvedRelease, err = release.Latest(context.Background(), branding.ReleaseTrust(), "core")
					if err == nil {
						set("Downloading EvenCord", "The download is verified before any installation files are changed.")
						err = installLatestBuilds()
					}
					if err == nil {
						set("Installing EvenCord", "Applying the verified build and preserving recovery files.")
						err = target.patch()
					}
				}
			}
			if err != nil {
				if target != nil {
					mu.Lock()
					ownership = target.ownership()
					mu.Unlock()
				}
				set("Could not complete "+strings.ToLower(operation), err.Error()+" Close Discord and retry if files are in use. Recovery data has been kept.")
				return
			}
			mu.Lock()
			ownership = target.ownership()
			mu.Unlock()
			if operation == "Uninstall" {
				set("EvenCord uninstalled", "Discord's original loader is restored. Your EvenCord settings are still saved.")
			} else {
				set(operation+" complete", "You can now open Discord. Your settings have been preserved.")
			}
		}()
	}
	go discover()
	win.Run(func() {
		mu.Lock()
		s, d, owned, working, target := state, detail, ownership, busy, chosen
		channelIndex, available := selectedChannel, detected
		mu.Unlock()
		eligible := target != nil && requireLive() == nil
		canInstall := eligible && (owned == "Not installed" || owned == "Installed")
		canRepair := eligible && (owned == "Installed" || owned == "Repair available")
		canRemove := eligible && (owned == "Installed" || owned == "Repair available")
		installLabel := "Install EvenCord"
		if owned == "Installed" {
			installLabel = "Update EvenCord"
		}
		choices := make([]g.Widget, len(channels))
		for i, channel := range channels {
			index := i
			label := channel.name
			if !available[i] {
				label += " (not detected)"
			}
			choices[i] = g.RadioButton(label, channelIndex == i).OnChange(func() { chooseChannel(index) })
		}
		location := "Selected channel is not installed"
		if target != nil {
			location = target.path
		}
		g.PushWindowPadding(32, 24)
		g.SingleWindow().Layout(
			g.Child().Size(g.Auto, g.Auto).Layout(
				g.Align(g.AlignCenter).To(g.Row(
					g.ImageWithRgba(icon).Size(64, 64),
					g.Column(g.Style().SetFontSize(30).To(g.Label("EvenCord Installer")), g.Label("Make Discord your own.")),
				)),
				g.Dummy(0, 24),
				g.Style().SetFontSize(20).To(g.Label("Choose your Discord installation")),
				g.Row(choices...),
				g.Label(location).Wrapped(true),
				g.Label("EvenCord settings: "+install.DataRoot(BaseDir, channels[channelIndex].branch)).Wrapped(true),
				g.Dummy(0, 12),
				g.Child().Size(g.Auto, 138).Border(true).Layout(
					g.Style().SetFontSize(20).To(g.Label(s)),
					g.Dummy(0, 8), g.Label(d).Wrapped(true),
				),
				g.Dummy(0, 16),
				g.Row(
					g.Button(installLabel).Size(244, 46).Disabled(working || !canInstall).OnClick(func() { launch("Install") }),
					g.Button("Repair").Size(166, 46).Disabled(working || !canRepair).OnClick(func() { launch("Repair") }),
					g.Button("Uninstall").Size(166, 46).Disabled(working || !canRemove).OnClick(func() { launch("Uninstall") }),
				),
				g.Dummy(0, 12),
				g.Row(
					g.Button("Refresh").Size(128, 34).Disabled(working).OnClick(func() { mu.Lock(); busy = true; mu.Unlock(); go discover() }),
					g.Button("Help").Size(128, 34).OnClick(func() { g.OpenURL(branding.Project.Links.Support) }),
					&CondWidget{owned == "Recovery available", func() g.Widget {
						return g.Button("Recover").Size(128, 34).Disabled(working).OnClick(func() { launch("Recover") })
					}, nil},
				),
				g.Dummy(0, 22), g.Separator(), g.Dummy(0, 10),
				g.Label("Windows x64 / Stable, PTB and Canary / "+buildinfo.InstallerTag),
				g.Label("Operations close the selected Discord channel. Save unsaved clips first. Other channels stay open.").Wrapped(true),
				g.Label("Based on Equilotl, Equicord and Vencord.").Wrapped(true),
			),
		)
		g.PopStyle()
	})
}

func applyBrandTheme() {
	c := branding.Project.Colors
	rgba := func(value string) imgui.Vec4 {
		bytes, err := hex.DecodeString(value[1:])
		if err != nil || len(bytes) != 3 {
			panic("invalid branding palette")
		}
		return imgui.Vec4{X: float32(bytes[0]) / 255, Y: float32(bytes[1]) / 255, Z: float32(bytes[2]) / 255, W: 1}
	}
	style := imgui.CurrentStyle()
	for id, value := range map[imgui.StyleColorID]string{
		imgui.StyleColorWindowBg: c.Dark, imgui.StyleColorChildBg: c.Dark,
		imgui.StyleColorPopupBg: c.Surface, imgui.StyleColorFrameBg: c.Surface,
		imgui.StyleColorText: c.Text, imgui.StyleColorButton: c.Primary,
		imgui.StyleColorButtonHovered: c.Surface, imgui.StyleColorButtonActive: c.Primary,
		imgui.StyleColorHeader: c.Primary, imgui.StyleColorHeaderHovered: c.Surface,
		imgui.StyleColorCheckMark: c.Text, imgui.StyleColorBorder: c.Surface,
		imgui.StyleColorSeparator: c.Surface, imgui.StyleColorTitleBgActive: c.Surface,
	} {
		style.SetColor(id, rgba(value))
	}
}
