package config

import "fmt"

// instance is the Singleton for config Pool.
var instance = &Pool{
	configurations: make(map[string]interface{}),
}

// Pool follows the pool design pattern, and contains the named configurations.
type Pool struct {
	configurations map[string]interface{}
}

// RegisterConfiguration registers a configuration with the Pool.  If configurationKey is already contained in the pool,
// an appropriate error is returned.
func (p *Pool) RegisterConfiguration(configurationKey string, configurationPayload interface{}) error {
	if _, ok := p.configurations[configurationKey]; ok {
		return fmt.Errorf("pool already contains a configuration for: %s", configurationKey)
	}
	p.configurations[configurationKey] = configurationPayload
	return nil
}

// GetConfigurations returns the raw configuration map.
func (p *Pool) GetConfigurations() map[string]interface{} {
	return p.configurations
}

// GetInstance returns the singleton Pool
func GetInstance() *Pool {
	return instance
}
