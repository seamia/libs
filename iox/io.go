// Copyright 2020 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iox

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/seamia/libs/zip"
)

func LoadJson(filename string) (interface{}, error) {

	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// potential de-compression
	raw, err = zip.Decompress(raw)
	if err != nil {
		return nil, err
	}

	var payload interface{}
	if err = json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func LoadJsonAsDictionary(filename string) (map[string]string, error) {
	raw, err := LoadJson(filename)
	if err != nil {
		return nil, err
	}

	if slice, converts := raw.(map[string]interface{}); converts {
		dict := make(map[string]string)
		for key, value := range slice {
			if txt, converts := value.(string); converts {
				dict[key] = txt
			} else {
				// todo: consider more elaborate convertion steps here, e.g. int to string ...
			}
		}
		return dict, nil
	}
	return nil, errors.New("wrong underlaying type")
}

func SaveJson(filename string, payload interface{}, compress bool) error {

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// potential compression
	if compress {
		data = zip.Compress(data)
	}

	if err := os.WriteFile(filename, data, 0666); err != nil {
		return err
	}

	return nil
}
