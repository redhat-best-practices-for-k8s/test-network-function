// Copyright (C) 2021 Red Hat, Inc.
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
	"github.com/test-network-function/test-network-function/test-network-function/common"
)

const (
	informativeResult = "informative"
	normativeResult   = "normative"
	url               = "http://test-network-function.com/testcases"
	versionOne        = "v1.0.0"
)

// TestCaseDescription describes a JUnit test case.
type TestCaseDescription struct {
	// Identifier is the unique test identifier.
	Identifier claim.Identifier `json:"identifier" yaml:"identifier"`

	// Description is a helpful description of the purpose of the test case.
	Description string `json:"description" yaml:"description"`

	// Remediation is an optional suggested remediation for passing the test.
	Remediation string `json:"remediation,omitempty" yaml:"remediation,omitempty"`

	// Type is the type of the test (i.e., normative).
	Type string `json:"type" yaml:"type"`
}

func formTestURL(suite, name string) string {
	return fmt.Sprintf("%s/%s/%s", url, suite, name)
}

var (
	// TestHostResourceIdentifier tests container best practices.
	TestHostResourceIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "host-resource"),
		Version: versionOne,
	}
	// TestContainerIsCertifiedIdentifier tests whether the container has passed Container Certification.
	TestContainerIsCertifiedIdentifier = claim.Identifier{
		Url:     formTestURL(common.AffiliatedCertTestKey, "container-is-certified"),
		Version: versionOne,
	}
	// TestExtractNodeInformationIdentifier is a test which extracts Node information.
	TestExtractNodeInformationIdentifier = claim.Identifier{
		Url:     formTestURL(common.DiagnosticTestKey, "extract-node-information"),
		Version: versionOne,
	}
	// TestListCniPluginsIdentifier retrieves list of CNI plugins.
	TestListCniPluginsIdentifier = claim.Identifier{
		Url:     formTestURL(common.DiagnosticTestKey, "list-cni-plugins"),
		Version: versionOne,
	}
	// TestNodesHwInfoIdentifier retrieves nodes HW info.
	TestNodesHwInfoIdentifier = claim.Identifier{
		Url:     formTestURL(common.DiagnosticTestKey, "nodes-hw-info"),
		Version: versionOne,
	}
	// TestHugepagesNotManuallyManipulated represents the test identifier testing hugepages have not been manipulated.
	TestHugepagesNotManuallyManipulated = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "hugepages-config"),
		Version: versionOne,
	}
	// TestICMPv4ConnectivityIdentifier tests icmpv4 connectivity.
	TestICMPv4ConnectivityIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "icmpv4-connectivity"),
		Version: versionOne,
	}
	// TestNamespaceBestPracticesIdentifier ensures the namespace has followed best namespace practices.
	TestNamespaceBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "namespace"),
		Version: versionOne,
	}
	// TestNonDefaultGracePeriodIdentifier tests best grace period practices.
	TestNonDefaultGracePeriodIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-termination-grace-period"),
		Version: versionOne,
	}
	// TestNonTaintedNodeKernelsIdentifier is the identifier for the test checking tainted nodes.
	TestNonTaintedNodeKernelsIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "tainted-node-kernel"),
		Version: versionOne,
	}
	// TestOperatorInstallStatusIdentifier tests Operator best practices.
	TestOperatorInstallStatusIdentifier = claim.Identifier{
		Url:     formTestURL(common.OperatorTestKey, "install-status"),
		Version: versionOne,
	}
	// TestOperatorIsCertifiedIdentifier tests that an Operator has passed Operator certification.
	TestOperatorIsCertifiedIdentifier = claim.Identifier{
		Url:     formTestURL(common.AffiliatedCertTestKey, "operator-is-certified"),
		Version: versionOne,
	}
	// TestOperatorIsInstalledViaOLMIdentifier tests that an Operator is installed via OLM.
	TestOperatorIsInstalledViaOLMIdentifier = claim.Identifier{
		Url:     formTestURL(common.OperatorTestKey, "install-source"),
		Version: versionOne,
	}
	// TestPodNodeSelectorAndAffinityBestPractices is the test ensuring nodeSelector and nodeAffinity are not used by a
	// Pod.
	TestPodNodeSelectorAndAffinityBestPractices = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-scheduling"),
		Version: versionOne,
	}
	// TestPodHighAvailabilityBestPractices is the test ensuring podAntiAffinity are used by a
	// Pod when pod replica # are great than 1
	TestPodHighAvailabilityBestPractices = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-high-availability"),
		Version: versionOne,
	}

	// TestPodClusterRoleBindingsBestPracticesIdentifier ensures Pod crb best practices.
	TestPodClusterRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "cluster-role-bindings"),
		Version: versionOne,
	}
	// TestPodDeploymentBestPracticesIdentifier ensures a CNF follows best Deployment practices.
	TestPodDeploymentBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-owner-type"),
		Version: versionOne,
	}
	// TestPodRecreationIdentifier ensures recreation best practices.
	TestPodRecreationIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-recreation"),
		Version: versionOne,
	}
	// TestPodRoleBindingsBestPracticesIdentifier represents rb best practices.
	TestPodRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-role-bindings"),
		Version: versionOne,
	}
	// TestPodServiceAccountBestPracticesIdentifier tests Pod SA best practices.
	TestPodServiceAccountBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-service-account"),
		Version: versionOne,
	}
	// TestServicesDoNotUseNodeportsIdentifier ensures Services don't utilize NodePorts.
	TestServicesDoNotUseNodeportsIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "service-type"),
		Version: versionOne,
	}
	// TestUnalteredBaseImageIdentifier ensures the base image is not altered.
	TestUnalteredBaseImageIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "base-image"),
		Version: versionOne,
	}
	// TestUnalteredStartupBootParamsIdentifier ensures startup boot params are not altered.
	TestUnalteredStartupBootParamsIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "boot-params"),
		Version: versionOne,
	}
	// TestLoggingIdentifier ensures stderr/stdout are used
	TestLoggingIdentifier = claim.Identifier{
		Url:     formTestURL(common.ObservabilityTestKey, "container-logging"),
		Version: versionOne,
	}
	// TestShudtownIdentifier ensures pre-stop lifecycle is defined
	TestShudtownIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "container-shutdown"),
		Version: versionOne,
	}
	// TestSysctlConfigsIdentifier ensures that the node's sysctl configs are consistent with the MachineConfig CR
	TestSysctlConfigsIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "sysctl-config"),
		Version: versionOne,
	}
	// TestScalingIdentifier ensures deployment scale in/out operations work correctly.
	TestScalingIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "scaling"),
		Version: versionOne,
	}
	// TestIsRedHatReleaseIdentifier ensures platform is defined
	TestIsRedHatReleaseIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "isredhat-release"),
		Version: versionOne,
	}
)

