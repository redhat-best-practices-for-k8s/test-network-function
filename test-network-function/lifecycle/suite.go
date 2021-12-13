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
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/scaling"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	dd "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsdrain"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/graceperiod"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeselector"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/owners"
	ps "github.com/test-network-function/test-network-function/pkg/tnf/handlers/podsets"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/test-network-function/results"
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

	// relativeimagepullpolicyTestPath is the relative path to the imagepullpolicy.json test case.
	imagepullpolicyTestPath         = path.Join("pkg", "tnf", "handlers", "imagepullpolicy", "imagepullpolicy.json")
	relativeimagepullpolicyTestPath = path.Join(common.PathRelativeToRoot, imagepullpolicyTestPath)
)

var drainTimeout = time.Duration(drainTimeoutMinutes) * time.Minute

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.LifecycleTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.LifecycleTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))

		})

		ginkgo.ReportAfterEach(results.RecordResult)

		testImagePolicy(env)

		testNodeSelector(env)

		testGracePeriod(env)

		testShutdown(env)

		testPodAntiAffinity(env)

		if common.Intrusive() {
			testPodsRecreation(env)

			testScaling(env)
			testStateFullSetScaling(env)
		}

		testOwner(env)
	}
})

func waitForAllDeploymentsReady(namespace string, timeout, pollingPeriod time.Duration, resourceType configsections.PodSetType) int { //nolint:unparam // it is fine to use always the same value for timeout
	var elapsed time.Duration
	var notReadyDeployments []string

	for elapsed < timeout {
		_, notReadyDeployments = GetPodSets(namespace, resourceType)
		log.Debugf("Waiting for deployments to get ready, remaining: %d deployments", len(notReadyDeployments))
		if len(notReadyDeployments) == 0 {
			break
		}
		time.Sleep(pollingPeriod)
		elapsed += pollingPeriod
	}
	return len(notReadyDeployments)
}

// restoreDeployments is the last attempt to restore the original test deployments' replicaCount
func restoreDeployments(env *config.TestEnvironment) {
	for i := range env.DeploymentsUnderTest {
		// For each test deployment in the namespace, refresh the current replicas and compare.
		refreshReplicas(&env.DeploymentsUnderTest[i], env)
	}
}

// restoreDeployments is the last attempt to restore the original test deployments' replicaCount
func restoreStateFullSet(env *config.TestEnvironment) {
	for i := range env.StateFullSetUnderTest {
		// For each test deployment in the namespace, refresh the current replicas and compare.
		refreshReplicas(&env.StateFullSetUnderTest[i], env)
	}
}

func refreshReplicas(podset *configsections.PodSet, env *config.TestEnvironment) {
	podsets, notReadypodsets := GetPodSets(podset.Namespace, podset.Type)

	if len(notReadypodsets) > 0 {
		// Wait until the deployment is ready
		notReady := waitForAllDeploymentsReady(podset.Namespace, scalingTimeout, scalingPollingPeriod, podset.Type)
		if notReady != 0 {
			collectNodeAndPendingPodInfo(podset.Namespace)
			log.Fatalf("Could not restore deployment replicaCount for namespace %s.", podset.Namespace)
		}
	}
	if podset.Hpa.HpaName != "" { // it have hpa and need to update the max min
		runHpaScalingTest(podset, podset.Hpa.MinReplicas, podset.Hpa.MaxReplicas)
	}
	key := podset.Namespace + ":" + podset.Name
	dep, ok := podsets[key]
	if ok {
		if dep.Replicas != podset.Replicas {
			log.Warn("Deployment ", podset.Name, " replicaCount (", podset.Replicas, ") needs to be restored.")

			// Try to scale to the original deployment's replicaCount.
			runScalingTest(podset)

			env.SetNeedsRefresh()
		}
	}
}

func closeOcSessionsByDeployment(containers map[configsections.ContainerIdentifier]*config.Container, deployment *configsections.PodSet) {
	log.Debug("close session for deployment=", deployment.Name, " start")
	defer log.Debug("close session for deployment=", deployment.Name, " done")
	for cid, c := range containers {
		if cid.Namespace == deployment.Namespace && strings.HasPrefix(cid.PodName, deployment.Name+"-") {
			log.Infof("Closing session to %s %s", cid.PodName, cid.ContainerName)
			c.CloseOc()
			delete(containers, cid)
		}
	}
}

