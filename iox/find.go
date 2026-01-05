// Copyright 2026 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Find given "filename" in one of provided (or constructed) locations,
// return first found
func Find(filename string, candidates []string) (string, error) {
	moduleName := ""
	if len(candidates) == 0 {
		if info, ok := debug.ReadBuildInfo(); ok {
			parts := strings.Split(info.Main.Path, ".")
			moduleName = strings.ToLower(parts[len(parts)-1] + "." + filename)
		}

		if configDir := os.Getenv("CONFIG_DIR"); len(configDir) > 0 {
			candidates = append(candidates, configDir)
		}
		if current, err := os.Getwd(); err == nil {
			candidates = append(candidates, current)
		}
		if executable, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Dir(executable))
		}
		if configDir, err := os.UserConfigDir(); err == nil {
			candidates = append(candidates, configDir)
		}
		if homeDir, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, homeDir)
		}
	}

	if len(moduleName) > 0 {
		for _, candidate := range candidates {
			fullPath := filepath.Join(candidate, moduleName)
			if fileExists(fullPath) {
				return fullPath, nil
			}
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