func formDescription(identifier claim.Identifier, description string) string {
	return fmt.Sprintf("%s %s", identifier.Url, description)
}

// Catalog is the JUnit testcase catalog of tests.
var Catalog = map[claim.Identifier]TestCaseDescription{

	TestHostResourceIdentifier: {
		Identifier: TestHostResourceIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that each Pod in the CNF abides by the suggested best practices listed in the test description.  In some rare
cases, not all best practices can be followed.  For example, some CNFs may be required to run as root.  Such exceptions
should be handled on a case-by-case basis, and should provide a proper justification as to why the best practice(s)
cannot be followed.`,
		Description: formDescription(TestHostResourceIdentifier,
			`tests several aspects of CNF best practices, including:
1. The Pod does not have access to Host Node Networking.
2. The Pod does not have access to Host Node Ports.
3. The Pod cannot access Host Node IPC space.
4. The Pod cannot access Host Node PID space.
5. The Pod is not granted NET_ADMIN SCC.
6. The Pod is not granted SYS_ADMIN SCC.
7. The Pod does not run as root.
8. The Pod does not allow privileged escalation.
`),
	},

	TestContainerIsCertifiedIdentifier: {
		Identifier:  TestContainerIsCertifiedIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your container has passed the Red Hat Container Certification Program (CCP).`,
		Description: formDescription(TestContainerIsCertifiedIdentifier,
			`tests whether container images have passed the Red Hat Container Certification Program (CCP).`),
	},

	TestExtractNodeInformationIdentifier: {
		Identifier: TestExtractNodeInformationIdentifier,
		Type:       informativeResult,
		Description: formDescription(TestExtractNodeInformationIdentifier,
			`extracts informational information about the cluster.`),
	},

	TestHugepagesNotManuallyManipulated: {
		Identifier: TestHugepagesNotManuallyManipulated,
		Type:       normativeResult,
		Remediation: `HugePage settings should be configured either directly through the MachineConfigOperator or indirectly using the
PeformanceAddonOperator.  This ensures that OpenShift is aware of the special MachineConfig requirements, and can
provision your CNF on a Node that is part of the corresponding MachineConfigSet.  Avoid making changes directly to an
underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings.`,
		Description: formDescription(TestHugepagesNotManuallyManipulated,
			`checks to see that HugePage settings have been configured through MachineConfig, and not manually on the
underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First,
the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled
for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if
they are the same.`),
	},

	TestICMPv4ConnectivityIdentifier: {
		Identifier: TestICMPv4ConnectivityIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that the CNF is able to communicate via the Default OpenShift network.  In some rare cases,
CNFs may require routing table changes in order to communicate over the Default network.  In other cases, if the
Container base image does not provide the "ip" or "ping" binaries, this test may not be applicable.  For instructions on
how to exclude a particular container from ICMPv4 connectivity tests, consult:
[README.md](https://github.com/test-network-function/test-network-function#issue-161-some-containers-under-test-do-nto-contain-ping-or-ip-binary-utilities).`,
		Description: formDescription(TestICMPv4ConnectivityIdentifier,
			`checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This
test case requires the Deployment of the
[CNF Certification Test Partner](https://github.com/test-network-function/cnf-certification-test-partner/blob/main/test-partner/partner.yaml).
The test ensures that all CNF containers respond to ICMPv4 requests from the Partner Pod, and vice-versa.
`),
	},

	TestNamespaceBestPracticesIdentifier: {
		Identifier: TestNamespaceBestPracticesIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that your CNF utilizes a CNF-specific namespace.  Additionally, the CNF-specific namespace
should not start with "openshift-", except in rare cases.`,
		Description: formDescription(TestNamespaceBestPracticesIdentifier,
			`tests that CNFs utilize a CNF-specific namespace, and that the namespace does not start with "openshift-".
OpenShift may host a variety of CNF and software applications, and multi-tenancy of such applications is supported
through namespaces.  As such, each CNF should be a good neighbor, and utilize an appropriate, unique namespace.`),
	},

	TestNonDefaultGracePeriodIdentifier: {
		Identifier: TestNonDefaultGracePeriodIdentifier,
		Type:       informativeResult,
		Remediation: `Choose a terminationGracePeriod that is appropriate for your given CNF.  If the default (30s) is appropriate, then feel
free to ignore this informative message.  This test is meant to raise awareness around how Pods are terminated, and to
suggest that a CNF is configured based on its requirements.  In addition to a terminationGracePeriod, consider utilizing
a termination hook in the case that your application requires special shutdown instructions.`,
		Description: formDescription(TestNonDefaultGracePeriodIdentifier,
			`tests whether the terminationGracePeriod is CNF-specific, or if the default (30s) is utilized.  This test is
informative, and will not affect CNF Certification.  In many cases, the default terminationGracePeriod is perfectly
acceptable for a CNF.`),
	},

	TestNonTaintedNodeKernelsIdentifier: {
		Identifier: TestNonTaintedNodeKernelsIdentifier,
		Type:       normativeResult,
		Remediation: `Test failure indicates that the underlying Node's' kernel is tainted.  Ensure that you have not altered underlying
Node(s) kernels in order to run the CNF.`,
		Description: formDescription(TestNonTaintedNodeKernelsIdentifier,
			`ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important
to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have
the same hacks.'`),
	},

	TestOperatorInstallStatusIdentifier: {
		Identifier:  TestOperatorInstallStatusIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your Operator abides by the Operator Best Practices mentioned in the description.`,
		Description: formDescription(TestOperatorInstallStatusIdentifier,
			`Ensures that CNF Operators abide by best practices.  The following is tested:
1. The Operator CSV reports "Installed" status.
2. TODO: Describe operator scc check.`),
	},

	TestOperatorIsCertifiedIdentifier: {
		Identifier:  TestOperatorIsCertifiedIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).`,
		Description: formDescription(TestOperatorIsCertifiedIdentifier,
			`tests whether CNF Operators have passed the Red Hat Operator Certification Program (OCP).`),
	},

	TestOperatorIsInstalledViaOLMIdentifier: {
		Identifier:  TestOperatorIsInstalledViaOLMIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your Operator is installed via OLM.`,
		Description: formDescription(TestOperatorIsInstalledViaOLMIdentifier,
			`tests whether a CNF Operator is installed via OLM.`),
	},

	TestPodNodeSelectorAndAffinityBestPractices: {
		Identifier: TestPodNodeSelectorAndAffinityBestPractices,
		Type:       informativeResult,
		Remediation: `In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity.  However, there are
cases in which CNFs require specialized hardware specific to a particular class of Node.  As such, this test is purely
informative, and will not prevent a CNF from being certified. However, one should have an appropriate justification as
to why nodeSelector and/or nodeAffinity is utilized by a CNF.`,
		Description: formDescription(TestPodNodeSelectorAndAffinityBestPractices,
			`ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for
instantiation on any underlying Node.`),
	},

	TestPodHighAvailabilityBestPractices: {
		Identifier:  TestPodHighAvailabilityBestPractices,
		Type:        informativeResult,
		Remediation: `In high availability cases, Pod podAntiAffinity rule should be specified for pod scheduling and pod replica value is set to more than 1 .`,
		Description: formDescription(TestPodHighAvailabilityBestPractices,
			`ensures that CNF Pods specify podAntiAffinity rules and replica value is set to more than 1.`),
	},

	TestPodClusterRoleBindingsBestPracticesIdentifier: {
		Identifier: TestPodClusterRoleBindingsBestPracticesIdentifier,
		Type:       normativeResult,
		Remediation: `In most cases, Pod's should not have ClusterRoleBindings.  The suggested remediation is to remove the need for
ClusterRoleBindings, if possible.`,
		Description: formDescription(TestPodClusterRoleBindingsBestPracticesIdentifier,
			`tests that a Pod does not specify ClusterRoleBindings.`),
	},

	TestPodDeploymentBestPracticesIdentifier: {
		Identifier:  TestPodDeploymentBestPracticesIdentifier,
		Type:        normativeResult,
		Remediation: `Deploy the CNF using DaemonSet or ReplicaSet.`,
		Description: formDescription(TestPodDeploymentBestPracticesIdentifier,
			`tests that CNF Pod(s) are deployed as part of a ReplicaSet(s).`),
	},

	TestPodRoleBindingsBestPracticesIdentifier: {
		Identifier:  TestPodRoleBindingsBestPracticesIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure the CNF is not configured to use RoleBinding(s) in a non-CNF Namespace.`,
		Description: formDescription(TestPodRoleBindingsBestPracticesIdentifier,
			`ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.`),
	},

	TestPodServiceAccountBestPracticesIdentifier: {
		Identifier:  TestPodServiceAccountBestPracticesIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that the each CNF Pod is configured to use a valid Service Account`,
		Description: formDescription(TestPodServiceAccountBestPracticesIdentifier,
			`tests that each CNF Pod utilizes a valid Service Account.`),
	},

	TestServicesDoNotUseNodeportsIdentifier: {
		Identifier:  TestServicesDoNotUseNodeportsIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure Services are not configured to not use NodePort(s).`,
		Description: formDescription(TestServicesDoNotUseNodeportsIdentifier,
			`tests that each CNF Service does not utilize NodePort(s).`),
	},

	TestUnalteredBaseImageIdentifier: {
		Identifier: TestUnalteredBaseImageIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that Container applications do not modify the Container Base Image.  In particular, ensure that the following
directories are not modified:
1) /var/lib/rpm
2) /var/lib/dpkg
3) /bin
4) /sbin
5) /lib
6) /lib64
7) /usr/bin
8) /usr/sbin
9) /usr/lib
10) /usr/lib64
Ensure that all required binaries are built directly into the container image, and are not installed post startup.`,
		Description: formDescription(TestUnalteredBaseImageIdentifier,
			`ensures that the Container Base Image is not altered post-startup.  This test is a heuristic, and ensures
