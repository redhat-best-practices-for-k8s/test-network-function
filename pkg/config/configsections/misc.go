// Copyright (C) 2020-2022 Red Hat, Inc.
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

// Types defined in this file are not currently in use. Move them out when starting to use.
// May remove this altogether in the future

// CNFType defines a type to be either Operator or Container
type CNFType string

// Permission defines roles and cluster roles resources
type Permission struct {
	// Name is the name of Roles and Cluster Roles that is specified in the CSV
	Name string `yaml:"name" json:"name"`

	// Role is the role type either CLUSTER_ROLE or ROLE
	Role string `yaml:"role" json:"role"`
}