// runScalingTest Runs a Scaling handler TC and waits for all the deployments to be ready.
func runScalingTest(podset *configsections.PodSet) {
	handler := scaling.NewScaling(common.DefaultTimeout, podset.Namespace, podset.Name, podset.Replicas)
	test, err := tnf.NewTest(common.GetContext().GetExpecter(), handler, []reel.Handler{handler}, common.GetContext().GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()

	// Wait until the deployment is ready
	notReady := waitForAllDeploymentsReady(podset.Namespace, scalingTimeout, scalingPollingPeriod, podset.Type)
	if notReady != 0 {
		collectNodeAndPendingPodInfo(podset.Namespace)
		ginkgo.Fail(fmt.Sprintf("Failed to scale deployment for namespace %s.", podset.Namespace))
	}
}

func runHpaScalingTest(podset *configsections.PodSet, min, max int) {
	handler := scaling.NewHpaScaling(common.DefaultTimeout, podset.Namespace, podset.Hpa.HpaName, min, max)
	test, err := tnf.NewTest(common.GetContext().GetExpecter(), handler, []reel.Handler{handler}, common.GetContext().GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()

	// Wait until the deployment is ready
	notReady := waitForAllDeploymentsReady(podset.Namespace, scalingTimeout, scalingPollingPeriod, podset.Type)
	if notReady != 0 {
		collectNodeAndPendingPodInfo(podset.Namespace)
		ginkgo.Fail(fmt.Sprintf("Failed to auto-scale deployment for namespace %s.", podset.Namespace))
	}
}

func testScaling(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestScalingIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing deployment scaling")
		defer restoreDeployments(env)
		defer env.SetNeedsRefresh()

		if len(env.DeploymentsUnderTest) == 0 {
			ginkgo.Skip("No test deployments found.")
		}
		for i := range env.DeploymentsUnderTest {
			runScalingfunc(&env.DeploymentsUnderTest[i], env)
		}
	})
}
func testStateFullSetScaling(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestScalingIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing StatefulSet scaling")
		defer restoreStateFullSet(env)
		defer env.SetNeedsRefresh()

		if len(env.StateFullSetUnderTest) == 0 {
			ginkgo.Skip("No test StatefulSet found.")
		}
		for i := range env.StateFullSetUnderTest {
			runScalingfunc(&env.StateFullSetUnderTest[i], env)
		}
	})
}

func runScalingfunc(podset *configsections.PodSet, env *config.TestEnvironment) {
	ginkgo.By(fmt.Sprintf("Scaling Deployment=%s, Replicas=%d (ns=%s)", podset.Name, podset.Replicas, podset.Namespace))

	closeOcSessionsByDeployment(env.ContainersUnderTest, podset)
	replicaCount := podset.Replicas
	podsetscale := podset
	if podsetscale.Hpa.HpaName != "" {
		min := replicaCount - 1
		max := replicaCount - 1
		runHpaScalingTest(podsetscale, min, max) // scale in
		min = replicaCount
		max = replicaCount
		runHpaScalingTest(podsetscale, min, max) // scale out
	} else {
		// ScaleIn, removing one pod from the replicaCount
		podset.Replicas = replicaCount - 1
		runScalingTest(podset)

		// Scaleout, restoring the original replicaCount number
		podset.Replicas = replicaCount
		runScalingTest(podset)
	}
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
			tester := nodeselector.NewNodeSelector(common.DefaultTimeout, podName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidateWithFailureCallback(func() {
				msg := fmt.Sprintf("The pod specifies nodeSelector/nodeAffinity field, you might want to change it, %s %s", podNamespace, podName)
				log.Warn(msg)
				_, err := ginkgo.GinkgoWriter.Write([]byte(msg))
				if err != nil {
					log.Errorf("Ginkgo writer could not write because: %s", err)
				}
			})
		}
	})
}

