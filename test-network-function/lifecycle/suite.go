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
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/scaling"
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
	scalingTimeout                = 60 * time.Second
	scalingPollingPeriod          = 1 * time.Second
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
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.LifecycleTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))

		})

		testNodeSelector(env)

		testGracePeriod(env)

		testShutdown(env)

		testPodAntiAffinity(env)

		if common.Intrusive() {
			testPodsRecreation(env)

			testScaling(env)
		}

		testOwner(env)
	}
})

func waitForAllDeploymentsReady(namespace string, timeout, pollingPeriod time.Duration) {
	gomega.Eventually(func() []string {
		_, notReadyDeployments := getDeployments(namespace)
		return notReadyDeployments
	}, timeout, pollingPeriod).Should(gomega.HaveLen(0))
}

// restoreDeployments is the last attempt to restore the original test deployments' replicaCount
func restoreDeployments(env *config.TestEnvironment) {
	for _, deployment := range env.DeploymentsUnderTest {
		// For each test deployment in the namespace, refresh the current replicas and compare.
		deployments, notReadyDeployments := getDeployments(deployment.Namespace)

		if len(notReadyDeployments) > 0 {
			// Wait until the deployment is ready
			waitForAllDeploymentsReady(deployment.Namespace, scalingTimeout, scalingPollingPeriod)
		}

		if deployments[deployment.Name].Replicas != deployment.Replicas {
			log.Warn("Deployment ", deployment.Name, " replicaCount (", deployment.Replicas, ") needs to be restored.")

			// Try to scale to the original deployment's replicaCount.
			runScalingTest(deployment)

			env.SetNeedsRefresh()
		}
	}
}

func closeOcSessionsByDeployment(containers map[configsections.ContainerIdentifier]*config.Container, deployment configsections.Deployment) {
	for cid, c := range containers {
		if cid.Namespace == deployment.Namespace && strings.HasPrefix(cid.PodName, deployment.Name+"-") {
			log.Infof("Closing session to %s %s", cid.PodName, cid.ContainerName)
			c.Oc.Close()
			delete(containers, cid)
		}
	}
}

// runScalingTest Runs a Scaling handler TC and waits for all the deployments to be ready.
func runScalingTest(deployment configsections.Deployment) {
	handler := scaling.NewScaling(common.DefaultTimeout, deployment.Namespace, deployment.Name, deployment.Replicas)
	test, err := tnf.NewTest(common.GetContext().GetExpecter(), handler, []reel.Handler{handler}, common.GetContext().GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)

	// Wait until the deployment is ready
	waitForAllDeploymentsReady(deployment.Namespace, scalingTimeout, scalingPollingPeriod)
}

func testScaling(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestScalingIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing deployment scaling")
		defer results.RecordResult(identifiers.TestScalingIdentifier)
		defer restoreDeployments(env)

		if len(env.DeploymentsUnderTest) == 0 {
			ginkgo.Skip("No test deployments found.")
		}

		for _, deployment := range env.DeploymentsUnderTest {
			ginkgo.By(fmt.Sprintf("Scaling Deployment=%s, Replicas=%d (ns=%s)",
				deployment.Name, deployment.Replicas, deployment.Namespace))

			closeOcSessionsByDeployment(env.ContainersUnderTest, deployment)

			replicaCount := deployment.Replicas

			// ScaleIn, removing one pod from the replicaCount
			deployment.Replicas = replicaCount - 1
			runScalingTest(deployment)

			// Scaleout, restoring the original replicaCount number
			deployment.Replicas = replicaCount
			runScalingTest(deployment)

			// Ensure next tests/test suites receive a refreshed config.
			env.SetNeedsRefresh()
		}
	})
}

