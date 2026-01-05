// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libs

type LogFuncType func(format string, args ...interface{})

var (
	Info    = empty
	Trace   = empty
	Warning = empty
	Failure = empty
	Alarm   = empty
)

func empty(format string, args ...interface{}) {}
