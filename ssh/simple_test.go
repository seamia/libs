// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"testing"
)

func TestStreamRecordSequenceLess(t *testing.T) {
	fmt.Printf("start\n")
	defer fmt.Printf("end\n")

	to := func(name string) {
		if err := SSHCopyFile("README.md", name); err != nil {
			panic(err)
		}
	}

	to("vova:dst")
	to("vova:222")
	to("vova:333")

	to("linode:dst")
	to("linode:222")
	to("linode:333")
	to("linode:plus/dst")

	to("castle:dst")
	to("dock:dst")
	/*
		to("castle:222")
		to("castle:333")
	*/
	to("bingen:222")
	to("bingen:333")
}
