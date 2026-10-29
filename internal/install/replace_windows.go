// SPDX-License-Identifier: GPL-3.0
package install

import (
	"fmt"
	"golang.org/x/sys/windows"
	"path/filepath"
)

func replaceFile(from, to string) error {
	a, e := windows.UTF16PtrFromString(from)
	if e != nil {
		return e
	}
	b, e := windows.UTF16PtrFromString(to)
	if e != nil {
		return e
	}
	return windows.MoveFileEx(a, b, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

func lock(root string) (func(), error) {
	p := filepath.Join(root, ".evencord-operation.lock")
	if e := SafePath(p); e != nil {
		return nil, e
	}
	name, e := windows.UTF16PtrFromString(p)
	if e != nil {
		return nil, e
	}
	h, e := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE|windows.DELETE, 0, nil, windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_DELETE_ON_CLOSE, 0)
	if e != nil {
		return nil, fmt.Errorf("another operation owns this path or a file is locked: %w", e)
	}
	return func() { windows.CloseHandle(h) }, nil
}
