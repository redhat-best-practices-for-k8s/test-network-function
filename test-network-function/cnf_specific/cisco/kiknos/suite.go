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

package kiknos

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/itc"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/generic"
	"time"
)

const (
	// Unfortunately, some busy waits are required since the command interaction with itc is asynchronous
	busyWaitSeconds = 5
	// The default rate used for any messages/sec required by itc
	messagesPerSecond = 100
	// The number of ICMP Requests to perform across the IPSEC tunnel
	numPings = 100
	// The number of UDP packets to send.
	numUDPPackets = 100
	// The number of new tunnels to create.
	numTunnels = 1
	// The size of the ICMP Requests that will traverse the IPSEC tunnel
	pingSize = 100
	// The target IP address, which is preconfigured in the setup.
	targetTunnelIP = "40.0.0.1"
	// The size of the UDP Packets
	udpSize = 100
)

var _ = ginkgo.Describe("cisco_kiknos", func() {
	// Extract some basic configuration parameters from the generic configuration.
	config := generic.GetTestConfiguration()
	partnerPodName := config.TestOrchestrator.PodName
	partnerPodNamespace := config.TestOrchestrator.Namespace
	partnerPodContainerName := config.TestOrchestrator.ContainerName

	// Run the only CNF-Specific Test Spec., which has several sub-tests.
	testTunnel(partnerPodName, partnerPodContainerName, partnerPodNamespace)
})

// Tests Kiknos IPSEC tunnel creation.
func testTunnel(partnerPodName, partnerPodContainerName, partnerPodNamespace string) {
	ginkgo.When(fmt.Sprintf("%s(%s) creates an IPSEC tunnel", partnerPodName, partnerPodContainerName), func() {
		var newTunnelIndex int
		ginkgo.It("should report the tunnel was created through the CLI", func() {
			newTunnelIndex = createAndVerifyTunnel(partnerPodName, partnerPodContainerName, partnerPodNamespace)
		})
		ginkgo.It("should pass ICMP traffic", func() {
			verifyICMPTraffic(partnerPodName, partnerPodContainerName, partnerPodNamespace, newTunnelIndex)
		})
		ginkgo.It("should pass UDP traffic", func() {
			verifyUDPTraffic(partnerPodName, partnerPodContainerName, partnerPodNamespace, newTunnelIndex)
		})
	})
}

// Verify UDP traffic
func verifyUDPTraffic(partnerPodName, partnerPodContainerName, partnerPodNamespace string, newTunnelIndex int) {
	// Take an initial snapshot of the ICMP received count.
	initialUDPCount, err := itc.GetUDPReceivedCount(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	expectNil(err)
	_, err = itc.SendData(partnerPodName, partnerPodContainerName, partnerPodNamespace, newTunnelIndex, messagesPerSecond, numUDPPackets, udpSize)
	expectNil(err)

	// TODO:  Could this be done asynchronously through a channel?
	time.Sleep(busyWaitSeconds * time.Second)

	// Check that numPings were received.
	updatedUDPCount, err := itc.GetUDPReceivedCount(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	expectNil(err)
	udpCountDiff := updatedUDPCount - initialUDPCount
	gomega.Expect(udpCountDiff).To(gomega.Equal(numUDPPackets))
}

// Verify ICMP traffic
func verifyICMPTraffic(partnerPodName, partnerPodContainerName, partnerPodNamespace string, newTunnelIndex int) {
	// Take an initial snapshot of the ICMP received count.
	initialIcmpCount, err := itc.GetItcIcmpReplyCount(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	expectNil(err)
	_, err = itc.Ping(partnerPodName, partnerPodContainerName, partnerPodNamespace, newTunnelIndex, targetTunnelIP, messagesPerSecond, numPings, pingSize)
	expectNil(err)

	// TODO:  Could this be done asynchronously through a channel?
	time.Sleep(busyWaitSeconds * time.Second)

	// Check that numPings were received.
	updatedIcmpCount, err := itc.GetItcIcmpReplyCount(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	expectNil(err)
	icmpCountDiff := updatedIcmpCount - initialIcmpCount
	gomega.Expect(icmpCountDiff).To(gomega.Equal(numPings))
}

// Create an IPSEC tunnel and verify it was successfully established.
func createAndVerifyTunnel(partnerPodName, partnerPodContainerName, partnerPodNamespace string) int {
	var newTunnelIndex int

	// Take a snapshot of the original number of tunnels that were created, and the number that successfully connected.
	initialNumInstantiatedTunnels, initialNumConnectedTunnels := extractNumTunnels(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	createTunnel(partnerPodName, partnerPodContainerName, partnerPodNamespace, numTunnels, messagesPerSecond)

	// TODO:  Could this be done asynchronously through a channel?
	time.Sleep(busyWaitSeconds * time.Second)

	// Check that the correct number of tunnels were created.
	updatedNumInstantiatedTunnels, updatedNumConnectedTunnels := extractNumTunnels(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	newTunnelIndex = updatedNumInstantiatedTunnels
	diffNumInstantigatedTunnels := updatedNumInstantiatedTunnels - initialNumInstantiatedTunnels
	checkTunnelCountUpdated(diffNumInstantigatedTunnels, numTunnels)
	diffNumConnectedTunnels := updatedNumConnectedTunnels - initialNumConnectedTunnels
	checkTunnelCountUpdated(diffNumConnectedTunnels, numTunnels)

	return newTunnelIndex
}

// make sure the number of tunnels is incremented
func checkTunnelCountUpdated(actual, expected int) {
	gomega.Expect(actual).To(gomega.Equal(expected))
}

// creates an IPSEC tunnel using ike-testctl
func createTunnel(partnerPodName, partnerPodContainerName, partnerPodNamespace string, numTunnels, messagesPerSecond int) {
	err := itc.Create(partnerPodName, partnerPodContainerName, partnerPodNamespace, numTunnels, messagesPerSecond)
	expectNil(err)
}

// extract the number of tunnels broadcast by kiknos
func extractNumTunnels(partnerPodName, partnerPodContainerName, partnerPodNamespace string) (numInstantiated, numConnected int) {
	numInstantiatedTunnels := 0
	numConnectedTunnels := 0
	var err error
	numInstantiatedTunnels, numConnectedTunnels, err = itc.IkeSummary(partnerPodName, partnerPodContainerName, partnerPodNamespace)
	expectNil(err)
	return numInstantiatedTunnels, numConnectedTunnels
}

// helper shorthand to expect nil out of an error
func expectNil(err error) {
	gomega.Expect(err).To(gomega.BeNil())
}
