// Copyright 2019 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package resolve

import (
	"os"
	"syscall"

	"github.com/seamia/libs/printer"
)

const (
	mappingFailureFormat = "${%s}"
)

type (
	m2s = map[string]string

	Filter func(string) (bool, string)

	Resolver interface {
		Text(from string) string
		Bytes(from []byte) []byte
		Add(key, value string)
		SetFilter(flt Filter, pre bool)
	}

	resolver struct {
		Resolver
		mapping    m2s
		preFilter  Filter
		postFilter Filter
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

func (self *resolver) SetFilter(flt Filter, pre bool) {
	if self == nil {
		self = defaultResolver
	}
	if flt == nil {
		flt = emptyFilter
	}
	if pre {
		self.preFilter = flt
	} else {
		self.postFilter = flt
	}
}

func (self *resolver) mappingFunc(from string) string {
	// first: run pre-filter
	if success, resolution := self.preFilter(from); success {
		return resolution
	}
	// second look into "custom"
	if value, exists := self.mapping[from]; exists {
		return value
	}

	// then, check the environment
	if value, exists := syscall.Getenv(from); exists {
		return value
	}

	// last: run post-filter
	if success, resolution := self.postFilter(from); success {
		return resolution
	}

	printer.Print("failed to resolve [%s]", from)
	return mappingFailure(from)
}

func mappingFailure(from string) string {
	return "" // return fmt.Sprintf(mappingFailureFormat, from)
}

func New() Resolver {
	return &resolver{
		mapping:    make(m2s),
		preFilter:  emptyFilter,
		postFilter: emptyFilter,
	}
}

func new() *resolver {
	return New().(*resolver)
}

func emptyFilter(src string) (bool, string) {
	return false, src
}
