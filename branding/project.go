// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.

package branding

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed project.json
var data []byte

type Contract struct {
	SchemaVersion int
	ID            string
	DisplayName   string
	InstallerName string
	Description   string
	Repositories  struct {
		Core struct {
			Slug string
			ID   int64
		}
		Installer struct {
			Slug string
			ID   int64
		}
	}
	Links struct {
		Source        string
		Documentation string
		Support       string
	}
	Storage struct {
		Directory         string
		PortableDirectory string
		Payload           string
		Namespace         string
		EnvironmentPrefix string
	}
	Release struct {
		Channel           string
		Platform          string
		Architecture      string
		DiscordChannel    string
		CoreAsset         string
		InstallerGuiAsset string
		InstallerCliAsset string
	}
	Capabilities struct {
		Distribution bool
		Updates      bool
		Installation bool
		Repair       bool
		SelfUpdate   bool
		HostRepatch  bool
		Cloud        bool
	}
}

var Project = load()

func load() Contract {
	var project Contract
	if err := json.Unmarshal(data, &project); err != nil {
		panic(fmt.Errorf("invalid embedded project identity: %w", err))
	}
	if project.SchemaVersion != 1 || project.ID != "evencord" {
		panic("unsupported project identity")
	}
	return project
}
