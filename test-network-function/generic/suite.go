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

package generic

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"

	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/bootconfigentries"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/cnffsdiff"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/containerid"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/currentkernelcmdlineargs"
	dp "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deployments"
	dd "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsdrain"
	dn "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsnodes"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/graceperiod"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/hugepages"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/mckernelarguments"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodehugepages"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodemcname"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodenames"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeport"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeselector"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodetainted"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/owners"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/podnodename"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/readbootconfig"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/sysctlallconfigsargs"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	utils "github.com/test-network-function/test-network-function/pkg/utils"
)

const (
	defaultNumPings               = 5
	defaultTimeoutSeconds         = 10
	defaultTerminationGracePeriod = 30
	multusTestsKey                = "multus"
	testsKey                      = "generic"
	drainTimeoutMinutes           = 5
)

var (
	// nodeUncordonTestPath is the file location of the uncordon.json test case relative to the project root.
	nodeUncordonTestPath = path.Join("pkg", "tnf", "handlers", "nodeuncordon", "uncordon.json")

	// loggingTestPath is the file location of the logging.json test case relative to the project root.
	loggingTestPath = path.Join("pkg", "tnf", "handlers", "logging", "logging.json")

	// pathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// relativeNodesTestPath is the relative path to the nodes.json test case.
	relativeNodesTestPath = path.Join(pathRelativeToRoot, nodeUncordonTestPath)

	// relativeLoggingTestPath is the relative path to the logging.json test case.
	relativeLoggingTestPath = path.Join(pathRelativeToRoot, loggingTestPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

// The default test timeout.
var defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

var drainTimeout = time.Duration(drainTimeoutMinutes) * time.Minute

// containersToExcludeFromConnectivityTests is a set used for storing the containers that should be excluded from
// connectivity testing.
var containersToExcludeFromConnectivityTests = make(map[configsections.ContainerIdentifier]interface{})

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
	containerConfiguration  configsections.Container
	oc                      *interactive.Oc
	defaultNetworkIPAddress string
	containerIdentifier     configsections.ContainerIdentifier
}

// createContainers contains the general steps involved in creating "oc" sessions and other configuration. A map of the
// aggregate information is returned.
func createContainers(containerDefinitions []configsections.Container) map[configsections.ContainerIdentifier]*container {
	createdContainers := make(map[configsections.ContainerIdentifier]*container)
	for _, c := range containerDefinitions {
		oc := getOcSession(c.PodName, c.ContainerName, c.Namespace, defaultTimeout, interactive.Verbose(true))
		var defaultIPAddress = "UNKNOWN"
		if _, ok := containersToExcludeFromConnectivityTests[c.ContainerIdentifier]; !ok {
			defaultIPAddress = getContainerDefaultNetworkIPAddress(oc, c.DefaultNetworkDevice)
		}
		createdContainers[c.ContainerIdentifier] = &container{
			containerConfiguration:  c,
			oc:                      oc,
			defaultNetworkIPAddress: defaultIPAddress,
			containerIdentifier:     c.ContainerIdentifier,
		}
	}
	return createdContainers
}

// createContainersUnderTest sets up the test containers.
func createContainersUnderTest(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*container {
	return createContainers(conf.ContainersUnderTest)
}

// createPartnerContainers sets up the partner containers.
func createPartnerContainers(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*container {
	return createContainers(conf.PartnerContainers)
}

//
// All actual test code belongs below here.  Utilities belong above.
//

// Runs the "generic" CNF test cases.
var _ = ginkgo.Describe(testsKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testsKey) {
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

		for _, containerUnderTest := range containersUnderTest {
			testNodeSelector(getContext(), containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodNamespace())
		}

		ginkgo.Context("Container does not have additional packages installed", func() {
			// use this boolean to turn off tests that require OS packages
			if !isMinikube() {
				if fsDiffContainer != nil {
					for _, containerUnderTest := range containersUnderTest {
						testFsDiff(fsDiffContainer.oc, containerUnderTest.oc)
					}
				} else {
					log.Warn("no fs diff container is configured, cannot run fs diff test")
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

		for _, containerUnderTest := range containersUnderTest {
			testIsRedHatRelease(containerUnderTest.oc)
		}
		testIsRedHatRelease(testOrchestrator.oc)

		for _, containerUnderTest := range containersUnderTest {
			testNamespace(containerUnderTest.oc)
		}

		for _, containerUnderTest := range containersUnderTest {
			testRoles(containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodNamespace())
		}

		for _, containerUnderTest := range containersUnderTest {
			testNodePort(containerUnderTest.oc.GetPodNamespace())
		}

		for _, containerUnderTest := range containersUnderTest {
			testGracePeriod(getContext(), containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodNamespace())
		}

		for _, containerUnderTest := range containersUnderTest {
			testLogging(containerUnderTest.oc.GetPodNamespace(), containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodContainerName())
		}
		testTainted()
		testHugepages()

		if !isMinikube() {
			for _, containersUnderTest := range containersUnderTest {
				testDeployments(containersUnderTest.oc.GetPodNamespace())
			}
		}

		for _, containerUnderTest := range containersUnderTest {
			testOwner(containerUnderTest.oc.GetPodNamespace(), containerUnderTest.oc.GetPodName())
		}

		if !isMinikube() {
			for _, containersUnderTest := range containersUnderTest {
				testBootParams(getContext(), containersUnderTest.oc.GetPodName(), containersUnderTest.oc.GetPodNamespace(), containersUnderTest.oc)
			}
		}

		if !isMinikube() {
			for _, containersUnderTest := range containersUnderTest {
				testSysctlConfigs(getContext(), containersUnderTest.oc.GetPodName(), containersUnderTest.oc.GetPodNamespace())
			}
		}
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
		defer results.RecordResult(identifiers.TestUnalteredBaseImageIdentifier)
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
			defer results.RecordResult(identifiers.TestICMPv4ConnectivityIdentifier)
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
	runAndValidateTest(test)
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
	runAndValidateTest(test)
	return ipTester.GetIPv4Address()
}

// GetTestConfiguration returns the cnf-certification-generic-tests test configuration.
func GetTestConfiguration() *configsections.TestConfiguration {
	conf := config.GetConfigInstance()
	return &conf.Generic
}

func testNamespace(oc *interactive.Oc) {
	pod := oc.GetPodName()
	container := oc.GetPodContainerName()
	ginkgo.When(fmt.Sprintf("Reading namespace of %s/%s", pod, container), func() {
		ginkgo.It("Should not be 'default' and should not begin with 'openshift-'", func() {
			defer results.RecordResult(identifiers.TestNamespaceBestPracticesIdentifier)
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.Equal("default")))
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.HavePrefix("openshift-")))
		})
	})
}

func testRoles(podName, podNamespace string) {
	var serviceAccountName string
	ginkgo.When(fmt.Sprintf("Testing roles and privileges of %s/%s", podNamespace, podName), func() {
		testServiceAccount(podName, podNamespace, &serviceAccountName)
		testRoleBindings(podNamespace, &serviceAccountName)
		testClusterRoleBindings(podNamespace, &serviceAccountName)
	})
}

func testGracePeriod(context *interactive.Context, podName, podNamespace string) {
	ginkgo.It(fmt.Sprintf("Testing pod terminationGracePeriod %s/%s", podNamespace, podName), func() {
		defer results.RecordResult(identifiers.TestNonDefaultGracePeriodIdentifier)
		infoWriter := tnf.CreateTestExtraInfoWriter()
		tester := graceperiod.NewGracePeriod(defaultTimeout, podName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
		gracePeriod := tester.GetGracePeriod()
		if gracePeriod == defaultTerminationGracePeriod {
			msg := fmt.Sprintf("%s %s has terminationGracePeriod set to 30, you might want to change it", podName, podNamespace)
			log.Warn(msg)
			infoWriter(msg)
		}
	})
}

func testNodeSelector(context *interactive.Context, podName, podNamespace string) {
	ginkgo.It(fmt.Sprintf("Testing pod nodeSelector %s/%s", podNamespace, podName), func() {
		defer results.RecordResult(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
		infoWriter := tnf.CreateTestExtraInfoWriter()
		tester := nodeselector.NewNodeSelector(defaultTimeout, podName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()

		gomega.Expect(err).To(gomega.BeNil())
		if testResult != tnf.SUCCESS {
			msg := fmt.Sprintf("The pod specifies nodeSelector/nodeAffinity field, you might want to change it, %s %s", podName, podNamespace)
			log.Warn(msg)
			infoWriter(msg)
		}
	})
}

func getMcKernelArguments(context *interactive.Context, mcName string) map[string]string {
	mcKernelArgumentsTester := mckernelarguments.NewMcKernelArguments(defaultTimeout, mcName)
	test, err := tnf.NewTest(context.GetExpecter(), mcKernelArgumentsTester, []reel.Handler{mcKernelArgumentsTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	mcKernelArguments := mcKernelArgumentsTester.GetKernelArguments()
	var mcKernelArgumentsJSON []string
	err = json.Unmarshal([]byte(mcKernelArguments), &mcKernelArgumentsJSON)
	gomega.Expect(err).To(gomega.BeNil())
	mcKernelArgumentsMap := utils.ArgListToMap(mcKernelArgumentsJSON)
	return mcKernelArgumentsMap
}

func getMcName(context *interactive.Context, nodeName string) string {
	mcNameTester := nodemcname.NewNodeMcName(defaultTimeout, nodeName)
	test, err := tnf.NewTest(context.GetExpecter(), mcNameTester, []reel.Handler{mcNameTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	return mcNameTester.GetMcName()
}

func getPodNodeName(context *interactive.Context, podName, podNamespace string) string {
	podNameTester := podnodename.NewPodNodeName(defaultTimeout, podName, podNamespace)
	test, err := tnf.NewTest(context.GetExpecter(), podNameTester, []reel.Handler{podNameTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	return podNameTester.GetNodeName()
}

func getCurrentKernelCmdlineArgs(targetPodOc *interactive.Oc) map[string]string {
	currentKernelCmdlineArgsTester := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(defaultTimeout)
	test, err := tnf.NewTest(targetPodOc.GetExpecter(), currentKernelCmdlineArgsTester, []reel.Handler{currentKernelCmdlineArgsTester}, targetPodOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	currnetKernelCmdlineArgs := currentKernelCmdlineArgsTester.GetKernelArguments()
	currentSplitKernelCmdlineArgs := strings.Split(currnetKernelCmdlineArgs, " ")
	return utils.ArgListToMap(currentSplitKernelCmdlineArgs)
}

func getBootEntryIndex(bootEntry string) (int, error) {
	return strconv.Atoi(strings.Split(bootEntry, "-")[1])
}

func getMaxIndexEntry(bootConfigEntries []string) string {
	maxIndex, err := getBootEntryIndex(bootConfigEntries[0])
	gomega.Expect(err).To(gomega.BeNil())
	maxIndexEntryName := bootConfigEntries[0]
	for _, bootEntry := range bootConfigEntries {
		if entryIndex, err2 := getBootEntryIndex(bootEntry); entryIndex > maxIndex {
			maxIndex = entryIndex
			gomega.Expect(err2).To(gomega.BeNil())
			maxIndexEntryName = bootEntry
		}
	}

	return maxIndexEntryName
}

func getGrubKernelArgs(context *interactive.Context, nodeName string) map[string]string {
	bootConfigEntriesTester := bootconfigentries.NewBootConfigEntries(defaultTimeout, nodeName)
	test, err := tnf.NewTest(context.GetExpecter(), bootConfigEntriesTester, []reel.Handler{bootConfigEntriesTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	bootConfigEntries := bootConfigEntriesTester.GetBootConfigEntries()

	maxIndexEntryName := getMaxIndexEntry(bootConfigEntries)

	readBootConfigTester := readbootconfig.NewReadBootConfig(defaultTimeout, nodeName, maxIndexEntryName)
	test, err = tnf.NewTest(context.GetExpecter(), readBootConfigTester, []reel.Handler{readBootConfigTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	bootConfig := readBootConfigTester.GetBootConfig()

	splitBootConfig := strings.Split(bootConfig, "\n")
	filteredBootConfig := utils.FilterArray(splitBootConfig, func(line string) bool {
		return strings.HasPrefix(line, "options")
	})
	gomega.Expect(len(filteredBootConfig)).To(gomega.Equal(1))
	grubKernelConfig := filteredBootConfig[0]
	grubSplitKernelConfig := strings.Split(grubKernelConfig, " ")
	grubSplitKernelConfig = grubSplitKernelConfig[1:]
	return utils.ArgListToMap(grubSplitKernelConfig)
}

// Creates a map describing the final sysctl key-value pair out of the results of "sysctl --system"
func parseSysctlSystemOutput(sysctlSystemOutput string) map[string]string {
	retval := make(map[string]string)
	splitConfig := strings.Split(sysctlSystemOutput, "\n")
	for _, line := range splitConfig {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "*") {
			continue
		}

		keyValRegexp := regexp.MustCompile(`( \S+)(\s*)=(\s*)(\S+)`) // A line is of the form "kernel.yama.ptrace_scope = 0"
		if !keyValRegexp.MatchString(line) {
			continue
		}
		regexResults := keyValRegexp.FindStringSubmatch(line)
		key := regexResults[1]
		val := regexResults[4]
		retval[key] = val
	}
	return retval
}

func getSysctlConfigArgs(context *interactive.Context, nodeName string) map[string]string {
	sysctlAllConfigsArgsTester := sysctlallconfigsargs.NewSysctlAllConfigsArgs(defaultTimeout, nodeName)
	test, err := tnf.NewTest(context.GetExpecter(), sysctlAllConfigsArgsTester, []reel.Handler{sysctlAllConfigsArgsTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	sysctlAllConfigsArgs := sysctlAllConfigsArgsTester.GetSysctlAllConfigsArgs()

	return parseSysctlSystemOutput(sysctlAllConfigsArgs)
}

func testBootParams(context *interactive.Context, podName, podNamespace string, targetPodOc *interactive.Oc) {
	ginkgo.It(fmt.Sprintf("Testing boot params for the pod's node %s/%s", podNamespace, podName), func() {
		defer results.RecordResult(identifiers.TestUnalteredStartupBootParamsIdentifier)
		nodeName := getPodNodeName(context, podName, podNamespace)
		mcName := getMcName(context, nodeName)
		mcKernelArgumentsMap := getMcKernelArguments(context, mcName)
		currentKernelArgsMap := getCurrentKernelCmdlineArgs(targetPodOc)
		grubKernelConfigMap := getGrubKernelArgs(context, nodeName)

		for key, mcVal := range mcKernelArgumentsMap {
			if currentVal, ok := currentKernelArgsMap[key]; ok {
				gomega.Expect(currentVal).To(gomega.Equal(mcVal))
			}
			if grubVal, ok := grubKernelConfigMap[key]; ok {
				gomega.Expect(grubVal).To(gomega.Equal(mcVal))
			}
		}
	})
}

func testSysctlConfigs(context *interactive.Context, podName, podNamespace string) {
	ginkgo.It(fmt.Sprintf("Testing sysctl config files for the pod's node %s/%s", podNamespace, podName), func() {
		nodeName := getPodNodeName(context, podName, podNamespace)
		combinedSysctlSettings := getSysctlConfigArgs(context, nodeName)
		mcName := getMcName(context, nodeName)
		mcKernelArgumentsMap := getMcKernelArguments(context, mcName)

		for key, sysctlConfigVal := range combinedSysctlSettings {
			if mcVal, ok := mcKernelArgumentsMap[key]; ok {
				gomega.Expect(mcVal).To(gomega.Equal(sysctlConfigVal))
			}
		}
	})
}

func testServiceAccount(podName, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should have a valid ServiceAccount name", func() {
		defer results.RecordResult(identifiers.TestPodServiceAccountBestPracticesIdentifier)
		context := getContext()
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

func testRoleBindings(podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have RoleBinding in other namespaces", func() {
		defer results.RecordResult(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		context := getContext()
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

func testClusterRoleBindings(podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have ClusterRoleBindings", func() {
		defer results.RecordResult(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		context := getContext()
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

func testNodePort(podNamespace string) {
	ginkgo.When(fmt.Sprintf("Testing services in namespace %s", podNamespace), func() {
		ginkgo.It("Should not have services of type NodePort", func() {
			defer results.RecordResult(identifiers.TestServicesDoNotUseNodeportsIdentifier)
			context := getContext()
			tester := nodeport.NewNodePort(defaultTimeout, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}

func testTainted() {
	if isMinikube() {
		return
	}
	var nodeNames []string
	ginkgo.When("Testing tainted nodes in cluster", func() {
		ginkgo.It("Should return list of node names", func() {
			context := getContext()
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
			defer results.RecordResult(identifiers.TestNonTaintedNodeKernelsIdentifier)
			if len(nodeNames) == 0 {
				ginkgo.Skip("Can't test tainted nodes when list of nodes is empty. Please check previous tests.")
			}
			var taintedNodes []string
			for _, node := range nodeNames {
				context := getContext()
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

func testHugepages() {
	if isMinikube() {
		return
	}
	var nodeNames []string
	var clusterHugepages, clusterHugepagesz int
	ginkgo.When("Testing worker nodes' hugepages configuration", func() {
		ginkgo.It("Should return list of worker node names", func() {
			context := getContext()
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
			context := getContext()
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
			defer results.RecordResult(identifiers.TestHugepagesNotManuallyManipulated)
			var badNodes []string
			for _, node := range nodeNames {
				context := getContext()
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

// testDeployments ensures that each Deployment has the correct number of "Ready" replicas.
func testDeployments(namespace string) {
	var deployments dp.DeploymentMap
	var notReadyDeployments []string
	ginkgo.When("Testing deployments in namespace", func() {
		ginkgo.It("Should return list of deployments", func() {
			deployments, notReadyDeployments = getDeployments(namespace)
			if len(deployments) == 0 {
				return
			}
			gomega.Expect(notReadyDeployments).To(gomega.BeEmpty())
		})
	})
}

func isMinikube() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))
	return b
}

//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
func nonIntrusive() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_NON_INTRUSIVE_ONLY"))
	return b
}

type node struct {
	name        string
	deployments map[string]bool
}

func sortNodesMap(nodesMap dn.NodesMap) []node {
	nodes := make([]node, 0, len(nodesMap))
	for n, d := range nodesMap {
		nodes = append(nodes, node{n, d})
	}
	sort.Slice(nodes, func(i, j int) bool { return len(nodes[i].deployments) > len(nodes[j].deployments) })
	return nodes
}

//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
func getDeploymentsNodes(namespace string) []node {
	context := getContext()
	tester := dn.NewDeploymentsNodes(defaultTimeout, namespace)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
	nodes := tester.GetNodes()
	gomega.Expect(nodes).NotTo(gomega.BeEmpty())
	return sortNodesMap(nodes)
}

// getDeployments returns map of deployments and names of not-ready deployments
func getDeployments(namespace string) (deployments dp.DeploymentMap, notReadyDeployments []string) {
	context := getContext()
	tester := dp.NewDeployments(defaultTimeout, namespace)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)

	deployments = tester.GetDeployments()

	for name, d := range deployments {
		if d.Unavailable != 0 || d.Ready != d.Replicas || d.Available != d.Replicas || d.UpToDate != d.Replicas {
			notReadyDeployments = append(notReadyDeployments, name)
		}
	}

	return deployments, notReadyDeployments
}

//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
func drainNode(node string) {
	context := getContext()
	tester := dd.NewDeploymentsDrain(drainTimeout, node)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	runAndValidateTest(test)
}
func testLogging(podNameSpace, podName, containerName string) {
	ginkgo.When("Testing PUT is emitting logs to stdout/stderr", func() {
		ginkgo.It("should return at least one line of log", func() {
			defer results.RecordResult(identifiers.TestLoggingIdentifier)
			loggingTest(podNameSpace, podName, containerName)
		})
	})
}
func loggingTest(podNamespace, podName, containerName string) {
	context := getContext()
	values := make(map[string]interface{})
	values["POD_NAMESPACE"] = podNamespace
	values["POD_NAME"] = podName
	values["CONTAINER_NAME"] = containerName
	test, handlers, result, err := generic.NewGenericFromMap(relativeLoggingTestPath, relativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())
	tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	testResult, err := tester.Run()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
}

// uncordonNode uncordons a Node.
//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
func uncordonNode(node string) {
	context := getContext()
	values := make(map[string]interface{})
	values["NODE"] = node
	test, handlers, result, err := generic.NewGenericFromMap(relativeNodesTestPath, relativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(len(handlers)).To(gomega.Equal(1))
	gomega.Expect(test).ToNot(gomega.BeNil())

	tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	testResult, err := tester.Run()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
}

func getContext() *interactive.Context {
	context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), defaultTimeout, interactive.Verbose(true))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(context).ToNot(gomega.BeNil())
	return context
}

func testOwner(podNamespace, podName string) {
	ginkgo.When("Testing owners of CNF pod", func() {
		ginkgo.It("Should be only ReplicaSet", func() {
			defer results.RecordResult(identifiers.TestPodDeploymentBestPracticesIdentifier)
			context := getContext()
			tester := owners.NewOwners(defaultTimeout, podNamespace, podName)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}

func runAndValidateTest(test *tnf.Test) {
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
}
