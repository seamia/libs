// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func jsonLoadUnmarshal(filename string, mapping any) error {
	raw, err := os.ReadFile(filename)
	if err != nil {
		onError("failed to open file: %v", err)
		return err
	}

	if err := json.Unmarshal(raw, mapping); err != nil {
		return err
	}
	return nil
}

func onError(format string, args ...any) {
	format = strings.TrimSuffix(format, "\n") + "\n"
	fmt.Fprintf(os.Stderr, format, args...)
}

func debugPrintf(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}