func testGracePeriod(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestNonDefaultGracePeriodIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Test terminationGracePeriod")
		context := common.GetContext()
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			ginkgo.By(fmt.Sprintf("Testing pod terminationGracePeriod %s %s", podNamespace, podName))
			tester := graceperiod.NewGracePeriod(common.DefaultTimeout, podName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			gracePeriod := tester.GetGracePeriod()
			if gracePeriod == defaultTerminationGracePeriod {
				msg := fmt.Sprintf("%s %s has terminationGracePeriod set to %d, you might want to change it", podNamespace, podName, defaultTerminationGracePeriod)
				log.Warn(msg)
				_, err := ginkgo.GinkgoWriter.Write([]byte(msg))
				if err != nil {
					log.Errorf("Ginkgo writer could not write because: %s", err)
				}
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
	tester, handlers := utils.NewGenericTesterAndValidate(relativeShutdownTestPath, common.RelativeSchemaPath, values)
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())

	test.RunAndValidate()
}

func cleanupNodeDrain(env *config.TestEnvironment, nodeName string) {
	uncordonNode(nodeName)
	for _, ns := range env.NameSpacesUnderTest {
		notReady := waitForAllDeploymentsReady(ns, scalingTimeout, scalingPollingPeriod, "deployment")
		if notReady != 0 {
			collectNodeAndPendingPodInfo(ns)
			log.Fatalf("Cleanup after node drain for %s failed, stopping tests to ensure cluster integrity", nodeName)
		}
	}
}

func testNodeDrain(env *config.TestEnvironment, nodeName string) {
	ginkgo.By(fmt.Sprintf("Testing node drain for %s\n", nodeName))
	// Ensure the node is uncordoned before exiting the function,
	// and all deployments are ready
	defer cleanupNodeDrain(env, nodeName)
	// drain node
	drainNode(nodeName)
	for _, ns := range env.NameSpacesUnderTest {
		notReady := waitForAllDeploymentsReady(ns, scalingTimeout, scalingPollingPeriod, "deployment")
		if notReady != 0 {
			collectNodeAndPendingPodInfo(ns)
			ginkgo.Fail(fmt.Sprintf("Failed to recover deployments on namespace %s after draining node %s.", ns, nodeName))
		}
	}
	// If we got this far, all deployments are ready after draining the node
	tnf.ClaimFilePrintf("Node drain for %s succeeded", nodeName)
}

func testPodsRecreation(env *config.TestEnvironment) {
	deployments := make(ps.PodSetMap)
	var notReadyDeployments []string

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodRecreationIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing node draining effect of deployment")
		ginkgo.By(fmt.Sprintf("test deployment in namespace %s", env.NameSpacesUnderTest))
		for _, ns := range env.NameSpacesUnderTest {
			var dps ps.PodSetMap
			dps, notReadyDeployments = GetPodSets(ns, "deployment")
			for dpKey, dp := range dps {
				deployments[dpKey] = dp
			}
			// We require that all deployments have the desired number of replicas and are all up to date
			if len(notReadyDeployments) != 0 {
				ginkgo.Skip("Can not test when deployments are not ready")
			}
		}
		if len(deployments) == 0 {
			ginkgo.Skip("no valid deployment")
		}
		defer env.SetNeedsRefresh()
		ginkgo.By("should create new replicas when node is drained")
		// We need to delete all Oc sessions because the drain operation is often deleting oauth-openshift pod
		// This results in lost connectivity for oc sessions
		env.ResetOc()
		for _, n := range env.NodesUnderTest {
			if !n.HasDeployment() {
				log.Debug("node ", n.Name, " has no deployment, skip draining")
				continue
			}
			testNodeDrain(env, n.Name)
		}
	})
}

// getDeployments returns map of deployments and names of not-ready deployments
func GetPodSets(namespace string, resourceType configsections.PodSetType) (podsets ps.PodSetMap, notReadypodsets []string) {
	context := common.GetContext()
	tester := ps.NewPodSets(common.DefaultTimeout, namespace, string(resourceType))
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()

	podsets = tester.GetPodSets()
	for name, d := range podsets {
		if d.Unavailable != 0 || d.Ready != d.Replicas || resourceType == configsections.Deployment && d.Available != d.Replicas || resourceType == configsections.StateFullSet && d.Current != d.Replicas || d.UpToDate != d.Replicas {
			notReadypodsets = append(notReadypodsets, name)
			log.Tracef("deployment %s: not ready", name)
		} else {
			log.Tracef("deployment %s: ready", name)
		}
	}

	return podsets, notReadypodsets
}

