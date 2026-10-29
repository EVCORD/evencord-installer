// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package release

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const MaxArtifact = 256 << 20
const MaxEnvelope = 128 << 10
const Domain = "EvenCord release v1\n"

type Envelope struct {
	KeyID     string `json:"keyId"`
	Signed    string `json:"signed"`
	Signature string `json:"signature"`
}
type Artifact struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type Manifest struct {
	Schema       int      `json:"schemaVersion"`
	Product      string   `json:"product"`
	Component    string   `json:"component"`
	Repository   string   `json:"repository"`
	RepositoryID int64    `json:"repositoryId"`
	Channel      string   `json:"channel"`
	Version      string   `json:"version"`
	Sequence     int64    `json:"sequence"`
	Source       string   `json:"sourceCommit"`
	Upstream     string   `json:"upstreamCommit"`
	Platform     string   `json:"platform"`
	Arch         string   `json:"architecture"`
	Discord      string   `json:"discordChannel"`
	Loader       int      `json:"loaderProtocol"`
	Published    string   `json:"publishedAt"`
	Expires      string   `json:"expiresAt"`
	Archives     []string `json:"discordArchives"`
	Notes        string   `json:"notes"`
	Artifact     Artifact `json:"artifact"`
}
type Trust struct {
	Keys    map[string]string `json:"keys"`
	Revoked []string          `json:"revoked"`
}
type Verified struct {
	Manifest Manifest
	Envelope []byte
}

var hex40 = regexp.MustCompile("^[a-f0-9]{40}$")
var hex64 = regexp.MustCompile("^[a-f0-9]{64}$")
var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func Decode(data []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}
func Verify(data []byte, trust Trust, component string, now time.Time, freshness bool) (Verified, error) {
	var out Verified
	if len(data) > MaxEnvelope {
		return out, errors.New("release envelope too large")
	}
	var env Envelope
	if err := Decode(data, &env); err != nil {
		return out, err
	}
	for _, id := range trust.Revoked {
		if id == env.KeyID {
			return out, errors.New("release signing key revoked")
		}
	}
	key, err := base64.StdEncoding.Strict().DecodeString(trust.Keys[env.KeyID])
	if err != nil || len(key) != ed25519.PublicKeySize {
		return out, errors.New("release signing key is not trusted")
	}
	body, err := base64.StdEncoding.Strict().DecodeString(env.Signed)
	if err != nil {
		return out, err
	}
	sig, err := base64.StdEncoding.Strict().DecodeString(env.Signature)
	if err != nil || !ed25519.Verify(key, append([]byte(Domain), body...), sig) {
		return out, errors.New("invalid release signature")
	}
	var m Manifest
	if err = Decode(body, &m); err != nil {
		return out, err
	}
	repo, id, name := "alrimii/evencord", int64(1438489516), "desktop.asar"
	if component == "installer-gui" {
		repo, id, name = "alrimii/evencord-installer", 1438489594, "EvenCordInstaller-x64.exe"
	} else if component == "installer-cli" {
		repo, id, name = "alrimii/evencord-installer", 1438489594, "EvenCordInstallerCli-x64.exe"
	} else if component != "core" {
		return out, errors.New("unsupported component")
	}
	if m.Schema != 1 || m.Product != "evencord" || m.Component != component || m.Repository != repo || m.RepositoryID != id || m.Channel != "stable" || m.Platform != "win32" || m.Arch != "x64" || m.Discord != "stable" || m.Loader != 1 || m.Artifact.Name != name {
		return out, errors.New("release identity or compatibility mismatch")
	}
	if !versionPattern.MatchString(m.Version) || m.Sequence < 1 || m.Sequence > 9007199254740991 || !hex40.MatchString(m.Source) || !hex40.MatchString(m.Upstream) || !hex64.MatchString(m.Artifact.SHA256) || m.Artifact.Size < 1 || m.Artifact.Size > MaxArtifact || len(m.Notes) > 16384 {
		return out, errors.New("invalid release metadata")
	}
	if _, e := Compare(m.Version, m.Version); e != nil {
		return out, e
	}
	if len(m.Archives) < 1 || len(m.Archives) > 64 {
		return out, errors.New("approved Discord archive hashes are required")
	}
	for _, hash := range m.Archives {
		if !hex64.MatchString(hash) {
			return out, errors.New("invalid Discord archive hash")
		}
	}
	published, e1 := time.Parse(time.RFC3339, m.Published)
	expires, e2 := time.Parse(time.RFC3339, m.Expires)
	if e1 != nil || e2 != nil || !expires.After(published) || expires.Sub(published) > 90*24*time.Hour {
		return out, errors.New("invalid release validity period")
	}
	if freshness && (published.After(now.Add(5*time.Minute)) || !expires.After(now)) {
		return out, errors.New("release metadata expired or not yet valid")
	}
	return Verified{m, append([]byte(nil), data...)}, nil
}
func Compare(a, b string) (int, error) {
	if !versionPattern.MatchString(a) || !versionPattern.MatchString(b) {
		return 0, errors.New("stable versions must use major.minor.patch")
	}
	aa, bb := strings.Split(a, "."), strings.Split(b, ".")
	for i := range aa {
		x, e := strconv.ParseUint(aa[i], 10, 32)
		if e != nil {
			return 0, e
		}
		y, e := strconv.ParseUint(bb[i], 10, 32)
		if e != nil {
			return 0, e
		}
		if x < y {
			return -1, nil
		}
		if x > y {
			return 1, nil
		}
	}
	return 0, nil
}
func Advance(next, old Manifest) error {
	c, e := Compare(next.Version, old.Version)
	if e != nil {
		return e
	}
	if next.Component != old.Component || next.Sequence < old.Sequence || c < 0 {
		return errors.New("release rollback rejected")
	}
	if c == 0 && next.Artifact.SHA256 != old.Artifact.SHA256 {
		return errors.New("release version reused for different artifact")
	}
	if next.Sequence == old.Sequence && (next.Artifact.SHA256 != old.Artifact.SHA256 || next.Version != old.Version || next.Source != old.Source) {
		return errors.New("release sequence reused")
	}
	if c > 0 && next.Sequence <= old.Sequence {
		return errors.New("new release requires higher sequence")
	}
	return nil
}
func SafeURL(raw string) bool {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return false
	}
	switch u.Hostname() {
	case "github.com", "api.github.com", "release-assets.githubusercontent.com", "objects.githubusercontent.com":
		return true
	}
	return false
}
func Client() *http.Client {
	return &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 5 || !SafeURL(r.URL.String()) {
			return errors.New("unsafe release redirect")
		}
		return nil
	}}
}
func Read(ctx context.Context, raw string, limit int64) ([]byte, error) {
	if !SafeURL(raw) {
		return nil, errors.New("unapproved release URL")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", raw, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("User-Agent", "EvenCord")
	res, e := Client().Do(req)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("release unavailable: HTTP %d", res.StatusCode)
	}
	if res.ContentLength > limit {
		return nil, errors.New("download exceeds size limit")
	}
	data, e := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if e != nil {
		return nil, e
	}
	if int64(len(data)) > limit {
		return nil, errors.New("download exceeds size limit")
	}
	return data, nil
}
func Latest(ctx context.Context, trust Trust, component string) (Verified, error) {
	repo := "alrimii/evencord"
	if component != "core" {
		repo = "alrimii/evencord-installer"
	}
	data, e := Read(ctx, "https://github.com/"+repo+"/releases/latest/download/"+component+".release.json", MaxEnvelope)
	if e != nil {
		return Verified{}, e
	}
	return Verify(data, trust, component, time.Now(), true)
}
func ArtifactURL(m Manifest) string {
	tag := m.Component
	if tag != "core" {
		tag = "installer"
	}
	return "https://github.com/" + m.Repository + "/releases/download/" + tag + "-v" + m.Version + "/" + m.Artifact.Name
}

