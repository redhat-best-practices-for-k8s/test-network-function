package configuration

import (
	"os"
)

const (
	configurationFilePathEnvironmentVariableKey = "TEST_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "./test-configuration.yaml"
)

// GetConfigurationFilePathFromEnvironment returns the test configuration file.
func GetConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

// ContainerIdentifier is a complex key representing a unique container.
type ContainerIdentifier struct {
	Namespace     string `yaml:"namespace" json:"namespace"`
	PodName       string `yaml:"podName" json:"podName"`
	ContainerName string `yaml:"containerName" json:"containerName"`
}

// Container contains the payload of container facets.
type Container struct {
	// OpenShift Default network interface name (i.e., eth0)
	DefaultNetworkDevice string `yaml:"defaultNetworkDevice" json:"defaultNetworkDevice"`
	// MultusIPAddresses are the overlay IPs.
	MultusIPAddresses []string `yaml:"multusIpAddresses,omitempty" json:"multusIpAddresses,omitempty"`
}

// TestConfiguration provides generic test related configuration
type TestConfiguration struct {
	ContainersUnderTest map[ContainerIdentifier]Container `yaml:"containersUnderTest" json:"containersUnderTest"`
	PartnerContainers   map[ContainerIdentifier]Container `yaml:"partnerContainers" json:"partnerContainers"`
	TestOrchestrator    ContainerIdentifier               `yaml:"testOrchestrator" json:"testOrchestrator"`
}
