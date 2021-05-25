// Copyright (C) 2021 Red Hat, Inc.
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

package identifiers

import (
	"fmt"

	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

const (
	genericSuite = "generic"
	url          = "http://test-network-function.com/testcases"
	versionOne   = "v1.0.0"
)

func formTestURL(name string) string {
	return fmt.Sprintf("%s/%s/%s", url, genericSuite, name)
}

var (
	// TestHugepagesNotManuallyManipulated represents the test identifier testing hugepages have not been manipulated.
	TestHugepagesNotManuallyManipulated = claim.Identifier{
		Url:     formTestURL("hugepages-not-manually-manipulated"),
		Version: versionOne,
	}
	// TestICMPv4ConnectivityIdentifier tests icmpv4 connectivity.
	TestICMPv4ConnectivityIdentifier = claim.Identifier{
		Url:     formTestURL("icmpv4-connectivity"),
		Version: versionOne,
	}
	// TestNamespaceBestPracticesIdentifier ensures the namespace has followed best namespace practices.
	TestNamespaceBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL("namespace-best-practices"),
		Version: versionOne,
	}
	// TestNonDefaultGracePeriodIdentifier tests best grace period practices.
	TestNonDefaultGracePeriodIdentifier = claim.Identifier{
		Url:     formTestURL("non-default-grace-period"),
		Version: versionOne,
	}
	// TestNonTaintedNodeKernelsIdentifier is the identifier for the test checking tainted nodes.
	TestNonTaintedNodeKernelsIdentifier = claim.Identifier{
		Url:     formTestURL("non-tainted-node-kernel"),
		Version: versionOne,
	}
	// TestPodNodeSelectorAndAffinityBestPractices is the test ensuring nodeSelector and nodeAffinity are not used by a
	// Pod.
	TestPodNodeSelectorAndAffinityBestPractices = claim.Identifier{
		Url:     formTestURL("pod-node-selector-node-affinity-best-practices"),
		Version: versionOne,
	}
	// TestPodClusterRoleBindingsBestPracticesIdentifier ensures Pod crb best practices.
	TestPodClusterRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL("pod-cluster-role-bindings-best-practices"),
		Version: versionOne,
	}
	// TestPodDeploymentBestPracticesIdentifier ensures Pod rb best practices.
	TestPodDeploymentBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL("pod-deployment-best-practices"),
		Version: versionOne,
	}
	// TestPodRecreationIdentifier ensures recreation best practices.
	TestPodRecreationIdentifier = claim.Identifier{
		Url:     formTestURL("pod-recreation"),
		Version: versionOne,
	}
	// TestPodRoleBindingsBestPracticesIdentifier represents rb best practices.
	TestPodRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL("pod-role-bindings-best-practices"),
		Version: versionOne,
	}
	// TestPodServiceAccountBestPracticesIdentifier tests Pod SA best practices.
	TestPodServiceAccountBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL("pod-service-account-best-practices"),
		Version: versionOne,
	}
	// TestServicesDoNotUseNodeportsIdentifier ensures Services don't utilize NodePorts.
	TestServicesDoNotUseNodeportsIdentifier = claim.Identifier{
		Url:     formTestURL("services-do-not-use-nodeports"),
		Version: versionOne,
	}
	// TestUnalteredBaseImageIdentifier ensures the base image is not altered.
	TestUnalteredBaseImageIdentifier = claim.Identifier{
		Url:     formTestURL("unaltered-base-image"),
		Version: versionOne,
	}
	// TestUnalteredStartupBootParamsIdentifier ensures startup boot params are not altered.
	TestUnalteredStartupBootParamsIdentifier = claim.Identifier{
		Url:     formTestURL("unaltered-startup-boot-params"),
		Version: versionOne,
	}
)
