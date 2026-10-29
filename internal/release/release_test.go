// SPDX-License-Identifier: GPL-3.0
package release

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T) (Trust, Manifest, ed25519.PrivateKey) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	return Trust{Keys: map[string]string{"fixture": base64.StdEncoding.EncodeToString(pub)}}, Manifest{Schema: 1, Product: "evencord", Component: "core", Repository: "alrimii/evencord", RepositoryID: 1438489516, Channel: "stable", Version: "1.2.3", Sequence: 2, Source: strings.Repeat("a", 40), Upstream: strings.Repeat("b", 40), Platform: "win32", Arch: "x64", Discord: "stable", Loader: 1, Published: time.Now().Add(-time.Hour).UTC().Format(time.RFC3339), Expires: time.Now().Add(time.Hour).UTC().Format(time.RFC3339), Archives: []string{strings.Repeat("c", 64)}, Artifact: Artifact{"desktop.asar", 3, "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}}, priv
}
func signed(m Manifest, p ed25519.PrivateKey) []byte {
	b, _ := json.Marshal(m)
	s := ed25519.Sign(p, append([]byte(Domain), b...))
	out, _ := json.Marshal(Envelope{"fixture", base64.StdEncoding.EncodeToString(b), base64.StdEncoding.EncodeToString(s)})
	return out
}
func TestVerification(t *testing.T) {
	trust, m, p := fixture(t)
	if _, e := Verify(signed(m, p), trust, "core", time.Now(), true); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"signature", "repository", "component", "channel", "expired", "digest", "key", "version"} {
		t.Run(name, func(t *testing.T) {
			tm := m
			tt := trust
			data := signed(tm, p)
			switch name {
			case "signature":
				var env Envelope
				json.Unmarshal(data, &env)
				env.Signature = base64.StdEncoding.EncodeToString(make([]byte, 64))
				data, _ = json.Marshal(env)
			case "repository":
				tm.Repository = "Equicord/Equicord"
				data = signed(tm, p)
			case "component":
				tm.Component = "installer-cli"
				data = signed(tm, p)
			case "channel":
				tm.Discord = "canary"
				data = signed(tm, p)
			case "expired":
				tm.Expires = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
				data = signed(tm, p)
			case "digest":
				tm.Artifact.SHA256 = "../wrong"
				data = signed(tm, p)
			case "key":
				tt = Trust{Keys: map[string]string{}}
			case "version":
				tm.Version = "1.99999999999.0"
				data = signed(tm, p)
			}
			if _, e := Verify(data, tt, "core", time.Now(), true); e == nil {
				t.Fatal("invalid release accepted")
			}
		})
	}
}

type interrupted struct{}

func (interrupted) Read([]byte) (int, error) { return 0, errors.New("connection interrupted") }
func TestStagingRejectsBrokenDownloads(t *testing.T) {
	trust, m, p := fixture(t)
	v, _ := Verify(signed(m, p), trust, "core", time.Now(), true)
	for _, r := range []io.Reader{strings.NewReader("ab"), strings.NewReader("abcd"), strings.NewReader("bad"), interrupted{}} {
		root := t.TempDir()
		if _, e := StageReader(root, v, r); e == nil {
			t.Fatal("bad download accepted")
		}
		files, _ := os.ReadDir(root)
		if len(files) != 0 {
			t.Fatal("partial download retained")
		}
	}
	root := t.TempDir()
	file, e := StageReader(root, v, bytes.NewBufferString("abc"))
	if e != nil {
		t.Fatal(e)
	}
	if e = CheckFile(file, m.Artifact); e != nil {
		t.Fatal(e)
	}
	if filepath.Base(file) != m.Artifact.SHA256+"-desktop.asar" {
		t.Fatal("unexpected payload name")
	}
}
func TestRollbackAndRedirects(t *testing.T) {
	_, m, _ := fixture(t)
	next := m
	next.Sequence--
	if Advance(next, m) == nil {
		t.Fatal("rollback accepted")
	}
	for _, u := range []string{"http://github.com/x", "https://evil.example/x", "https://github.com.evil/x", "https://user@github.com/x", "https://github.com:443/x"} {
		if SafeURL(u) {
			t.Fatal(u)
		}
	}
	if !SafeURL("https://release-assets.githubusercontent.com/x") {
		t.Fatal("valid redirect refused")
	}
}

func TestSelfUpdateFailurePreservesExecutable(t *testing.T) {
	trust, m, key := fixture(t)
	m.Component = "installer-cli"
	m.Repository = "alrimii/evencord-installer"
	m.RepositoryID = 1438489594
	m.Artifact.Name = "EvenCordInstallerCli-x64.exe"
	dir := t.TempDir()
	old := filepath.Join(dir, "EvenCordInstallerCli-x64.exe")
	os.WriteFile(old, []byte("working executable"), 0600)
	if _, err := StageSelfUpdate(filepath.Join(dir, "updates"), signed(m, key), trust, "installer-cli", "1.0.0", interrupted{}); err == nil {
		t.Fatal("failed update accepted")
	}
	if b, _ := os.ReadFile(old); string(b) != "working executable" {
		t.Fatal("working executable changed")
	}
	for i := 0; i < 2; i++ {
		envelope := signed(m, key)
		candidate, err := StageSelfUpdate(filepath.Join(dir, "updates"), envelope, trust, "installer-cli", "1.0.0", strings.NewReader("abc"))
		if err != nil {
			t.Fatal(err)
		}
		proof, err := os.ReadFile(candidate + ".release.json")
		if err != nil || !bytes.Equal(proof, envelope) {
			t.Fatal("signed proof was not completely staged")
		}
	}
}

func TestSharedNodeSignatureVector(t *testing.T) {
	data, err := os.ReadFile("../../tests/protocol-vector.json")
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		Trust    Trust
		Envelope json.RawMessage
	}
	if err = json.Unmarshal(data, &vector); err != nil {
		t.Fatal(err)
	}
	v, err := Verify(vector.Envelope, vector.Trust, "core", time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if v.Manifest.Artifact.Size != 3 {
		t.Fatal("cross-language manifest differs")
	}
}
func TestBetaChannelAndVersions(t *testing.T) {
	trust, m, p := fixture(t)
	trust.Channels = []string{"beta"}
	m.Channel = "beta"
	m.Version = "0.1.0-beta.1"
	if _, e := Verify(signed(m, p), trust, "core", time.Now(), true); e != nil {
		t.Fatal(e)
	}
	trust.Channels = []string{"stable"}
	if _, e := Verify(signed(m, p), trust, "core", time.Now(), true); e == nil {
		t.Fatal("cross-channel release accepted")
	}
	for _, pair := range [][2]string{{"0.1.0-beta.10", "0.1.0-beta.2"}, {"0.1.0", "0.1.0-beta.10"}} {
		if c, e := Compare(pair[0], pair[1]); e != nil || c != 1 {
			t.Fatal(pair, c, e)
		}
	}
	if _, e := Compare("0.1.0-beta.01", "0.1.0-beta.1"); e == nil {
		t.Fatal("invalid beta accepted")
	}
}
