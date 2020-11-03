package config_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetInstance(t *testing.T) {
	assert.NotNil(t, config.GetInstance())
}

// Also tests GetConfigurations
func TestPool_RegisterConfiguration(t *testing.T) {
	type arbitraryConfig struct {
		name string
		id   int
	}
	assert.Nil(t, config.GetInstance().RegisterConfiguration("someKey", &arbitraryConfig{}))
	assert.Contains(t, config.GetInstance().GetConfigurations(), "someKey")
	assert.Equal(t, &arbitraryConfig{}, config.GetInstance().GetConfigurations()["someKey"])
	assert.Equal(t, fmt.Errorf("pool already contains a configuration for: someKey"),
		config.GetInstance().RegisterConfiguration("someKey", &arbitraryConfig{}))
}
