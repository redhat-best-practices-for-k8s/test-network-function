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

package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
)

const (
	casaCNFTestConfigurationFilePathEnvironmentVariableKey = "CASA_CNF_TEST_CONFIGURATION_PATH"
)

var defaultConfigurationFilePath = path.Join("cnf_specific", "casa", "casa-cnf-test-configuration.yaml")

// GetCasaCNFTestConfiguration returns the Casa CNF specific test configuration.
func GetCasaCNFTestConfiguration() (*CasaCNFConfiguration, error) {
	config := &CasaCNFConfiguration{}
	configFilePath := getCasaCNFConfigurationFilePathFromEnvironment()
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getCasaCNFConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(casaCNFTestConfigurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

// CasaCNFConfiguration stores the Casa CNF specific test configuration.
type CasaCNFConfiguration struct {
	NRFName   string   `json:"nrfName" yaml:"nrfName"`
	CNFTypes  []string `json:"cnfTypes" yaml:"cnfTypes"`
	Namespace string   `json:"namespace" yaml:"namespace"`
}
