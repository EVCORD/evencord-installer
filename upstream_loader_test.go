// SPDX-License-Identifier: GPL-3.0
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUpstreamLoaderRoundTrip(t *testing.T) {
	root := t.TempDir()
	before := EquicordDirectory
	EquicordDirectory = filepath.Join(root, "profile", "evencord.asar")
	t.Cleanup(func() { EquicordDirectory = before })
	original := []byte("original Discord fixture")
	app := filepath.Join(root, "app.asar")
	if err := os.WriteFile(app, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := patchAppAsar(root, false); err != nil {
		t.Fatal(err)
	}
	if owned, err := isEquicordLoaderAppAsar(app); err != nil || !owned {
		t.Fatalf("loader not detected: %v", err)
	}
	if err := unpatchAppAsar(root, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(app)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("original not restored: %v", err)
	}
}

func TestTransitionDirectoryLoaderRoundTrip(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "app.asar")
	if err := os.Mkdir(app, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "index.js"), []byte(`require("profile/evencord.asar/patcher.js");`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "_app.asar"), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := unpatchAppAsar(root, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(app)
	if err != nil || string(got) != "original" {
		t.Fatalf("original not restored: %v", err)
	}
}

func TestInterruptedTransitionRetiresLegacyOwnership(t *testing.T) {
	root := t.TempDir()
	resources := filepath.Join(root, "resources")
	app := filepath.Join(resources, "app.asar")
	profile := filepath.Join(root, "profile")
	if err := os.MkdirAll(app, 0700); err != nil {
		t.Fatal(err)
	}
	owner, _ := json.Marshal(map[string]string{"product": "evencord", "dataRoot": profile})
	if err := os.WriteFile(filepath.Join(resources, "evencord-install.json"), owner, 0600); err != nil {
		t.Fatal(err)
	}
	r, _ := json.Marshal(profile)
	p, _ := json.Marshal(filepath.Join(profile, "evencord.asar", "patcher.js"))
	index := "process.env.EVENCORD_USER_DATA_DIR=" + string(r) + ";require(" + string(p) + ");"
	if err := os.WriteFile(filepath.Join(app, "index.js"), []byte(index), 0600); err != nil {
		t.Fatal(err)
	}
	di := DiscordInstall{appPath: filepath.Join(resources, "app"), branch: "stable"}
	if err := di.prepareLegacy(); err != nil {
		t.Fatal(err)
	}
	if !ExistsFile(filepath.Join(resources, "evencord-legacy-install.json")) {
		t.Fatal("legacy recovery record was not preserved")
	}
	if err := di.prepareLegacy(); err != nil {
		t.Fatal(err)
	}
}
