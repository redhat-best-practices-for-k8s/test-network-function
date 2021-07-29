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

package lifecycle

import (
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	dp "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deployments"
	dd "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsdrain"
	dn "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsnodes"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/graceperiod"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeselector"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/owners"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	defaultTerminationGracePeriod = 30
	drainTimeoutMinutes           = 5
	partnerPod                    = "partner"
)

var (
	// nodeUncordonTestPath is the file location of the uncordon.json test case relative to the project root.
	nodeUncordonTestPath = path.Join("pkg", "tnf", "handlers", "nodeuncordon", "uncordon.json")

	// shutdownTestPath is the file location of shutdown.json test case relative to the project root.
	shutdownTestPath = path.Join("pkg", "tnf", "handlers", "shutdown", "shutdown.json")

	// shutdownTestDirectoryPath is the directory of the shutdown test
	shutdownTestDirectoryPath = path.Join("pkg", "tnf", "handlers", "shutdown")

	// relativeNodesTestPath is the relative path to the nodes.json test case.
	relativeNodesTestPath = path.Join(common.PathRelativeToRoot, nodeUncordonTestPath)

	// relativeShutdownTestPath is the relative path to the shutdown.json test case.
	relativeShutdownTestPath = path.Join(common.PathRelativeToRoot, shutdownTestPath)

	// relativeShutdownTestDirectoryPath is the directory of the shutdown directory
	relativeShutdownTestDirectoryPath = path.Join(common.PathRelativeToRoot, shutdownTestDirectoryPath)

	// podAntiAffinityTestPath is the file location of the podantiaffinity.json test case relative to the project root.
	podAntiAffinityTestPath = path.Join("pkg", "tnf", "handlers", "podantiaffinity", "podantiaffinity.json")

	// relativePodTestPath is the relative path to the podantiaffinity.json test case.
	relativePodTestPath = path.Join(common.PathRelativeToRoot, podAntiAffinityTestPath)
)

var drainTimeout = time.Duration(drainTimeoutMinutes) * time.Minute

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.LifecycleTestKey, func() {
	configData := common.ConfigurationData{}
	ginkgo.BeforeSuite(func() {
		common.Loadconfiguration(&configData)
		log.Info(configData.ContainersUnderTest)
	})
	ginkgo.BeforeEach(func() {
		common.ReloadConfiguration(&configData)
	})
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.LifecycleTestKey) {

		testNodeSelector(&configData)

		testGracePeriod(&configData)

		testShutdown(&configData)

		testPodAntiAffinity(&configData)
		if common.NonIntrusive() {
			testPodsRecreation(&configData)
		}

		testOwner(&configData)
	}
})

