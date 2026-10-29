// SPDX-License-Identifier: GPL-3.0
package release

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"
)

func StageSelfUpdate(directory string, envelope []byte, trust Trust, component, current string, reader io.Reader) (string, error) {
	if component != "installer-cli" && component != "installer-gui" {
		return "", errors.New("installer component required")
	}
	v, err := Verify(envelope, trust, component, time.Now(), true)
	if err != nil {
		return "", err
	}
	newer, err := Compare(v.Manifest.Version, current)
	if err != nil || newer <= 0 {
		return "", errors.New("installer release is not newer")
	}
	candidate, err := StageReader(directory, v, reader)
	if err != nil {
		return "", err
	}
	proof := filepath.Join(directory, filepath.Base(candidate)+".release.json")
	file, err := os.CreateTemp(directory, ".evencord-proof-")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err = file.Write(envelope); err != nil {
		return "", err
	}
	if err = file.Sync(); err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	if err = os.Rename(file.Name(), proof); err != nil {
		return "", err
	}
	return candidate, nil
}