func CheckFile(file string, a Artifact) error {
	s, e := os.Lstat(file)
	if e != nil {
		return e
	}
	if !s.Mode().IsRegular() || s.Size() != a.Size {
		return errors.New("artifact size or file type mismatch")
	}
	f, e := os.Open(file)
	if e != nil {
		return e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return e
	}
	if hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
		return errors.New("artifact SHA-256 mismatch")
	}
	return nil
}
func StageReader(root string, v Verified, reader io.Reader) (string, error) {
	if e := os.MkdirAll(root, 0700); e != nil {
		return "", e
	}
	f, e := os.CreateTemp(root, ".evencord-download-")
	if e != nil {
		return "", e
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(io.MultiWriter(f, h), io.LimitReader(reader, v.Manifest.Artifact.Size+1))
	if e != nil {
		return "", e
	}
	if n != v.Manifest.Artifact.Size || hex.EncodeToString(h.Sum(nil)) != v.Manifest.Artifact.SHA256 {
		return "", errors.New("artifact size or SHA-256 mismatch")
	}
	if e = f.Sync(); e != nil {
		return "", e
	}
	if e = f.Close(); e != nil {
		return "", e
	}
	dest := filepath.Join(root, v.Manifest.Artifact.SHA256+"-"+v.Manifest.Artifact.Name)
	if _, e = os.Lstat(dest); e == nil {
		if e = CheckFile(dest, v.Manifest.Artifact); e != nil {
			return "", e
		}
		return dest, nil
	}
	if e = os.Rename(name, dest); e != nil {
		return "", e
	}
	return dest, nil
}
func Download(ctx context.Context, root string, v Verified) (string, error) {
	raw := ArtifactURL(v.Manifest)
	req, e := http.NewRequestWithContext(ctx, "GET", raw, nil)
	if e != nil {
		return "", e
	}
	req.Header.Set("User-Agent", "EvenCord")
	res, e := Client().Do(req)
	if e != nil {
		return "", e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("artifact unavailable: HTTP %d", res.StatusCode)
	}
	if res.ContentLength >= 0 && res.ContentLength != v.Manifest.Artifact.Size {
		return "", errors.New("artifact Content-Length mismatch")
	}
	return StageReader(root, v, res.Body)
}
