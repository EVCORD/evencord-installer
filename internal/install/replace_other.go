//go:build !windows

// SPDX-License-Identifier: GPL-3.0
package install

import "os"

func replaceFile(from, to string) error { return os.Rename(from, to) }

func lock(root string) (func(), error) { return lockFallback(root) }
