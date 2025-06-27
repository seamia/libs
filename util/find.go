// Copyright 2019 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	root   = "../../dist"
	suffix = ".html"
)

func main1() {

	absolute, err := filepath.Abs(root)
	if err != nil {
		fmt.Println(err)
	}

	//var files []string
	err = filepath.Walk(absolute, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
			// files = append(files, path)
			processOneFile(path, info.Name())
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(files)
}

func processOneFile(src, name string) error {
	raw, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	txt := string(raw)
	vars := findVariables(txt)

	if strings.HasSuffix(name, suffix) {
		name = name[:len(name)-len(suffix)]
	}

	// var known = map[string][]string {
	//  "": [
	fmt.Fprintf(os.Stdout, "\t\"%s\": {", name)
	for _, one := range vars {
		fmt.Fprintf(os.Stdout, "\"%s\", ", one)
	}
	fmt.Fprintf(os.Stdout, "},\n")

	return nil
}

func findVariables(src string) []string {
	return findInBetween(src, "{{", "}}", true, removeKeywords)
}

type Transform func(string) (string, bool)

func findInBetween(src, left, right string, unique bool, sanitize Transform) []string {
	results := Slice{}
	if sanitize == nil {
		sanitize = func(src string) (string, bool) {
			return src, true
		}
	}
	for {
		l := strings.Index(src, left)
		if l < 0 {
			break
		}
		src = src[l+len(left):]
		r := strings.Index(src, right)
		if r < 0 {
			break
		}
		found := src[:r]
		src = src[r+len(right):]

		if found, okay := sanitize(found); okay {
			results.Append(found)
		}
	}

	if unique && len(results) > 1 {
		return results.Distinct()
	}

	return results
}

var exclusions = map[string]bool{
	"#each":   true,
	"/each":   true,
	"#unless": true,
	"/unless": true,
	"@last":   true,
	"this":    true,
}

func removeKeywords(src string) (string, bool) {
	parts := strings.Split(src, " ")
	for _, one := range parts {
		if _, present := exclusions[one]; present {
			continue
		}
		return one, true
	}
	return "", false
}

func ExludeThese(exclude []string) Transform {
	exludes := map[string]struct{}{}
	for _, one := range exclude {
		exludes[one] = struct{}{}
	}
	return func(src string) (string, bool) {
		if _, found := exludes[src]; found {
			return "", false
		}
		return src, true
	}
}
