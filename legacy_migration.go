// SPDX-License-Identifier: GPL-3.0
package main

import (
	"bytes"
	"encoding/json"
	"equilotl/branding"
	"equilotl/internal/install"
	"os"
	"path/filepath"
)

func secureEngine() install.Engine {
	return install.Engine{Trust: branding.ReleaseTrust(), TrustData: branding.TrustData, LegacyTrustData: branding.LegacyTrustData, Runtime: branding.Runtime, Loader: branding.Loader, ChannelArchives: branding.ChannelArchives()}
}
func (di *DiscordInstall) prepareLegacy() error {
	resources := filepath.Dir(di.appPath)
	if !ExistsFile(filepath.Join(resources, "evencord-install.json")) {
		return nil
	}
	var owner struct{ Product, DataRoot, OriginalSHA256 string }
	b, err := os.ReadFile(filepath.Join(resources, "evencord-install.json"))
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &owner); err != nil {
		return err
	}
	root, _ := json.Marshal(owner.DataRoot)
	payload, _ := json.Marshal(filepath.Join(owner.DataRoot, "evencord.asar", "patcher.js"))
	expected := []byte("process.env.EVENCORD_USER_DATA_DIR=" + string(root) + ";require(" + string(payload) + ");")
	index, _ := os.ReadFile(filepath.Join(resources, "app.asar", "index.js"))
	if owner.Product == "evencord" && bytes.Equal(index, expected) {
		return os.Rename(filepath.Join(resources, "evencord-install.json"), filepath.Join(resources, "evencord-legacy-install.json"))
	}
	target := install.Target{Resources: resources, Channel: di.branch}
	if err := secureEngine().Recover(target); err != nil {
		return err
	}
	if err := secureEngine().Uninstall(target); err != nil {
		return err
	}
	di.isPatched = false
	return nil
}
