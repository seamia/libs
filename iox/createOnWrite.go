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

type streamState byte

const (
	streamNotOpened streamState = iota
	streamOpened
	streamClosed
)


type createOnWriteWriter struct {
	name    string
	handler *os.File
	guard   sync.Mutex
	state 	streamState
}

func (w *createOnWriteWriter) Close() error {
	w.guard.Lock()
	defer w.guard.Unlock()

	if w.handler != nil {
		err := w.handler.Close()
		w.state = streamClosed
		w.handler = nil
		return err
	}
	return nil
}

func (w *createOnWriteWriter) Write(p []byte) (n int, err error) {
	w.guard.Lock()
	defer w.guard.Unlock()

	switch w.state {
	case streamNotOpened:
		handler, err := os.Create(w.name)
		if err != nil {
			onError("failed to create file [%s], due to: %v", w.name, err)
			return 0, err
		}
		w.handler = handler
		w.state = streamOpened
		fallthrough

	case streamOpened:
		return w.handler.Write(p)

	case streamClosed:
		return 0, fmt.Errorf("", w.name)
	}
	panic("unreachable")
}


func CreateWriter(fileName string) io.WriteCloser {
	return &createOnWriteWriter{
		name: fileName,
		state: streamNotOpened,
	}
}

func onError(format string, args ...any) {
	format = strings.TrimSuffix(format, "\n") + "\n"
	fmt.Fprintf(os.Stderr, format, args...)
}

type dummyWriter struct {
}

func (d *dummyWriter) Close() error {
	return nil
}

func (d *dummyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func CreateDummyWriter() io.WriteCloser {
	return &dummyWriter{}
}
