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

// CertifiedContainerRequestInfo contains all certified images request info
type CertifiedContainerRequestInfo struct {
	// Name is the name of the `operator bundle package name` or `image-version` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Repository is the name of the repository `rhel8` of the container
	// This is valid for container only and required field
	Repository string `yaml:"repository" json:"repository"`
}

// CertifiedOperatorRequestInfo contains all certified operator request info
type CertifiedOperatorRequestInfo struct {

	// Name is the name of the `operator bundle package name` that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Organization as understood by the operator publisher , e.g. `redhat-marketplace`
	Organization string `yaml:"organization" json:"organization"`
}
