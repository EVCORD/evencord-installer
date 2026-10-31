/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"equilotl/internal/install"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	path "path/filepath"
	"regexp"
	"strings"
	"time"
)

type GithubRelease struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
		Size        int64  `json:"size"`
		Digest      string `json:"digest"`
	} `json:"assets"`
}

var ReleaseData GithubRelease
var GithubError error
var GithubDoneChan chan bool

var InstalledHash = "None"
var LatestHash = "Unknown"
var IsDevInstall bool

func GetGithubRelease(url, fallbackUrl string) (*GithubRelease, error) {
	Log.Debug("Fetching", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Error("Failed to create Request", err)
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Log.Error("Failed to send Request", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 300 {
		isRateLimitedOrBlocked := res.StatusCode == 401 || res.StatusCode == 403 || res.StatusCode == 429
		triedFallback := url == fallbackUrl

		if isRateLimitedOrBlocked && !triedFallback {
			Log.Error(fmt.Sprintf("Failed to fetch %s (status code %d). Trying fallback url %s", url, res.StatusCode, fallbackUrl))
			return GetGithubRelease(fallbackUrl, fallbackUrl)
		}

		err = errors.New(res.Status)
		Log.Error(url, "returned Non-OK status", GithubError)
		return nil, err
	}

	var data GithubRelease

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		Log.Error("Failed to decode GitHub JSON Response", err)
		return nil, err
	}

	return &data, nil
}

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("EVENCORD_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install: ", IsDevInstall)
	if IsDevInstall {
		GithubDoneChan <- true
		return
	}

	go func() {
		// Make sure UI updates once the request either finished or failed
		defer func() {
			GithubDoneChan <- GithubError == nil
		}()

		data, err := GetGithubRelease(ReleaseUrl, ReleaseUrlFallback)
		if err != nil {
			GithubError = err
			return
		}

		ReleaseData = *data

		i := strings.LastIndex(data.Name, " ") + 1
		LatestHash = data.Name[i:]
		Log.Debug("Finished fetching GitHub Data")
		Log.Debug("Latest hash is", LatestHash, "Local Install is", Ternary(LatestHash == InstalledHash, "up to date!", "outdated!"))
	}()

	// either .asar file or directory with main.js file (in DEV)
	EquicordFile := EquicordDirectory

	stat, err := os.Stat(EquicordFile)
	if err != nil {
		return
	}

	// dev
	if stat.IsDir() {
		EquicordFile = path.Join(EquicordFile, "main.js")
	}

	// Check hash of installed version if exists
	b, err := os.ReadFile(EquicordFile)
	if err != nil {
		return
	}

	Log.Debug("Found existing EvenCord Install. Checking for hash...")

	re := regexp.MustCompile(`// EvenCord (\w+)`)
	match := re.FindSubmatch(b)
	if match != nil {
		InstalledHash = string(match[1])
		Log.Debug("Existing hash is", InstalledHash)

	} else {
		Log.Debug("Didn't find hash")

	}
}

func installLatestBuilds() error {
	if IsDevInstall {
		return nil
	}
	for _, asset := range ReleaseData.Assets {
		if asset.Name != "desktop.asar" {
			continue
		}
		if asset.Size <= 0 || asset.Size > 134217728 || len(asset.Digest) != 71 || !strings.HasPrefix(asset.Digest, "sha256:") {
			return errors.New("release payload digest is unavailable")
		}
		expected := "https://github.com/alrimii/evencord/releases/download/" + ReleaseData.TagName + "/desktop.asar"
		if asset.DownloadURL != expected {
			return errors.New("unexpected release payload URL")
		}
		client := &http.Client{Timeout: 2 * time.Minute}
		res, err := client.Get(asset.DownloadURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed: %s", res.Status)
		}
		data, err := io.ReadAll(io.LimitReader(res.Body, asset.Size+1))
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		if int64(len(data)) != asset.Size || "sha256:"+hex.EncodeToString(digest[:]) != asset.Digest {
			return errors.New("release payload size or digest mismatch")
		}
		if err = install.SafePath(EquicordDirectory); err != nil {
			return err
		}
		if err = install.Atomic(EquicordDirectory, data); err != nil {
			return err
		}
		_ = FixOwnership(EquicordDirectory)
		InstalledHash = LatestHash
		return nil
	}
	return errors.New("the release has no desktop payload")
}
