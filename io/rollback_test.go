// Copyright 2020 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestRollBack(t *testing.T) {

	payload := "A cookie is a small piece of text that allows a website to recognize your device and maintain a consistent, cohesive experience throughout multiple sessions. If you use the Stack Overflow Network, both Stack Overflow and third parties will use cookies to track and monitor some of your activities on and off the Stack Overflow Network, and store and access some data about you, your browsing history, and your usage of the Stack Overflow Network."
	src := []byte("[446|some irrelevan stuff here]" + payload)

	r := bytes.NewReader(src)
	var from io.Reader = r
	rb := NewRollbackReader(from)

	buffer := make([]byte, 80)
	read, err := rb.Read(buffer)
	if err != nil {
		t.Fatalf("failed to read (%v)", err)
	} else if read < len(buffer) {
		buffer = buffer[:read]
	}

	left := bytes.Index(buffer, []byte("["))
	right := bytes.Index(buffer, []byte("]"))

	if left < 0 || right < left {
		t.Fatalf("missing reqired markers")
	}

	header := string(buffer[left+1 : right])
	rollback := read - (right + 1)

	parts := strings.Split(header, "|")
	size, err := strconv.Atoi(parts[0])
	if err != nil {
		t.Fatalf("failed to convert (%s) to int (%v)", parts[0], err)
	}

	if size < 0 {
		t.Fatalf("size is negative")
	}

	if err := rb.Rollback(rollback); err != nil {
		t.Fatalf("failed to rollback (%v)", err)
	}

	text := make([]byte, size)
	textread, err := rb.Read(text)
	if err != nil {
		t.Fatalf("failed to read payload (%v)", err)
	} else if textread != size {
		t.Fatalf("failed to read full payload (%d instead of %d)", textread, size)
	}

	if string(text) != payload {
		t.Fatalf("mismatch in received data. got [%s], expected [%s]", string(text), payload)
	}
}
