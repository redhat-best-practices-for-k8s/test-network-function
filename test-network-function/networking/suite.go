// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeport"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ping"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/podnodename"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/utils"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	commandportdeclared = "oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].ports'"
	commandportlisten   = "ss -tulwnH"
	ocCommandTimeOut    = time.Second * 10
	indexprotocolname   = 0
	indexport           = 4
	defaultNumPings     = 5
)

type ipVersion string

const (
	IPv4 ipVersion = "IPv4"
	IPv6 ipVersion = "IPv6"
)

type key struct {
	port     int
	protocol string
}

type Port []struct {
	ContainerPort int    `json:"containerPort"`
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
}

// netTestContext this is a data structure describing a network test context for a given subnet (e.g. network attachment)
// The test context defines a tester or test initiator, that is initiating the pings. It is selected randomly (first container in the list)
// It also defines a list of destination ping targets corresponding to the other containers IPs on this subnet
type netTestContext struct {
	// testerContainerNodeOc session context to access the node running the container selected to initiate tests
	testerContainerNodeOc *interactive.Oc
	// testerSource is the container select to initiate the ping tests on this given network
	testerSource containerIP
	// ipDestTargets List of containers to be pinged by the testerSource on this given network
	destTargets []containerIP
}

// containerIP holds a container identification and its IP for networking tests.
type containerIP struct {
	// ip address of the target container
	ip string
	// targetContainerIdentifier container identifier including namespace, pod name, container name, node name, and container UID
	containerIdentifier *configsections.ContainerIdentifier
}

func (testContext netTestContext) String() string {
	output := fmt.Sprintf("From initiating container: %s\n", testContext.testerSource.String())
	if len(testContext.destTargets) == 0 {
		output = "--> No target containers to test for this network" //nolint:goconst // this is only one time
	}
	for _, target := range testContext.destTargets {
		output += fmt.Sprintf("--> To target container: %s\n", target.String())
	}
	return output
}

func (cip *containerIP) String() string {
	return fmt.Sprintf("%s ( %s )",
		cip.ip,
		cip.containerIdentifier.String(),
	)
}

func printNetTestContextMap(netsUnderTest map[string]netTestContext) string {
	var output string
	if len(netsUnderTest) == 0 {
		output = "No networks to test.\n" //nolint:goconst // this is only one time
	}
	for netName, netUnderTest := range netsUnderTest {
		output += fmt.Sprintf("***Test for Network attachment: %s\n", netName)
		output += fmt.Sprintf("%s\n", netUnderTest.String())
	}
	return output
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
		ginkgo.AfterEach(env.CloseLocalShellContext)

		ginkgo.Context("Both Pods are on the Default network", func() {
			testDefaultNetworkConnectivity(env, defaultNumPings, IPv4)
			testDefaultNetworkConnectivity(env, defaultNumPings, IPv6)
		})

		ginkgo.Context("Both Pods are connected via a Multus Overlay Network", func() {
			testMultusNetworkConnectivity(env, defaultNumPings, IPv4)
			testMultusNetworkConnectivity(env, defaultNumPings, IPv6)
		})
		ginkgo.Context("Should not have type of nodePort", func() {
			testNodePort(env)
		})
		ginkgo.Context("Should not have type of listen port and declared port", func() {
			testListenAndDeclared(env)
		})
	}
})

