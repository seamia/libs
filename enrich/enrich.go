// Copyright 2020 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package enrich

import (
	"github.com/sirupsen/logrus"
)

type (
	Fields = logrus.Fields
	Entry  = logrus.Entry

	Enrichment interface {
		Enrich(entry *Entry) *Entry
	}
)

func Enrich(what *Entry, with interface{}) *Entry {
	if what == nil || with == nil {
		// we don't have enough data to do the enrichment
		return what
	}

	if rich, converts := with.(Enrichment); converts && rich != nil {
		return rich.Enrich(what)
	}
	return what
}

/*
	Why:
		by implementing the Enrichment interface on your 'data' and subsequently calling Enrich function you can
		streamline augmenting your log entries with relevant 'fields' of your data structures.

*/