func testNodeSelector(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing pod nodeSelector")
		context := common.GetContext()
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			ginkgo.By(fmt.Sprintf("Testing pod nodeSelector %s/%s", podNamespace, podName))
			defer results.RecordResult(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
			infoWriter := tnf.CreateTestExtraInfoWriter()
			tester := nodeselector.NewNodeSelector(common.DefaultTimeout, podName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
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

func testGracePeriod(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestNonDefaultGracePeriodIdentifier)
	ginkgo.It(testID+" ", func() {
		ginkgo.By("Test terminationGracePeriod")
		context := common.GetContext()
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
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
}

func testShutdown(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestShudtownIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing PUTs are configured with pre-stop lifecycle")
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			ginkgo.By(fmt.Sprintf("should have pre-stop configured %s/%s", podNamespace, podName))
			defer results.RecordResult(identifiers.TestShudtownIdentifier)
			shutdownTest(podNamespace, podName)
		}
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

func testPodsRecreation(env *config.TestEnvironment) {
	var deployments dp.DeploymentMap
	var notReadyDeployments []string

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodRecreationIdentifier)
	ginkgo.It(testID, func() {
		env.SetNeedsRefresh()
		ginkgo.By("Testing node draining effect of deployment")

		ginkgo.By(fmt.Sprintf("test deployment in namespace %s", env.NameSpaceUnderTest))
		deployments, notReadyDeployments = getDeployments(env.NameSpaceUnderTest)
		if len(deployments) == 0 {
			return
		}
		// We require that all deployments have the desired number of replicas and are all up to date
		if len(notReadyDeployments) != 0 {
			ginkgo.Skip("Can not test when deployments are not ready")
		}

		ginkgo.By("Should return map of nodes to deployments")
		nodesSorted := getDeploymentsNodes(env.NameSpaceUnderTest)

		ginkgo.By("should create new replicas when node is drained")
		defer results.RecordResult(identifiers.TestPodRecreationIdentifier)
		for _, n := range nodesSorted {
			closeOcSessionsByNode(env.ContainersUnderTest, n.name)
			closeOcSessionsByNode(env.PartnerContainers, n.name)
			// drain node
			drainNode(n.name) // should go in this
			// verify deployments are ready again
			_, notReadyDeployments = getDeployments(env.NameSpaceUnderTest)
			if len(notReadyDeployments) != 0 {
				uncordonNode(n.name)
				ginkgo.Fail(fmt.Sprintf("did not create replicas when node %s is drained", n.name))
			}

			uncordonNode(n.name)
		}
	})
}

func closeOcSessionsByNode(containers map[configsections.ContainerIdentifier]*config.Container, nodeName string) {
	for cid, c := range containers {
		if cid.NodeName == nodeName {
			log.Infof("Closing session to %s %s", cid.PodName, cid.ContainerName)
			c.Oc.Close()
			delete(containers, cid)
		}
	}
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
func testPodAntiAffinity(env *config.TestEnvironment) {
	ginkgo.When("CNF is designed in high availability mode ", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHighAvailabilityBestPractices)
		ginkgo.It(testID, func() {
			ginkgo.By("Should set pod replica number greater than 1 and corresponding pod anti-affinity rules in deployment")
			if len(env.DeploymentsUnderTest) == 0 {
				ginkgo.Skip("No test deployments found.")
			}

			for _, deployment := range env.DeploymentsUnderTest {
				defer results.RecordResult(identifiers.TestPodHighAvailabilityBestPractices)
				ginkgo.By(fmt.Sprintf("Testing Pod AntiAffinity on Deployment=%s, Replicas=%d (ns=%s)",
					deployment.Name, deployment.Replicas, deployment.Namespace))
				podAntiAffinity(deployment.Name, deployment.Namespace, deployment.Replicas)
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

func testOwner(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodDeploymentBestPracticesIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing owners of CNF pod, should be replicas Set")
		context := common.GetContext()
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			ginkgo.By(fmt.Sprintf("Should be ReplicaSet %s %s", podNamespace, podName))
			defer results.RecordResult(identifiers.TestPodDeploymentBestPracticesIdentifier)
			tester := owners.NewOwners(common.DefaultTimeout, podNamespace, podName)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		}
	})
}
