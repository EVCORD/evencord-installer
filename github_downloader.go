/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"context"
	"equilotl/branding"
	"equilotl/internal/release"
	"errors"
	"os"
	"path/filepath"
	"time"
)

var GithubError error
var GithubDoneChan chan bool
var InstalledHash = "Unavailable"
var LatestHash = "Unavailable"
var IsDevInstall = false
var approvedRelease release.Verified
var stagedPayload string

func requireLive() error {
	if !branding.Project.Capabilities.Distribution || !branding.Project.Capabilities.Installation {
		return ErrDistributionDisabled
	}
	if len(branding.ReleaseTrust().Keys) == 0 {
		return errors.New("no production release signing keys are configured")
	}
	return nil
}
func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)
	if err := requireLive(); err != nil {
		GithubError = err
		GithubDoneChan <- false
		return
	}
	go func() {
		defer func() { GithubDoneChan <- GithubError == nil }()
		approvedRelease, GithubError = release.Latest(context.Background(), branding.ReleaseTrust(), "core")
		if GithubError == nil {
			LatestHash = approvedRelease.Manifest.Source
		}
	}()
	if current, err := secureEngine().Current(BaseDir); err == nil {
		InstalledHash = current.Manifest.Source
	}
}
func installLatestBuilds() error {
	if err := requireLive(); err != nil {
		return err
	}
	if len(approvedRelease.Envelope) == 0 {
		return errors.New("no authenticated release is available")
	}
	verified, err := release.Verify(approvedRelease.Envelope, branding.ReleaseTrust(), "core", time.Now(), true)
	if err != nil {
		return err
	}
	// Each attempt has separate staging so a damaged cache cannot prevent repair.
	dir, err := os.MkdirTemp(BaseDir, ".evencord-staging-")
	if os.IsNotExist(err) {
		if err = os.MkdirAll(BaseDir, 0700); err != nil {
			return err
		}
		dir, err = os.MkdirTemp(BaseDir, ".evencord-staging-")
	}
	if err != nil {
		return err
	}
	stagedPayload, err = release.Download(context.Background(), filepath.Clean(dir), verified)
	if err != nil {
		os.RemoveAll(dir)
		return err
	}
	approvedRelease = verified
	return nil
}
