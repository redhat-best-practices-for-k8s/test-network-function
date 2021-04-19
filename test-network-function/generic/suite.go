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

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	tnfConfig "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/cnffsdiff"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/containerid"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/hugepages"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodehugepages"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodenames"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeport"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodetainted"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	defaultNumPings       = 5
	defaultTimeoutSeconds = 10
	multusTestsKey        = "multus"
	testsKey              = "generic"
)

// The default test timeout.
var defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// containersToExcludeFromConnectivityTests is a set used for storing the containers that should be excluded from
// connectivity testing.
var containersToExcludeFromConnectivityTests = make(map[tnfConfig.ContainerIdentifier]interface{})

// Helper used to instantiate an OpenShift Client Session.
func getOcSession(pod, container, namespace string, timeout time.Duration, options ...interactive.Option) *interactive.Oc {
	// Spawn an interactive OC shell using a goroutine (needed to avoid cross expect.Expecter interaction).  Extract the
	// Oc reference from the goroutine through a channel.  Performs basic sanity checking that the Oc session is set up
	// correctly.
	var containerOc *interactive.Oc
	ocChan := make(chan *interactive.Oc)
	var chOut <-chan error

	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner

	go func() {
		oc, outCh, err := interactive.SpawnOc(&spawner, pod, container, namespace, timeout, options...)
		gomega.Expect(outCh).ToNot(gomega.BeNil())
		gomega.Expect(err).To(gomega.BeNil())
		ocChan <- oc
	}()

	// Set up a go routine which reads from the error channel
	go func() {
		err := <-chOut
		gomega.Expect(err).To(gomega.BeNil())
	}()

	containerOc = <-ocChan

	gomega.Expect(containerOc).ToNot(gomega.BeNil())

	return containerOc
}

// container is an internal construct which follows the Container design pattern.  Essentially, a container holds the
// pertinent information to perform a test against or using an Operating System container.  This includes facets such
// as the reference to the interactive.Oc instance, the reference to the test configuration, and the default network
// IP address.
type container struct {
	containerConfiguration  tnfConfig.Container
	oc                      *interactive.Oc
	defaultNetworkIPAddress string
	containerIdentifier     tnfConfig.ContainerIdentifier
}

// createContainers contains the general steps involved in creating "oc" sessions and other configuration. A map of the
// aggregate information is returned.
func createContainers(containerDefinitions map[tnfConfig.ContainerIdentifier]tnfConfig.Container) map[tnfConfig.ContainerIdentifier]*container {
	createdContainers := map[tnfConfig.ContainerIdentifier]*container{}
	for containerID, containerConfig := range containerDefinitions {
		oc := getOcSession(containerID.PodName, containerID.ContainerName, containerID.Namespace, defaultTimeout, interactive.Verbose(true))
		var defaultIPAddress = "UNKNOWN"
		if _, ok := containersToExcludeFromConnectivityTests[containerID]; !ok {
			defaultIPAddress = getContainerDefaultNetworkIPAddress(oc, containerConfig.DefaultNetworkDevice)
		}
		createdContainers[containerID] = &container{
			containerConfiguration:  containerConfig,
			oc:                      oc,
			defaultNetworkIPAddress: defaultIPAddress,
			containerIdentifier:     containerID,
		}
	}
	return createdContainers
}

// createContainersUnderTest sets up the test containers.
func createContainersUnderTest(config *tnfConfig.TestConfiguration) map[tnfConfig.ContainerIdentifier]*container {
	return createContainers(config.ContainersUnderTest)
}

// createPartnerContainers sets up the partner containers.
func createPartnerContainers(config *tnfConfig.TestConfiguration) map[tnfConfig.ContainerIdentifier]*container {
	return createContainers(config.PartnerContainers)
}

//
// All actual test code belongs below here.  Utilities belong above.
//

