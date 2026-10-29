// SPDX-License-Identifier: GPL-3.0
// Copyright (c) 2026 EvenCord contributors.
package install

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows"
)

func checkRenameAccess(path string) error {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(name, windows.DELETE|windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return errors.New("Discord or another process is using app.asar. Fully quit the selected Discord channel, including its tray icon, then retry. The original archive has not been moved")
	}
	if err != nil {
		return fmt.Errorf("cannot prepare the original Discord archive for installation: %w", err)
	}
	return windows.CloseHandle(handle)
}
