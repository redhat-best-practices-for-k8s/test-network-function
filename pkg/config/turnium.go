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

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	turniumConfigurationKey = "turnium"
)

// GetTurniumConfiguration returns the Turnium test configuration.
func GetTurniumConfiguration(filepath string) (*TurniumConfiguration, error) {
	config := &TurniumConfiguration{}

	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, config)
	if err != nil {
		return nil, err
	}

	//nolint:errcheck // Even if not run, each of the suites attempts to initialise the config. This results in
	// RegisterConfigurations erroring due to duplicate keys.
	(*GetInstance()).RegisterConfiguration(turniumConfigurationKey, config)

	return config, err
}

// TurniumConfiguration provides turnium related configuration
type TurniumConfiguration struct {
	// TurniumBonder identifies the container in which to run the `legids -v` test
	TurniumBonder ContainerIdentifier `yaml:"turniumBonder" json:"turniumBonder"`
}