func testNodeSelector(configData *common.ConfigurationData) {
	ginkgo.It("Testing pod nodeSelector", func() {
		for _, cut := range configData.ContainersUnderTest {
			podName := cut.Oc.GetPodName()
			podNamespace := cut.Oc.GetPodNamespace()
			ginkgo.By(fmt.Sprintf("Testing pod nodeSelector %s/%s", cut.Oc.GetPodNamespace(), podName))
			defer results.RecordResult(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
			infoWriter := tnf.CreateTestExtraInfoWriter()
			tester := nodeselector.NewNodeSelector(common.DefaultTimeout, podName, podNamespace)
			test, err := tnf.NewTest(cut.Oc.GetExpecter(), tester, []reel.Handler{tester}, cut.Oc.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			if testResult != tnf.SUCCESS {
				msg := fmt.Sprintf("The pod specifies nodeSelector/nodeAffinity field, you might want to change it, %s %s", podNamespace, podName)
				log.Warn(msg)
				infoWriter(msg)
			}
		}
	})
}

func testGracePeriod(configData *common.ConfigurationData) {
	ginkgo.When("Test terminationGracePeriod ", func() {
		ginkgo.It("Testing pod terminationGracePeriod", func() {
			for _, cut := range configData.ContainersUnderTest {
				context := common.GetContext()
				podName := cut.Oc.GetPodName()
				podNamespace := cut.Oc.GetPodNamespace()
				ginkgo.By(fmt.Sprintf("Testing pod terminationGracePeriod %s %s", podNamespace, podName))
				defer results.RecordResult(identifiers.TestNonDefaultGracePeriodIdentifier)
				infoWriter := tnf.CreateTestExtraInfoWriter()
				tester := graceperiod.NewGracePeriod(common.DefaultTimeout, podName, podNamespace)
				test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
				gomega.Expect(err).To(gomega.BeNil())
				gracePeriod := tester.GetGracePeriod()
				if gracePeriod == defaultTerminationGracePeriod {
					msg := fmt.Sprintf("%s %s has terminationGracePeriod set to %d, you might want to change it", podNamespace, podName, defaultTerminationGracePeriod)
					log.Warn(msg)
					infoWriter(msg)
				}
			}
		})
	})
}

func testShutdown(configData *common.ConfigurationData) {
	ginkgo.When("Testing PUTs are configured with pre-stop lifecycle", func() {
		ginkgo.It("should have pre-stop configured", func() {
			for _, cut := range configData.ContainersUnderTest {
				podName := cut.Oc.GetPodName()
				podNamespace := cut.Oc.GetPodNamespace()
				ginkgo.By(fmt.Sprintf("should have pre-stop configured %s/%s", podNamespace, podName))
				defer results.RecordResult(identifiers.TestShudtownIdentifier)
				shutdownTest(podNamespace, podName)
			}
		})
	})
}

func shutdownTest(podNamespace, podName string) {
	context := common.GetContext()
	values := make(map[string]interface{})
	values["POD_NAMESPACE"] = podNamespace
	values["POD_NAME"] = podName
	values["GO_TEMPLATE_PATH"] = relativeShutdownTestDirectoryPath
	test, handlers, result, err := generic.NewGenericFromMap(relativeShutdownTestPath, common.RelativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())
	tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	testResult, err := tester.Run()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
}

func testPodsRecreation(configData *common.ConfigurationData) {
	var deployments dp.DeploymentMap
	var notReadyDeployments []string
	var nodesSorted []node // A slice version of nodes sorted by number of deployments descending
	ginkgo.It("Testing node draining effect of deplyment", func() {
		configData.SetNeedsRefresh()
		for _, cut := range configData.ContainersUnderTest {
			namespace := cut.Oc.GetPodNamespace()
			ginkgo.By(fmt.Sprintf("test deployment in namespace %s", namespace))
			deployments, notReadyDeployments = getDeployments(namespace)
			if len(deployments) == 0 {
				return
			}
			// We require that all deployments have the desired number of replicas and are all up to date
			if len(notReadyDeployments) == 0 {
				ginkgo.Skip("Can not test when deployments are not ready")
			}
			gomega.Expect(notReadyDeployments).To(gomega.BeEmpty())
			ginkgo.By("Should return map of nodes to deployments")
			nodesSorted = getDeploymentsNodes(namespace)
			if !common.NonIntrusive() {
				ginkgo.By("should create new replicas when node is drained")
				defer results.RecordResult(identifiers.TestPodRecreationIdentifier)
				testedDeployments := map[string]bool{}
				for _, n := range nodesSorted {
					oldLen := len(testedDeployments) // this starts with zero
					// mark tested deployments
					for d := range n.deployments {
						testedDeployments[d] = true
					}
					if oldLen == len(testedDeployments) {
						// If node does not add new deployments then skip it
						continue
					}
					// drain node
					drainNode(n.name) // should go in this
					// verify deployments are ready again
					_, notReadyDeployments = getDeployments(namespace)
					gomega.Expect(notReadyDeployments).To(gomega.BeEmpty()) // this is to make sure pods are created again
					uncordonNode(n.name)
					if len(testedDeployments) == len(deployments) {
						break
					}
				}
			}
		}
	})
}

type node struct {
	name        string
	deployments map[string]bool
}

func sortNodesMap(nodesMap dn.NodesMap) []node {
	nodes := make([]node, 0, len(nodesMap))
	for n, d := range nodesMap {
		nodes = append(nodes, node{n, d})
	}
	sort.Slice(nodes, func(i, j int) bool { return len(nodes[i].deployments) > len(nodes[j].deployments) })
	return nodes
}

func getDeploymentsNodes(namespace string) []node {
	context := common.GetContext()
	tester := dn.NewDeploymentsNodes(common.DefaultTimeout, namespace)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	nodes := tester.GetNodes()
	gomega.Expect(nodes).NotTo(gomega.BeEmpty())
	return sortNodesMap(nodes)
}

// getDeployments returns map of deployments and names of not-ready deployments
func getDeployments(namespace string) (deployments dp.DeploymentMap, notReadyDeployments []string) {
	context := common.GetContext()
	tester := dp.NewDeployments(common.DefaultTimeout, namespace)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)

	deployments = tester.GetDeployments()

	for name, d := range deployments {
		if d.Unavailable != 0 || d.Ready != d.Replicas || d.Available != d.Replicas || d.UpToDate != d.Replicas {
			notReadyDeployments = append(notReadyDeployments, name)
		}
	}

	return deployments, notReadyDeployments
}

func drainNode(node string) {
	context := common.GetContext()
	tester := dd.NewDeploymentsDrain(drainTimeout, node)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
}

func uncordonNode(node string) {
	context := common.GetContext()
	values := make(map[string]interface{})
	values["NODE"] = node
	test, handlers, result, err := generic.NewGenericFromMap(relativeNodesTestPath, common.RelativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(len(handlers)).To(gomega.Equal(1))
	gomega.Expect(test).ToNot(gomega.BeNil())

	tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	testResult, err := tester.Run()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
}

// Pod antiaffinity test for all deployments
func testPodAntiAffinity(configData *common.ConfigurationData) {
	var deployments dp.DeploymentMap
	ginkgo.When("CNF is designed in high availability mode ", func() {
		ginkgo.It("Should set pod replica number greater than 1 and corresponding pod anti-affinity rules in deployment", func() {
			for _, cut := range configData.ContainersUnderTest {
				podNamespace := cut.Oc.GetPodNamespace()
				defer results.RecordResult(identifiers.TestPodHighAvailabilityBestPractices)
				deployments, _ = getDeployments(podNamespace)
				if len(deployments) == 0 {
					return
				}
				for name, d := range deployments {
					if name != partnerPod {
						podAntiAffinity(name, podNamespace, d.Replicas)
					}
				}
			}
		})
	})
}

// check pod antiaffinity definition for a deployment
func podAntiAffinity(deployment, podNamespace string, replica int) {
	context := common.GetContext()
	values := make(map[string]interface{})
	values["DEPLOYMENT_NAME"] = deployment
	values["DEPLOYMENT_NAMESPACE"] = podNamespace
	infoWriter := tnf.CreateTestExtraInfoWriter()
	test, handlers, result, err := generic.NewGenericFromMap(relativePodTestPath, common.RelativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(len(handlers)).To(gomega.Equal(1))
	gomega.Expect(test).ToNot(gomega.BeNil())
	tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	testResult, err := tester.Run()
	if testResult != tnf.SUCCESS {
		if replica > 1 {
			msg := fmt.Sprintf("The deployment replica count is %d, but a podAntiAffinity rule is not defined, "+
				"you might want to change it in deployment %s in namespace %s", replica, deployment, podNamespace)
			log.Warn(msg)
			infoWriter(msg)
		} else {
			msg := fmt.Sprintf("The deployment replica count is %d. Pod replica should be > 1 with an "+
				"podAntiAffinity rule defined . You might want to change it in deployment %s in namespace %s",
				replica, deployment, podNamespace)
			log.Warn(msg)
			infoWriter(msg)
		}
	}
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
}

func testOwner(configData *common.ConfigurationData) {
	ginkgo.When("Testing owners of CNF pod", func() {
		ginkgo.It("Should be only ReplicaSet", func() {
			for _, cut := range configData.ContainersUnderTest {
				podNamespace := cut.Oc.GetPodNamespace()
				podName := cut.Oc.GetPodName()
				ginkgo.By(fmt.Sprintf("Should be ReplicaSet %s %s", podNamespace, podName))
				defer results.RecordResult(identifiers.TestPodDeploymentBestPracticesIdentifier)
				context := common.GetContext()
				tester := owners.NewOwners(common.DefaultTimeout, podNamespace, podName)
				test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
				gomega.Expect(err).To(gomega.BeNil())
			}
		})
	})
}
