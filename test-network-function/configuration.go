package cnftests

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
			MultusIpAddress string `yaml:"multusIpAddress"`
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
			MultusIpAddress string `yaml:"multusIpAddress"`
		} `yaml:"containerConfiguration"`
	} `yaml:"partnerPod"`
}