func collectNodeAndPendingPodInfo(ns string) {
	context := common.GetContext()

	nodeStatus, _ := utils.ExecuteCommand("oc get nodes -o json | jq '.items[]|{name:.metadata.name, taints:.spec.taints}'", common.DefaultTimeout, context)
	common.TcClaimLogPrintf("Namespace: %s\nNode status:\n%s", ns, nodeStatus)

	cmd := fmt.Sprintf("oc get pods -n %s --field-selector=status.phase!=Running,status.phase!=Succeeded -o json | jq '.items[]|{name:.metadata.name, status:.status}'", ns)
	podStatus, _ := utils.ExecuteCommand(cmd, common.DefaultTimeout, context)
	common.TcClaimLogPrintf("Pending Pods:\n%s", podStatus)

	cmd = fmt.Sprintf("oc get events -n %s --field-selector type!=Normal -o json --sort-by='.lastTimestamp' | jq '.items[]|{object:.involvedObject, reason:.reason, type:.type, message:.message, lastSeen:.lastTimestamp}'", ns)
	events, _ := utils.ExecuteCommand(cmd, common.DefaultTimeout, context)
	common.TcClaimLogPrintf("Events:\n%s", events)
}

func drainNode(node string) {
	context := common.GetContext()
	tester := dd.NewDeploymentsDrain(drainTimeout, node)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	result, err := test.Run()
	if err != nil || result == tnf.ERROR {
		log.Fatalf("Test skipped because of draining node failure - platform issue")
	}
}

func uncordonNode(node string) {
	context := common.GetContext()
	values := make(map[string]interface{})
	values["NODE"] = node
	tester, handlers := utils.NewGenericTesterAndValidate(relativeNodesTestPath, common.RelativeSchemaPath, values)
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())

	test.RunAndValidate()
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
	tester, handlers := utils.NewGenericTesterAndValidate(relativePodTestPath, common.RelativeSchemaPath, values)
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())

	test.RunAndValidateWithFailureCallback(func() {
		if replica > 1 {
			msg := fmt.Sprintf("The deployment replica count is %d, but a podAntiAffinity rule is not defined, "+
				"you might want to change it in deployment %s in namespace %s", replica, deployment, podNamespace)
			log.Warn(msg)
			_, err := ginkgo.GinkgoWriter.Write([]byte(msg))
			if err != nil {
				log.Errorf("Ginkgo writer could not write because: %s", err)
			}
		} else {
			msg := fmt.Sprintf("The deployment replica count is %d. Pod replica should be > 1 with an "+
				"podAntiAffinity rule defined . You might want to change it in deployment %s in namespace %s",
				replica, deployment, podNamespace)
			log.Warn(msg)
			_, err := ginkgo.GinkgoWriter.Write([]byte(msg))
			if err != nil {
				log.Errorf("Ginkgo writer could not write because: %s", err)
			}
		}
	})
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
			tester := owners.NewOwners(common.DefaultTimeout, podNamespace, podName)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
		}
	})
}

func testImagePolicy(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestImagePullPolicyIdentifier)
	ginkgo.It(testID, func() {
		context := common.GetContext()
		for _, podUnderTest := range env.PodsUnderTest {
			values := make(map[string]interface{})
			ContainerCount := podUnderTest.ContainerCount
			values["POD_NAMESPACE"] = podUnderTest.Namespace
			values["POD_NAME"] = podUnderTest.Name
			for i := 0; i < ContainerCount; i++ {
				values["CONTAINER_NUM"] = i
				tester, handlers := utils.NewGenericTesterAndValidate(relativeimagepullpolicyTestPath, common.RelativeSchemaPath, values)
				test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())

				test.RunAndValidate()
			}
		}
	})
}