// Runs the "generic" CNF test cases.
var _ = ginkgo.Describe(testsKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testsKey) {

		context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), defaultTimeout, interactive.Verbose(true))
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(context).ToNot(gomega.BeNil())

		config := GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		for _, cid := range config.ExcludeContainersFromConnectivityTests {
			containersToExcludeFromConnectivityTests[cid] = ""
		}
		containersUnderTest := createContainersUnderTest(config)
		partnerContainers := createPartnerContainers(config)
		testOrchestrator := partnerContainers[config.TestOrchestrator]
		fsDiffContainer := partnerContainers[config.FsDiffMasterContainer]
		log.Info(testOrchestrator)
		log.Info(containersUnderTest)

		ginkgo.Context("Container does not have additional packages installed", func() {
			if os.Getenv("FSDIFF") == "1" {
				for _, containerUnderTest := range containersUnderTest {
					testFsDiff(fsDiffContainer.oc, containerUnderTest.oc)
				}
			}
		})

		ginkgo.Context("Both Pods are on the Default network", func() {
			// for each container under test, ensure bidirectional ICMP traffic between the container and the orchestrator.
			for _, containerUnderTest := range containersUnderTest {
				if _, ok := containersToExcludeFromConnectivityTests[containerUnderTest.containerIdentifier]; !ok {
					testNetworkConnectivity(containerUnderTest.oc, testOrchestrator.oc, testOrchestrator.defaultNetworkIPAddress, defaultNumPings)
					testNetworkConnectivity(testOrchestrator.oc, containerUnderTest.oc, containerUnderTest.defaultNetworkIPAddress, defaultNumPings)
				}
			}
		})

		for _, containersUnderTest := range containersUnderTest {
			testIsRedHatRelease(containersUnderTest.oc)
		}
		testIsRedHatRelease(testOrchestrator.oc)

		for _, containersUnderTest := range containersUnderTest {
			testNamespace(containersUnderTest.oc)
		}

		for _, containersUnderTest := range containersUnderTest {
			testRoles(context, containersUnderTest.oc.GetPodName(), containersUnderTest.oc.GetPodNamespace())
		}

		for _, containersUnderTest := range containersUnderTest {
			testNodePort(context, containersUnderTest.oc.GetPodNamespace())
		}

		testTainted(context)
		testHugepages(context)
	}
})

