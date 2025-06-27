// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libs

import (
	"strconv"
)

func GetText(headers Msi, key string) string {
	if headers != nil {
		if entry, found := headers[key]; found {
			if txt, okay := entry.(string); okay {
				return txt
			}
		}
	}
	Trace("failed to find text entry for key %s", key)
	return ""
}

func GetInt(headers Msi, key string, fallback int) int {
	if headers != nil {
		if entry, found := headers[key]; found {
			switch actual := entry.(type) {
			case int:
				return actual
			case string:
				if value, err := strconv.Atoi(actual); err != nil {
					Warning("failed to convert entry (%v) for key (%s) into int", actual, key)
					return fallback
				} else {
					return value
				}
			case float64:
				return int(actual)

			default:
				Warning("unhandled type (%T) for key (%s)", entry, key)
				return fallback
			}
		}
	}

	Trace("failed to find numeric entry for key %s", key)
	return fallback
}
