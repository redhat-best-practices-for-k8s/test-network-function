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

/*
Package config provides test-network-function configuration through a central place. Configuration data
is automatically included in the claim. Configuration should all be contained in a single yaml file, with each
configuration area under its own key. The different sections of the config file are independent and can have different
structures. The config file MUST Unmarshal into a `map[string]interface{}` using `yaml.Unmarshal()`.
The env var "TEST_CONFIGURATION_PATH" identifies the config file. If not set, the default of `tnf_config.yml` is used.
The config file is loaded by this package, and sections can be requested by the different test specs using the
`GetConfigSection()` method, and providing a struct into which the section is to be loaded.
Go `struct`s used for configuration should each be defined in their own files in the package that uses that config,
such as `generic_config.go`. They MUST Marshal successfully into JSON using the `json.Marshal()` method.
*/
package config
