package itc

import (
	"errors"
	"github.com/redhat-nfvpe/test-network-function/internal/oc"
	"regexp"
	"strconv"
)

const (
	// Return value when there is an error calculating the number of tunnels
	DefaultNumTunnels = -1
	// Return value when there is an error calculating the number of packets
	DefaultPacketCount         = -1
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

// Creates IPSEC tunnels through wrapping "itc" (ike-testctl alias) on the remote container.
func Create(pod string, container string, namespace string, numTunnels, messagesPerSecond int) error {
	remoteCommand := []string{itcCommand, itcCreateCommand, strconv.Itoa(numTunnels), strconv.Itoa(messagesPerSecond)}
	// Ignore stdout, since "itc create" has no output.
	_, err := oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
	return err
}

// Helper to count the number of matching tunnels from "itc summary" output.
func countTunnelMatches(output string, regex string) int {
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
func parseItcSummaryOutput(output string) (int, int, error) {
	successfulIkeOutput := regexp.MustCompile(itcIkeSuccessRegex)
	successfulIkeOutputMatch := successfulIkeOutput.FindStringSubmatch(output)
	if successfulIkeOutputMatch != nil {
		numInstantiatedTunnels := countTunnelMatches(output, itcIkeTunnelRegex)
		numCreatedTunnels := countTunnelMatches(output, itcIkeTunnelConnectedRegex)
		return numInstantiatedTunnels, numCreatedTunnels, nil
	}
	return DefaultNumTunnels, DefaultNumTunnels, errors.New("itc ike command failed")
}

// Extracts the number of instantiated tunnels and the number of successfully connected tunnels.
func IkeSummary(pod string, container string, namespace string) (int, int, error) {
	remoteCommand := []string{itcCommand, itcIkeCommand}
	stdout, err := oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
	if err != nil {
		return DefaultNumTunnels, DefaultNumTunnels, err
	}
	return parseItcSummaryOutput(stdout)
}

// Extracts the result of running "itc summary" on a remote container.
func Summary(pod string, container string, namespace string) (string, error) {
	remoteCommand := []string{itcCommand, itcSummaryCommand}
	return oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
}

// Helper for filtering "itc summary" output
func getPacketTypeRegex(packetType string) string {
	return `(?m)^\s+` + packetType + `\:\s+(\d+).*`
}

// Extracts packet count for a particular type from "itc summary".
func GetPacketCount(pod string, container string, namespace string, packetType string) (int, error) {
	packetTypeRegex := getPacketTypeRegex(packetType)
	re := regexp.MustCompile(packetTypeRegex)
	summary, err := Summary(pod, container, namespace)
	if err != nil {
		return DefaultPacketCount, err
	}
	match := re.FindStringSubmatch(summary)
	if match == nil || len(match) <= 1 {
		return DefaultPacketCount, errors.New("couldn't find a match for " + packetType)
	}
	return strconv.Atoi(match[1])
}

// Extracts ICMP Reply count.
func GetItcIcmpReplyCount(pod string, container string, namespace string) (int, error) {
	return GetPacketCount(pod, container, namespace, itcIcmpRepliesRecieved)
}

// Extracts the number of UDP replies received.
func GetUdpReceivedCount(pod string, container string, namespace string) (int, error) {
	return GetPacketCount(pod, container, namespace, udpReceivedCount)
}

// A wrapper for "itc ping".
func Ping(pod string, container string, namespace string, tunnelIndex int, targetAddress string, messagesPerSecond int, packetCount int, dataLength int) (string, error) {
	remoteCommand := []string{itcCommand, itcPingCommand, strconv.Itoa(tunnelIndex), targetAddress,
		strconv.Itoa(messagesPerSecond), strconv.Itoa(packetCount), strconv.Itoa(dataLength)}
	return oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
}

// A wrapper for "itc senddata".
func SendData(pod string, container string, namespace string, tunnelIndex int, messagesPerSecond int, packetCount int, dataLength int) (string, error) {
	remoteCommand := []string{itcCommand, itcSendDataCommand, strconv.Itoa(tunnelIndex),
		strconv.Itoa(messagesPerSecond), strconv.Itoa(packetCount), strconv.Itoa(dataLength)}
	return oc.InvokeOCCommand(pod, container, namespace, remoteCommand)
}
