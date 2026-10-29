// SPDX-License-Identifier: GPL-3.0
package install

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"equilotl/internal/release"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixture struct {
	engine      Engine
	target      Target
	root, stage string
	verified    release.Verified
	private     ed25519.PrivateKey
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	base := t.TempDir()
	r := filepath.Join(base, "Discord", "app-1.0.1", "resources")
	if err := os.MkdirAll(r, 0700); err != nil {
		t.Fatal(err)
	}
	exe := make([]byte, 128)
	copy(exe, "MZ")
	binary.LittleEndian.PutUint32(exe[60:], 64)
	copy(exe[64:], "PE\x00\x00")
	binary.LittleEndian.PutUint16(exe[68:], 0x8664)
	os.WriteFile(filepath.Join(r, "..", "Discord.exe"), exe, 0600)
	os.WriteFile(filepath.Join(r, "build_info.json"), []byte("{\"releaseChannel\":\"stable\"}"), 0600)
	original := []byte("synthetic approved Discord bootstrap")
	os.WriteFile(filepath.Join(r, "app.asar"), original, 0600)
	pub, private, _ := ed25519.GenerateKey(rand.Reader)
	trust := release.Trust{Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(pub)}, Revoked: []string{}}
	trustData, _ := json.Marshal(trust)
	engine := Engine{Trust: trust, TrustData: trustData, Runtime: []byte("fixture runtime"), Loader: []byte("fixture loader")}
	payload := []byte("verified synthetic payload")
	h := sha256.Sum256(payload)
	m := release.Manifest{Schema: 1, Product: "evencord", Component: "core", Repository: "alrimii/evencord", RepositoryID: 1438489516, Channel: "stable", Version: "1.0.0", Sequence: 1, Source: strings.Repeat("a", 40), Upstream: strings.Repeat("b", 40), Platform: "win32", Arch: "x64", Discord: "stable", Loader: 1, Published: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339), Expires: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), Notes: "Fixture", Archives: []string{digest(original)}, Artifact: release.Artifact{Name: "desktop.asar", Size: int64(len(payload)), SHA256: hex.EncodeToString(h[:])}}
	body, _ := json.Marshal(m)
	env, _ := json.Marshal(release.Envelope{KeyID: "test", Signed: base64.StdEncoding.EncodeToString(body), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, append([]byte(release.Domain), body...)))})
	v, err := release.Verify(env, trust, "core", time.Now(), true)
	if err != nil {
		t.Fatal(err)
	}
	stage, err := release.StageReader(filepath.Join(base, "download"), v, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	return fixture{engine, Target{r, "stable"}, filepath.Join(base, "data"), stage, v, private}
}
func (f fixture) install(t *testing.T) {
	t.Helper()
	if err := f.engine.Install(f.target, f.root, f.verified, f.stage); err != nil {
		t.Fatal(err)
	}
}
func TestFreshRepairUninstallPreservesSettings(t *testing.T) {
	f := newFixture(t)
	f.install(t)
	s, e := f.engine.Inspect(f.target)
	if e != nil || s != "Installed" {
		t.Fatalf("%s %v", s, e)
	}
	os.WriteFile(filepath.Join(f.root, "settings.json"), []byte("keep"), 0600)
	if e = f.engine.Install(f.target, f.root, f.verified, f.stage); e != nil {
		t.Fatal(e)
	}
	payload := filepath.Join(f.root, "releases", f.verified.Manifest.Artifact.SHA256+"-desktop.asar")
	os.Remove(payload)
	s, e = f.engine.Inspect(f.target)
	if e != nil || s != "Repair available" {
		t.Fatalf("%s %v", s, e)
	}
	if e = f.engine.Install(f.target, f.root, f.verified, f.stage); e != nil {
		t.Fatal(e)
	}
	if e = f.engine.Uninstall(f.target); e != nil {
		t.Fatal(e)
	}
	if b, _ := os.ReadFile(filepath.Join(f.target.Resources, "app.asar")); string(b) != "synthetic approved Discord bootstrap" {
		t.Fatal("original not restored")
	}
	if b, _ := os.ReadFile(filepath.Join(f.root, "settings.json")); string(b) != "keep" {
		t.Fatal("settings lost")
	}
}
func TestForeignAndDamagedInstallations(t *testing.T) {
	for _, name := range []string{"Equicord", "Vencord", "OpenAsar", "unknown", "directory", "missing-original", "damaged-loader", "unsupported-channel"} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			r := f.target.Resources
			switch name {
			case "missing-original":
				f.install(t)
				os.Remove(filepath.Join(r, "_app.asar"))
			case "damaged-loader":
				f.install(t)
				os.WriteFile(filepath.Join(r, "app.asar", "index.js"), []byte("foreign"), 0600)
			case "unsupported-channel":
				f.target.Channel = "canary"
			case "directory":
				os.Remove(filepath.Join(r, "app.asar"))
				os.Mkdir(filepath.Join(r, "app.asar"), 0700)
			default:
				os.WriteFile(filepath.Join(r, "app.asar"), []byte(name), 0600)
				os.WriteFile(filepath.Join(r, "_app.asar"), []byte("foreign recovery"), 0600)
			}
			before, _ := os.ReadFile(filepath.Join(r, "_app.asar"))
			if e := f.engine.Install(f.target, f.root, f.verified, f.stage); e == nil {
				t.Fatal("unsafe install accepted")
			}
			if after, _ := os.ReadFile(filepath.Join(r, "_app.asar")); !bytes.Equal(before, after) {
				t.Fatal("foreign recovery changed")
			}
			if e := f.engine.Uninstall(f.target); e == nil {
				t.Fatal("unsafe uninstall accepted")
			}
		})
	}
}
func TestInterruptedInstallRecovery(t *testing.T) {
	for _, step := range []string{"original-preserved", "loader-activated"} {
		t.Run(step, func(t *testing.T) {
			f := newFixture(t)
			f.engine.AfterStep = func(s string) error {
				if s == step {
					return errors.New("simulated interruption")
				}
				return nil
			}
			if e := f.engine.Install(f.target, f.root, f.verified, f.stage); e == nil {
				t.Fatal("failure not reported")
			}
			f.engine.AfterStep = nil
			if e := f.engine.Recover(f.target); e != nil {
				t.Fatal(e)
			}
			state, e := f.engine.Inspect(f.target)
			if e != nil || (state != "Installed" && state != "Not installed") {
				t.Fatalf("%s %v", state, e)
			}
		})
	}
}
func TestInterruptedUninstallRecovery(t *testing.T) {
	f := newFixture(t)
	f.install(t)
	f.engine.AfterStep = func(s string) error {
		if s == "loader-removed" {
			return errors.New("interrupted")
		}
		return nil
	}
	if e := f.engine.Uninstall(f.target); e == nil {
		t.Fatal("expected failure")
	}
	f.engine.AfterStep = nil
	if e := f.engine.Recover(f.target); e != nil {
		t.Fatal(e)
	}
	if state, e := f.engine.Inspect(f.target); e != nil || state != "Not installed" {
		t.Fatalf("%s %v", state, e)
	}
}
func TestFailedRepairPreservesCurrent(t *testing.T) {
	f := newFixture(t)
	f.install(t)
	before, _ := os.ReadFile(filepath.Join(f.root, "current.json"))
	os.WriteFile(f.stage, []byte("bad"), 0600)
	if e := f.engine.Install(f.target, f.root, f.verified, f.stage); e == nil {
		t.Fatal("corrupt repair accepted")
	}
	after, _ := os.ReadFile(filepath.Join(f.root, "current.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("current release changed")
	}
}
func TestInterruptedUpdateKeepsPrevious(t *testing.T) {
	f := newFixture(t)
	f.install(t)
	before, _ := os.ReadFile(filepath.Join(f.root, "current.json"))
	f.engine.AfterStep = func(string) error { return errors.New("interrupted update") }
	if e := f.engine.Activate(f.root, f.verified, f.stage); e == nil {
		t.Fatal("expected error")
	}
	after, _ := os.ReadFile(filepath.Join(f.root, "current.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("current changed")
	}
}
func TestRecoveryRefusesUnknownReplacement(t *testing.T) {
	f := newFixture(t)
	f.engine.AfterStep = func(string) error { return errors.New("interrupt") }
	_ = f.engine.Install(f.target, f.root, f.verified, f.stage)
	os.WriteFile(filepath.Join(f.target.Resources, "app.asar"), []byte("foreign"), 0600)
	f.engine.AfterStep = nil
	if e := f.engine.Recover(f.target); e == nil {
		t.Fatal("unknown replacement accepted")
	}
}

func TestRootLockAndPayloadRecovery(t *testing.T) {
	f := newFixture(t)
	f.install(t)
	unlock, err := lock(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lock(f.root); err == nil {
		t.Fatal("concurrent operation accepted")
	}
	unlock()
	data, err := os.ReadFile(filepath.Join(f.root, "current.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = Atomic(filepath.Join(f.root, "previous.json"), data); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(f.root, "current.json"), []byte("interrupted write"), 0600)
	if err = f.engine.Recover(f.target); err != nil {
		t.Fatal(err)
	}
	if _, err = f.engine.Current(f.root); err != nil {
		t.Fatal(err)
	}
}
func TestDamagedPayloadRepair(t *testing.T) {
	f := newFixture(t)
	f.install(t)
	p := filepath.Join(f.root, "releases", f.verified.Manifest.Artifact.SHA256+"-desktop.asar")
	os.WriteFile(p, []byte("damaged"), 0600)
	if err := f.engine.Install(f.target, f.root, f.verified, f.stage); err != nil {
		t.Fatal(err)
	}
	if _, err := f.engine.Current(f.root); err != nil {
		t.Fatal(err)
	}
	kept, _ := filepath.Glob(p + ".damaged-*")
	if len(kept) != 1 {
		t.Fatal("damaged payload evidence not preserved")
	}
}
