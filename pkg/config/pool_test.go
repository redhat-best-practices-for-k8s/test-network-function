// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package config_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config"
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
