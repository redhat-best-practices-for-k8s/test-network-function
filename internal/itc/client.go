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

package itc

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/redhat-nfvpe/test-network-function/internal/oc"
)

const (
	// ErrorNumTunnels is the value returned when there is an error calculating the number of tunnels.
	ErrorNumTunnels = -1
	// ErrorPacketCount is the return value when there is an error calculating the number of desired packets.
	ErrorPacketCount           = -1
	itcCommand                 = "itc"
	itcCreateCommand           = "create"
	itcIcmpRepliesRecieved     = `ESPRawPacketsReceivedICMPReplyPayloadReceived`
	itcIkeCommand              = "ike"
	itcIkeSuccessRegex         = `^\s+TUNNEL\s+(\|\s+[a-zA-Z\s\-]+\s+){13}`
	itcIkeTunnelRegex          = `(?m)^\s+(\d+)\s+(\|\s+((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))\s+){2}`
	itcIkeTunnelConnectedRegex = `(?m)^\s+(\d+)\s+(\|\s+((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))\s+){2}.*Connected`
	itcPingCommand             = "ping"
	itcSendDataCommand         = "senddata"
	itcSummaryCommand          = "summary"
	udpReceivedCount           = `ESPRawPacketsReceivedUDP`
)

// Create creates IPSEC tunnels through wrapping "itc" (ike-testctl alias) on the remote container.
func Create(pod, container, namespace string, numTunnels, messagesPerSecond int) error {
	remoteCommand := []string{itcCommand, itcCreateCommand, strconv.Itoa(numTunnels), strconv.Itoa(messagesPerSecond)}
	// Ignore stdout, since "itc create" has no output.
	_, err := oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
	return err
}

// Helper to count the number of matching tunnels from "itc summary" output.
func countTunnelMatches(output, regex string) int {
	re := regexp.MustCompile(regex)
	startIndex := -1
	match := re.FindAllString(output, startIndex)
	if match == nil {
		return 0
	}
	return len(match)
}

// Parse the itc ike output and extract the number of instantiated tunnels and the number of successfully connected
// tunnels.
func parseItcSummaryOutput(output string) (numInstantiated, numCreated int, err error) {
	successfulIkeOutput := regexp.MustCompile(itcIkeSuccessRegex)
	successfulIkeOutputMatch := successfulIkeOutput.FindStringSubmatch(output)
	if successfulIkeOutputMatch != nil {
		numInstantiatedTunnels := countTunnelMatches(output, itcIkeTunnelRegex)
		numCreatedTunnels := countTunnelMatches(output, itcIkeTunnelConnectedRegex)
		return numInstantiatedTunnels, numCreatedTunnels, nil
	}
	return ErrorNumTunnels, ErrorNumTunnels, errors.New("itc ike command failed")
}

// IkeSummary extracts the number of instantiated tunnels and the number of successfully connected tunnels.
func IkeSummary(pod, container, namespace string) (numInstantiated, numCreated int, err error) {
	remoteCommand := []string{itcCommand, itcIkeCommand}
	stdout, err := oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
	if err != nil {
		return ErrorNumTunnels, ErrorNumTunnels, err
	}
	return parseItcSummaryOutput(stdout)
}

// Summary extracts the result of running "itc summary" on a remote container.
func Summary(pod, container, namespace string) (string, error) {
	remoteCommand := []string{itcCommand, itcSummaryCommand}
	return oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
}

// Helper for filtering "itc summary" output
func getPacketTypeRegex(packetType string) string {
	return `(?m)^\s+` + packetType + `\:\s+(\d+).*`
}

// GetPacketCount extracts packet count for a particular type from "itc summary".
func GetPacketCount(pod, container, namespace, packetType string) (int, error) {
	packetTypeRegex := getPacketTypeRegex(packetType)
	re := regexp.MustCompile(packetTypeRegex)
	summary, err := Summary(pod, container, namespace)
	if err != nil {
		return ErrorPacketCount, err
	}
	match := re.FindStringSubmatch(summary)
	if match == nil || len(match) <= 1 {
		return ErrorPacketCount, errors.New("couldn't find a match for " + packetType)
	}
	return strconv.Atoi(match[1])
}

// GetItcIcmpReplyCount extracts ICMP Reply count.
func GetItcIcmpReplyCount(pod, container, namespace string) (int, error) {
	return GetPacketCount(pod, container, namespace, itcIcmpRepliesRecieved)
}

// GetUDPReceivedCount extracts the number of UDP replies received.
func GetUDPReceivedCount(pod, container, namespace string) (int, error) {
	return GetPacketCount(pod, container, namespace, udpReceivedCount)
}

// Ping is a wrapper for "itc ping".
func Ping(pod, container, namespace string, tunnelIndex int, targetAddress string, messagesPerSecond, packetCount, dataLength int) (string, error) {
	remoteCommand := []string{itcCommand, itcPingCommand, strconv.Itoa(tunnelIndex), targetAddress,
		strconv.Itoa(messagesPerSecond), strconv.Itoa(packetCount), strconv.Itoa(dataLength)}
	return oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
}

// SendData is a wrapper for "itc senddata".
func SendData(pod, container, namespace string, tunnelIndex, messagesPerSecond, packetCount, dataLength int) (string, error) {
	remoteCommand := []string{itcCommand, itcSendDataCommand, strconv.Itoa(tunnelIndex),
		strconv.Itoa(messagesPerSecond), strconv.Itoa(packetCount), strconv.Itoa(dataLength)}
	return oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
}
