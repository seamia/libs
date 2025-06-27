package util

import (
	"testing"
)

func TestAbandoned(t *testing.T) {
	one := Slice{}

	if !one.Empty() {
		t.Fatal("should've been empty")
	}

	if one.Len() != 0 {
		t.Fatal("should've been 0")
	}

	one.Append("apple")

	if one.Empty() {
		t.Fatal("should've not been empty")
	}

	if one.Len() != 1 {
		t.Fatal("should've been 1")
	}
}

/*

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



*/
