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

	ginkgoconfig "github.com/onsi/ginkgo/config"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"gopkg.in/yaml.v2"
)

const (
	configurationFilePathEnvironmentVariableKey = "TNF_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "tnf_config.yml"
)

const (
	containerTestSpecName = "container"
	operatorTestSpecName  = "operator"
)

// File is the top level of the config file. All new config sections must be added here
type File struct {
	// Custom Pod labels for discovering containers under test for generic and container suites
	TargetPodLabels []configsections.Label `yaml:"targetPodLabels,omitempty" json:"targetPodLabels,omitempty"`

	Generic configsections.TestConfiguration `yaml:"generic,omitempty" json:"generic,omitempty"`

	// Operator is the list of operator objects that needs to be tested.
	Operators []configsections.Operator `yaml:"operators,omitempty"  json:"operators,omitempty"`

	// CNFs is the list of the CNFs that needs to be tested. Each entry is a single pod to be tested.
	CNFs []configsections.Cnf `yaml:"cnfs,omitempty" json:"cnfs,omitempty"`

	// CertifiedContainerInfo is the list of container images to be checked for certification status.
	CertifiedContainerInfo []configsections.CertifiedContainerRequestInfo `yaml:"certifiedcontainerinfo,omitempty" json:"certifiedcontainerinfo,omitempty"`

	// CertifiedOperatorInfo is list of operator bundle names that are queried for certification status.
	CertifiedOperatorInfo []configsections.CertifiedOperatorRequestInfo `yaml:"certifiedoperatorinfo,omitempty" json:"certifiedoperatorinfo,omitempty"`

	// CnfAvailableTestCases list the available test cases for  reference.
	CnfAvailableTestCases []string `yaml:"cnfavailabletestcases,omitempty" json:"cnfavailabletestcases,omitempty"`
}

var (
	// configInstance is the singleton instance of loaded config, accessed through GetConfigInstance
	configInstance File
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

// doAutodiscovery will autodiscover config for any enabled test spec. Specs which are not selected will be skipped to
// avoid unnecessary noise in the logs.
func doAutodiscovery() {
	if genericTestConfigRequired() {
		configInstance.Generic = autodiscover.BuildGenericConfig()
	}
	if podTestConfigRequired() {
		configInstance.CNFs = autodiscover.BuildCNFsConfig()
	}
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, operatorTestSpecName) {
		configInstance.Operators = autodiscover.BuildOperatorConfig()
	}
}

// GetConfigInstance provides access to the singleton ConfigFile instance.
func GetConfigInstance() File {
	if !loaded {
		filePath := getConfigurationFilePathFromEnvironment()
		log.Debugf("GetConfigInstance before config loaded, loading from file: %s", filePath)
		err := loadConfigFromFile(filePath)
		if err != nil {
			log.Fatalf("unable to load configuration file: %s", err)
		}

		BuildConfig()
	} else if needsRefresh {
		BuildConfig()
	}
	return configInstance
}

func findContainersByLabels(labels []configsections.Label) (containers []configsections.Container) {
	for _, l := range labels {
		list, err := autodiscover.GetContainersByLabel(l)
		if err == nil {
			containers = append(containers, list...)
		} else {
			log.Warnf("failed to query by label: %v %v", l, err)
		}
	}
	return containers
}

func findPodsByLabels(labels []configsections.Label) (cnfs []configsections.Cnf) {
	for _, l := range labels {
		pods, err := autodiscover.GetPodsByLabel(l)
		if err == nil {
			for i := range pods.Items {
				cnfs = append(cnfs, autodiscover.BuildCnfFromPodResource(&pods.Items[i]))
			}
		} else {
			log.Warnf("failed to query by label: %v %v", l, err)
		}
	}
	return cnfs
}

// BuildConfig does auto discovery based on default labels if enabled and additional target pod/container
// discovery based on custom labels
func BuildConfig() {
	if autodiscover.PerformAutoDiscovery() {
		log.Warn("doing configuration autodiscovery. Currently this WILL override parts of the configuration file")
		doAutodiscovery()
	}
	if genericTestConfigRequired() {
		configInstance.Generic.ContainersUnderTest = append(configInstance.Generic.ContainersUnderTest, findContainersByLabels(configInstance.TargetPodLabels)...)
	}
	if podTestConfigRequired() {
		configInstance.CNFs = append(configInstance.CNFs, findPodsByLabels(configInstance.TargetPodLabels)...)
	}
	needsRefresh = false
}

func genericTestConfigRequired() bool {
	// TODO clean up as part of config api refactoring task
	return true
}

func podTestConfigRequired() bool {
	return testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, containerTestSpecName)
}

// SetNeedsRefresh marks the config stale so that the next getInstance call will redo discovery
func SetNeedsRefresh() {
	needsRefresh = true
}