// testIsRedHatRelease tests whether the container attached to oc is Red Hat based.
func testIsRedHatRelease(oc *interactive.Oc) {
	pod := oc.GetPodName()
	container := oc.GetPodContainerName()
	ginkgo.When(fmt.Sprintf("%s(%s) is checked for Red Hat version", pod, container), func() {
		ginkgo.It("Should report a proper Red Hat version", func() {
			versionTester := redhat.NewRelease(defaultTimeout)
			test, err := tnf.NewTest(oc.GetExpecter(), versionTester, []reel.Handler{versionTester}, oc.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}

// TODO: Multus is not applicable to every CNF, so in some regards it is CNF-specific.  On the other hand, it is likely
// a useful test across most CNFs.  Should "multus" be considered generic, cnf_specific, or somewhere in between.
var _ = ginkgo.Describe(multusTestsKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, multusTestsKey) {
		config := GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		containersUnderTest := createContainersUnderTest(config)
		partnerContainers := createPartnerContainers(config)
		testOrchestrator := partnerContainers[config.TestOrchestrator]

		ginkgo.Context("Both Pods are connected via a Multus Overlay Network", func() {
			// Unidirectional test;  for each container under test, attempt to ping the target Multus IP addresses.
			for _, containerUnderTest := range containersUnderTest {
				for _, multusIPAddress := range containerUnderTest.containerConfiguration.MultusIPAddresses {
					testNetworkConnectivity(testOrchestrator.oc, containerUnderTest.oc, multusIPAddress, defaultNumPings)
				}
			}
		})
	}
})

// Helper to test that the PUT didn't install new packages after starting, and report through Ginkgo.
func testFsDiff(masterPodOc, targetPodOc *interactive.Oc) {
	ginkgo.It(fmt.Sprintf("%s(%s) should not install new packages after starting", targetPodOc.GetPodName(), targetPodOc.GetPodContainerName()), func() {
		targetPodOc.GetExpecter()
		containerIDTester := containerid.NewContainerID(defaultTimeout)
		test, err := tnf.NewTest(targetPodOc.GetExpecter(), containerIDTester, []reel.Handler{containerIDTester}, targetPodOc.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
		containerID := containerIDTester.GetID()

		fsDiffTester := cnffsdiff.NewFsDiff(defaultTimeout, containerID)
		test, err = tnf.NewTest(masterPodOc.GetExpecter(), fsDiffTester, []reel.Handler{fsDiffTester}, masterPodOc.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err = test.Run()
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
	})
}

// Helper to test that a container can ping a target IP address, and report through Ginkgo.
func testNetworkConnectivity(initiatingPodOc, targetPodOc *interactive.Oc, targetPodIPAddress string, count int) {
	ginkgo.When(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", initiatingPodOc.GetPodName(),
		initiatingPodOc.GetPodContainerName(), targetPodOc.GetPodName(), targetPodOc.GetPodContainerName(),
		targetPodIPAddress), func() {
		ginkgo.It(fmt.Sprintf("%s(%s) should reply", targetPodOc.GetPodName(), targetPodOc.GetPodContainerName()), func() {
			testPing(initiatingPodOc, targetPodIPAddress, count)
		})
	})
}

// Test that a container can ping a target IP address.
func testPing(initiatingPodOc *interactive.Oc, targetPodIPAddress string, count int) {
	log.Infof("Sending ICMP traffic(%s to %s)", initiatingPodOc.GetPodName(), targetPodIPAddress)
	pingTester := ping.NewPing(defaultTimeout, targetPodIPAddress, count)
	test, err := tnf.NewTest(initiatingPodOc.GetExpecter(), pingTester, []reel.Handler{pingTester}, initiatingPodOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	transmitted, received, errors := pingTester.GetStats()
	gomega.Expect(received).To(gomega.Equal(transmitted))
	gomega.Expect(errors).To(gomega.BeZero())
}

// Extract a container IP address for a particular device.  This is needed since container default network IP address
// is served by dhcp, and thus is ephemeral.
func getContainerDefaultNetworkIPAddress(oc *interactive.Oc, dev string) string {
	log.Infof("Getting IP Information for: %s(%s) in ns=%s", oc.GetPodName(), oc.GetPodContainerName(), oc.GetPodNamespace())
	ipTester := ipaddr.NewIPAddr(defaultTimeout, dev)
	test, err := tnf.NewTest(oc.GetExpecter(), ipTester, []reel.Handler{ipTester}, oc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	return ipTester.GetIPv4Address()
}

// GetTestConfiguration returns the cnf-certification-generic-tests test configuration.
func GetTestConfiguration() *tnfConfig.TestConfiguration {
	config, err := tnfConfig.GetConfiguration(tnfConfig.UseDefaultConfigurationFilePath)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(config).ToNot(gomega.BeNil())
	return config
}

func testNamespace(oc *interactive.Oc) {
	pod := oc.GetPodName()
	container := oc.GetPodContainerName()
	ginkgo.When(fmt.Sprintf("Reading namespace of %s/%s", pod, container), func() {
		ginkgo.It("Should not be 'default' and should not begin with 'openshift-'", func() {
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.Equal("default")))
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.HavePrefix("openshift-")))
		})
	})
}

func testRoles(context *interactive.Context, podName, podNamespace string) {
	var serviceAccountName string

	ginkgo.When(fmt.Sprintf("Testing roles and privileges of %s/%s", podNamespace, podName), func() {
		testServiceAccount(context, podName, podNamespace, &serviceAccountName)
		testRoleBindings(context, podNamespace, &serviceAccountName)
		testClusterRoleBindings(context, podNamespace, &serviceAccountName)
	})
}

func testServiceAccount(context *interactive.Context, podName, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should have a valid ServiceAccount name", func() {
		tester := serviceaccount.NewServiceAccount(defaultTimeout, podName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
		*serviceAccountName = tester.GetServiceAccountName()
		gomega.Expect(*serviceAccountName).ToNot(gomega.BeEmpty())
	})
}

func testRoleBindings(context *interactive.Context, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have RoleBinding in other namespaces", func() {
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		rbTester := rolebinding.NewRoleBinding(defaultTimeout, *serviceAccountName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), rbTester, []reel.Handler{rbTester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		if rbTester.Result() == tnf.FAILURE {
			log.Info("RoleBindings: ", rbTester.GetRoleBindings())
		}
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
	})
}

