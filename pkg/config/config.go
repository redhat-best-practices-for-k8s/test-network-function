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

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	configurationFilePathEnvironmentVariableKey = "TEST_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "tnf_config.yml"
)

var (
	configInstance  map[string]interface{}
	retrievedConfig map[string]interface{} = make(map[string]interface{})
	rawConfig       []byte
)

// getConfigurationFilePathFromEnvironment returns the test configuration file.
func getConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

func loadConfigFromFile(filePath string) error {
	if configInstance != nil {
		return fmt.Errorf("cannot load config from file when a config is already loaded")
	}
	log.Info("Loading config from file: ", filePath)

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	conf := make(map[string]interface{})
	err = yaml.Unmarshal(contents, conf)
	if err != nil {
		return err
	}

	configInstance = conf
	rawConfig = contents
	return nil
}

// getConfigInstance returns the singleton Config instance.
func getConfigInstance() map[string]interface{} {
	if configInstance == nil {
		filePath := getConfigurationFilePathFromEnvironment()
		log.Debugf("getConfigInstance before config loaded, loading from file: %s", filePath)
		err := loadConfigFromFile(filePath)
		if err != nil {
			log.Fatalf("unable to load configuration file: %s", err)
		}
	}
	return configInstance
}

// GetConfigSection loads a top-level section from the config file into the object referenced by `out`
func GetConfigSection(configSection string, out interface{}) error {
	conf := getConfigInstance()
	if _, ok := conf[configSection]; !ok {
		return fmt.Errorf("config section not found for: %s", configSection)
	}

	targetType := reflect.ValueOf(out).Type().String()
	log.Infof("request to load section '%s' into type '%s'", configSection, targetType)

	// Converting an unknown data structure to a known one is not trivial in Go. Instead of
	// handling this manually using `reflect` we leverage the existing capabilities of the `yaml`
	// package to re-marshal the specific data we are interested in, to allow us to then unmarshal
	// that data into the `out` value.
	section, err := yaml.Marshal(conf[configSection])
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(section, out)
	if err != nil {
		return err
	}

	retrievedKey := fmt.Sprintf("%s:%s", configSection, targetType)
	retrievedConfig[retrievedKey] = out
	return nil
}

// GetConfigForClaim returns the config representations for the claim file
func GetConfigForClaim() ([]byte, map[string]interface{}) {
	return rawConfig, retrievedConfig
}
