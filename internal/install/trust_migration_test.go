// SPDX-License-Identifier: GPL-3.0
package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTrustMigrationRecovery(t *testing.T) {
	for _, step := range []string{"migrate-trust.json", "migrate-evencord-owner.json", "migrate-owner"} {
		t.Run(step, func(t *testing.T) {
			f := newFixture(t)
			if err := f.engine.Install(f.target, f.root, f.verified, f.stage); err != nil {
				t.Fatal(err)
			}
			f.engine.LegacyTrustData = f.engine.TrustData
			f.engine.TrustData = append(append([]byte{}, f.engine.TrustData...), '\n')
			f.engine.AfterStep = func(s string) error {
				if s == step {
					return errors.New("interrupted")
				}
				return nil
			}
			if err := f.engine.Install(f.target, f.root, f.verified, f.stage); err == nil {
				t.Fatal("expected migration interruption")
			}
			f.engine.AfterStep = nil
			if err := f.engine.Recover(f.target); err != nil {
				t.Fatal(err)
			}
			o, err := readOwner(filepath.Join(f.target.Resources, ownerName))
			if err != nil {
				t.Fatal(err)
			}
			if o.Files["trust.json"] != digest(f.engine.TrustData) {
				t.Fatal("trust not migrated")
			}
			b, err := os.ReadFile(filepath.Join(f.target.Resources, "app.asar", "evencord-owner.json"))
			if err != nil {
				t.Fatal(err)
			}
			var inner Owner
			if json.Unmarshal(b, &inner) != nil || inner.Files["trust.json"] != o.Files["trust.json"] {
				t.Fatal("owner records differ")
			}
			if err = f.engine.Uninstall(f.target); err != nil {
				t.Fatal(err)
			}
		})
	}
}
