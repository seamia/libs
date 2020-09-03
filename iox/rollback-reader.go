// Copyright 2020 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
)

var (
	errNil          = errors.New("nil")
	errNegative     = errors.New("negative rollback value")
	errInsufficient = errors.New("rollback is too big")
	errImpossible   = errors.New("inconceivable!")

	success error = nil
)

// RollbackReader is a helper interface that allows a limited "rollback" on onethewise one-way stream
type RollbackReader interface {
	io.Reader
	Rollback(int) error
}

type rollbackReader struct {
	reader   io.Reader
	lastRead []byte
	rollback int
}

func NewRollbackReader(from io.Reader) RollbackReader {
	return &rollbackReader{
		reader: from,
	}
}

const (
	maxRollBackSize = 1024
)

var (
	report = log.Infof
)

func minimumOf(one, two int) int {
	if one < two {
		return one
	} else {
		return two
	}
}

// io.Reader
func (r *rollbackReader) Read(receiver []byte) (n int, err error) {
	if r == nil || r.reader == nil || receiver == nil {
		return 0, errNil
	}

	if r.rollback == 0 {
		read, err := r.read(receiver)
		if err == nil && read > 0 {
			preserve := minimumOf(read, maxRollBackSize)
			r.lastRead = make([]byte, preserve)
			if saved := copy(r.lastRead, receiver[read-preserve:]); saved != preserve {
				report("failed to copy (%d; %d)", saved, preserve)
				return 0, errImpossible
			}
		}
		return read, err
	}

	if r.rollback > len(r.lastRead) {
		// aka: cannot roll back more than we have in store
		// this should not happen due to check inside Rollback func below
		report("rollback is too big (%d; %d)", r.rollback, len(r.lastRead))
		return 0, errImpossible
	} else if len(receiver) <= r.rollback {
		// all the data we need is inside `lastRead` buffer
		read := copy(receiver, r.lastRead[len(r.lastRead)-r.rollback:])
		r.rollback -= read

		return read, success
	} else {
		// we need more than we hold inside `lastRead` buffer
		// #1. copy all from `lastRead` buffer
		copied := copy(receiver, r.lastRead[len(r.lastRead)-r.rollback:])
		if copied != r.rollback {
			report("failed to copy #2 (%d; %d)", copied, r.rollback)
			return 0, errImpossible
		}
		r.lastRead = r.lastRead[:0]
		r.rollback -= copied

		// #2. append the remaining data from the 'real' stream
		read, err := r.read(receiver[copied:])
		if err != nil {
			// eof?
			return read, err
		}

		// #3. save read data?
		preserve := minimumOf(read, maxRollBackSize)
		r.lastRead = make([]byte, preserve)
		if saved := copy(r.lastRead, receiver[copied+read-preserve:]); saved != read {
			report("failed to copy #3 (%d; %d)", saved, read)
			return 0, errImpossible
		}
		return copied + read, success
	}
}

func (r *rollbackReader) Rollback(back int) error {
	if r == nil {
		return errNil
	}

	if back < 0 {
		return errNegative
	}
	if len(r.lastRead) < (r.rollback + back) {
		return errInsufficient
	}

	r.rollback += back

	return success
}

func (r *rollbackReader) read(receiver []byte) (n int, err error) {
	if r == nil || r.reader == nil || receiver == nil {
		return 0, errNil
	}

	return io.ReadFull(r.reader, receiver)
}
