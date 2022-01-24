package configsections

// csi defines the list of the operators on the cluster
type Csi struct {
	// Name is the name of a single Pod to test
	Org    string `yaml:"org" json:"org"`
	Packag string `yaml:"package" json:"package"`
}
