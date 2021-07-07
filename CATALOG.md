# test-network-function test case catalog

test-network-function contains a variety of `Test Cases`, as well as `Test Case Building Blocks`.
* Test Cases:  Traditional JUnit testcases, which are specified internally using `Ginkgo.It`.  Test cases often utilize several Test Case Building Blocks.
* Test Case Building Blocks:  Self-contained building blocks, which perform a small task in the context of `oc`, `ssh`, `shell`, or some other `Expecter`.
## Test Case Catalog

Test Cases are the specifications used to perform a meaningful test.  Test cases may run once, or several times against several targets.  CNF Certification includes a number of normative and informative tests to ensure CNFs follow best practices.  Here is the list of available Test Cases:
### http://test-network-function.com/testcases/operator/operator-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/operator-best-practices Ensures that CNF Operators abide by best practices.  The following is tested: 1. The Operator CSV reports "Installed" status. 2. The Operator is installed using through an Operator subscription catalog.
Result Type|normative
Suggested Remediation|Ensure that your Operator abides by the Operator Best Practices mentioned in the description.
### http://test-network-function.com/testcases/generic/pod-role-bindings-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/pod-role-bindings-best-practices ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.
Result Type|normative
Suggested Remediation|Ensure the CNF is not configured to use RoleBinding(s) in a non-CNF Namespace.
### http://test-network-function.com/testcases/generic/namespace-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/namespace-best-practices tests that CNFs utilize a CNF-specific namespace, and that the namespace does not start with "openshift-". OpenShift may host a variety of CNF and software applications, and multi-tenancy of such applications is supported through namespaces.  As such, each CNF should be a good neighbor, and utilize an appropriate, unique namespace.
Result Type|normative
Suggested Remediation|Ensure that your CNF utilizes a CNF-specific namespace.  Additionally, the CNF-specific namespace should not start with "openshift-", except in rare cases.
### http://test-network-function.com/testcases/diagnostic/extract-node-information

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/diagnostic/extract-node-information extracts informational information about the cluster.
Result Type|informative
Suggested Remediation|
### http://test-network-function.com/testcases/generic/icmpv4-connectivity

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/icmpv4-connectivity checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This test case requires the Deployment of the [CNF Certification Test Partner](https://github.com/test-network-function/cnf-certification-test-partner/blob/main/test-partner/partner.yaml). The test ensures that all CNF containers respond to ICMPv4 requests from the Partner Pod, and vice-versa. 
Result Type|normative
Suggested Remediation|Ensure that the CNF is able to communicate via the Default OpenShift network.  In some rare cases, CNFs may require routing table changes in order to communicate over the Default network.  In other cases, if the Container base image does not provide the "ip" or "ping" binaries, this test may not be applicable.  For instructions on how to exclude a particular container from ICMPv4 connectivity tests, consult: [README.md](https://github.com/test-network-function/test-network-function#issue-161-some-containers-under-test-do-nto-contain-ping-or-ip-binary-utilities).
### http://test-network-function.com/testcases/generic/non-tainted-node-kernel

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/non-tainted-node-kernel ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have the same hacks.'
Result Type|normative
Suggested Remediation|Test failure indicates that the underlying Node's' kernel is tainted.  Ensure that you have not altered underlying Node(s) kernels in order to run the CNF.
### http://test-network-function.com/testcases/operator/operator-is-certified

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/operator-is-certified tests whether CNF Operators have passed the Red Hat Operator Certification Program (OCP).
Result Type|normative
Suggested Remediation|Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).
### http://test-network-function.com/testcases/generic/pod-node-selector-node-affinity-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/pod-node-selector-node-affinity-best-practices ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for instantiation on any underlying Node.
Result Type|informative
Suggested Remediation|In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity.  However, there are cases in which CNFs require specialized hardware specific to a particular class of Node.  As such, this test is purely informative,  and will not prevent a CNF from being certified.  However, one should have an appropriate justification as to why nodeSelector and/or nodeAffinity is utilized by a CNF.
### http://test-network-function.com/testcases/container/container-is-certified

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/container/container-is-certified tests whether container images have passed the Red Hat Container Certification Program (CCP).
Result Type|normative
Suggested Remediation|Ensure that your container has passed the Red Hat Container Certification Program (CCP).
### http://test-network-function.com/testcases/operator/operator-is-installed-via-olm

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/operator/operator-is-installed-via-olm tests whether a CNF Operator is installed via OLM.
Result Type|normative
Suggested Remediation|Ensure that your Operator is installed via OLM.
### http://test-network-function.com/testcases/generic/pod-deployment-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/pod-deployment-best-practices tests that CNF Pod(s) are deployed as part of either DaemonSet(s) or a ReplicaSet(s).
Result Type|normative
Suggested Remediation|Deploy the CNF using DaemonSet or ReplicaSet.
### http://test-network-function.com/testcases/generic/services-do-not-use-nodeports

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/services-do-not-use-nodeports tests that each CNF Service does not utilize NodePort(s).
Result Type|normative
Suggested Remediation|Ensure Services are not configured to not use NodePort(s).
### http://test-network-function.com/testcases/generic/list-cni-plugins

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/list-cni-plugins lists CNI plugins
Result Type|normative
Suggested Remediation|
### http://test-network-function.com/testcases/generic/non-default-grace-period

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/non-default-grace-period tests whether the terminationGracePeriod is CNF-specific, or if the default (30s) is utilized.  This test is informative, and will not affect CNF Certification.  In many cases, the default terminationGracePeriod is perfectly acceptable for a CNF.
Result Type|informative
Suggested Remediation|Choose a terminationGracePeriod that is appropriate for your given CNF.  If the default (30s) is appropriate, then feel free to ignore this informative message.  This test is meant to raise awareness around how Pods are terminated, and to suggest that a CNF is configured based on its requirements.  In addition to a terminationGracePeriod, consider utilizing a termination hook in the case that your application requires special shutdown instructions.
### http://test-network-function.com/testcases/generic/hugepages-not-manually-manipulated

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/hugepages-not-manually-manipulated checks to see that HugePage settings have been configured through MachineConfig, and not manually on the underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First, the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if they are the same.
Result Type|normative
Suggested Remediation|HugePage settings should be configured either directly through the MachineConfigOperator or indirectly using the PeformanceAddonOperator.  This ensures that OpenShift is aware of the special MachineConfig requirements, and can provision your CNF on a Node that is part of the corresponding MachineConfigSet.  Avoid making changes directly to an underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings.
### http://test-network-function.com/testcases/generic/pod-cluster-role-bindings-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/pod-cluster-role-bindings-best-practices tests that a Pod does not specify ClusterRoleBindings.
Result Type|normative
Suggested Remediation|In most cases, Pod's should not have ClusterRoleBindings.  The suggested remediation is to remove the need for ClusterRoleBindings, if possible.
### http://test-network-function.com/testcases/generic/pod-service-account-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/pod-service-account-best-practices tests that each CNF Pod utilizes a valid Service Account.
Result Type|normative
Suggested Remediation|Ensure that the each CNF Pod is configured to use a valid Service Account
### http://test-network-function.com/testcases/generic/unaltered-base-image

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/unaltered-base-image ensures that the Container Base Image is not altered post-startup.  This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64
Result Type|normative
Suggested Remediation|Ensure that Container applications do not modify the Container Base Image.  In particular, ensure that the following directories are not modified: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64 Ensure that all required binaries are built directly into the container image, and are not installed post startup.
### http://test-network-function.com/testcases/generic/unaltered-startup-boot-params

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/unaltered-startup-boot-params tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.
Result Type|normative
Suggested Remediation|Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the PerfromanceAddonOperator.  Boot parameters should not be changed directly through the Node, as OpenShift should manage the changes for you.
### http://test-network-function.com/testcases/generic/nodes-hw-info

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/generic/nodes-hw-info list nodes HW info
Result Type|normative
Suggested Remediation|
### http://test-network-function.com/testcases/container/container-best-practices

Property|Description
---|---
Version|v1.0.0
Description|http://test-network-function.com/testcases/container/container-best-practices tests several aspects of CNF best practices, including: 1. The Pod does not have access to Host Node Networking. 2. The Pod does not have access to Host Node Ports. 3. The Pod cannot access Host Node IPC space. 4. The Pod cannot access Host Node PID space. 5. The Pod is not granted NET_ADMIN SCC. 6. The Pod is not granted SYS_ADMIN SCC. 7. The Pod does not run as root. 8. The Pod does not allow privileged escalation. 
Result Type|normative
Suggested Remediation|Ensure that each Pod in the CNF abides by the suggested best practices listed in the test description.  In some rare cases, not all best practices can be followed.  For example, some CNFs may be required to run as root.  Such exceptions should be handled on a case-by-case basis, and should provide a proper justification as to why the best practice(s) cannot be followed.


## Test Case Building Blocks Catalog

A number of Test Case Building Blocks, or `tnf.Test`s, are included out of the box.  This is a summary of the available implementations:
### http://test-network-function.com/tests/serviceaccount
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to extract the CNF pod's ServiceAccount name.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`

### http://test-network-function.com/tests/nodeport
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to test services of CNF pod's namespace.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### http://test-network-function.com/tests/deploymentsnodes
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to read node names of pods owned by deployments in namespace
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### http://test-network-function.com/tests/nodemcname
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get a node's current mc
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### http://test-network-function.com/tests/nodes
Property|Description
---|---
Version|v1.0.0
Description|Polls the state of the OpenShift cluster nodes using "oc get nodes -o json".
Result Type|
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### http://test-network-function.com/tests/container/pod
Property|Description
---|---
Version|v1.0.0
Description|A container-specific test suite used to verify various aspects of the underlying container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

### http://test-network-function.com/tests/readRemoteFile
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to read a specified file at a specified node
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`echo`

### http://test-network-function.com/tests/logging
Property|Description
---|---
Version|v1.0.0
Description|A test used to check logs are redirected to stderr/stdout
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `wc`

### http://test-network-function.com/tests/operator
Property|Description
---|---
Version|v1.0.0
Description|An operator-specific test used to exercise the behavior of a given operator.  In the current offering, we check if the operator ClusterServiceVersion (CSV) is installed properly.  A CSV is a YAML manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running the Operator.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`jq`, `oc`

### http://test-network-function.com/tests/generic/containerId
Property|Description
---|---
Version|v1.0.0
Description|A test used to check what is the id of the crio generated container this command is run from
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### http://test-network-function.com/tests/deploymentsnodes
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to drain node from its deployment pods
Result Type|normative
Intrusive|true
Modifications Persist After Test|true
Runtime Binaries Required|`jq`, `echo`

### http://test-network-function.com/tests/currentKernelCmdlineArgs
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get node's /proc/cmdline
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### http://test-network-function.com/tests/grubKernelCmdlineArgs
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get node's next boot kernel args
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ls`, `sort`, `head`, `cut`, `oc`

### http://test-network-function.com/tests/sysctlConfigFilesList
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get node's list of sysctl config files
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### http://test-network-function.com/tests/nodehugepages
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to verify a pod's nodeSelector and nodeAffinity configuration
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### http://test-network-function.com/tests/nodedebug
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to execute a command in a node
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `echo`

### http://test-network-function.com/tests/clusterrolebinding
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to test ClusterRoleBindings of CNF pod's ServiceAccount.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### http://test-network-function.com/tests/hugepages
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to read cluster's hugepages configuration
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`, `oc`, `grep`

### http://test-network-function.com/tests/podnodename
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get a pod's node
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### http://test-network-function.com/tests/mckernelarguments
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get an mc's kernel arguments
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `jq`, `echo`

### http://test-network-function.com/tests/generic/cnf_fs_diff
Property|Description
---|---
Version|v1.0.0
Description|A test used to check if there were no installation during container runtime
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`

### http://test-network-function.com/tests/nodehugepages
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to verify a node's hugepages configuration
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `grep`

### http://test-network-function.com/tests/generic/version
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to determine if a target container/machine is based on RHEL.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### http://test-network-function.com/tests/rolebinding
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to test RoleBindings of CNF pod's ServiceAccount.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`, `oc`

### http://test-network-function.com/tests/nodetainted
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to test whether node is tainted
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`, `cat`, `echo`

### http://test-network-function.com/tests/gracePeriod
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to extract the CNF pod's terminationGracePeriod.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`grep`, `cut`

### http://test-network-function.com/tests/deployments
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to read namespace's deployments
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### http://test-network-function.com/tests/node/uncordon
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to uncordon a node
Result Type|normative
Intrusive|true
Modifications Persist After Test|true
Runtime Binaries Required|`oc`

### http://test-network-function.com/tests/ipaddr
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to derive the default network interface IP address of a target container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ip`

### http://test-network-function.com/tests/ping
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to test ICMP connectivity from a source machine/container to a target destination.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`ping`

### http://test-network-function.com/tests/operator/check-subscription
Property|Description
---|---
Version|v1.0.0
Description|A test used to check the subscription of a given operator
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

### http://test-network-function.com/tests/owners
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to verify pod is managed by a ReplicaSet
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`cat`

### http://test-network-function.com/tests/hostname
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to check the hostname of a target machine/container.
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`hostname`

### http://test-network-function.com/tests/nodenames
Property|Description
---|---
Version|v1.0.0
Description|A generic test used to get node names
Result Type|normative
Intrusive|false
Modifications Persist After Test|false
Runtime Binaries Required|`oc`

