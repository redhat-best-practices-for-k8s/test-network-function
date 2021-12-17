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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"

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
	CommandPortDeclared = "oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].ports'"
	CommandPortListen   = "ss -tulwnH"
	defaultNumPings     = 5
	ocCommandTimeOut    = time.Second * 10
	indexProtocolName   = 0
	indexPort           = 4
)

var (
	expectersVerboseModeEnabled = false
	PortDeclaredAndListen       []PortDeclared
	nodeListen                  PortListen
	protocolName                = ""
	port                        = ""
	name                        = ""
	flag                        = 0
)

type PortDeclared struct {
	containerPort int
	name          string
	protocol      string
}

type PortListen struct {
	port     int
	protocol string
}

// netTestContext this is a data structure describing a network test context for a given subnet (e.g. network attachment)
// The test context defines a tester or test initiator, that is initating the pings. It is selected randomly (first container in the list)
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

func (testContext netTestContext) String() (output string) {
	output = fmt.Sprintf("From initiating container: %s\n", testContext.testerSource.String())
	if len(testContext.destTargets) == 0 {
		output = "--> No target containers to test for this network" //nolint:goconst // this is only one time
	}
	for _, target := range testContext.destTargets {
		output += fmt.Sprintf("--> To target container: %s\n", target.String())
	}
	return
}

func (cip *containerIP) String() (output string) {
	output = fmt.Sprintf("%s ( %s )",
		cip.ip,
		cip.containerIdentifier.String(),
	)
	return
}

func printNetTestContextMap(netsUnderTest map[string]netTestContext) (output string) {
	if len(netsUnderTest) == 0 {
		output = "No networks to test.\n" //nolint:goconst // this is only one time
	}
	for netName, netUnderTest := range netsUnderTest {
		output += fmt.Sprintf("***Test for Network attachment: %s\n", netName)
		output += fmt.Sprintf("%s\n", netUnderTest.String())
	}
	return
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
		ginkgo.Context("Should not have type of listen port and declared port", func() {
			testListenAndDeclared(env)
		})
	}
})

// processContainerIpsPerNet takes a container ip addresses for a given network attachment's and uses it as a test target.
// The first container in the loop is selected as the test initiator. the Oc context of the container is used to initiate the pings
func processContainerIpsPerNet(containerID *configsections.ContainerIdentifier,
	netKey string,
	ipAddress []string,
	netsUnderTest map[string]netTestContext,
	containerNodeOc *interactive.Oc) {
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
	if entry.testerContainerNodeOc == nil {
		tnf.ClaimFilePrintf("Pod %s, container %s selected to initiate ping tests", containerID.PodName, containerID.ContainerName)
		entry.testerSource.containerIdentifier = containerID
		entry.testerContainerNodeOc = containerNodeOc
		// if multiple interfaces are present for this network on this container/pod, pick the first one as the tester source ip
		entry.testerSource.ip = ipAddress[firstIPIndex]
		// do no include tester's IP in the list of destination IPs to ping
		firstIPIndex++
	}

	for _, aIP := range ipAddress[firstIPIndex:] {
		ipDestEntry := containerIP{}
		ipDestEntry.containerIdentifier = containerID
		ipDestEntry.ip = aIP
		entry.destTargets = append(entry.destTargets, ipDestEntry)
	}

	// Then reassign map entry
	netsUnderTest[netKey] = entry
}

