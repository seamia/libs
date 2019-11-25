// Copyright 2019 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"sort"
)

type Slice []string

func (src Slice) Distinct() Slice {
	sort.Strings(src)
	distinct := make([]string, 0, len(src))
	for index, value := range src {
		if index == 0 || value != src[index-1] {
			distinct = append(distinct, value)
		}
	}
	return distinct
}

func (src Slice) Distinct() Slice {
	sort.Strings(src)
	return src
}

func (src Slice) Append(one string) Slice {
	src = append(src, one)
	return src
}
