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

	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeport"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
)

const (
	defaultNumPings = 5
)

//
// All actual test code belongs below here.  Utilities belong above.
//

// Runs the "generic" CNF test cases.
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.NetworkingTestKey) {
		config := common.GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		for _, cid := range config.ExcludeContainersFromConnectivityTests {
			common.ContainersToExcludeFromConnectivityTests[cid] = ""
		}
		containersUnderTest := common.CreateContainersUnderTest(config)
		partnerContainers := common.CreatePartnerContainers(config)
		testOrchestrator := partnerContainers[config.TestOrchestrator]
		log.Info(testOrchestrator)
		log.Info(containersUnderTest)

		ginkgo.Context("Both Pods are on the Default network", func() {
			// for each container under test, ensure bidirectional ICMP traffic between the container and the orchestrator.
			for _, containerUnderTest := range containersUnderTest {
				if _, ok := common.ContainersToExcludeFromConnectivityTests[containerUnderTest.ContainerIdentifier]; !ok {
					testNetworkConnectivity(containerUnderTest.Oc, testOrchestrator.Oc, testOrchestrator.DefaultNetworkIPAddress, defaultNumPings)
					testNetworkConnectivity(testOrchestrator.Oc, containerUnderTest.Oc, containerUnderTest.DefaultNetworkIPAddress, defaultNumPings)
				}
			}
		})

		ginkgo.Context("Both Pods are connected via a Multus Overlay Network", func() {
			// Unidirectional test;  for each container under test, attempt to ping the target Multus IP addresses.
			for _, containerUnderTest := range containersUnderTest {
				for _, multusIPAddress := range containerUnderTest.ContainerConfiguration.MultusIPAddresses {
					testNetworkConnectivity(testOrchestrator.Oc, containerUnderTest.Oc, multusIPAddress, defaultNumPings)
				}
			}
		})

		conf := configpkg.GetConfigInstance()

		log.Info(conf.CNFs )
		for _, podUnderTest := range conf.CNFs {
			testNodePort(podUnderTest.Namespace)
		}

	}
})

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
	pingTester := ping.NewPing(common.DefaultTimeout, targetPodIPAddress, count)
	test, err := tnf.NewTest(initiatingPodOc.GetExpecter(), pingTester, []reel.Handler{pingTester}, initiatingPodOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	transmitted, received, errors := pingTester.GetStats()
	gomega.Expect(received).To(gomega.Equal(transmitted))
	gomega.Expect(errors).To(gomega.BeZero())
}

func testNodePort(podNamespace string) {
	ginkgo.When(fmt.Sprintf("Testing services in namespace %s", podNamespace), func() {
		ginkgo.It("Should not have services of type NodePort", func() {
			defer results.RecordResult(identifiers.TestServicesDoNotUseNodeportsIdentifier)
			context := common.GetContext()
			tester := nodeport.NewNodePort(common.DefaultTimeout, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}