// processContainerIpsPerNet takes a container ip addresses for a given network attachment's and uses it as a test target.
// The first container in the loop is selected as the test initiator. the Oc context of the container is used to initiate the pings
func processContainerIpsPerNet(containerID *configsections.ContainerIdentifier,
	netKey string,
	ipAddresses []string,
	netsUnderTest map[string]netTestContext,
	containerNodeOc *interactive.Oc,
	aIPVersion ipVersion) {
	ipAddressesFiltered := FilterIPListPerVersion(ipAddresses, aIPVersion)
	if len(ipAddressesFiltered) == 0 {
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
	if entry.testerContainerNodeOc == nil {
		tnf.ClaimFilePrintf("Pod %s, container %s selected to initiate ping tests", containerID.PodName, containerID.ContainerName)
		entry.testerSource.containerIdentifier = containerID
		entry.testerContainerNodeOc = containerNodeOc
		// if multiple interfaces are present for this network on this container/pod, pick the first one as the tester source ip
		entry.testerSource.ip = ipAddressesFiltered[firstIPIndex]
		// do no include tester's IP in the list of destination IPs to ping
		firstIPIndex++
	}

	for _, aIP := range ipAddressesFiltered[firstIPIndex:] {
		ipDestEntry := containerIP{}
		ipDestEntry.containerIdentifier = containerID
		ipDestEntry.ip = aIP
		entry.destTargets = append(entry.destTargets, ipDestEntry)
	}

	// Then reassign map entry
	netsUnderTest[netKey] = entry
}

func FilterIPListPerVersion(ipList []string, aIPVersion ipVersion) []string {
	var filteredIPList []string
	for _, aIP := range ipList {
		if ver, _ := getIPVersion(aIP); aIPVersion == ver {
			filteredIPList = append(filteredIPList, aIP)
		}
	}
	return filteredIPList
}

func getIPVersion(aIP string) (ipVersion, error) {
	ip := net.ParseIP(aIP)
	if ip == nil {
		return "", fmt.Errorf("%s is Not an IPv4 or an IPv6", aIP)
	}
	if ip.To4() != nil {
		return IPv4, nil
	}
	return IPv6, nil
}

// runNetworkingTests takes a map netTestContext, e.g. one context per network attachment
// and runs pings test with it. Returns a network name to a slice of bad target IPs map.
func runNetworkingTests(netsUnderTest map[string]netTestContext, count int, aIPVersion ipVersion) map[string][]string {
	tnf.ClaimFilePrintf("%s", printNetTestContextMap(netsUnderTest))
	log.Debugf("%s", printNetTestContextMap(netsUnderTest))
	if len(netsUnderTest) == 0 {
		ginkgo.Skip(fmt.Sprintf("There are no %s networks to test, skipping test", aIPVersion))
	}
	// maps a net name to a list of failed destination IPs
	badNets := map[string][]string{}

	// if no network can be tested, then we need to skip the test entirely.
	// If at least one network can be tested (e.g. > 2 IPs/ interfaces present), then we do not skip the test
	atLeastOneNetworkTested := false
	for netName, netUnderTest := range netsUnderTest {
		if len(netUnderTest.destTargets) == 0 {
			log.Warnf("There are no containers to ping for %s network %s. A minimum of 2 containers is needed to run a ping test (a source and a destination) Skipping test", aIPVersion, netName)
			tnf.ClaimFilePrintf("There are no containers to ping for %s network %s. Skip testing this network", aIPVersion, netName)
			continue
		}
		atLeastOneNetworkTested = true
		ginkgo.By(fmt.Sprintf("%s Ping tests on network %s. Number of target IPs: %d", aIPVersion, netName, len(netUnderTest.destTargets)))
		for _, aDestIP := range netUnderTest.destTargets {
			ginkgo.By(fmt.Sprintf("a %s Ping is issued from %s(%s) %s to %s(%s) %s",
				aIPVersion,
				netUnderTest.testerSource.containerIdentifier.PodName,
				netUnderTest.testerSource.containerIdentifier.ContainerName,
				netUnderTest.testerSource.ip, aDestIP.containerIdentifier.PodName,
				aDestIP.containerIdentifier.ContainerName,
				aDestIP.ip))
			testPass := testPing(netUnderTest.testerContainerNodeOc, netUnderTest.testerSource.containerIdentifier, aDestIP, count)
			if !testPass {
				if failedDestIps, netFound := badNets[netName]; netFound {
					badNets[netName] = append(failedDestIps, aDestIP.ip)
				} else {
					badNets[netName] = []string{aDestIP.ip}
				}
			}
		}
	}
	if !atLeastOneNetworkTested {
		ginkgo.Skip(fmt.Sprintf("There are no network to test for any %s networks, skipping test", aIPVersion))
	}
	return badNets
}
func testDefaultNetworkConnectivity(env *config.TestEnvironment, count int, aIPVersion ipVersion) {
	ginkgo.When("Testing Default network connectivity", func() {
		identifier := identifiers.TestICMPv4ConnectivityIdentifier
		if aIPVersion == IPv6 {
			identifier = identifiers.TestICMPv6ConnectivityIdentifier
		}
		testID := identifiers.XformToGinkgoItIdentifier(identifier)
		ginkgo.It(testID, ginkgo.Label(testID), func() {
			netsUnderTest := make(map[string]netTestContext)
			for _, pod := range env.PodsUnderTest {
				// The first container is used to get the network namespace
				aContainerInPod := pod.ContainerList[0]
				if _, ok := env.ContainersToExcludeFromConnectivityTests[aContainerInPod.ContainerIdentifier]; ok {
					tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from all connectivity tests", pod.Name)
					continue
				}
				netKey := "default" //nolint:goconst // only used once
				defaultIPAddress := pod.DefaultNetworkIPAddresses
				gomega.Expect(env).To(gomega.Not(gomega.BeNil()))
				gomega.Expect(env.NodesUnderTest[aContainerInPod.NodeName]).To(gomega.Not(gomega.BeNil()))
				gomega.Expect(env.NodesUnderTest[aContainerInPod.NodeName].DebugContainer.GetOc()).To(gomega.Not(gomega.BeNil()))
				nodeOc := env.NodesUnderTest[aContainerInPod.NodeName].DebugContainer.GetOc()
				processContainerIpsPerNet(&aContainerInPod.ContainerIdentifier, netKey, defaultIPAddress, netsUnderTest, nodeOc, aIPVersion)
			}
			badNets := runNetworkingTests(netsUnderTest, count, aIPVersion)

			if n := len(badNets); n > 0 {
				log.Warnf("Failed nets: %+v", badNets)
				ginkgo.Fail(fmt.Sprintf("%d nets failed the default network %s ping test.", n, aIPVersion))
			}
		})
	})
}
func testMultusNetworkConnectivity(env *config.TestEnvironment, count int, aIPVersion ipVersion) {
	identifier := identifiers.TestICMPv4ConnectivityMultusIdentifier
	if aIPVersion == IPv6 {
		identifier = identifiers.TestICMPv6ConnectivityMultusIdentifier
	}
	ginkgo.When("Testing Multus network connectivity", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifier)
		ginkgo.It(testID, ginkgo.Label(testID), func() {
			netsUnderTest := make(map[string]netTestContext)
			for _, pod := range env.PodsUnderTest {
				// The first container is used to get the network namespace
				aContainerInPod := pod.ContainerList[0]
				if _, ok := env.ContainersToExcludeFromConnectivityTests[aContainerInPod.ContainerIdentifier]; ok {
					tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from all connectivity tests", pod.Name)
					continue
				}
				if _, ok := env.ContainersToExcludeFromMultusConnectivityTests[aContainerInPod.ContainerIdentifier]; ok {
					tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from multus connectivity tests only", pod.Name)
					continue
				}
				for netKey, multusIPAddress := range pod.MultusIPAddressesPerNet {
					gomega.Expect(env).To(gomega.Not(gomega.BeNil()))
					gomega.Expect(env.NodesUnderTest[aContainerInPod.NodeName]).To(gomega.Not(gomega.BeNil()))
					gomega.Expect(env.NodesUnderTest[aContainerInPod.NodeName].DebugContainer.GetOc()).To(gomega.Not(gomega.BeNil()))
					nodeOc := env.NodesUnderTest[aContainerInPod.NodeName].DebugContainer.GetOc()
					processContainerIpsPerNet(&aContainerInPod.ContainerIdentifier, netKey, multusIPAddress, netsUnderTest, nodeOc, aIPVersion)
				}
			}
			badNets := runNetworkingTests(netsUnderTest, count, aIPVersion)

			if n := len(badNets); n > 0 {
				log.Warnf("Failed nets: %+v", badNets)
				ginkgo.Fail(fmt.Sprintf("%d nets failed the multus %s ping test.", n, aIPVersion))
			}
		})
	})
}

