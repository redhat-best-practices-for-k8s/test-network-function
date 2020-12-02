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

package generic

import (
	"encoding/json"
	"fmt"
	configpool "github.com/redhat-nfvpe/test-network-function/pkg/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	configurationFilePathEnvironmentVariableKey = "TEST_CONFIGURATION_PATH"
	configurationKey                            = "generic"
	containerNameJSONKey                        = "containerName"
	defaultConfigurationFilePath                = "./test-configuration.yaml"
	namespaceJSONKey                            = "namespace"
	podNameJSONKey                              = "podName"
	// UseDefaultConfigurationFilePath is the sentinel used to indicate extracting the config filepath from the
	// environment.
	UseDefaultConfigurationFilePath = ""
	yamlExtension                   = ".yaml"
	ymlExtension                    = ".yml"
)

// getConfigurationFilePathFromEnvironment returns the test configuration file.
func getConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

// isYAMLFile is an heuristic to determine whether a file is likely a YAML file (i.e., has a `.yaml` or `.yml`
// extension).
func isYAMLFile(filepath string) bool {
	return strings.HasSuffix(filepath, yamlExtension) || strings.HasSuffix(filepath, ymlExtension)
}

// GetConfiguration returns the cnf-certification-generic-tests test configuration.  GetConfiguration supports reading
// JSON and YAML configurations.
func GetConfiguration(filepath string) (*TestConfiguration, error) {
	config := &TestConfiguration{}
	if filepath == "" {
		filepath = getConfigurationFilePathFromEnvironment()
	}

	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if isYAMLFile(filepath) {
		err = yaml.Unmarshal(contents, config)
	} else {
		err = json.Unmarshal(contents, config)
	}

	(*configpool.GetInstance()).RegisterConfiguration(configurationKey, config)

	return config, err
}

// ContainerIdentifier is a complex key representing a unique container.
type ContainerIdentifier struct {
	Namespace     string `yaml:"namespace" json:"namespace"`
	PodName       string `yaml:"podName" json:"podName"`
	ContainerName string `yaml:"containerName" json:"containerName"`
}

// MarshalText is a custom Marshal function needed since ContainerIdentifier is used as a complex (composite) map key.
func (c ContainerIdentifier) MarshalText() ([]byte, error) {
	type cid ContainerIdentifier
	return json.Marshal(cid(c))
}

// UnmarshalText is a custom Unmarshal function needed since ContainerIdentifier is used as a complex (composite) map
// key.
func (c *ContainerIdentifier) UnmarshalText(text []byte) error {
	type cid ContainerIdentifier
	err := json.Unmarshal(text, (*cid)(c))
	return err
}

// unquoteBytes contains the logic for unquoting raw bytes for Unmarshall operations.
func unquoteBytes(bytes []byte) ([]byte, error) {
	sBytes := string(bytes)
	if len(bytes) < 2 {  //nolint:gomnd
		return nil, fmt.Errorf("cannot decode bytes: %s", sBytes)
	}
	str := sBytes[1 : len(sBytes)-1]
	str = strings.ReplaceAll(str, "\\\"", "\"")
	return []byte(str), nil
}

// extractsField extracts the payload for key from a map of json.RawMessage.
func extractField(data map[string]json.RawMessage, key string) (string, error) {
	if quotedBytes, ok := data[key]; ok {
		quotedString := string(quotedBytes)
		return strconv.Unquote(quotedString)
	}
	return "", fmt.Errorf("couldn't Unmarshal key: %s from %s", key, data)
}

// UnmarshalJSON is a custom Unmarshal function that knows how to reconstruct a ContainerIdentifier from raw bytes.
func (c *ContainerIdentifier) UnmarshalJSON(bytes []byte) error {
	data := make(map[string]json.RawMessage)
	convertedBytes, err := unquoteBytes(bytes)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(convertedBytes, &data); err != nil {
		return err
	}

	containerName, err := extractField(data, containerNameJSONKey)
	if err != nil {
		return err
	}
	c.ContainerName = containerName

	podName, err := extractField(data, podNameJSONKey)
	if err != nil {
		return err
	}
	c.PodName = podName

	namespace, err := extractField(data, namespaceJSONKey)
	if err != nil {
		return err
	}
	c.Namespace = namespace
	return nil
}

// Container contains the payload of container facets.
type Container struct {
	// OpenShift Default network interface name (i.e., eth0)
	DefaultNetworkDevice string `yaml:"defaultNetworkDevice" json:"defaultNetworkDevice"`
	// MultusIPAddresses are the overlay IPs.
	MultusIPAddresses []string `yaml:"multusIpAddresses" json:"multusIpAddresses"`
}

// TestConfiguration provides generic test related configuration
type TestConfiguration struct {
	ContainersUnderTest map[ContainerIdentifier]Container `yaml:"containersUnderTest" json:"containersUnderTest"`
	PartnerContainers   map[ContainerIdentifier]Container `yaml:"partnerContainers" json:"partnerContainers"`
	TestOrchestrator    ContainerIdentifier               `yaml:"testOrchestrator" json:"testOrchestrator"`
	Hosts               []string                          `yaml:"hosts" json:"hosts"`
}