func testClusterRoleBindings(context *interactive.Context, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have ClusterRoleBindings", func() {
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		crbTester := clusterrolebinding.NewClusterRoleBinding(defaultTimeout, *serviceAccountName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), crbTester, []reel.Handler{crbTester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		if crbTester.Result() == tnf.FAILURE {
			log.Info("ClusterRoleBindings: ", crbTester.GetClusterRoleBindings())
		}
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	})
}

func testNodePort(context *interactive.Context, podNamespace string) {
	ginkgo.When(fmt.Sprintf("Testing services in namespace %s", podNamespace), func() {
		ginkgo.It("Should not have services of type NodePort", func() {
			tester := nodeport.NewNodePort(defaultTimeout, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}

func testTainted(context *interactive.Context) {
	if isMinikube() {
		return
	}
	var nodeNames []string
	ginkgo.When("Testing tainted nodes in cluster", func() {
		ginkgo.It("Should return list of node names", func() {
			tester := nodenames.NewNodeNames(defaultTimeout, nil)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
			nodeNames = tester.GetNodeNames()
			gomega.Expect(nodeNames).NotTo(gomega.BeNil())
		})

		ginkgo.It("Should not have tainted nodes", func() {
			if len(nodeNames) == 0 {
				ginkgo.Skip("Can't test tainted nodes when list of nodes is empty. Please check previous tests.")
			}
			var taintedNodes []string
			for _, node := range nodeNames {
				tester := nodetainted.NewNodeTainted(defaultTimeout, node)
				test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(testResult).NotTo(gomega.Equal(tnf.ERROR))
				gomega.Expect(err).To(gomega.BeNil())
				if testResult == tnf.FAILURE {
					taintedNodes = append(taintedNodes, node)
				}
			}
			gomega.Expect(taintedNodes).To(gomega.BeNil())
		})
	})
}

func testHugepages(context *interactive.Context) {
	if isMinikube() {
		return
	}
	var nodeNames []string
	var clusterHugepages, clusterHugepagesz int
	ginkgo.When("Testing worker nodes' hugepages configuration", func() {
		ginkgo.It("Should return list of worker node names", func() {
			tester := nodenames.NewNodeNames(defaultTimeout, map[string]*string{"node-role.kubernetes.io/worker": nil})
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
			nodeNames = tester.GetNodeNames()
			gomega.Expect(nodeNames).NotTo(gomega.BeNil())
		})
		ginkgo.It("Should return cluster's hugepages configuration", func() {
			tester := hugepages.NewHugepages(defaultTimeout)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
			clusterHugepages = tester.GetHugepages()
			clusterHugepagesz = tester.GetHugepagesz()
		})
		ginkgo.It("Should have same configuration as cluster", func() {
			var badNodes []string
			for _, node := range nodeNames {
				tester := nodehugepages.NewNodeHugepages(defaultTimeout, node, clusterHugepagesz, clusterHugepages)
				test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(err).To(gomega.BeNil())
				if testResult != tnf.SUCCESS {
					badNodes = append(badNodes, node)
				}
			}
			gomega.Expect(badNodes).To(gomega.BeNil())
		})
	})
}

func isMinikube() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))
	return b
}
