/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"equilotl/branding"
	"equilotl/internal/install"
	"equilotl/internal/release"
	"errors"
	"github.com/ProtonMail/go-appdir"
	"os"
	"path/filepath"
	"runtime"
)

var BaseDir string
var BaseDirErr error
var EquicordDirectory string
var ErrAlreadyReported = errors.New("already reported")

func init() {
	prefix := branding.Project.Storage.EnvironmentPrefix
	BaseDir = os.Getenv(prefix + "_USER_DATA_DIR")
	if BaseDir == "" {
		if portable := os.Getenv("DISCORD_USER_DATA_DIR"); portable != "" {
			BaseDir = filepath.Join(portable, "..", branding.Project.Storage.PortableDirectory)
		} else {
			BaseDir = appdir.New(branding.Project.Storage.Directory).UserConfig()
		}
	}
	BaseDir, BaseDirErr = filepath.Abs(BaseDir)
	EquicordDirectory = filepath.Join(BaseDir, branding.Project.Storage.Payload)
}

type DiscordInstall struct {
	path             string // the base path
	branch           string // canary / stable / ...
	appPath          string // List of app folder to patch
	isPatched        bool
	isFlatpak        bool
	isSystemElectron bool // Needs special care https://aur.archlinux.org/packages/discord_arch_electron
	isOpenAsar       *bool
}

func secureEngine() install.Engine {
	return install.Engine{Trust: branding.ReleaseTrust(), TrustData: branding.TrustData, Runtime: branding.Runtime, Loader: branding.Loader}
}
func (di *DiscordInstall) target() (install.Target, error) {
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" || di.branch != "stable" || di.isSystemElectron || di.isFlatpak {
		return install.Target{}, errors.New("only Windows x64 Discord Stable is supported")
	}
	t := install.Target{Resources: filepath.Dir(di.appPath), Channel: di.branch}
	return t, install.Validate(t)
}
func (di *DiscordInstall) patch() error {
	if err := requireLive(); err != nil {
		return err
	}
	target, err := di.target()
	if err != nil {
		return err
	}
	if stagedPayload == "" || release.CheckFile(stagedPayload, approvedRelease.Manifest.Artifact) != nil {
		if err = InstallLatestBuilds(); err != nil {
			return err
		}
	}
	if err = secureEngine().Install(target, BaseDir, approvedRelease, stagedPayload); err != nil {
		return err
	}
	clearStagedPayload()
	di.isPatched = true
	InstalledHash = approvedRelease.Manifest.Source
	return nil
}
func (di *DiscordInstall) unpatch() error {
	if err := requireLive(); err != nil {
		return err
	}
	target, err := di.target()
	if err != nil {
		return err
	}
	if err = secureEngine().Uninstall(target); err != nil {
		return err
	}
	di.isPatched = false
	return nil
}
func (di *DiscordInstall) recover() error {
	if err := requireLive(); err != nil {
		return err
	}
	target, err := di.target()
	if err != nil {
		return err
	}
	return secureEngine().Recover(target)
}
func (di *DiscordInstall) ownership() string {
	target, err := di.target()
	if err != nil {
		return "Unsupported installation"
	}
	state, _ := secureEngine().Inspect(target)
	return state
}