that there are no changes to the following directories:
1) /var/lib/rpm
2) /var/lib/dpkg
3) /bin
4) /sbin
5) /lib
6) /lib64
7) /usr/bin
8) /usr/sbin
9) /usr/lib
10) /usr/lib64`),
	},

	TestUnalteredStartupBootParamsIdentifier: {
		Identifier: TestUnalteredStartupBootParamsIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the
PerfromanceAddonOperator.  Boot parameters should not be changed directly through the Node, as OpenShift should manage
the changes for you.`,
		Description: formDescription(TestUnalteredStartupBootParamsIdentifier,
			`tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.`),
	},
	TestListCniPluginsIdentifier: {
		Identifier:  TestListCniPluginsIdentifier,
		Type:        normativeResult,
		Remediation: "",
		Description: formDescription(TestListCniPluginsIdentifier,
			`lists CNI plugins`),
	},
	TestNodesHwInfoIdentifier: {
		Identifier:  TestNodesHwInfoIdentifier,
		Type:        normativeResult,
		Remediation: "",
		Description: formDescription(TestNodesHwInfoIdentifier,
			`list nodes HW info`),
	},

	TestShudtownIdentifier: {
		Identifier: TestShudtownIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestShudtownIdentifier,
			`Ensure that the containers lifecycle pre-stop management feature is configured.`),
		Remediation: `
		It's considered best-practices to define prestop for proper management of container lifecycle.
		The prestop can be used to gracefully stop the container and clean resources (e.g., DB connexion).
		
		The prestop can be configured using :
		 1) Exec : executes the supplied command inside the container
		 2) HTTP : executes HTTP request against the specified endpoint.
		
		When defined. K8s will handle shutdown of the container using the following:
		1) K8s first execute the preStop hook inside the container.
		2) K8s will wait for a grace perdiod.
		3) K8s will clean the remaining processes using KILL signal.		
			`,
	},
	TestPodRecreationIdentifier: {
		Identifier: TestPodRecreationIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestPodRecreationIdentifier,
			`tests that a CNF is configured to support High Availability.  
			First, this test cordons and drains a Node that hosts the CNF Pod.  
			Next, the test ensures that OpenShift can re-instantiate the Pod on another Node, 
			and that the actual replica count matches the desired replica count.`),
		Remediation: `Ensure that CNF Pod(s) utilize a configuration that supports High Availability.  
			Additionally, ensure that there are available Nodes in the OpenShift cluster that can be utilized in the event that a host Node fails.`,
	},
	TestSysctlConfigsIdentifier: {
		Identifier: TestSysctlConfigsIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestPodRecreationIdentifier,
			`tests that no one has changed the node's sysctl configs after the node
			was created, the tests works by checking if the sysctl configs are consistent with the
			MachineConfig CR which defines how the node should be configured`),
		Remediation: `You should recreate the node or change the sysctls, recreating is recommended because there might be other unknown changes`,
	},
	TestScalingIdentifier: {
		Identifier: TestScalingIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestScalingIdentifier,
			`tests that CNF deployments support scale in/out operations. 
			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the 
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.`),
		Remediation: `Make sure CNF deployments/replica sets can scale in/out successfully.`,
	},
	TestIsRedHatReleaseIdentifier: {
		Identifier: TestIsRedHatReleaseIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestIsRedHatReleaseIdentifier,
			`This test is meant for the end user to figure out what's wrong with the test case and some hints on how to fix it.`),
		Remediation: `You should recreate the nodes, recreating is recommended because there might be other unknown changes`,
	},
}
