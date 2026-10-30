// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package main

import (
	"context"
	"equilotl/branding"
	"equilotl/internal/install"
	"equilotl/internal/release"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func run() error {
	channel := flag.String("channel", "stable", "Fixture Discord channel")
	bundle := flag.String("bundle", "", "Signed local bundle")
	previousBundle := flag.String("previous-bundle", "", "Earlier signed public bundle to install before upgrading the disposable fixture")
	component := flag.String("component", "core", "Component to verify")
	output := flag.String("output", "", "Staging directory inside this checkout's .cache")
	network := flag.Bool("network", false, "Retrieve the approved public release")
	fixture := flag.String("fixture", "", "Disposable Discord app directory inside this checkout's .cache")
	flag.Parse()
	cwd, e := os.Getwd()
	if e != nil {
		return e
	}
	cache := filepath.Join(cwd, ".cache")
	checked := func(p string) (string, error) {
		a, e := filepath.Abs(p)
		if e != nil {
			return "", e
		}
		rel, e := filepath.Rel(cache, a)
		if e != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return "", fmt.Errorf("path must be inside this checkout's .cache")
		}
		if e = install.SafePath(a); e != nil {
			return "", e
		}
		return a, nil
	}
	if *output == "" {
		return fmt.Errorf("explicit workspace output is required")
	}
	out, e := checked(*output)
	if e != nil {
		return e
	}
	var v release.Verified
	var payload string
	if *network {
		v, e = release.Latest(context.Background(), branding.ReleaseTrust(), *component)
		if e != nil {
			return e
		}
		payload, e = release.Download(context.Background(), out, v)
		if e != nil {
			return e
		}
	} else {
		if *bundle == "" {
			return fmt.Errorf("bundle or network required")
		}
		data, e := os.ReadFile(filepath.Join(*bundle, *component+".release.json"))
		if e != nil {
			return e
		}
		v, e = release.Verify(data, branding.ReleaseTrust(), *component, time.Now(), true)
		if e != nil {
			return e
		}
		payload = filepath.Join(*bundle, v.Manifest.Artifact.Name)
		if e = release.CheckFile(payload, v.Manifest.Artifact); e != nil {
			return e
		}
	}
	fmt.Println("Verified", v.Manifest.Component, v.Manifest.Version, v.Manifest.Source, v.Manifest.Artifact.SHA256)
	if *fixture != "" {
		if *component != "core" {
			return fmt.Errorf("only core supports installation fixture")
		}
		app, e := checked(*fixture)
		if e != nil {
			return e
		}
		if _, e = os.Stat(filepath.Join(app, "EVENCORD_DISPOSABLE_FIXTURE")); e != nil {
			return fmt.Errorf("fixture marker required")
		}
		target := install.Target{Resources: filepath.Join(app, "resources"), Channel: *channel}
		engine := install.Engine{Trust: branding.ReleaseTrust(), TrustData: branding.TrustData, LegacyTrustData: branding.LegacyTrustData, Runtime: branding.Runtime, Loader: branding.Loader, ChannelArchives: branding.ChannelArchives()}
		root := install.DataRoot(filepath.Join(out, "data"), *channel)
		if e = os.MkdirAll(root, 0700); e != nil {
			return e
		}
		settings := filepath.Join(root, "fixture-settings.txt")
		if e = os.WriteFile(settings, []byte("preserve"), 0600); e != nil {
			return e
		}
		if *previousBundle != "" {
			data, err := os.ReadFile(filepath.Join(*previousBundle, "core.release.json"))
			if err != nil {
				return err
			}
			previous, err := release.Verify(data, branding.ReleaseTrust(), "core", time.Now(), true)
			if err != nil {
				return err
			}
			if err = engine.Install(target, root, previous, filepath.Join(*previousBundle, previous.Manifest.Artifact.Name)); err != nil {
				return fmt.Errorf("previous public beta fixture: %w", err)
			}
		}
		if e = engine.Install(target, root, v, payload); e != nil {
			return e
		}
		state, e := engine.Inspect(target)
		if e != nil || state != "Installed" {
			return fmt.Errorf("ownership check: %s %v", state, e)
		}
		if e = engine.Install(target, root, v, payload); e != nil {
			return fmt.Errorf("repair: %w", e)
		}
		if e = engine.Uninstall(target); e != nil {
			return e
		}
		restored, e := os.ReadFile(filepath.Join(target.Resources, "app.asar"))
		if e != nil {
			return e
		}
		_ = restored
		state, e = engine.Inspect(target)
		if e != nil || state != "Not installed" {
			return fmt.Errorf("restoration: %s %v", state, e)
		}
		data, e := os.ReadFile(settings)
		if e != nil || string(data) != "preserve" {
			return fmt.Errorf("fixture settings changed")
		}
		fmt.Println("Disposable fixture installation, ownership, repair and uninstall passed; settings preserved.")
	}
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
