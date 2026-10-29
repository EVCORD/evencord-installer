// SPDX-License-Identifier: GPL-3.0
package install

import (
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLockedOriginalIsUnchanged(t *testing.T) {
	f := newFixture(t)
	app := filepath.Join(f.target.Resources, "app.asar")
	name, err := windows.UTF16PtrFromString(app)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	err = f.engine.Install(f.target, f.root, f.verified, f.stage)
	if err == nil || !strings.Contains(err.Error(), "using app.asar") {
		t.Fatalf("expected clear lock error, got %v", err)
	}
	for _, path := range []string{filepath.Join(f.target.Resources, journalName), filepath.Join(f.target.Resources, "_app.asar"), filepath.Join(f.root, "current.json")} {
		if _, err = os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("locked original changed install state: %s", path)
		}
	}
}
