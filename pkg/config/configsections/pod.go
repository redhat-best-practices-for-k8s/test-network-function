// Copyright (C) 2020-2021 Red Hat, Inc.
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

package configsections

// Pod defines cloud network function in the cluster
type Pod struct {
	// Name is the name of a single Pod to test
	Name string `yaml:"name" json:"name"`

	// Namespace where the Pod is deployed
	Namespace string `yaml:"namespace" json:"namespace"`

	// Tests this is list of test that need to run against the Pod.
	Tests []string `yaml:"tests" json:"tests"`
}
