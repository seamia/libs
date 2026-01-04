// Copyright 2026 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"errors"
	"os"
	"path/filepath"
)

func Find(filename string, candidates []string) (string, error) {
	if len(candidates) == 0 {
		candidates = append(candidates, "./")
		if executable, err := os.Executable(); err == nil {
			candidates = append(candidates, executable)
		}
		if configDir, err := os.UserConfigDir(); err == nil {
			candidates = append(candidates, configDir)
		}

		if homeDir, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, homeDir)
		}
	}

	for _, candidate := range candidates {
		fullPath := filepath.Join(candidate, filename)
		if fileExists(fullPath) {
			return fullPath, nil
		}
	}

	return "", errors.New("file not found")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
