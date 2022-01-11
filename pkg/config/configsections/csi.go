package configsections

// Pod defines cloud network function in the cluster
type Csi struct {
	// Name is the name of a single Pod to test
	Name string `yaml:"name" json:"name"`

	Organization string `yaml:"organization" json:"organization"`

	Packag string `yaml:"package" json:"package"`
}
