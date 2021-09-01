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

// Label ns/name/value for resource lookup
type Label struct {
	Namespace string `yaml:"namespace" json:"namespace"`
	Name      string `yaml:"name" json:"name"`
	Value     string `yaml:"value" json:"value"`
}

// Operator struct defines operator manifest for testing
type Operator struct {

	// Name is a required field, Name of the csv .
	Name string `yaml:"name" json:"name"`

	// Namespace is a required field , namespace is where the csv is installed.
	// If its all namespace then you can replace it with ALL_NAMESPACE TODO: add check for ALL_NAMESPACE
	Namespace string `yaml:"namespace" json:"namespace"`

	// Tests this is list of test that need to run against the operator.
	Tests []string `yaml:"tests" json:"tests"`

	// Subscription name is required field, Name of used subscription.
	SubscriptionName string `yaml:"subscriptionName" json:"subscriptionName"`
}

// TestConfiguration provides test related configuration
type TestConfiguration struct {
	// Custom Pod labels for discovering containers/pods under test
	TargetPodLabels []Label `yaml:"targetPodLabels,omitempty" json:"targetPodLabels,omitempty"`
	// TestTarget contains k8s resources that can be targeted by tests
	TestTarget `yaml:"testTarget" json:"testTarget"`
	// TestPartner contains the helper containers that can be used to facilitate tests
	Partner TestPartner `yaml:"testPartner" json:"testPartner"`
	// CertifiedContainerInfo is the list of container images to be checked for certification status.
	CertifiedContainerInfo []CertifiedContainerRequestInfo `yaml:"certifiedcontainerinfo,omitempty" json:"certifiedcontainerinfo,omitempty"`
	// CertifiedOperatorInfo is list of operator bundle names that are queried for certification status.
	CertifiedOperatorInfo []CertifiedOperatorRequestInfo `yaml:"certifiedoperatorinfo,omitempty" json:"certifiedoperatorinfo,omitempty"`
}

// TestPartner contains the helper containers that can be used to facilitate tests
type TestPartner struct {
	// ContainerConfigList is the list parter containers that facilitates tests
	ContainerConfigList []ContainerConfig `yaml:"partnerContainers" json:"partnerContainers"`
	// TestOrchestratorID is the id of the partner container for conducting connectivity tests
	TestOrchestratorID ContainerIdentifier `yaml:"testOrchestrator" json:"testOrchestrator"`
}

// TestTarget is a collection of resources under test
type TestTarget struct {
	// PodsUnderTest is the list of the pods that needs to be tested. Each entry is a single pod to be tested.
	PodsUnderTest []Pod `yaml:"podsUnderTest,omitempty" json:"podsUnderTest,omitempty"`
	// ContainerConfigList is the list of containers that needs to be tested.
	ContainerConfigList []ContainerConfig `yaml:"containersUnderTest" json:"containersUnderTest"`
	// ExcludeContainersFromConnectivityTests excludes specific containers from network connectivity tests.  This is particularly useful for containers that don't have ping available.
	ExcludeContainersFromConnectivityTests []ContainerIdentifier `yaml:"excludeContainersFromConnectivityTests" json:"excludeContainersFromConnectivityTests"`
	// Operator is the list of operator objects that needs to be tested.
	Operators []Operator `yaml:"operators,omitempty"  json:"operators,omitempty"`
}
