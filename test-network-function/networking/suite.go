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

package networking

import (
	"fmt"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeport"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	defaultNumPings = 5
)

type netTestContext struct {
	testerContainerOc         *interactive.Oc
	testerContainerIdentifier configsections.ContainerIdentifier
	ipSource                  string
	ipDestTargets             []ipDest
}
type ipDest struct {
	ip                        string
	targetContainerIdentifier configsections.ContainerIdentifier
}

//
// All actual test code belongs below here.  Utilities belong above.
//

// Runs the "generic" CNF test cases.
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.NetworkingTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})

		ginkgo.ReportAfterEach(results.RecordResult)

		ginkgo.Context("Both Pods are on the Default network", func() {
			// for each container under test, ensure bidirectional ICMP traffic between the container and the orchestrator.
			testDefaultNetworkConnectivity(env, defaultNumPings)
		})

		ginkgo.Context("Both Pods are connected via a Multus Overlay Network", func() {
			// Unidirectional test;  for each container under test, attempt to ping the target Multus IP addresses.
			testMultusNetworkConnectivity(env, defaultNumPings)
		})
		ginkgo.Context("Should not have type of nodePort", func() {
			testNodePort(env)
		})
	}
})

// processContainerIpsPerNet takes a container ip addresses for a given network attachment's network and uses it as a test target.
// The first container in the loop is selected as the test initiator. the Oc context of the container is used to initiate the pings
func processContainerIpsPerNet(containerID configsections.ContainerIdentifier, netKey string, ipAddress []string, netsUnderTest map[string]netTestContext, containerOc *interactive.Oc) {
	if len(ipAddress) == 0 {
		// if no multus addresses found, skip this container
		tnf.ClaimFilePrintf("Skipping container %s, Network %s because no multus IPs are present", containerID.PodName, netKey)
		return
	}
	// Create an entry at "key" if it is not present
	if _, ok := netsUnderTest[netKey]; !ok {
		netsUnderTest[netKey] = netTestContext{}
	}
	// get a copy of the content
	entry := netsUnderTest[netKey]
	// Then modify the copy
	firstIPIndex := 0
	if entry.testerContainerOc == nil {
		tnf.ClaimFilePrintf("Pod %s, container %s selected to initiate ping tests", containerID.PodName, containerID.ContainerName)
		entry.testerContainerIdentifier = containerID
		entry.testerContainerOc = containerOc
		// if multiple interfaces are present for this network on this container/pod, pick the first one as the tester source ip
		entry.ipSource = ipAddress[firstIPIndex]
		// do no include tester's IP in the list of destination IPs to ping
		firstIPIndex++
	}

	for _, aIP := range ipAddress[firstIPIndex:] {
		ipDestEntry := ipDest{}
		ipDestEntry.targetContainerIdentifier = containerID
		ipDestEntry.ip = aIP
		entry.ipDestTargets = append(entry.ipDestTargets, ipDestEntry)
	}

	// Then reassign map entry
	netsUnderTest[netKey] = entry
}
func runNetworkingTests(netsUnderTest map[string]netTestContext, count int) {
	for _, netUnderTest := range netsUnderTest {
		for _, aDestIP := range netUnderTest.ipDestTargets {
			ginkgo.By(fmt.Sprintf("a Ping is issued from %s(%s) %s to %s(%s) %s",
				netUnderTest.testerContainerIdentifier.PodName,
				netUnderTest.testerContainerIdentifier.ContainerName,
				netUnderTest.ipSource, aDestIP.targetContainerIdentifier.PodName,
				aDestIP.targetContainerIdentifier.ContainerName,
				aDestIP.ip))
			testPing(netUnderTest.testerContainerOc, aDestIP.ip, count)
		}
	}
}
func testDefaultNetworkConnectivity(env *config.TestEnvironment, count int) {
	ginkgo.When("Testing Default network connectivity", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityIdentifier)
		ginkgo.It(testID, func() {
			netsUnderTest := make(map[string]netTestContext)
			for _, cut := range env.ContainersUnderTest {
				if _, ok := env.ContainersToExcludeFromConnectivityTests[cut.ContainerIdentifier]; ok {
					tnf.ClaimFilePrintf("Skipping container %s because it is excluded from connectivity tests (default interface)", cut.ContainerConfiguration.PodName)
					continue
				}
				netKey := "default" //nolint:goconst // only used once
				defaultIPAddress := []string{cut.DefaultNetworkIPAddress}
				processContainerIpsPerNet(cut.ContainerIdentifier, netKey, defaultIPAddress, netsUnderTest, cut.Oc)
			}
			runNetworkingTests(netsUnderTest, count)
		})
	})
}
func testMultusNetworkConnectivity(env *config.TestEnvironment, count int) {
	ginkgo.When("Testing Multus network connectivity", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityIdentifier)
		ginkgo.It(testID, func() {
			netsUnderTest := make(map[string]netTestContext)
			for _, cut := range env.ContainersUnderTest {
				if _, ok := env.ContainersToExcludeFromConnectivityTests[cut.ContainerIdentifier]; ok {
					tnf.ClaimFilePrintf("Skipping container %s because it is excluded from connectivity tests (multus interface)", cut.ContainerConfiguration.PodName)
					continue
				}
				for netKey, multusIPAddress := range cut.ContainerConfiguration.MultusIPAddressesPerNet {
					processContainerIpsPerNet(cut.ContainerIdentifier, netKey, multusIPAddress, netsUnderTest, cut.Oc)
				}
			}
			runNetworkingTests(netsUnderTest, count)
		})
	})
}

// Test that a container can ping a target IP address.
func testPing(initiatingPodOc *interactive.Oc, targetPodIPAddress string, count int) {
	log.Infof("Sending ICMP traffic(%s to %s)", initiatingPodOc.GetPodName(), targetPodIPAddress)
	pingTester := ping.NewPing(common.DefaultTimeout, targetPodIPAddress, count)
	test, err := tnf.NewTest(initiatingPodOc.GetExpecter(), pingTester, []reel.Handler{pingTester}, initiatingPodOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	transmitted, received, errors := pingTester.GetStats()
	gomega.Expect(received).To(gomega.Equal(transmitted))
	gomega.Expect(errors).To(gomega.BeZero())
}

func testNodePort(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestServicesDoNotUseNodeportsIdentifier)
	ginkgo.It(testID, func() {
		context := common.GetContext()
		for _, ns := range env.NameSpacesUnderTest {
			ginkgo.By(fmt.Sprintf("Testing services in namespace %s", ns))
			tester := nodeport.NewNodePort(common.DefaultTimeout, ns)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
		}
	})
}
