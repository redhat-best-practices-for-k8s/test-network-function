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

func testDefaultNetworkConnectivity(env *config.TestEnvironment, count int) {
	ginkgo.When("Testing network connectivity", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityIdentifier)
		ginkgo.It(testID, func() {
			if env.TestOrchestrator == nil {
				ginkgo.Skip("Orchestrator is not deployed, skip this test")
			}
			found := false
			for _, cut := range env.ContainersUnderTest {
				if _, ok := env.ContainersToExcludeFromConnectivityTests[cut.ContainerIdentifier]; ok {
					tnf.ClaimFilePrintf("Skipping container %s because it is excluded from connectivity tests (default)", cut.ContainerConfiguration.PodName)

					continue
				}
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
			if !found {
				ginkgo.Skip("No container found suitable for connectivity test")
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
					tnf.ClaimFilePrintf("Skipping container %s because it is excluded from connectivity tests (multus)", cut.ContainerConfiguration.PodName)
					continue
				}
				if len(cut.ContainerConfiguration.MultusIPAddresses) == 0 {
					tnf.ClaimFilePrintf("Skipping container %s for multus test because no multus IPs are present", cut.ContainerConfiguration.PodName)
					continue
				}
				found = true

				for _, multusIPAddress := range cut.ContainerConfiguration.MultusIPAddresses {
					testOrchestrator := env.TestOrchestrator
					ginkgo.By(fmt.Sprintf("a Ping is issued from %s(%s) to %s(%s) %s", testOrchestrator.Oc.GetPodName(),
						testOrchestrator.Oc.GetPodContainerName(), cut.Oc.GetPodName(), cut.Oc.GetPodContainerName(),
						multusIPAddress))
					testPing(testOrchestrator.Oc, multusIPAddress, count)
				}
			}
			if !found {
				ginkgo.Skip("No container found suitable for Multus connectivity test")
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
	res := utils.ExecuteCommand(ocCommandToExecute, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("can't run command: ", command)
	})
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
	res := utils.ExecuteCommand(ocCommand, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("can't run command: ", ocCommand)
	})
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