// Test that a container can ping a target IP address.
func testPing(initiatingPodNodeOc *interactive.Oc, sourceContainerID *configsections.ContainerIdentifier, targetContainerIP containerIP, count int) bool {
	log.Infof("Sending ICMP traffic(%s to %s)", initiatingPodNodeOc.GetPodName(), targetContainerIP.ip)
	env := config.GetTestEnvironment()
	gomega.Expect(env).To(gomega.Not(gomega.BeNil()))
	gomega.Expect(env.NodesUnderTest[sourceContainerID.NodeName]).To(gomega.Not(gomega.BeNil()))
	gomega.Expect(env.NodesUnderTest[sourceContainerID.NodeName].DebugContainer.GetOc()).To(gomega.Not(gomega.BeNil()))
	nodeOc := env.NodesUnderTest[sourceContainerID.NodeName].DebugContainer.GetOc()
	containerPID := utils.GetContainerPID(sourceContainerID.NodeName, nodeOc, sourceContainerID.ContainerUID, sourceContainerID.ContainerRuntime)
	pingTester := ping.NewPingNsenter(common.DefaultTimeout, containerPID, targetContainerIP.ip, count)
	test, err := tnf.NewTest(initiatingPodNodeOc.GetExpecter(), pingTester, []reel.Handler{pingTester}, initiatingPodNodeOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())

	sourcePodName := initiatingPodNodeOc.GetPodName()
	targetPodName := targetContainerIP.containerIdentifier.PodName

	testResult := false
	test.RunWithCallbacks(func() {
		transmitted, received, errors := pingTester.GetStats()
		if received == transmitted && errors == 0 {
			log.Infof("Ping test from pod %s to pod %s (ip %s) succeeded. Tx/Rx/Err: %d/%d/%d",
				sourcePodName, targetPodName, targetContainerIP.ip, transmitted, received, errors)
			testResult = true
		} else {
			tnf.ClaimFilePrintf("Ping test from pod %s to pod %s (ip: %s) failed. Tx/Rx/Err: %d/%d/%d",
				sourcePodName, targetPodName, targetContainerIP.ip, transmitted, received, errors)
		}
	}, func() {
		tnf.ClaimFilePrintf("FAILURE: Ping test from pod %s to pod %s (ip: %s) failed.",
			sourcePodName, targetPodName, targetContainerIP.ip)
	}, func(err error) {
		tnf.ClaimFilePrintf("ERROR: Ping test from pod %s to pod %s (ip: %s) failed. Error: %v",
			sourcePodName, targetPodName, targetContainerIP.ip, err)
		if reel.IsTimeout(err) {
			env.NodesUnderTest[sourceContainerID.NodeName].DebugContainer.CloseOc()
		}
	})

	return testResult
}

