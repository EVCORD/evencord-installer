// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"equilotl/branding"
	"equilotl/internal/install"
	"equilotl/internal/release"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var trustSHA256 string

func run() error {
	bundle := flag.String("bundle", "", "Local signed preview directory")
	resources := flag.String("resources", "", "Explicit Discord Stable resources directory")
	action := flag.String("action", "inspect", "inspect, install, uninstall or recover")
	flag.Parse()
	if runtime.GOOS != "windows" || runtime.GOARCH != "amd64" || *bundle == "" || *resources == "" {
		return errors.New("Windows x64 and explicit bundle/resources paths are required")
	}
	if branding.Project.Capabilities.Distribution || branding.Project.Capabilities.Installation || branding.Project.Capabilities.Updates || branding.Project.Capabilities.SelfUpdate || branding.Project.Capabilities.HostRepatch {
		return errors.New("local preview requires production distribution to remain disabled")
	}
	data, err := os.ReadFile(filepath.Join(*bundle, "trust.json"))
	if err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	if len(trustSHA256) != 64 || hex.EncodeToString(digest[:]) != trustSHA256 {
		return errors.New("local trust does not match the fingerprint compiled into this installer")
	}
	var trust release.Trust
	if err = json.Unmarshal(data, &trust); err != nil {
		return err
	}
	if len(trust.Keys) != 1 {
		return errors.New("one local preview key is required")
	}
	for id := range trust.Keys {
		if !strings.HasPrefix(id, "local-preview-") {
			return errors.New("not a local preview trust identity")
		}
	}
	root := filepath.Join(os.Getenv("APPDATA"), "EvenCord")
	if !filepath.IsAbs(root) {
		return errors.New("APPDATA must resolve to an absolute path")
	}
	target := install.Target{Resources: filepath.Clean(*resources), Channel: "stable"}
	if !filepath.IsAbs(target.Resources) {
		return errors.New("resources must be an absolute path")
	}
	engine := install.Engine{Trust: trust, TrustData: data, Runtime: branding.Runtime, Loader: branding.Loader}
	switch *action {
	case "inspect":
		status, err := engine.Inspect(target)
		fmt.Println(status)
		return err
	case "uninstall":
		return engine.Uninstall(target)
	case "recover":
		return engine.Recover(target)
	case "install":
		envelope, err := os.ReadFile(filepath.Join(*bundle, "core.release.json"))
		if err != nil {
			return err
		}
		verified, err := release.Verify(envelope, trust, "core", time.Now(), true)
		if err != nil {
			return err
		}
		payload := filepath.Join(*bundle, "desktop.asar")
		if err = release.CheckFile(payload, verified.Manifest.Artifact); err != nil {
			return err
		}
		if err = engine.Install(target, root, verified, payload); err != nil {
			return fmt.Errorf("installation failed: %w; preserve files and use the local recover action", err)
		}
		status, err := engine.Inspect(target)
		if err != nil {
			return err
		}
		if status != "Installed" {
			return fmt.Errorf("installation verification returned %s", status)
		}
		fmt.Println("Verified local EvenCord installation. Public updates remain disabled.")
		return nil
	default:
		return errors.New("unsupported local operation")
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
