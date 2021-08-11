// Copyright (C) 2020-2021 Red Hat, Inc.
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

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"gopkg.in/yaml.v2"
)

const (
	configurationFilePathEnvironmentVariableKey = "TNF_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "tnf_config.yml"
)

var (
	// configInstance is the singleton instance of loaded config, accessed through GetConfigInstance
	configInstance configsections.TestConfiguration
	// loaded tracks if the config has been loaded to prevent it being reloaded.
	loaded = false
	// set when an intrusive test has done something that would cause Pod/Container to be recreated
	needsRefresh = false
)

// getConfigurationFilePathFromEnvironment returns the test configuration file.
func getConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

// loadConfigFromFile loads a config file once.
func loadConfigFromFile(filePath string) error {
	if loaded {
		return fmt.Errorf("cannot load config from file when a config is already loaded")
	}
	log.Info("Loading config from file: ", filePath)

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(contents, &configInstance)
	if err != nil {
		return err
	}
	loaded = true
	return nil
}

// GetConfigInstance provides access to the singleton ConfigFile instance.
func GetConfigInstance() configsections.TestConfiguration {
	if !loaded {
		filePath := getConfigurationFilePathFromEnvironment()
		log.Debugf("GetConfigInstance before config loaded, loading from file: %s", filePath)
		err := loadConfigFromFile(filePath)
		if err != nil {
			log.Fatalf("unable to load configuration file: %s", err)
		}
		doAutodiscover()
	} else if needsRefresh {
		configInstance.TestPartner = configsections.TestPartner{}
		doAutodiscover()
	}
	return configInstance
}

func doAutodiscover() {
	if autodiscover.PerformAutoDiscovery() {
		configInstance.TestTarget = autodiscover.FindTestTarget(configInstance.TargetPodLabels)
		autodiscover.FillTestPartner(&configInstance.TestPartner)
	}
	needsRefresh = false
}

// SetNeedsRefresh marks the config stale so that the next getInstance call will redo discovery
func SetNeedsRefresh() {
	needsRefresh = true
}
