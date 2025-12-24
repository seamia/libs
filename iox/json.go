// Copyright 2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"encoding/json"
	"fmt"
	"os"
)

func JsonSaveMarshal(filename string, mapping any) error {
	raw, err := json.MarshalIndent(mapping, "", "\t")
	if err != nil {
		// we can't marshal this, let's see if we can get anything out of it:
		if stringer, ok := mapping.(fmt.Stringer); ok {
			raw = []byte(stringer.String())
		} else {
			onError("failed to marshal [%s], due to: %v", filename, err)
			return err
		}
	}
	if err := os.WriteFile(filename, raw, 0644); err != nil {
		onError("failed to write file [%s], due to: %v", filename, err)
		return err
	}

	return nil
}

func JsonLoadUnmarshal(filename string, mapping any) error {
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
