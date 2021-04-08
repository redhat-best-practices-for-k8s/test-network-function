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

package generic

// ContainerIdentifier is a complex key representing a unique container.
type ContainerIdentifier struct {
	Namespace     string `yaml:"namespace" json:"namespace"`
	PodName       string `yaml:"podName" json:"podName"`
	ContainerName string `yaml:"containerName" json:"containerName"`
}

// Container contains the payload of container facets.
type Container struct {
	ContainerIdentifier `yaml:",inline"`
	// OpenShift Default network interface name (i.e., eth0)
	DefaultNetworkDevice string `yaml:"defaultNetworkDevice" json:"defaultNetworkDevice"`
	// MultusIPAddresses are the overlay IPs.
	MultusIPAddresses []string `yaml:"multusIpAddresses" json:"multusIpAddresses"`
}

// TestConfiguration provides generic test related configuration
type TestConfiguration struct {
	ContainersUnderTest []Container         `yaml:"containersUnderTest" json:"containersUnderTest"`
	PartnerContainers   []Container         `yaml:"partnerContainers" json:"partnerContainers"`
	TestOrchestrator    ContainerIdentifier `yaml:"testOrchestrator" json:"testOrchestrator"`
	// ExcludeContainersFromConnectivityTests excludes specific containers from network connectivity tests.  This is particularly useful for containers that don't have ping available.
	ExcludeContainersFromConnectivityTests []ContainerIdentifier `yaml:"excludeContainersFromConnectivityTests" json:"excludeContainersFromConnectivityTests"`
}
