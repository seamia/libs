// Copyright 2017-2025 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	affirmative           = "yes"
	defaultConfigFileName = "./config"
	keyLiveReload         = "live.reload"
)

type (
	Config = map[string]string
)

var (
	configFileName = defaultConfigFileName // this can be changed externally (prior to the first use)
	configData     Config
	configDebug    bool
	configTrace    bool
)

func loadConfigFile(name string) (Config, error) {
	raw, err := os.ReadFile(name)
	if err != nil {
		workdir, _ := os.Getwd()
		alarm("failed to find/open config file (%s): %v (work dir: %v)", name, err, workdir)

		if configFileName == defaultConfigFileName {
			appDir, _ := path.Split(os.Args[0])
			_, configName := path.Split(configFileName)
			absConfigName := path.Join(appDir, configName)
			if absConfigName != name {
				return loadConfigFile(absConfigName)
			}
		}

		return nil, err
	}
	
	var data Config
	if err := json.Unmarshal(raw, &data); err != nil {
		alarm("failed to process config file (%s): %v", name, err)
		return nil, err
	}

	if value, found := data["debug"]; found {
		configDebug = value == affirmative
	}

	if value, found := data["trace"]; found {
		configTrace = value == affirmative
	}

	transform := func(s string) string {
		return os.Expand(s, func(key string) string {
			// 1. see if we have the key in data
			if value, found := data["$"+key]; found {
				return value
			}

			// 2. see if there is such an env var
			if value, found := os.LookupEnv(s); found {
				return value
			}

			return s
		})
	}

	for k, v := range data {
		if strings.HasPrefix(k, "$") {
			data[k] = transform(v)
		}
	}

	for k, v := range data {
		if !strings.HasPrefix(k, "$") {
			if expanded := transform(v); expanded != v {
				if configDebug {
					trace("config: changing (%s) to (%s)", v, expanded)
				}
				data[k] = expanded
			}
		}
	}

	return data, nil
}

func Get(key string) string {
	if configData == nil {
		data, err := loadConfigFile(configFileName)
		if err != nil {
			alarm("failed to find/open/process config file (%s): %v", configFileName, err)
			os.Exit(13)
		}

		configData = data
		if value, found := data[keyLiveReload]; found && value == affirmative {
			trace("config: live.reload is requested")
			go configLiveReload(configFileName)
		}
	}

	if value, ok := configData[key]; ok {
		return value
	}

	if configDebug {
		trace("failed to find key (%s) in config", key)
	}
	return ""
}

func Flag(key string) bool {
	return Get(key) == affirmative
}

func GetInt(key string, fallback int) int {
	if txt := Get(key); len(txt) > 0 {
		if value, err := strconv.Atoi(txt); err != nil {
			warning("failed to convert value (%v) for key (%s) into int", txt, key)
		} else {
			return value
		}
	}
	return fallback
}

func configLiveReload(name string) {

	const scope = "live.config: "
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		failure(scope+"failed to create a watcher, err: %v", err)
		return
	}
	defer watcher.Close()

	if err = watcher.Add(name); err != nil {
		failure(scope+"failed to add a file (%s) to watcher, err: %v", name, err)
		return
	}

	for {
		select {
		case event, operational := <-watcher.Events:
			if !operational {
				warning(scope + "events channel closed")
				return
			}

			// trace(scope+"got event: %v", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				trace(scope+"modified file: %s", event.Name)
				if !reloadConfig(name) {
					return
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				if err = watcher.Add(name); err != nil {
					failure(scope+"failed to add a file (%s) to watcher, err: %v", name, err)
					return
				}
				if !reloadConfig(name) {
					return
				}
			}
		case err, operational := <-watcher.Errors:
			if !operational {
				warning(scope + "errors channel closed")
				return
			}
			warning(scope+"watch error: %v", err)
		}
	}

}

func reloadConfig(name string) bool {
	const scope = "reload.config: "
	time.Sleep(1 * time.Second)
	if data, err := loadConfigFile(name); err == nil {
		if value, found := data[keyLiveReload]; found && value != affirmative {
			trace(scope + " no more live reloads")
			return false
		}

		if configDebug {
			for key, oldValue := range configData {
				if newValue, found := data[key]; found {
					if oldValue != newValue {
						trace(scope+"value for key [%s] changed from [%s] to [%s]", key, oldValue, newValue)
					}
				} else {
					trace(scope+"removed key [%s] from config", key)
				}
			}
			for key, value := range data {
				if _, found := configData[key]; !found {
					trace(scope+"added new key [%s] with value [%s]", key, value)
				}
			}
		}

		configData = data

	} else {
		warning(scope+"failed to reload config file (%s), err: %v", name, err)
	}

	return true
}
