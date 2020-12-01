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

const (
	casaNRFCheckRegistrationIdentifierURL = "http://test-network-function.com/tests/casa/nrf/checkregistration"
	casaNRFIDIdentifierURL                = "http://test-network-function.com/tests/casa/nrf/id"
	hostnameIdentifierURL                 = "http://test-network-function.com/tests/hostname"
	ipAddrIdentifierURL                   = "http://test-network-function.com/tests/ipaddr"
	operatorIdentifierURL                 = "http://test-network-function.com/tests/operator"
	pingIdentifierURL                     = "http://test-network-function.com/tests/ping"
	podIdentifierURL                      = "http://test-network-function.com/tests/container/pod"
	versionIdentifierURL                  = "http://test-network-function.com/tests/generic/version"
	versionOne                            = "v1.0.0"
)

const (
	// Normative is the test type used for a test that returns normative results.
	Normative = "normative"
	// TODO: Informative = "informative" once we have informative tests.
)

// TestCatalogEntry is a container for required test facets.
type TestCatalogEntry struct {

	// Identifier is the unique test identifier.
	Identifier Identifier

	// Description is a helpful description of the purpose of the test.
	Description string

	// Type is the type of the test (i.e., normative).
	Type string
}

// Catalog is the test catalog.
var Catalog = map[string]TestCatalogEntry{
	casaNRFIDIdentifierURL: {
		Identifier:  CasaNRFIDIdentifier,
		Description: "A Casa cnf-specific test which checks for the existence of the AMF and SMF CNFs.  The UUIDs are gathered and stored by introspecting the \"nfregistrations.mgmt.casa.io\" Custom Resource.",
		Type:        Normative,
	},
	casaNRFCheckRegistrationIdentifierURL: {
		Identifier:  CasaNRFRegistrationIdentifier,
		Description: "A Casa cnf-specific test which checks the Registration status of the AMF and SMF from the NRF.  This is done by making sure the \"nfStatus\" field in the \"nfregistrations.mgmt.casa.io\" Custom Resource reports as \"REGISTERED\"",
		Type:        Normative,
	},
	hostnameIdentifierURL: {
		Identifier:  HostnameIdentifier,
		Description: "A generic test used to check the hostname of a target machine/container.",
		Type:        Normative,
	},
	ipAddrIdentifierURL: {
		Identifier:  IPAddrIdentifier,
		Description: "A generic test used to derive the default network interface IP address of a target container.",
		Type:        Normative,
	},
	operatorIdentifierURL: {
		Identifier:  OperatorIdentifier,
		Description: "An operator-specific test used to exercise the behavior of a given operator.  Currently, this test just checks that the Custom Resource Definition (CRD) of a resource is properly installed.",
		Type:        Normative,
	},
	pingIdentifierURL: {
		Identifier:  PingIdentifier,
		Description: "A generic test used to test ICMP connectivity from a source machine/container to a target destination.",
		Type:        Normative,
	},
	podIdentifierURL: {
		Identifier:  PodIdentifier,
		Description: "A container-specific test suite used to verify various aspects of the underlying container.",
		Type:        Normative,
	},
	versionIdentifierURL: {
		Identifier:  VersionIdentifier,
		Description: "A generic test used to determine if a target container/machine is based on RHEL.",
		Type:        Normative,
	},
}

// CasaNRFIDIdentifier is the Identifier used to represent the Casa CNF-specific test case which checks AMF/SMF uuids.
var CasaNRFIDIdentifier = Identifier{
	URL:             casaNRFIDIdentifierURL,
	SemanticVersion: versionOne,
}

// CasaNRFRegistrationIdentifier is the Identifier used to represent the CASA CNF-specific registration test case.
var CasaNRFRegistrationIdentifier = Identifier{
	URL:             casaNRFCheckRegistrationIdentifierURL,
	SemanticVersion: versionOne,
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
