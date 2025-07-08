// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import "sync"

var (
	knownConnetions      map[string]*Connection
	knownConnetionsGuard sync.Mutex
)
