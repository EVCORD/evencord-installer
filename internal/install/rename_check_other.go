//go:build !windows

// SPDX-License-Identifier: GPL-3.0
package install

func checkRenameAccess(_ string) error { return nil }
