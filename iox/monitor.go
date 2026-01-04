// Copyright 2026 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"io"
	"time"
)

type monitorWriter struct {
	total               int64
	written             int64
	lastShownPercentage int
	startedAt           time.Time
	writer              io.WriteCloser
	monitor             Monitor
}

func (w *monitorWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)

	w.written += int64(n)
	if w.total != 0 {
		percent := int(float64(w.written) * 100 / float64(w.total))
		if percent > w.lastShownPercentage {
			w.lastShownPercentage = percent
			took := time.Since(w.startedAt)
			w.monitor(w.written, percent, took)
		}
	}

	return n, err
}

func (w *monitorWriter) Close() error {
	return w.writer.Close()
}

type Monitor func(bytes int64, percent int, duration time.Duration)

func GetMonitorWriter(base io.WriteCloser, total int64, monitor Monitor) io.WriteCloser {
	return &monitorWriter{
		writer:              base,
		total:               total,
		lastShownPercentage: -1,
		startedAt:           time.Now(),
		monitor:             monitor,
	}
}
