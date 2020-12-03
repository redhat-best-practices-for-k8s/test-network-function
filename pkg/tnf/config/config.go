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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// String that contains the configuration path for tnf
var configPath = flag.String("config", "config.yml", "path to config file")

const (
	filePerm = 0644
)

// CNFType defines a type to be either Operator or Container
type CNFType string

const (
	// ContainerType is a `Container Image` CNF type
	ContainerType = "CONTAINER"
	// OperatorType is a `Operator` CNF type
	OperatorType = "OPERATOR"
)

// CertifiedContainerRequestInfo contains all certified images request info
type CertifiedContainerRequestInfo struct {
	// Name is the name of the `operator bundle package name` or `image-version` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`
	// Repository is the name of the repository `rhel8` of the container
	// This is valid for container only and required field
	Repository string `yaml:"repository" json:"repository"`
}

// CertifiedOperatorRequestInfo contains all certified operator request info
type CertifiedOperatorRequestInfo struct {
	// Name is the name of the `operator bundle package name` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`
	// Organization as understood by the operator publisher , e.g. `redhat-marketplace`
	Organization string `yaml:"organization" json:"organization"`
}

// Operator struct defines operator manifest for testing
type Operator struct {
	// Name is a required field, Name of the csv .
	Name string `yaml:"name" json:"name"`
	// Namespace is a required field , namespace is where the csv is installed.
	// If its all namespace then you can replace it with ALL_NAMESPACE TODO: add check for ALL_NAMESPACE
	Namespace string `yaml:"namespace" json:"namespace"`
	// Status is a required field , specified what status of the csv to be checked.
	Status string `yaml:"status" json:"status"`
	// AutoGenerate if set to true will generate the config with operator related artifacts.
	AutoGenerate string `yaml:"autogenerate,omitempty" json:"autogenerate"`
	// CRDs If AutoGenerate is set to true, then the program will populate the CRD data from the CSV file.
	CRDs []Crd `yaml:"crds" json:"crds"`
	// Deployments If AutoGenerate is set to true, then the program will populate the Deployment data from the CSV file.
	Deployments []Deployment `yaml:"deployments" json:"deployments"`
	// CNFs If AutoGenerate is set to true, then the program will populate the CNFs data from the CSV file.
	CNFs []Cnf `yaml:"cnfs" json:"cnfs"`
	// Permissions If AutoGenerate is set to true, then the program will populate the Permission data from the CSV file.
	Permissions []Permission `yaml:"permissions" json:"permissions"`
	// Tests this is list of test that need to run against the operator.
	Tests []string `yaml:"tests" json:"tests"`
	// CertifiedOperatorRequestInfos  is list of  operator bundle names (`package-name`)
	// that are queried for certificate status
	CertifiedOperatorRequestInfos []CertifiedOperatorRequestInfo `yaml:"certifiedoperatorrequestinfo,omitempty" json:"certifiedoperatorrequestinfo,omitempty"`
}

// Crd struct defines Custom Resource Definition of the operator
type Crd struct {
	// Name is the name of the CRD populated by the operator config generator
	Name string `yaml:"name" json:"name"`
	// Namespace is the namespace where above CRD is installed(For all namespace this will be ALL_NAMESPACE)
	Namespace string `yaml:"namespace" json:"namespace"`
	// Instances is the instance of CR matching for the above CRD KIND
	Instances []Instance `yaml:"instances" json:"instances"`
}

// Deployment defines deployment resources
type Deployment struct {
	// Name is the name of the deployment specified in the CSV
	Name string `yaml:"name" json:"name"`
	// Replicas is no of replicas that are expected for this deployment as specified in the CSV
	Replicas string `yaml:"replicas" json:"replicas"`
}

// Permission defines roles and cluster roles resources
type Permission struct {
	// Name is the name of Roles and Cluster Roles that is specified in the CSV
	Name string `yaml:"name" json:"name"`
	// Role is the role type either CLUSTER_ROLE or ROLE
	Role string `yaml:"role" json:"role"`
}

// Cnf defines cloud network function in the cluster
type Cnf struct {
	// Name is the name of the CNF (TODO: This should also take cnf labels in case name is dynamically created)
	Name string `yaml:"name" json:"name"`
	// Namespace where the CNF is deployed
	Namespace string `yaml:"namespace" json:"namespace"`
	// Status is the status of the CNF
	Status string `yaml:"status" json:"status"`
	// Tests this is list of test that need to run against the CNF.
	Tests []string `yaml:"tests" json:"tests"`
	// CertifiedContainerRequestInfos  is list of images (`repo/image-version`)
	// that are queried for certificate status
	CertifiedContainerRequestInfos []CertifiedContainerRequestInfo `yaml:"certifiedcontainerrequestinfo,omitempty" json:"certifiedcontainerrequestinfo,omitempty"`
}

// Instance defines crd instances in the cluster
type Instance struct {
	// Name is the name of the instance of custom resource (Auto populated)
	Name string `yaml:"name" json:"name"`
}

// TnfContainerOperatorTestConfig the main configuration struct for tnf
type TnfContainerOperatorTestConfig struct {
	// Operator is the lis of operator objects that needs to be tested.
	Operator []Operator `yaml:"operators,omitempty"  json:"operators,omitempty"`
	// CNFs is the list of the CNFs that needs to be tested.
	CNFs []Cnf `yaml:"cnfs,omitempty" json:"cnfs,omitempty"`
	// CnfAvailableTestCases list the available test cases for  reference.
	CnfAvailableTestCases []string `yaml:"cnfavailabletestcases,omitempty" json:"cnfavailabletestcases,omitempty"`
}

// SaveConfig writes configuration to a file at the given config path
func (c *TnfContainerOperatorTestConfig) SaveConfig(configPath string) (err error) {
	bytes, _ := yaml.Marshal(c)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	return
}

// SaveConfigAsJSON writes configuration to a file in json format
func (c *TnfContainerOperatorTestConfig) SaveConfigAsJSON(configPath string) (err error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	return
}

// NewConfig  returns a new decoded TnfContainerOperatorTestConfig struct
func NewConfig(configPath string) (*TnfContainerOperatorTestConfig, error) {
	var file *os.File
	var err error
	// Create config structure
	config := &TnfContainerOperatorTestConfig{}
	// Open config file
	if file, err = os.Open(configPath); err != nil {
		return nil, err
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

// parseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func parseFlags() (string, error) {
	flag.Parse()
	// Validate the path first
	if err := ValidateConfigPath(*configPath); err != nil {
		return "", err
	}
	// Return the configuration path
	return *configPath, nil
}

// GetConfig returns the Operator TestConfig configuration.
func GetConfig() (*TnfContainerOperatorTestConfig, error) {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgPath, err := parseFlags()
	if err != nil {
		return nil, err
	}
	cfg, err := NewConfig(cfgPath)
	return cfg, err
}
