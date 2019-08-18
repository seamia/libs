// Copyright 2019 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolve

import (
	"fmt"
	"os"
	"syscall"

	"github.com/seamia/libs/printer"
)

const (
	mappingFailureFormat = "${%s}"
)

type (
	m2s = map[string]string

	Resolver interface {
		Text(from string) string
		Bytes(from []byte) []byte
		Add(key, value string)
	}
	resolver struct {
		Resolver
		mapping m2s
	}
)

var (
	defaultResolver = new()
)

func Text(from string) string {
	return defaultResolver.Text(from)
}

func Bytes(from []byte) []byte {
	return defaultResolver.Bytes(from)
}

func Add(key, value string) {
	defaultResolver.Add(key, value)
}

func (self *resolver) Text(from string) string {
	if self == nil {
		self = defaultResolver
	}
	return os.Expand(from, self.mappingFunc)
}

func (self *resolver) Bytes(from []byte) []byte {
	if self == nil {
		self = defaultResolver
	}
	return []byte(self.Text(string(from)))
}

func (self *resolver) Add(key, value string) {
	if self == nil {
		self = defaultResolver
	}
	if len(key)+len(value) == 0 {
		return
	}

	if len(value) == 0 {
		delete(self.mapping, key)
	} else {
		self.mapping[key] = value
	}
}

func (self *resolver) mappingFunc(from string) string {
	// first look into "custom"
	if value, exists := self.mapping[from]; exists {
		return value
	}

	// then, check the environment
	if value, exists := syscall.Getenv(from); exists {
		return value
	}

	printer.Print("failed to resolve [%s]", from)
	return mappingFailure(from)
}

func mappingFailure(from string) string {
	return fmt.Sprintf(mappingFailureFormat, from)
}

func New() Resolver {
	return &resolver{mapping: make(m2s)}
}

func new() *resolver {
	return New().(*resolver)
}