func testNodePort(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestServicesDoNotUseNodeportsIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badNamespaces := []string{}
		context := env.GetLocalShellContext()
		for _, ns := range env.NameSpacesUnderTest {
			ginkgo.By(fmt.Sprintf("Testing services in namespace %s", ns))
			tester := nodeport.NewNodePort(common.DefaultTimeout, ns)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("Namespace %s has one or more nodePort/s", ns)
				badNamespaces = append(badNamespaces, ns)
			}, func(err error) {
				tnf.ClaimFilePrintf("nodePort test on namespace %s failed. Error: %v", ns, err)
				badNamespaces = append(badNamespaces, ns)
			})
		}

		if n := len(badNamespaces); n > 0 {
			log.Warnf("Failed namespaces: %+v", badNamespaces)
			ginkgo.Fail(fmt.Sprintf("%d namespaces have nodePort/s.", n))
		}
	})
}

func parseVariables(res string, declaredPorts map[key]bool) error {
	var p Port
	err := json.Unmarshal([]byte(res), &p)
	if err != nil {
		return err
	}

	for element := range p {
		var k key
		k.port = p[element].ContainerPort
		k.protocol = p[element].Protocol
		declaredPorts[k] = true
	}
	return nil
}
func declaredPortList(container int, podName, podNamespace string, declaredPorts map[key]bool) error {
	ocCommandToExecute := fmt.Sprintf(commandportdeclared, podName, podNamespace, container)
	res, err := utils.ExecuteCommand(ocCommandToExecute, ocCommandTimeOut, interactive.GetContext(false))
	if err != nil {
		return err
	}
	err = parseVariables(res, declaredPorts)
	return err
}

