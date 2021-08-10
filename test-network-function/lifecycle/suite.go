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
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
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
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.LifecycleTestKey) {

		conf := configpkg.GetConfigInstance()

		log.Info(conf.CNFs )

		for _, podUnderTest := range conf.CNFs {
			testNodeSelector(common.GetContext(), podUnderTest.Name, podUnderTest.Namespace)
		}

		for _, podUnderTest := range conf.CNFs  {
			testGracePeriod(common.GetContext(), podUnderTest.Name, podUnderTest.Namespace)
		}

		for _, podUnderTest := range conf.CNFs  {
			testShutdown(podUnderTest.Namespace, podUnderTest.Name)
		}

		for _, podUnderTest := range conf.CNFs  {
			testDeployments(podUnderTest.Namespace)
		}

		for _, podUnderTest := range conf.CNFs  {
			testOwner(podUnderTest.Namespace, podUnderTest.Name)
		}

		for _, podUnderTest := range conf.CNFs  {
			testPodAntiAffinity(podUnderTest.Namespace)

		}
	}
})

func testGracePeriod(context *interactive.Context, podName, podNamespace string) {
	ginkgo.It(fmt.Sprintf("Testing pod terminationGracePeriod %s/%s", podNamespace, podName), func() {
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
			msg := fmt.Sprintf("%s %s has terminationGracePeriod set to 30, you might want to change it", podName, podNamespace)
			log.Warn(msg)
			infoWriter(msg)
		}
	})
}

func testNodeSelector(context *interactive.Context, podName, podNamespace string) {
	ginkgo.It(fmt.Sprintf("Testing pod nodeSelector %s/%s", podNamespace, podName), func() {
		defer results.RecordResult(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
		infoWriter := tnf.CreateTestExtraInfoWriter()
		tester := nodeselector.NewNodeSelector(common.DefaultTimeout, podName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()

		gomega.Expect(err).To(gomega.BeNil())
		if testResult != tnf.SUCCESS {
			msg := fmt.Sprintf("The pod specifies nodeSelector/nodeAffinity field, you might want to change it, %s %s", podName, podNamespace)
			log.Warn(msg)
			infoWriter(msg)
		}
	})
}

// testDeployments ensures that each Deployment has the correct number of "Ready" replicas.
func testDeployments(namespace string) {
	var deployments dp.DeploymentMap
	var notReadyDeployments []string
	ginkgo.When("Testing deployments in namespace", func() {
		ginkgo.It("Should return list of deployments", func() {
			deployments, notReadyDeployments = getDeployments(namespace)
			if len(deployments) == 0 {
				return
			}
			gomega.Expect(notReadyDeployments).To(gomega.BeEmpty())
		})
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

//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
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

//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
func drainNode(node string) {
	context := common.GetContext()
	tester := dd.NewDeploymentsDrain(drainTimeout, node)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
}

func testShutdown(podNamespace, podName string) {
	ginkgo.When("Testing PUT is configured with pre-stop lifecycle", func() {
		ginkgo.It("should have pre-stop configured", func() {
			defer results.RecordResult(identifiers.TestShudtownIdentifier)
			shutdownTest(podNamespace, podName)
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

// uncordonNode uncordons a Node.
//nolint:deadcode // Taken out of v2.0.0 for CTONET-1022.
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
func testPodAntiAffinity(podNamespace string) {
	var deployments dp.DeploymentMap
	ginkgo.When("CNF is designed in high availability mode ", func() {
		ginkgo.It("Should set pod replica number greater than 1 and corresponding pod anti-affinity rules in deployment", func() {
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

func testOwner(podNamespace, podName string) {
	ginkgo.When("Testing owners of CNF pod", func() {
		ginkgo.It("Should be only ReplicaSet", func() {
			defer results.RecordResult(identifiers.TestPodDeploymentBestPracticesIdentifier)
			context := common.GetContext()
			tester := owners.NewOwners(common.DefaultTimeout, podNamespace, podName)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}
