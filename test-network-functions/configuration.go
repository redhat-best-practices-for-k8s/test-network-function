package setup


type TestConfiguration struct {
	ContainerUnderTest struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"containerUnderTest"`
	PartnerContainer struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"partnerContainer"`
}
