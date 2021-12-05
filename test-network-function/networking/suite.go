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

func testDefaultNetworkConnectivity(env *config.TestEnvironment, count int) {
	ginkgo.When("Testing network connectivity", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityIdentifier)
		ginkgo.It(testID, func() {
			if env.TestOrchestrator == nil {
				ginkgo.Skip("Orchestrator is not deployed, skip this test")
			}
			found := false
			containerping := false

			for _, cut := range env.ContainersUnderTest {
				if _, ok := env.ContainersToExcludeFromConnectivityTests[cut.ContainerIdentifier]; ok {
					continue
				}
				if cut.ContainerConfiguration.HasPing {
					containerping = true
					found = true
					context := cut.Oc
					testOrchestrator := env.TestOrchestrator
					ginkgo.By(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", testOrchestrator.Oc.GetPodName(),
						testOrchestrator.Oc.GetPodContainerName(), cut.Oc.GetPodName(), cut.Oc.GetPodContainerName(),
						cut.DefaultNetworkIPAddress))
					testPing(testOrchestrator.Oc, cut.DefaultNetworkIPAddress, count)
					ginkgo.By(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", cut.Oc.GetPodName(),
						cut.Oc.GetPodContainerName(), testOrchestrator.Oc.GetPodName(), testOrchestrator.Oc.GetPodContainerName(),
						testOrchestrator.DefaultNetworkIPAddress))
					testPing(context, testOrchestrator.DefaultNetworkIPAddress, count)
				}
			}
			if !found || !containerping {
				ginkgo.Skip("No container found suitable for connectivity test or no ping installed")
			}
		})
	})
}

func testMultusNetworkConnectivity(env *config.TestEnvironment, count int) {
	ginkgo.When("Testing network connectivity", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityIdentifier)
		ginkgo.It(testID, func() {
			if env.TestOrchestrator == nil {
				ginkgo.Skip("Orchestrator is not deployed, skip this test")
			}
			found := false
			for _, cut := range env.ContainersUnderTest {
				if _, ok := env.ContainersToExcludeFromConnectivityTests[cut.ContainerIdentifier]; ok {
					continue
				}
				found = true
				if len(cut.ContainerConfiguration.MultusIPAddresses) == 0 {
					ginkgo.Skip("No Multus IPs detected")
				}

				for _, multusIPAddress := range cut.ContainerConfiguration.MultusIPAddresses {
					testOrchestrator := env.TestOrchestrator
					ginkgo.By(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", testOrchestrator.Oc.GetPodName(),
						testOrchestrator.Oc.GetPodContainerName(), cut.Oc.GetPodName(), cut.Oc.GetPodContainerName(),
						multusIPAddress))
					testPing(testOrchestrator.Oc, multusIPAddress, count)
				}
				if !found {
					ginkgo.Skip("No container found suitable for Multus connectivity test")
				}
			}
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
	ginkgo.By("err err err err err err err err errerr err err err err err")
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
