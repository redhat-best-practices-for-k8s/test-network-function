package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

//String that contains the configured configuration path
var configPath = flag.String("config", "./config/config.yml", "path to config file")

//Config struct for configuring operator test
type Config struct {
	//Csv is a clusterServiceVersion which contains the packaging details of the operator
	Csv struct {
		//Name of csv  operator package with version name
		Name string `yaml:"name" json:"name"`
		//Namespace where the operator will be running
		Namespace string `yaml:"namespace" json:"namespace"`
		//Expected status of the Csv
		Status string `yaml:"status" json:"status"`
	} `yaml:"csv" json:"csv"`
}

//NewConfig  returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	var file *os.File
	var err error
	// Create config structure
	config := &Config{}
	// Open config file
	if file, err = os.Open(configPath); err != nil {
		return nil, err
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func validateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

// parseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func parseFlags() (string, error) {
	var err error
	flag.Parse()
	// Validate the path first
	if err = validateConfigPath(*configPath); err != nil {
		return "", err
	}
	// Return the configuration path
	return *configPath, nil
}

// GetConfig returns the Operator TestConfig configuration.
func GetConfig() (*Config, error) {
	// Generate our config based on the config supplied
	// by the user in the flags
	cfgPath, err := parseFlags()
	if err != nil {
		return nil, err
	}
	cfg, err := NewConfig(cfgPath)
	return cfg, err
}
