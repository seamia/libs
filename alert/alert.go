// Copyright 2017 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alert

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/seamia/libs/zip"
)

const (
	agentName           = "no-agent"
	headerContentType   = "Content-Type"
	headerUserAgent     = "User-Agent"
	headerAuthorization = "Authorization"
	authorizationBasic  = "Basic "
	mimeRaw             = "custom/kirland"
	mimeCompressed      = "custom/renton"
	usernameLength      = 16
	defaultHost         = "https://alerts.nontechno.com/"
)

type ApplicationConfig struct {
	Name string `json:"name,omitempty"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Host string `json:"host,omitempty"`
}

func PostData(config string, data []byte) {
	if len(data) == 0 {
		log.Printf("nothing to post")
		return
	}
	if len(config) == 0 {
		config = "default"
	}

	configFileName := "./" + config + ".config"
	raw, err := ioutil.ReadFile(configFileName)
	if err != nil {
		log.Printf("failed to open config (%s), err: %v", configFileName, err)
		return
	}
	var ac ApplicationConfig
	if err := json.Unmarshal(raw, &ac); err != nil {
		log.Printf("failed to parse config (%s), err: %v", configFileName, err)
		return
	}

	if len(ac.Name) == 0 {
		ac.Name = config
	}
	if len(ac.Host) == 0 {
		ac.Host = defaultHost
	}

	app := strings.Trim(ac.Name, "/\\ \t\r\n")

	mime := mimeRaw
	zipped := zip.Compress(data)
	if len(zipped) < len(data) {
		data = zipped
		mime = mimeCompressed
	}

	client := &http.Client{
		//		Jar: cookieJar,
		//		CheckRedirect: redirectPolicyFunc,
	}

	postUrl := ac.Host + "app/" + app
	req, err := http.NewRequest(http.MethodPost, postUrl, bytes.NewReader(data))
	req.Header.Add(headerAuthorization, authorizationBasic+basicAuth(ac.User, ac.Pass))
	req.Header.Add(headerContentType, mime)
	req.Header.Add(headerUserAgent, agentName)
	resp, err := client.Do(req)

	// check for response error
	if err != nil {
		log.Printf("failed to post, err: %v", err)
		return
	}
	_ = resp
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
