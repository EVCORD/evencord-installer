//go:build !windows

// SPDX-License-Identifier: GPL-3.0
package main

func (di *DiscordInstall) closeForOperation() error { return nil }
