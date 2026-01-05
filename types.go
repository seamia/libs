// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libs

import (
	"time"
)

type (
	Msi           = map[string]interface{}
	WriteDeadline interface {
		SetWriteDeadline(t time.Time) error
	}
	BinaryPacket struct {
		Header Msi
		Blob   []byte
	}
	Tracer func(format string, args ...any)
)
