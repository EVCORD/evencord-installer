// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.

package main

import (
	"encoding/json"
	"equilotl/branding"
	"equilotl/buildinfo"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	cli := flag.Bool("cli", false, "Generate CLI product metadata")
	flag.Parse()
	if err := generate(*cli); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate(cli bool) error {
	data, err := os.ReadFile("winres/winres.json")
	if err != nil {
		return err
	}
	var resources map[string]json.RawMessage
	if err = json.Unmarshal(data, &resources); err != nil {
		return err
	}
	name := branding.Project.InstallerName
	filename := branding.Project.Release.InstallerGuiAsset
	if cli {
		name += " CLI"
		filename = branding.Project.Release.InstallerCliAsset
	}
	icon := "../winres/icon.ico"
	resources["RT_GROUP_ICON"], err = json.Marshal(map[string]map[string]string{"APP": {"0000": icon}})
	if err != nil {
		return err
	}
	version := map[string]any{
		"fixed": map[string]string{"file_version": "1.0.0.0", "product_version": "1.0.0.0"},
		"info": map[string]map[string]string{"0409": {
			"CompanyName": "EvenCord contributors", "ProductName": branding.Project.InstallerName,
			"FileDescription": name, "ProductVersion": buildinfo.InstallerTag,
			"OriginalFilename": filename, "FileVersion": "1.0.0.0", "InternalName": "evencord-installer",
			"LegalCopyright": "© 2026 Vencord and Equicord contributors - GPL3.0. Additional branding copyright 2026 EvenCord contributors.",
		}},
	}
	resources["RT_VERSION"], err = json.Marshal(map[string]map[string]any{"#1": {"0000": version}})
	if err != nil {
		return err
	}
	data, err = json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll("build", 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join("build", "winres.json"), append(data, '\n'), 0644)
}
