// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package install

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"equilotl/internal/release"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Target struct {
	Resources string
	Channel   string
}
type Owner struct {
	Schema   int               `json:"schemaVersion"`
	Product  string            `json:"product"`
	DataRoot string            `json:"dataRoot"`
	Original string            `json:"originalSHA256"`
	Files    map[string]string `json:"files"`
}
type Journal struct {
	Operation string `json:"operation"`
	Owner     Owner  `json:"owner"`
}
type Engine struct {
	Trust                      release.Trust
	TrustData, Runtime, Loader []byte
	AfterStep                  func(string) error
	ChannelArchives            map[string][]string
}

const ownerName = "evencord-install.json"
const journalName = ".evencord-transaction.json"

var appVersion = regexp.MustCompile(`^app-[0-9]+\.[0-9]+\.[0-9]+$`)

func exists(p string) bool   { _, e := os.Lstat(p); return e == nil }
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func writeJSON(p string, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	return Atomic(p, b)
}
func (e Engine) step(name string) error {
	if e.AfterStep != nil {
		return e.AfterStep(name)
	}
	return nil
}
func Validate(t Target) error {
	if (t.Channel != "stable" && t.Channel != "ptb" && t.Channel != "canary") || filepath.Base(t.Resources) != "resources" || !appVersion.MatchString(filepath.Base(filepath.Dir(t.Resources))) {
		return errors.New("unsupported Discord channel or installation layout")
	}
	if err := SafePath(t.Resources); err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(t.Resources, "build_info.json"))
	if err != nil {
		return err
	}
	var info struct {
		Channel string `json:"releaseChannel"`
	}
	if json.Unmarshal(b, &info) != nil || info.Channel != t.Channel {
		return errors.New("Discord build metadata must match the selected channel")
	}
	exe, err := os.ReadFile(filepath.Join(t.Resources, "..", map[string]string{"stable": "Discord.exe", "ptb": "DiscordPTB.exe", "canary": "DiscordCanary.exe"}[t.Channel]))
	if err != nil {
		return err
	}
	if len(exe) < 64 || string(exe[:2]) != "MZ" {
		return errors.New("invalid Discord executable")
	}
	off := int(binary.LittleEndian.Uint32(exe[60:64]))
	if off < 64 || off > len(exe)-6 || string(exe[off:off+4]) != "PE\x00\x00" || binary.LittleEndian.Uint16(exe[off+4:off+6]) != 0x8664 {
		return errors.New("Windows x64 Discord is required")
	}
	if exists(filepath.Join(t.Resources, "app")) {
		return errors.New("another app loader directory is present")
	}
	return nil
}
func readOwner(p string) (Owner, error) {
	var o Owner
	b, err := os.ReadFile(p)
	if err != nil {
		return o, err
	}
	err = release.Decode(b, &o)
	if err == nil && (o.Schema != 1 || o.Product != "evencord" || !filepath.IsAbs(o.DataRoot) || len(o.Original) != 64) {
		err = errors.New("invalid EvenCord ownership record")
	}
	return o, err
}
func (e Engine) expected(o Owner) map[string][]byte {
	b, _ := json.Marshal(o)
	return map[string][]byte{"package.json": []byte("{\"name\":\"discord\",\"main\":\"index.js\"}"), "index.js": e.Loader, "runtime.cjs": e.Runtime, "trust.json": e.TrustData, "evencord-owner.json": b}
}
func (e Engine) validLoader(dir string, o Owner) error {
	if len(o.Files) != 3 || o.Files["index.js"] != digest(e.Loader) || o.Files["runtime.cjs"] != digest(e.Runtime) || o.Files["trust.json"] != digest(e.TrustData) {
		return errors.New("loader identity does not match this installer")
	}
	if err := SafePath(dir); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	files := e.expected(o)
	if len(entries) != len(files) {
		return errors.New("unknown files in loader")
	}
	for name, want := range files {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || !bytes.Equal(b, want) {
			return errors.New("loader contents changed: " + name)
		}
	}
	return nil
}
func foreignLabel(app string) string {
	s, err := os.Lstat(app)
	if err == nil && s.IsDir() {
		app = filepath.Join(app, "index.js")
	}
	f, err := os.Open(app)
	if err != nil {
		return "Another modification detected (unknown)"
	}
	defer f.Close()
	b, _ := io.ReadAll(io.LimitReader(f, 128<<10))
	text := strings.ToLower(string(b))
	for _, name := range []string{"Equicord", "Vencord", "OpenAsar"} {
		if strings.Contains(text, strings.ToLower(name)) {
			return "Another modification detected (" + name + ")"
		}
	}
	return "Another modification detected (unknown)"
}
func (e Engine) Inspect(t Target) (string, error) {
	if err := Validate(t); err != nil {
		return "Unsupported installation", err
	}
	if exists(filepath.Join(t.Resources, journalName)) {
		return "Recovery available", nil
	}
	if exists(filepath.Join(t.Resources, ownerName)) {
		o, err := readOwner(filepath.Join(t.Resources, ownerName))
		if err != nil {
			return "Damaged loader", err
		}
		if err = e.validLoader(filepath.Join(t.Resources, "app.asar"), o); err != nil {
			return "Damaged loader", err
		}
		h, err := hashFile(filepath.Join(t.Resources, "_app.asar"))
		if err != nil || h != o.Original {
			return "Recovery archive missing or changed", errors.New("original archive cannot be verified")
		}
		_, err = e.Current(o.DataRoot)
		if err != nil {
			return "Repair available", nil
		}
		return "Installed", nil
	}
	if exists(filepath.Join(t.Resources, "_app.asar")) {
		return foreignLabel(filepath.Join(t.Resources, "app.asar")), nil
	}
	s, err := os.Lstat(filepath.Join(t.Resources, "app.asar"))
	if err != nil {
		return "Unsupported installation", err
	}
	if !s.Mode().IsRegular() {
		return foreignLabel(filepath.Join(t.Resources, "app.asar")), nil
	}
	return "Not installed", nil
}
func (e Engine) Current(root string) (release.Verified, error) {
	b, err := os.ReadFile(filepath.Join(root, "current.json"))
	if err != nil {
		return release.Verified{}, err
	}
	v, err := release.Verify(b, e.Trust, "core", time.Now(), false)
	if err != nil {
		return v, err
	}
	return v, release.CheckFile(filepath.Join(root, "releases", v.Manifest.Artifact.SHA256+"-"+v.Manifest.Artifact.Name), v.Manifest.Artifact)
}
func (e Engine) Activate(root string, v release.Verified, staged string) error {
	v, err := release.Verify(v.Envelope, e.Trust, "core", time.Now(), true)
	if err != nil {
		return err
	}
	if err = SafePath(root); err != nil {
		return err
	}
	if err = release.CheckFile(staged, v.Manifest.Artifact); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(root, "releases"), 0700); err != nil {
		return err
	}
	unlock, err := lock(root)
	if err != nil {
		return err
	}
	defer unlock()
	failure := struct{ SHA256 string }{}
	if b, readErr := os.ReadFile(filepath.Join(root, "failed.json")); readErr == nil {
		if json.Unmarshal(b, &failure) != nil {
			return errors.New("damaged startup recovery marker")
		}
	}
	if failure.SHA256 == v.Manifest.Artifact.SHA256 {
		return errors.New("release failed startup; use a newer release or reviewed recovery")
	}
	accepted := filepath.Join(root, "accepted.json")
	if data, readErr := os.ReadFile(accepted); readErr == nil {
		highest, err := release.Verify(data, e.Trust, "core", time.Now(), false)
		if err != nil {
			return err
		}
		if err = release.Advance(v.Manifest, highest.Manifest); err != nil {
			return err
		}
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	pointer := filepath.Join(root, "current.json")
	if exists(pointer) {
		oldData, err := os.ReadFile(pointer)
		if err != nil {
			return err
		}
		old, err := release.Verify(oldData, e.Trust, "core", time.Now(), false)
		if err != nil {
			return err
		}
		if err = release.Advance(v.Manifest, old.Manifest); err != nil {
			return err
		}
		if old.Manifest.Artifact.SHA256 != v.Manifest.Artifact.SHA256 && old.Manifest.Artifact.SHA256 != failure.SHA256 {
			if _, err = e.Current(root); err != nil {
				return fmt.Errorf("current build is damaged; repair it before upgrading: %w", err)
			}
			if err = Atomic(filepath.Join(root, "previous.json"), oldData); err != nil {
				return err
			}
		}
	}
	data, err := os.ReadFile(staged)
	if err != nil {
		return err
	}
	if int64(len(data)) != v.Manifest.Artifact.Size || digest(data) != v.Manifest.Artifact.SHA256 {
		return errors.New("staged payload changed before activation")
	}
	target := filepath.Join(root, "releases", v.Manifest.Artifact.SHA256+"-"+v.Manifest.Artifact.Name)
	if exists(target) {
		if err = release.CheckFile(target, v.Manifest.Artifact); err != nil {
			quarantine := target + ".damaged-" + fmt.Sprint(time.Now().UnixNano())
			if err = os.Rename(target, quarantine); err != nil {
				return err
			}
			if err = Atomic(target, data); err != nil {
				os.Rename(quarantine, target)
				return err
			}
		}
	} else if err = Atomic(target, data); err != nil {
		return err
	}
	if err = e.step("payload-staged"); err != nil {
		return err
	}
	if err = Atomic(accepted, v.Envelope); err != nil {
		return err
	}
	return Atomic(pointer, v.Envelope)
}
func (e Engine) hostAllowed(channel, hash string) bool {
	if e.ChannelArchives == nil {
		return channel == "stable"
	}
	for _, approved := range e.ChannelArchives[channel] {
		if hash == approved {
			return true
		}
	}
	return false
}
func DataRoot(base, channel string) string {
	if channel == "stable" {
		return base
	}
	if channel == "ptb" || channel == "canary" {
		return filepath.Join(base, "channels", channel)
	}
	panic("unsupported Discord channel")
}
func (e Engine) Install(t Target, root string, v release.Verified, staged string) error {
	if err := Validate(t); err != nil {
		return err
	}
	unlock, err := lock(t.Resources)
	if err != nil {
		return err
	}
	defer unlock()
	if exists(filepath.Join(t.Resources, journalName)) {
		return errors.New("unfinished transaction; recover before installing")
	}
	v, err = release.Verify(v.Envelope, e.Trust, "core", time.Now(), true)
	if err != nil {
		return err
	}
	app, backup := filepath.Join(t.Resources, "app.asar"), filepath.Join(t.Resources, "_app.asar")
	record := filepath.Join(t.Resources, ownerName)
	if exists(record) {
		o, err := readOwner(record)
		if err != nil {
			return err
		}
		if err = e.validLoader(app, o); err != nil {
			return err
		}
		h, err := hashFile(backup)
		if err != nil || h != o.Original {
			return errors.New("original archive missing or changed")
		}
		compatible := false
		for _, approved := range v.Manifest.Archives {
			if h == approved && e.hostAllowed(t.Channel, h) {
				compatible = true
			}
		}
		if !compatible {
			return errors.New("release does not support the installed Discord archive")
		}
		if filepath.Clean(o.DataRoot) != filepath.Clean(root) {
			return errors.New("installed data root differs")
		}
		return e.Activate(root, v, staged)
	}
	if exists(backup) {
		return errors.New("another modification owns the recovery archive")
	}
	original, err := hashFile(app)
	if err != nil {
		return err
	}
	approved := false
	for _, h := range v.Manifest.Archives {
		if h == original && e.hostAllowed(t.Channel, h) {
			approved = true
		}
	}
	if !approved {
		return errors.New("unrecognized Discord archive; refusing to overwrite a possible modification")
	}
	if err = checkRenameAccess(app); err != nil {
		return err
	}
	if err = e.Activate(root, v, staged); err != nil {
		return err
	}
	o := Owner{Schema: 1, Product: "evencord", DataRoot: root, Original: original, Files: map[string]string{"index.js": digest(e.Loader), "runtime.cjs": digest(e.Runtime), "trust.json": digest(e.TrustData)}}
	stage := filepath.Join(t.Resources, ".evencord-loader")
	if exists(stage) {
		return errors.New("staged loader exists; review recovery before retrying")
	}
	if err = os.Mkdir(stage, 0700); err != nil {
		return err
	}
	for name, data := range e.expected(o) {
		if err = Atomic(filepath.Join(stage, name), data); err != nil {
			return err
		}
	}
	if err = writeJSON(filepath.Join(t.Resources, journalName), Journal{"install", o}); err != nil {
		return err
	}
	if err = os.Rename(app, backup); err != nil {
		return err
	}
	if err = e.step("original-preserved"); err != nil {
		return err
	}
	if err = os.Rename(stage, app); err != nil {
		return err
	}
	if err = e.step("loader-activated"); err != nil {
		return err
	}
	if err = writeJSON(record, o); err != nil {
		return err
	}
	return os.Remove(filepath.Join(t.Resources, journalName))
}
func (e Engine) Uninstall(t Target) error {
	if err := Validate(t); err != nil {
		return err
	}
	unlock, err := lock(t.Resources)
	if err != nil {
		return err
	}
	defer unlock()
	if exists(filepath.Join(t.Resources, journalName)) {
		return errors.New("unfinished transaction; recover first")
	}
	o, err := readOwner(filepath.Join(t.Resources, ownerName))
	if err != nil {
		return err
	}
	if err = e.validLoader(filepath.Join(t.Resources, "app.asar"), o); err != nil {
		return err
	}
	h, err := hashFile(filepath.Join(t.Resources, "_app.asar"))
	if err != nil || h != o.Original {
		return errors.New("original archive missing or changed")
	}
	if exists(filepath.Join(t.Resources, ".evencord-removed")) {
		return errors.New("prior removal needs recovery")
	}
	if err = writeJSON(filepath.Join(t.Resources, journalName), Journal{"uninstall", o}); err != nil {
		return err
	}
	return e.recover(t)
}
func (e Engine) Recover(t Target) error {
	if err := Validate(t); err != nil {
		return err
	}
	unlock, err := lock(t.Resources)
	if err != nil {
		return err
	}
	defer unlock()
	if !exists(filepath.Join(t.Resources, journalName)) {
		o, err := readOwner(filepath.Join(t.Resources, ownerName))
		if err != nil {
			return err
		}
		if err = e.validLoader(filepath.Join(t.Resources, "app.asar"), o); err != nil {
			return err
		}
		h, err := hashFile(filepath.Join(t.Resources, "_app.asar"))
		if err != nil || h != o.Original {
			return errors.New("recovery archive missing or changed")
		}
		return e.RecoverPayload(o.DataRoot)
	}
	return e.recover(t)
}
func (e Engine) recover(t Target) error {
	jp := filepath.Join(t.Resources, journalName)
	data, err := os.ReadFile(jp)
	if err != nil {
		return err
	}
	var j Journal
	if err = release.Decode(data, &j); err != nil {
		return err
	}
	o := j.Owner
	if o.Schema != 1 || o.Product != "evencord" || len(o.Original) != 64 {
		return errors.New("invalid recovery journal")
	}
	app, backup := filepath.Join(t.Resources, "app.asar"), filepath.Join(t.Resources, "_app.asar")
	stage, removed := filepath.Join(t.Resources, ".evencord-loader"), filepath.Join(t.Resources, ".evencord-removed")
	if j.Operation == "install" {
		if e.validLoader(app, o) == nil {
			h, err := hashFile(backup)
			if err != nil || h != o.Original {
				return errors.New("cannot finalize without original archive")
			}
			if err = writeJSON(filepath.Join(t.Resources, ownerName), o); err != nil {
				return err
			}
		} else {
			if exists(app) {
				h, err := hashFile(app)
				if err != nil || h != o.Original {
					return errors.New("unknown loader blocks recovery")
				}
			} else {
				h, err := hashFile(backup)
				if err != nil || h != o.Original {
					return errors.New("recovery archive missing or changed")
				}
				if err = os.Rename(backup, app); err != nil {
					return err
				}
			}
			if exists(stage) {
				if err = e.validLoader(stage, o); err != nil {
					return err
				}
				if err = os.RemoveAll(stage); err != nil {
					return err
				}
			}
		}
	} else if j.Operation == "uninstall" {
		h, _ := hashFile(app)
		if h != o.Original {
			h, err := hashFile(backup)
			if err != nil || h != o.Original {
				return errors.New("recovery archive missing or changed")
			}
			if exists(app) {
				if err = e.validLoader(app, o); err != nil {
					return err
				}
				if exists(removed) {
					return errors.New("ambiguous removal state")
				}
				if err = os.Rename(app, removed); err != nil {
					return err
				}
			}
			if err = e.step("loader-removed"); err != nil {
				return err
			}
			if err = os.Rename(backup, app); err != nil {
				return err
			}
		}
		if exists(removed) {
			if err = e.validLoader(removed, o); err != nil {
				return err
			}
			if err = os.RemoveAll(removed); err != nil {
				return err
			}
		}
		if exists(filepath.Join(t.Resources, ownerName)) {
			if err = os.Remove(filepath.Join(t.Resources, ownerName)); err != nil {
				return err
			}
		}
	} else {
		return errors.New("unknown recovery operation")
	}
	return os.Remove(jp)
}

func (e Engine) RecoverPayload(root string) error {
	if err := SafePath(root); err != nil {
		return err
	}
	unlock, err := lock(root)
	if err != nil {
		return err
	}
	defer unlock()
	if _, err = e.Current(root); err == nil {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(root, "previous.json"))
	if err != nil {
		return errors.New("no verified previous build is available; preserve recovery data and use a reviewed repair release")
	}
	v, err := release.Verify(data, e.Trust, "core", time.Now(), false)
	if err != nil {
		return err
	}
	if err = release.CheckFile(filepath.Join(root, "releases", v.Manifest.Artifact.SHA256+"-"+v.Manifest.Artifact.Name), v.Manifest.Artifact); err != nil {
		return err
	}
	pointer := filepath.Join(root, "current.json")
	if old, readErr := os.ReadFile(pointer); readErr == nil {
		if err = Atomic(pointer+".damaged-"+fmt.Sprint(time.Now().UnixNano()), old); err != nil {
			return err
		}
	}
	return Atomic(pointer, data)
}
