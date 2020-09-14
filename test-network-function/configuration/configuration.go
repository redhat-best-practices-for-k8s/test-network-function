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

// TestConfiguration provides a generic test related configuration
type TestConfiguration struct {
	// PodUnderTest is the CNF pod brought by our partner
	PodUnderTest struct {
		// Name is the Pod name
		Name string `yaml:"name"`
		// Namespace is the Pod namespace
		Namespace string `yaml:"namespace"`
		// ContainerConfiguration contains the Container facets
		ContainerConfiguration struct {
			// Name is the Container name
			Name string `yaml:"name"`
			// DefaultNetworkDevice is the OpenShift Default network interface name (i.e., eth0)
			DefaultNetworkDevice string `yaml:"defaultNetworkDevice"`
			// MultusIPAddresses holds the Container overlay IP addresses
			MultusIPAddresses []string `yaml:"multusIPAddresses"`
		} `yaml:"containerConfiguration"`
	} `yaml:"podUnderTest"`
	// PartnerPod is the test partner.
	PartnerPod struct {
		// Name is the Partner Pod name
		Name string `yaml:"name"`
		// Namespace is the Partner Pod namespace
		Namespace string `yaml:"namespace"`
		// ContainerConfiguration contains the Container facets
		ContainerConfiguration struct {
			// Name is the Container name
			Name string `yaml:"name"`
			// DefaultNetworkDevice is the OpenShift Default network interface name (i.e., eth0)
			DefaultNetworkDevice string `yaml:"defaultNetworkDevice"`
			// MultusIPAddresses holds the Container overlay IP
			MultusIPAddresses []string `yaml:"multusIPAddresses"`
		} `yaml:"containerConfiguration"`
	} `yaml:"partnerPod"`
}
