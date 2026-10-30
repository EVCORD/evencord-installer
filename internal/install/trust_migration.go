// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const trustJournal = ".evencord-trust-migration.json"

func (e Engine) upgradeTrust(t Target) error {
	if len(e.LegacyTrustData) == 0 {
		return nil
	}
	jp := filepath.Join(t.Resources, trustJournal)
	var old Owner
	if exists(jp) {
		var j Journal
		b, err := os.ReadFile(jp)
		if err != nil {
			return err
		}
		if json.Unmarshal(b, &j) != nil || j.Operation != "trust" {
			return errors.New("unknown loader migration")
		}
		old = j.Owner
	} else {
		var err error
		old, err = readOwner(filepath.Join(t.Resources, ownerName))
		if err != nil {
			return err
		}
		if err = e.validLoader(filepath.Join(t.Resources, "app.asar"), old); err != nil {
			return err
		}
		if old.Files["trust.json"] == digest(e.TrustData) {
			return nil
		}
	}
	if old.Schema != 1 || old.Product != "evencord" || !filepath.IsAbs(old.DataRoot) || len(old.Files) != 3 || old.Files["index.js"] != digest(e.Loader) || old.Files["runtime.cjs"] != digest(e.Runtime) || old.Files["trust.json"] != digest(e.LegacyTrustData) {
		return errors.New("unknown legacy loader ownership")
	}
	next := old
	next.Files = map[string]string{"index.js": digest(e.Loader), "runtime.cjs": digest(e.Runtime), "trust.json": digest(e.TrustData)}
	before, after := e.expected(old), e.expected(next)
	app := filepath.Join(t.Resources, "app.asar")
	entries, err := os.ReadDir(app)
	if err != nil {
		return err
	}
	if len(entries) != len(before) {
		return errors.New("unknown loader files block migration")
	}
	for name, a := range before {
		file := filepath.Join(app, name)
		if err = SafePath(file); err != nil {
			return err
		}
		b, readErr := os.ReadFile(file)
		if readErr != nil || (!bytes.Equal(b, a) && !bytes.Equal(b, after[name])) {
			return errors.New("loader changed during migration: " + name)
		}
	}
	record := filepath.Join(t.Resources, ownerName)
	if err = SafePath(record); err != nil {
		return err
	}
	b, err := os.ReadFile(record)
	if err != nil || (!bytes.Equal(b, before["evencord-owner.json"]) && !bytes.Equal(b, after["evencord-owner.json"])) {
		return errors.New("ownership changed during migration")
	}
	h, err := hashFile(filepath.Join(t.Resources, "_app.asar"))
	if err != nil || h != old.Original {
		return errors.New("recovery archive changed")
	}
	if !exists(jp) {
		if err = writeJSON(jp, Journal{"trust", old}); err != nil {
			return err
		}
	}
	for _, name := range []string{"trust.json", "evencord-owner.json"} {
		if err = Atomic(filepath.Join(app, name), after[name]); err != nil {
			return err
		}
		if err = e.step("migrate-" + name); err != nil {
			return err
		}
	}
	if err = Atomic(record, after["evencord-owner.json"]); err != nil {
		return err
	}
	if err = e.step("migrate-owner"); err != nil {
		return err
	}
	if err = e.validLoader(app, next); err != nil {
		return err
	}
	return os.Remove(jp)
}
