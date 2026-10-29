/*
 * SPDX-License-Identifier: GPL-3.0
 * Vencord Installer, a cross platform gui/cli app for installing Vencord
 * Copyright (c) 2023 Vendicated and Vencord contributors
 */

package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	path "path/filepath"
)

func FindAsarFile(dir string) (*os.File, error) {
	for _, file := range []string{"_app.asar", "app.asar"} {
		f, err := os.Open(path.Join(dir, file))
		if err != nil {
			continue
		}
		stats, err := f.Stat()
		if err == nil && !stats.IsDir() {
			return f, nil
		}
		_ = f.Close()
	}
	return nil, errors.New("Install at " + dir + " has no asar file")
}

func (di *DiscordInstall) IsOpenAsar() (retBool bool) {
	if di.isOpenAsar != nil {
		return *di.isOpenAsar
	}

	defer func() {
		Log.Debug("Checking if", di.path, "is using OpenAsar:", retBool)
		di.isOpenAsar = &retBool
	}()

	asarFile, err := FindAsarFile(path.Join(di.appPath, ".."))
	if err != nil {
		Log.Error(err.Error())
		return false
	}

	b, err := io.ReadAll(asarFile)
	_ = asarFile.Close()
	if err != nil {
		Log.Error(err.Error())
		return false
	}

	if bytes.Contains(b, []byte("OpenAsar")) {
		return true
	}

	return false
}

func (di *DiscordInstall) InstallOpenAsar() error {
	return errors.New("OpenAsar modification is unsupported by EvenCord.")
}

func (di *DiscordInstall) UninstallOpenAsar() error {
	return errors.New("OpenAsar modification is unsupported by EvenCord.")
}
