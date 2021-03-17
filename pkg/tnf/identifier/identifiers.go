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

package identifier

import "github.com/test-network-function/test-network-function/pkg/tnf/dependencies"

const (
	hostnameIdentifierURL = "http://test-network-function.com/tests/hostname"
	ipAddrIdentifierURL   = "http://test-network-function.com/tests/ipaddr"
	nodesIdentifierURL    = "http://test-network-function.com/tests/nodes"
	operatorIdentifierURL = "http://test-network-function.com/tests/operator"
	pingIdentifierURL     = "http://test-network-function.com/tests/ping"
	podIdentifierURL      = "http://test-network-function.com/tests/container/pod"
	versionIdentifierURL  = "http://test-network-function.com/tests/generic/version"

	versionOne = "v1.0.0"
)

const (
	// Normative is the test type used for a test that returns normative results.
	Normative = "normative"
	// TODO: Informative = "informative" once we have informative tests.
)

// TestCatalogEntry is a container for required test facets.
type TestCatalogEntry struct {

	// Identifier is the unique test identifier.
	Identifier Identifier `json:"identifier" yaml:"identifier"`

	// Description is a helpful description of the purpose of the test.
	Description string `json:"description" yaml:"description"`

	// Type is the type of the test (i.e., normative).
	Type string `json:"type" yaml:"type"`

	// IntrusionSettings is used to specify test intrusion behavior into a target system.
	IntrusionSettings IntrusionSettings `json:"intrusionSettings" yaml:"intrusionSettings"`

	// BinaryDependencies tracks the needed binaries to complete tests, such as `ping`.
	BinaryDependencies []string `json:"binaryDependencies" yaml:"binaryDependencies"`
}

// IntrusionSettings is used to specify test intrusion behavior into a target system.
type IntrusionSettings struct {
	// ModifiesSystem records whether the test makes changes to target systems.
	ModifiesSystem bool `json:"modifiesSystem" yaml:"modifiesSystem"`

	// ModificationIsPersistent records whether the test makes a modification to the system that persists after the test
	// completes.  This is not always negative, and could involve something like setting up a tunnel that is used in
	// future tests.
	ModificationIsPersistent bool `json:"modificationIsPersistent" yaml:"modificationIsPersistent"`
}

// Catalog is the test catalog.
var Catalog = map[string]TestCatalogEntry{
	hostnameIdentifierURL: {
		Identifier:  HostnameIdentifier,
		Description: "A generic test used to check the hostname of a target machine/container.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.HostnameBinaryName,
		},
	},
	ipAddrIdentifierURL: {
		Identifier:  IPAddrIdentifier,
		Description: "A generic test used to derive the default network interface IP address of a target container.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.IPBinaryName,
		},
	},
	nodesIdentifierURL: {
		Identifier:  NodesIdentifier,
		Description: "Polls the state of the OpenShift cluster nodes using \"oc get nodes -o json\".",
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	operatorIdentifierURL: {
		Identifier: OperatorIdentifier,
		Description: "An operator-specific test used to exercise the behavior of a given operator.  In the current " +
			"offering, we check if the operator ClusterServiceVersion (CSV) is installed properly.  A CSV is a YAML " +
			"manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running " +
			"the Operator.",
		Type: Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.JqBinaryName,
			dependencies.OcBinaryName,
		},
	},
	pingIdentifierURL: {
		Identifier:  PingIdentifier,
		Description: "A generic test used to test ICMP connectivity from a source machine/container to a target destination.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.PingBinaryName,
		},
	},
	podIdentifierURL: {
		Identifier:  PodIdentifier,
		Description: "A container-specific test suite used to verify various aspects of the underlying container.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.JqBinaryName,
			dependencies.OcBinaryName,
		},
	},
	versionIdentifierURL: {
		Identifier:  VersionIdentifier,
		Description: "A generic test used to determine if a target container/machine is based on RHEL.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
		},
	},
}

// HostnameIdentifier is the Identifier used to represent the generic hostname test case.
var HostnameIdentifier = Identifier{
	URL:             hostnameIdentifierURL,
	SemanticVersion: versionOne,
}

// IPAddrIdentifier is the Identifier used to represent the generic IP Addr test case.
var IPAddrIdentifier = Identifier{
	URL:             ipAddrIdentifierURL,
	SemanticVersion: versionOne,
}

// NodesIdentifier is the Identifier used to represent the nodes test case.
var NodesIdentifier = Identifier{
	URL:             nodesIdentifierURL,
	SemanticVersion: versionOne,
}

// OperatorIdentifier is the Identifier used to represent the operator-specific test suite.
var OperatorIdentifier = Identifier{
	URL:             operatorIdentifierURL,
	SemanticVersion: versionOne,
}

// PingIdentifier is the Identifier used to represent the generic Ping test.
var PingIdentifier = Identifier{
	URL:             pingIdentifierURL,
	SemanticVersion: versionOne,
}

// PodIdentifier is the Identifier used to represent the container-specific test suite.
var PodIdentifier = Identifier{
	URL:             podIdentifierURL,
	SemanticVersion: versionOne,
}

// VersionIdentifier is the Identifier used to represent the generic container base image test.
var VersionIdentifier = Identifier{
	URL:             versionIdentifierURL,
	SemanticVersion: versionOne,
}
