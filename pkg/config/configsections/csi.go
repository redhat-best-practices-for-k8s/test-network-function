package configsections

// Pod defines cloud network function in the cluster
type Csi struct {
	// Name is the name of a single Pod to test
	Name string `yaml:"name" json:"name"`
	Packag string `yaml:"package" json:"package"`
}
