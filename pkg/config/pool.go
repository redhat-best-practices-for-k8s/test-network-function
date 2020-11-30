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
