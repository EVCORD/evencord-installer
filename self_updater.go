/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"context"
	"equilotl/branding"
	"equilotl/buildinfo"
	"equilotl/internal/release"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var selfMu sync.Mutex
var IsSelfOutdated = false
var SelfUpdateCheckDoneChan = make(chan bool, 1)
var selfRelease release.Verified
var pendingInstaller string

func installerComponent() string {
	if buildinfo.UiType == buildinfo.UiTypeCli {
		return "installer-cli"
	}
	return "installer-gui"
}
func selfEnabled() bool {
	return branding.Project.Capabilities.Distribution && branding.Project.Capabilities.SelfUpdate && len(branding.ReleaseTrust().Keys) > 0 && runtime.GOOS == "windows" && runtime.GOARCH == "amd64"
}
func init() {
	if !selfEnabled() {
		SelfUpdateCheckDoneChan <- false
		return
	}
	go func() {
		next, err := release.Latest(context.Background(), branding.ReleaseTrust(), installerComponent())
		if err == nil {
			var order int
			order, err = release.Compare(next.Manifest.Version, buildinfo.InstallerTag)
			selfMu.Lock()
			selfRelease = next
			IsSelfOutdated = err == nil && order > 0
			selfMu.Unlock()
		}
		SelfUpdateCheckDoneChan <- err == nil
	}()
}
func CanUpdateSelf() bool {
	selfMu.Lock()
	defer selfMu.Unlock()
	return selfEnabled() && IsSelfOutdated
}
func GetInstallerDownloadLink() string {
	if !CanUpdateSelf() {
		return ""
	}
	selfMu.Lock()
	defer selfMu.Unlock()
	return release.ArtifactURL(selfRelease.Manifest)
}
func UpdateSelf() error {
	if !CanUpdateSelf() {
		return errors.New("no authenticated newer installer is available")
	}
	selfMu.Lock()
	envelope := append([]byte(nil), selfRelease.Envelope...)
	selfMu.Unlock()
	v, err := release.Verify(envelope, branding.ReleaseTrust(), installerComponent(), time.Now(), true)
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// Side-by-side handoff keeps the running, locked executable intact.
	directory := filepath.Join(filepath.Dir(exe), ".evencord-installer-updates")
	pendingInstaller, err = release.Download(context.Background(), directory, v)
	if err != nil {
		return err
	}
	file, err := os.Open(pendingInstaller)
	if err != nil {
		return err
	}
	defer file.Close()
	pendingInstaller, err = release.StageSelfUpdate(directory, v.Envelope, branding.ReleaseTrust(), installerComponent(), buildinfo.InstallerTag, file)
	return err
}
func DeleteOldExecutable() {}
func RelaunchSelf() error {
	if !selfEnabled() || pendingInstaller == "" {
		return errors.New("no verified installer handoff is staged")
	}
	proof, err := os.ReadFile(pendingInstaller + ".release.json")
	if err != nil {
		return err
	}
	v, err := release.Verify(proof, branding.ReleaseTrust(), installerComponent(), time.Now(), true)
	if err != nil {
		return err
	}
	if err = release.CheckFile(pendingInstaller, v.Manifest.Artifact); err != nil {
		return err
	}
	proc, err := os.StartProcess(pendingInstaller, []string{pendingInstaller}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err != nil {
		return err
	}
	if err = proc.Release(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
