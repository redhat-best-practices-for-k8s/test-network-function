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

// Generic test related configuration
type TestConfiguration struct {
	// The CNF pod brought by our partner
	PodUnderTest struct {
		// Pod name
		Name string `yaml:"name"`
		// Pod namespace
		Namespace string `yaml:"namespace"`
		// Container facets
		ContainerConfiguration struct {
			// Container name
			Name string `yaml:"name"`
			// OpenShift Default network interface name (i.e., eth0)
			DefaultNetworkDevice string `yaml:"defaultNetworkDevice"`
			// Container overlay IP
			MultusIpAddresses []string `yaml:"multusIpAddresses"`
		} `yaml:"containerConfiguration"`
	} `yaml:"podUnderTest"`
	PartnerPod struct {
		// Partner Pod name
		Name string `yaml:"name"`
		// Partner Pod namespace
		Namespace string `yaml:"namespace"`
		// Container facets
		ContainerConfiguration struct {
			// Container name
			Name string `yaml:"name"`
			// OpenShift Default network interface name (i.e., eth0)
			DefaultNetworkDevice string `yaml:"defaultNetworkDevice"`
			// Container overlay IP
			MultusIpAddresses []string `yaml:"multusIpAddresses"`
		} `yaml:"containerConfiguration"`
	} `yaml:"partnerPod"`
}
