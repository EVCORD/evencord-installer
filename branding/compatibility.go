// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package branding

import (
	_ "embed"
	"encoding/json"
)

//go:embed discord-compatibility.json
var compatibilityData []byte

func ChannelArchives() map[string][]string {
	var catalog struct {
		Channels []struct {
			Channel string `json:"discordChannel"`
			Hash    string `json:"archiveSHA256"`
		}
	}
	if err := json.Unmarshal(compatibilityData, &catalog); err != nil {
		panic(err)
	}
	hashes := make(map[string][]string)
	for _, entry := range catalog.Channels {
		hashes[entry.Channel] = append(hashes[entry.Channel], entry.Hash)
	}
	return hashes
}
