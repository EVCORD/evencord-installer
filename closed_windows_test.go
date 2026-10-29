// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package main

import (
	"golang.org/x/sys/windows"
	"testing"
)

func TestCloseProcessPermissions(t *testing.T) {
	root := `C:\Fixture\DiscordCanary`
	for _, scenario := range []struct {
		name                       string
		openErr, queryErr, stopErr error
		image, secondImage         string
		denyStop, timeout          bool
		wantOpens, wantStops       int
		wantErr                    bool
	}{
		{name: "inaccessible process does not block archive checks", openErr: windows.ERROR_ACCESS_DENIED, wantOpens: 1},
		{name: "exited process", openErr: windows.ERROR_INVALID_PARAMETER, wantOpens: 1},
		{name: "query denied", queryErr: windows.ERROR_ACCESS_DENIED, wantOpens: 1},
		{name: "other installation", image: `C:\Other\DiscordCanary\DiscordCanary.exe`, wantOpens: 1},
		{name: "adjacent directory", image: root + `Other\DiscordCanary.exe`, wantOpens: 1},
		{name: "termination denied defers to archive checks", denyStop: true, wantOpens: 2},
		{name: "PID reused", secondImage: `C:\Other\Discord.exe`, wantOpens: 2},
		{name: "selected installation", wantOpens: 2, wantStops: 1},
		{name: "termination race", stopErr: windows.ERROR_ACCESS_DENIED, wantOpens: 2, wantStops: 1},
		{name: "exit timeout", timeout: true, wantOpens: 2, wantStops: 1, wantErr: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			opens, stops, closes, queries := 0, 0, 0, 0
			api := processAccess{
				open: func(rights uint32, inherit bool, pid uint32) (windows.Handle, error) {
					opens++
					if opens == 1 && rights != windows.PROCESS_QUERY_LIMITED_INFORMATION {
						t.Fatalf("inspection requested excess permissions: %x", rights)
					}
					if inherit || pid != 123 {
						t.Fatal("unexpected process identity")
					}
					if scenario.openErr != nil {
						return 0, scenario.openErr
					}
					if opens == 2 && scenario.denyStop {
						return 0, windows.ERROR_ACCESS_DENIED
					}
					return windows.Handle(opens), nil
				},
				image: func(windows.Handle) (string, error) {
					queries++
					if queries == 2 && scenario.secondImage != "" {
						return scenario.secondImage, nil
					}
					if scenario.image != "" {
						return scenario.image, scenario.queryErr
					}
					return root + `\app-1.0.1286\DiscordCanary.exe`, scenario.queryErr
				},
				wait: func(_ windows.Handle, ms uint32) (uint32, error) {
					if ms == 0 || scenario.timeout {
						return uint32(windows.WAIT_TIMEOUT), nil
					}
					return windows.WAIT_OBJECT_0, nil
				},
				stop:  func(windows.Handle, uint32) error { stops++; return scenario.stopErr },
				close: func(windows.Handle) error { closes++; return nil },
			}
			err := closeProcessInRoot(123, root, api)
			if (err != nil) != scenario.wantErr || opens != scenario.wantOpens || stops != scenario.wantStops {
				t.Fatalf("error=%v opens=%d stops=%d", err, opens, stops)
			}
			wantCloses := opens
			if scenario.openErr != nil || scenario.denyStop {
				wantCloses--
			}
			if closes != wantCloses {
				t.Fatalf("handles not closed: %d != %d", closes, wantCloses)
			}
		})
	}
}
