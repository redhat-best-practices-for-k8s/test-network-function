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
	"time"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	expect "github.com/ryandgoulding/goexpect"
	log "github.com/sirupsen/logrus"
	tnfConfig "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
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
func getOcSession(pod, container, namespace string, timeout time.Duration, options ...expect.Option) *interactive.Oc {
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
		oc := getOcSession(containerID.PodName, containerID.ContainerName, containerID.Namespace, defaultTimeout, expect.Verbose(true))
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

		config := GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		for _, cid := range config.ExcludeContainersFromConnectivityTests {
			containersToExcludeFromConnectivityTests[cid] = ""
		}
		containersUnderTest := createContainersUnderTest(config)
		partnerContainers := createPartnerContainers(config)
		testOrchestrator := partnerContainers[config.TestOrchestrator]
		log.Info(testOrchestrator)
		log.Info(containersUnderTest)

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
