//go:build !cli

// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package main

import (
	"bytes"
	"context"
	"equilotl/branding"
	"equilotl/internal/release"
	"errors"
	g "github.com/AllenDang/giu"
	"image"
	"sync"
)

func secureGUI() {
	win = g.NewMasterWindow(branding.Project.InstallerName, 940, 570, 0)
	if icon, _, err := image.Decode(bytes.NewReader(iconBytes)); err == nil {
		win.SetIcon([]image.Image{icon})
	}
	var mu sync.Mutex
	state, detail, busy := "Not checked", "Detect Discord Stable to inspect installation ownership.", false
	var chosen *DiscordInstall
	set := func(s, d string) { mu.Lock(); state, detail = s, d; mu.Unlock(); g.Update() }
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
				err = errors.New("detect a supported Discord installation first")
			} else if err = requireLive(); err == nil {
				switch operation {
				case "Uninstall":
					set("Uninstalling", "Restoring the verified original Discord archive.")
					err = target.unpatch()
				case "Recover":
					set("Recovering", "Checking the transaction journal and recovery archive.")
					err = target.recover()
				default:
					set("Downloading", "Retrieving signed EvenCord release metadata.")
					approvedRelease, err = release.Latest(context.Background(), branding.ReleaseTrust(), "core")
					if err == nil {
						set("Verifying", "Downloading and verifying artifact size, signature and SHA-256.")
						err = installLatestBuilds()
					}
					if err == nil {
						set("Installing", "Preserving recovery data and activating the verified build.")
						err = target.patch()
					}
				}
			}
			if err != nil {
				set("Failed", err.Error()+" Recovery data is preserved. Close Discord yourself before retrying locked-file failures; use Recover for an unfinished transaction.")
			} else {
				set("Completed", operation+" completed successfully. Settings were preserved.")
			}
		}()
	}
	win.Run(func() {
		mu.Lock()
		s, d, working := state, detail, busy
		target := chosen
		mu.Unlock()
		blocked := working || target == nil || requireLive() != nil
		g.SingleWindow().Layout(
			g.Style().SetFontSize(32).To(g.Label(branding.Project.InstallerName)),
			g.Label("Windows x64 ? Discord Stable ? Verified EvenCord releases"),
			g.Dummy(0, 12), g.Label(s), g.Label(d).Wrapped(true),
			g.Dummy(0, 12),
			g.Button("Detect Discord Stable").Disabled(working).OnClick(func() {
				found := FindDiscords()
				if len(found) != 1 {
					set("Unsupported installation", "No single supported Discord Stable installation was found.")
					return
				}
				target := found[0].(*DiscordInstall)
				mu.Lock()
				chosen = target
				mu.Unlock()
				set(target.ownership(), target.path)
			}),
			g.Button("Check approved release").Disabled(working).OnClick(func() {
				if err := requireLive(); err != nil {
					set("Releases unavailable", err.Error())
					return
				}
				mu.Lock()
				busy = true
				mu.Unlock()
				go func() {
					defer func() { mu.Lock(); busy = false; mu.Unlock(); g.Update() }()
					v, err := release.Latest(context.Background(), branding.ReleaseTrust(), "core")
					if err != nil {
						set("Release check failed", err.Error())
						return
					}
					old, err := secureEngine().Current(BaseDir)
					if err != nil {
						set("Release available", v.Manifest.Version)
						return
					}
					order, err := release.Compare(v.Manifest.Version, old.Manifest.Version)
					if err != nil {
						set("Release check failed", err.Error())
					} else if order > 0 {
						set("Update available", v.Manifest.Version)
					} else {
						set("Installed", "No newer approved release is available.")
					}
				}()
			}),
			g.Row(g.Button("Install / Update").Disabled(blocked).OnClick(func() { launch("Install") }), g.Button("Repair").Disabled(blocked).OnClick(func() { launch("Repair") }), g.Button("Uninstall").Disabled(blocked).OnClick(func() { launch("Uninstall") }), g.Button("Recover").Disabled(blocked).OnClick(func() { launch("Recover") })),
			g.Button("Stage installer update").Disabled(working || !CanUpdateSelf()).OnClick(func() {
				mu.Lock()
				busy = true
				mu.Unlock()
				go func() {
					defer func() { mu.Lock(); busy = false; mu.Unlock(); g.Update() }()
					set("Downloading", "Verifying the signed installer update.")
					if err := UpdateSelf(); err != nil {
						set("Failed", err.Error())
					} else {
						set("Installer update staged", "Verified executable: "+pendingInstaller+". Your existing executable remains available; launch the staged copy when ready.")
					}
				}()
			}),
			g.Dummy(0, 18), g.Label("Production signing and release activation are not configured. Live operations remain disabled.").Wrapped(true),
			g.Label("Based on Equilotl, Equicord and Vencord. "+branding.Project.Links.Support).Wrapped(true),
		)
	})
}
