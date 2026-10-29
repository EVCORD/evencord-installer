// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"path/filepath"
	"strings"
	"time"
	"unsafe"
)

type processAccess struct {
	open  func(uint32, bool, uint32) (windows.Handle, error)
	image func(windows.Handle) (string, error)
	wait  func(windows.Handle, uint32) (uint32, error)
	stop  func(windows.Handle, uint32) error
	close func(windows.Handle) error
}

func processImage(process windows.Handle) (string, error) {
	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(process, 0, &buffer[0], &size); err != nil {
		return "", err
	}
	return strings.TrimPrefix(windows.UTF16ToString(buffer[:size]), `\\?\`), nil
}

func closeProcessInRoot(pid uint32, root string, api processAccess) error {
	inspect := uint32(windows.PROCESS_QUERY_LIMITED_INFORMATION)
	process, err := api.open(inspect, false, pid)
	if err != nil {
		Log.Debug("Skipping inaccessible or exited Discord process", pid, err)
		return nil
	}
	image, queryErr := api.image(process)
	api.close(process)
	inRoot := func(image string) bool {
		return strings.HasPrefix(strings.ToLower(filepath.Clean(image)), strings.ToLower(filepath.Clean(root)+string(filepath.Separator)))
	}
	if queryErr != nil || !inRoot(image) {
		return nil
	}
	process, err = api.open(inspect|windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		Log.Warn("Could not automatically close Discord process", pid, err, "Locked installation files will prevent changes. Quit Discord from its tray icon if asked.")
		return nil
	}
	defer api.close(process)
	image, queryErr = api.image(process)
	if queryErr != nil || !inRoot(image) {
		return nil
	}
	result, err := api.wait(process, 0)
	if err != nil || result == windows.WAIT_OBJECT_0 {
		return nil
	}
	if err = api.stop(process, 0); err != nil {
		Log.Warn("Could not automatically close Discord process", pid, err)
		return nil
	}
	result, err = api.wait(process, 5000)
	if err != nil || result != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("%s has not exited yet. Quit it from the tray, wait a moment, then retry", filepath.Base(image))
	}
	return nil
}

func (di *DiscordInstall) closeForOperation() error {
	if _, err := di.target(); err != nil {
		return err
	}
	name := windowsNames[di.branch] + ".exe"
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("cannot check running Discord processes: %w", err)
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	root, err := filepath.Abs(di.path)
	if err != nil {
		return err
	}
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if !strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), name) {
			continue
		}
		if closeErr := closeProcessInRoot(entry.ProcessID, root, processAccess{windows.OpenProcess, processImage, windows.WaitForSingleObject, windows.TerminateProcess, windows.CloseHandle}); closeErr != nil {
			return closeErr
		}
	}
	if err != windows.ERROR_NO_MORE_FILES {
		return fmt.Errorf("cannot finish checking Discord processes: %w", err)
	}
	time.Sleep(150 * time.Millisecond)
	return nil
}
