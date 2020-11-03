package version

import (
	"encoding/json"
	"io/ioutil"
	"path"
)

var (
	defaultVersionFile = path.Join("..", "version.json")
)

// Version refers to the `test-network-function` version tag.
type Version struct {
	// Tag is the Git tag for the version.
	Tag string `json:"tag" yaml:"tag"`
}

// GetVersion extracts the test-network-function version.
func GetVersion() (*Version, error) {
	contents, err := ioutil.ReadFile(defaultVersionFile)
	if err != nil {
		return nil, err
	}
	version := &Version{}
	err = json.Unmarshal(contents, version)
	return version, err
}
