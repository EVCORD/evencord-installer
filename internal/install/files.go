// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package install

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func SafePath(p string) error {
	p, e := filepath.Abs(p)
	if e != nil {
		return e
	}
	for {
		s, e := os.Lstat(p)
		if e == nil && s.Mode()&os.ModeSymlink != 0 {
			return errors.New("linked filesystem paths are unsupported")
		}
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		parent := filepath.Dir(p)
		if parent == p {
			break
		}
		p = parent
	}
	return nil
}
func hashFile(p string) (string, error) {
	if e := SafePath(p); e != nil {
		return "", e
	}
	f, e := os.Open(p)
	if e != nil {
		return "", e
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() {
		return "", errors.New("expected a regular file")
	}
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func Atomic(p string, data []byte) error {
	if e := SafePath(p); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(p), ".evencord-write-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, e = f.Write(data); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return replaceFile(f.Name(), p)
}
func lockFallback(root string) (func(), error) {
	p := filepath.Join(root, ".evencord-operation.lock")
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return nil, errors.New("operation locked; close other installers, then review recovery instructions")
	}
	return func() { f.Close(); os.Remove(p) }, nil
}
