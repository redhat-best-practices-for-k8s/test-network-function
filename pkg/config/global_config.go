// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	configurationFilePathEnvironmentVariableKey = "TEST_CONFIGURATION_PATH"
	defaultConfigurationFilePath = "tnf_config.yml"
	defaultFilename = ""
	yamlSuffix                      = ".yaml"
	ymlSuffix                       = ".yml"
)

var (
	configInstance configPool
	once sync.Once
)

type configPool map[string]interface{}

// getConfigurationFilePathFromEnvironment returns the test configuration file.
func getConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

// isYAMLFile is an heuristic to determine whether a file is likely a YAML file
func isYAMLFile(filepath string) bool {
	return strings.HasSuffix(filepath, yamlSuffix) || strings.HasSuffix(filepath, ymlSuffix)
}

func loadConfigFromFile() (error) {
	filepath := getConfigurationFilePathFromEnvironment()

	log.Info("Loading config from file: ", filepath)
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	if isYAMLFile(filepath) {
		err = yaml.Unmarshal(contents, configInstance)
	} else {
		return fmt.Errorf("Must be a YAML file: %s", filepath)
	}

	return err
}

// GetConfigSection loads a top-level section from the config file into the object referenced by `out`
func GetConfigSection(configSection string, out interface{}) {
	conf := GetConfigInstance()

	log.Infof("request to load section '%s' into type '%s'", configSection, reflect.ValueOf(out).Type().String())

	section, err := yaml.Marshal(conf[configSection])
	if err != nil {
	}

	err = yaml.Unmarshal(section, out)
	if err != nil {
	}

	fmt.Println(out)
}

// GetInstance returns the singleton Config instance.
func GetConfigInstance() configPool  {
	once.Do(func() {
		configInstance = make(configPool)
		err := loadConfigFromFile()
		if err != nil {
			log.Fatal(err)
		}
	})
	log.Infof("%#v\n", configInstance)
	return configInstance
}
