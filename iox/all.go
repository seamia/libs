// Copyright 2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"os"
	"io"
	"strings"
	"fmt"
	"sync"
)

func ReadAll(from io.Reader, buffer []byte) error {
	trace("reading %v bytes...", len(buffer))
	for len(buffer) > 0 {
		bytesRead, err := from.Read(buffer)
		if err != nil {
			if err != io.EOF {
				trace("read error: %v", err)
				return err
			}
		}
		// trace("    read %v bytes...", bytesRead)
		buffer = buffer[bytesRead:]
	}
	return nil
}

func WriteAll(to io.Writer, buffer []byte) error {
	trace("writing %v bytes...", len(buffer))
	for len(buffer) > 0 {
		bytesWritten, err := to.Write(buffer)
		if err != nil {
			trace("writeAll error: %v", err)
			return err
		}
		trace("    wrote %v bytes...", bytesWritten)
		buffer = buffer[bytesWritten:]
	}
	return nil
}