// runNetworkingTests takes a map netTestContext, e.g. one context per network attachment
// and runs pings test with it
func runNetworkingTests(netsUnderTest map[string]netTestContext, count int) {
	tnf.ClaimFilePrintf("%s", printNetTestContextMap(netsUnderTest))
	log.Debugf("%s", printNetTestContextMap(netsUnderTest))
	if len(netsUnderTest) == 0 {
		ginkgo.Skip("There are no networks to test, skipping test")
	}
	for netName, netUnderTest := range netsUnderTest {
		if len(netUnderTest.destTargets) == 0 {
			ginkgo.Skip(fmt.Sprintf("There are no containers to ping for network %s. A minimum of 2 containers is needed to run a ping test (a source and a destination) Skipping test", netName))
		}
		m := make(map[string]bool)
		for _, aDestIP := range netUnderTest.destTargets {
			podName := aDestIP.containerIdentifier.PodName
			if _, ok := m[podName]; ok {
				continue
			}
			m[podName] = true
			ginkgo.By(fmt.Sprintf("a Ping is issued from %s(%s) %s to %s(%s) %s",
				netUnderTest.testerSource.containerIdentifier.PodName,
				netUnderTest.testerSource.containerIdentifier.ContainerName,
				netUnderTest.testerSource.ip, aDestIP.containerIdentifier.PodName,
				aDestIP.containerIdentifier.ContainerName,
				aDestIP.ip))
			testPing(netUnderTest.testerContainerNodeOc,
				netUnderTest.testerSource.containerIdentifier.NodeName,
				netUnderTest.testerSource.containerIdentifier.ContainerUID,
				netUnderTest.testerSource.containerIdentifier.ContainerRuntime,
				aDestIP.ip,
				count)
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
					tnf.ClaimFilePrintf("Skipping container %s because it is excluded from connectivity tests (default interface)", cut.PodName)
					continue
				}
				netKey := "default" //nolint:goconst // only used once
				defaultIPAddress := []string{cut.DefaultNetworkIPAddress}
				gomega.Expect(env).To(gomega.Not(gomega.BeNil()))
				gomega.Expect(env.NodesUnderTest[cut.NodeName]).To(gomega.Not(gomega.BeNil()))
				gomega.Expect(env.NodesUnderTest[cut.NodeName].DebugContainer.GetOc()).To(gomega.Not(gomega.BeNil()))
				nodeOc := env.NodesUnderTest[cut.NodeName].DebugContainer.GetOc()
				processContainerIpsPerNet(&cut.ContainerIdentifier, netKey, defaultIPAddress, netsUnderTest, nodeOc)
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
					tnf.ClaimFilePrintf("Skipping container %s because it is excluded from connectivity tests (multus interface)", cut.PodName)
					continue
				}
				for netKey, multusIPAddress := range cut.MultusIPAddressesPerNet {
					gomega.Expect(env).To(gomega.Not(gomega.BeNil()))
					gomega.Expect(env.NodesUnderTest[cut.NodeName]).To(gomega.Not(gomega.BeNil()))
					gomega.Expect(env.NodesUnderTest[cut.NodeName].DebugContainer.GetOc()).To(gomega.Not(gomega.BeNil()))
					nodeOc := env.NodesUnderTest[cut.NodeName].DebugContainer.GetOc()
					processContainerIpsPerNet(&cut.ContainerIdentifier, netKey, multusIPAddress, netsUnderTest, nodeOc)
				}
			}
			runNetworkingTests(netsUnderTest, count)
		})
	})
}

