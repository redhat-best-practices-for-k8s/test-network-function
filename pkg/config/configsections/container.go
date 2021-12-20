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

import (
	"fmt"
)

// ContainerIdentifier is a complex key representing a unique container.
type ContainerIdentifier struct {
	Namespace        string `yaml:"namespace" json:"namespace"`
	PodName          string `yaml:"podName" json:"podName"`
	ContainerName    string `yaml:"containerName" json:"containerName"`
	NodeName         string `yaml:"nodeName" json:"nodeName"`
	ContainerUID     string `yaml:"containerUID" json:"containerUID"`
	ContainerRuntime string `yaml:"containerRuntime" json:"containerRuntime"`
}

// ContainerConfig contains the payload of container facets.
type ContainerConfig struct {
	ContainerIdentifier `yaml:",inline"`
	// OpenShift Default network interface name (i.e., eth0)
	DefaultNetworkDevice string `yaml:"defaultNetworkDevice" json:"defaultNetworkDevice"`
	// MultusIPAddressesPerNet are the overlay IPs.
	MultusIPAddressesPerNet map[string][]string `yaml:"multusIpAddressesPerNet" json:"multusIpAddressesPerNet"`
}

func (cid *ContainerIdentifier) String() string {
	return fmt.Sprintf("node:%s ns:%s podName:%s containerName:%s containerUID:%s containerRuntime:%s",
		cid.NodeName,
		cid.Namespace,
		cid.PodName,
		cid.ContainerName,
		cid.ContainerUID,
		cid.ContainerRuntime,
	)
}
