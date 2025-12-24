// Copyright 2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"fmt"
	"os"
	"strings"
)

func onError(format string, args ...any) {
	format = strings.TrimSuffix(format, "\n") + "\n"
	fmt.Fprintf(os.Stderr, format, args...)
}
