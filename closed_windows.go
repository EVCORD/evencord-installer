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
		process, openErr := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, entry.ProcessID)
		if openErr == windows.ERROR_INVALID_PARAMETER {
			continue
		}
		if openErr != nil {
			return fmt.Errorf("close %s, including its tray icon, then retry; its running process could not be inspected", windowsNames[di.branch])
		}
		if status, waitErr := windows.WaitForSingleObject(process, 0); waitErr == nil && status == windows.WAIT_OBJECT_0 {
			windows.CloseHandle(process)
			continue
		}
		buffer := make([]uint16, 32768)
		size := uint32(len(buffer))
		queryErr := windows.QueryFullProcessImageName(process, 0, &buffer[0], &size)
		if queryErr != nil {
			windows.CloseHandle(process)
			return fmt.Errorf("close %s, including its tray icon, then retry; its running process could not be inspected", windowsNames[di.branch])
		}
		image := strings.TrimPrefix(windows.UTF16ToString(buffer[:size]), `\\?\`)
		if strings.HasPrefix(strings.ToLower(filepath.Clean(image)), strings.ToLower(filepath.Clean(root)+string(filepath.Separator))) {
			result, waitErr := windows.WaitForSingleObject(process, 0)
			if waitErr == nil && result != windows.WAIT_OBJECT_0 {
				if stopErr := windows.TerminateProcess(process, 0); stopErr != nil {
					windows.CloseHandle(process)
					return fmt.Errorf("could not close %s: %w. Quit it from the tray and retry", name, stopErr)
				}
				result, waitErr = windows.WaitForSingleObject(process, 5000)
				if waitErr != nil || result != windows.WAIT_OBJECT_0 {
					windows.CloseHandle(process)
					return fmt.Errorf("%s has not exited yet. Wait a moment, then retry", name)
				}
			}
			if waitErr != nil {
				windows.CloseHandle(process)
				return waitErr
			}
		}
		windows.CloseHandle(process)
	}
	if err != windows.ERROR_NO_MORE_FILES {
		return fmt.Errorf("cannot finish checking Discord processes: %w", err)
	}
	time.Sleep(150 * time.Millisecond)
	return nil
}