// Test that a container can ping a target IP address.
func testPing(initiatingPodNodeOc *interactive.Oc, nodeName, containerID, runtime, targetPodIPAddress string, count int) {
	log.Infof("Sending ICMP traffic(%s to %s)", initiatingPodNodeOc.GetPodName(), targetPodIPAddress)
	env := config.GetTestEnvironment()
	gomega.Expect(env).To(gomega.Not(gomega.BeNil()))
	gomega.Expect(env.NodesUnderTest[nodeName]).To(gomega.Not(gomega.BeNil()))
	gomega.Expect(env.NodesUnderTest[nodeName].DebugContainer.GetOc()).To(gomega.Not(gomega.BeNil()))
	nodeOc := env.NodesUnderTest[nodeName].DebugContainer.GetOc()
	containerPID := utils.GetContainerPID(nodeName, nodeOc, containerID, runtime)
	pingTester := ping.NewPingNsenter(common.DefaultTimeout, containerPID, targetPodIPAddress, count)
	test, err := tnf.NewTest(initiatingPodNodeOc.GetExpecter(), pingTester, []reel.Handler{pingTester}, initiatingPodNodeOc.GetErrorChannel())
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

func getDeclaredPortList(command string, containerNum int, podName, podNamespace string) []PortDeclared {
	var result []PortDeclared
	ocCommandToExecute := fmt.Sprintf(command, podName, podNamespace, containerNum)
	res, _ := utils.ExecuteCommand(ocCommandToExecute, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled))
	x := strings.Split(res, "\n")
	for _, i := range x {
		fmt.Println(i)
		if strings.Contains(i, "containerPort") {
			s := strings.Split(i, ":")
			s = strings.Split(s[1], ",")
			port = s[0]
		}
		if strings.Contains(i, "name") {
			s := strings.Split(i, ":")
			s = strings.Split(s[1], ",")
			name = s[0]
		}
		if strings.Contains(i, "protocol") {
			s := strings.Split(i, ":")
			protocolName = s[1]
			flag = 1
		}
		if flag < 1 {
			continue
		}
		p, _ := strconv.Atoi(strings.TrimSpace(port))
		noderes := PortDeclared{containerPort: p, name: name, protocol: protocolName}
		result = append(result, noderes)
		flag = 0
	}
	return result
}

func getListeningPortList(ocCommand string) []PortListen {
	res, _ := utils.ExecuteCommand(ocCommand, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled))
	splitRes := strings.Split(res, "\n")
	var result []PortListen
	protocolName := ""
	for _, line := range splitRes {
		fields := strings.Fields(line)
		protocolName = fields[indexProtocolName]
		s := strings.Split(fields[indexPort], ":")
		p, _ := strconv.Atoi(s[1])
		nodeListen = PortListen{port: p, protocol: protocolName}
		result = append(result, nodeListen)
	}
	return result
}

func compareList(declaredPortList []PortDeclared, listeningPortList []PortListen) []PortDeclared {
	var result []PortDeclared
	var temp PortDeclared
	indexElement1 := 0
	indexElement2 := 0

	for indexElement1 < len(declaredPortList) && indexElement2 < len(listeningPortList) {
		if declaredPortList[indexElement1].containerPort == listeningPortList[indexElement2].port {
			temp = declaredPortList[indexElement1]
			result = append(result, temp)
			indexElement1++
			indexElement2++
			continue
		}
		if declaredPortList[indexElement1].containerPort < listeningPortList[indexElement2].port {
			indexElement1++
			continue
		}
		indexElement2++
	}
	return result
}

func testListenAndDeclared(env *config.TestEnvironment) []PortDeclared {
	var declaredPort []PortDeclared
	var listeningPort []PortListen
	var PortDeclaredAndListen []PortDeclared
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestServicesDoNotUseNodeportsIdentifier)
	ginkgo.It(testID, func() {
		for _, podUnderTest := range env.PodsUnderTest {
			ContainerCount := podUnderTest.ContainerCount
			for i := 0; i < ContainerCount; i++ {
				podName := podUnderTest.Name
				podNamespace := podUnderTest.Namespace
				temp := getDeclaredPortList(CommandPortDeclared, i, podName, podNamespace)
				declaredPort = append(declaredPort, temp...)
			}
			temp := getListeningPortList(CommandPortListen)
			listeningPort = append(listeningPort, temp...)
			// sorting the lists
			sort.SliceStable(declaredPort, func(i, j int) bool {
				return declaredPort[i].containerPort < declaredPort[j].containerPort
			})
			sort.SliceStable(listeningPort, func(i, j int) bool {
				return listeningPort[i].port < listeningPort[j].port
			})
			// compare between declaredPort,listeningPort and return the common.
			res := compareList(declaredPort, listeningPort)
			PortDeclaredAndListen = append(PortDeclaredAndListen, res...)
		}
	})
	return PortDeclaredAndListen
}
