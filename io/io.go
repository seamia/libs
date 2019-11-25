package io

import (
	"encoding/json"
	"io/ioutil"
	"github.com/seamia/libs/zip"
)

func LoadJson(filename string) (interface{}, error) {

	raw, err := ioutil.ReadFile(filename)
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
