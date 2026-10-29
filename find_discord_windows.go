/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"errors"
	"os"
	path "path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var windowsNames = map[string]string{
	"stable": "Discord",
	"ptb":    "DiscordPTB",
	"canary": "DiscordCanary",
	"dev":    "DiscordDevelopment",
}

func ParseDiscordNew(p, branch string, isFlatpak bool) *DiscordInstall {
	return nil
}

func ParseDiscord(p, branch string) *DiscordInstall {
	entries, err := os.ReadDir(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			Log.Warn("Error during readdir "+p+":", err)
		}
		return nil
	}

	isPatched := false
	appPath := ""
	var latestVer []int
	for _, dir := range entries {
		if !dir.IsDir() || !strings.HasPrefix(dir.Name(), "app-") {
			continue
		}
		resources := path.Join(p, dir.Name(), "resources")
		if !ExistsFile(resources) {
			continue
		}
		dirIsPatched := ExistsFile(path.Join(resources, "_app.asar"))
		if !dirIsPatched && !ExistsFile(path.Join(resources, "app.asar")) {
			continue
		}
		ver := ParseAppVersion(dir.Name())
		if ver == nil {
			continue
		}
		if appPath == "" || CompareAppVersion(ver, latestVer) > 0 {
			appPath = path.Join(resources, "app")
			isPatched = dirIsPatched
			latestVer = ver
		}
	}

	if appPath == "" {
		return nil
	}

	if branch == "" {
		branch = GetBranch(p)
	}

	result := &DiscordInstall{
		path:             p,
		branch:           branch,
		appPath:          appPath,
		isPatched:        isPatched,
		isFlatpak:        false,
		isSystemElectron: false,
	}
	state := result.ownership()
	result.isPatched = state == "Installed" || state == "Repair available"
	return result
}

func FindDiscords() []any {
	var discords []any

	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		Log.Error("%LOCALAPPDATA% is empty???????")
		return discords
	}

	for branch, dirname := range map[string]string{"stable": "Discord", "ptb": "DiscordPTB", "canary": "DiscordCanary"} {
		p := path.Join(appData, dirname)
		if discord := ParseDiscord(p, branch); discord != nil {
			Log.Debug("Found Discord install at ", p)
			discords = append(discords, discord)
		}
	}
	return discords
}

func PreparePatch(di *DiscordInstall) {
	// EvenCord never terminates Discord. Locked files produce a recoverable operation error.
}

func FixOwnership(_ string) error {
	return nil
}

// https://github.com/Vencord/Installer/issues/9

func CheckScuffedInstall() bool {
	username := os.Getenv("USERNAME")
	programData := os.Getenv("PROGRAMDATA")
	for _, discordName := range windowsNames {
		if ExistsFile(path.Join(programData, username, discordName)) {
			HandleScuffedInstall()
			return true
		}
	}
	return false
}

func findProcessIdByName(name string) uint32 {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}

	defer windows.CloseHandle(snapshot)
	procEntry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for {
		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			return 0
		}
		if windows.UTF16ToString(procEntry.ExeFile[:]) == name {
			return procEntry.ProcessID
		}
	}
}