func listeningPortList(commandlisten []string, nodeOc *interactive.Context, listeningPorts map[key]bool) error {
	var k key
	listeningPortCommand := strings.Join(commandlisten, " ")
	res, err := utils.ExecuteCommand(listeningPortCommand, ocCommandTimeOut, nodeOc)
	if err != nil {
		return err
	}
	lines := strings.Split(res, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		if indexprotocolname > len(fields) || indexport > len(fields) {
			return err
		}
		s := strings.Split(fields[indexport], ":")
		p, _ := strconv.Atoi(s[1])
		k.port = p
		k.protocol = strings.ToUpper(fields[indexprotocolname])
		k.protocol = strings.ReplaceAll(k.protocol, "\"", "")
		listeningPorts[k] = true
	}
	return nil
}

func checkIfListenIsDeclared(listeningPorts, declaredPorts map[key]bool) map[key]bool {
	res := make(map[key]bool)
	if len(listeningPorts) == 0 {
		return res
	}
	for k := range listeningPorts {
		_, ok := declaredPorts[k]
		if !ok {
			res[k] = listeningPorts[k]
		}
	}
	return res
}

func testListenAndDeclared(env *config.TestEnvironment) {
	var failedcontainers []string
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUndeclaredContainerPortsUsage)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		for _, cut := range env.ContainersUnderTest {
			declaredPorts := make(map[key]bool)
			listeningPorts := make(map[key]bool)
			container, _ := strconv.Atoi(cut.ContainerUID)
			err := declaredPortList(container, cut.PodName, cut.Namespace, declaredPorts)
			if err != nil {
				tnf.ClaimFilePrintf("Failed to get declared port for container %d due to %v, skipping pod %s", cut.ContainerUID, err, cut.PodName)
				continue
			}
			nodeName := podnodename.NewPodNodeName(common.DefaultTimeout, cut.PodName, cut.Namespace)
			context := env.GetLocalShellContext()
			test, err := tnf.NewTest(context.GetExpecter(), nodeName, []reel.Handler{nodeName}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			nodeOc := env.NodesUnderTest[nodeName.GetNodeName()].DebugContainer.GetOc()
			containerPID := utils.GetContainerPID(nodeName.GetNodeName(), nodeOc, cut.ContainerUID, cut.ContainerRuntime)

			commandlisten := []string{utils.AddNsenterPrefix(containerPID), commandportlisten}

			err = listeningPortList(commandlisten, nodeOc.Context, listeningPorts)
			if err != nil {
				tnf.ClaimFilePrintf("Failed to get listening port for pod name %s in pod namespace %s on container %s due to %v, skipping this pod", cut.PodName, cut.Namespace, cut.ContainerUID, err)
				failedcontainers = append(failedcontainers, fmt.Sprintf("podName: %s, Namespace: %s, container:%s", cut.PodName, cut.Namespace, cut.ContainerUID))
				continue
			}
			// compare between declaredPort,listeningPort
			undeclaredPorts := checkIfListenIsDeclared(listeningPorts, declaredPorts)
			for k := range undeclaredPorts {
				tnf.ClaimFilePrintf("The port %d on protocol %s in pod name %s and namespace is %s not declared on container %s.", k.port, k.protocol, cut.PodName, cut.Namespace, containerPID)
			}
			if len(undeclaredPorts) != 0 {
				failedcontainers = append(failedcontainers, fmt.Sprintf("podName: %s, Namespace: %s, container:%s", cut.PodName, cut.Namespace, cut.ContainerUID))
			}
		}

		if n := len(failedcontainers); n > 0 {
			ginkgo.Fail(fmt.Sprintf("Found %d pods with listening ports that had not been declared", n))
		}
	})
}
