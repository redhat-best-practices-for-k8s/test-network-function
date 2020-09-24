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

var defaultConfigurationFilePath = path.Join("cnf-specific", "casa", "cnf", "casa-cnf-test-configuration.yaml")

func GetCasaCNFTestConfiguration() (*CasaCNFConfiguration, error) {
	config := &CasaCNFConfiguration{}
	yamlFile, err := ioutil.ReadFile(getCasaCNFConfigurationFilePathFromEnvironment())
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

type CasaCNFConfiguration struct {
	NRFName   string   `json:"nrfName" yaml:"nrfName"`
	CNFTypes  []string `json:"cnfTypes" yaml:"cnfTypes"`
	Namespace string   `json:"namespace" yaml:"namespace"`
}
