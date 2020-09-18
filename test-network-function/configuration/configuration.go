package configuration

import (
	"os"
)

const (
	configurationFilePathEnvironmentVariableKey = "TEST_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "./test-configuration.yaml"
)

func GetConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

type ContainerIdentifier struct {
	Namespace     string `yaml:"namespace" json:"namespace"`
	PodName       string `yaml:"podName" json:"podName"`
	ContainerName string `yaml:"containerName" json:"containerName"`
}

type Container struct {
	// OpenShift Default network interface name (i.e., eth0)
	DefaultNetworkDevice string `yaml:"defaultNetworkDevice" json:"defaultNetworkDevice"`
	// Container overlay IP
	MultusIpAddresses []string `yaml:"multusIpAddresses,omitempty" json:"multusIpAddresses,omitempty"`
}

// Generic test related configuration
type TestConfiguration struct {
	ContainersUnderTest map[ContainerIdentifier]Container `yaml:"containersUnderTest" json:"containersUnderTest"`
	PartnerContainers   map[ContainerIdentifier]Container `yaml:"partnerContainers" json:"partnerContainers"`
	TestOrchestrator    ContainerIdentifier               `yaml:"testOrchestrator" json:"testOrchestrator"`
}
