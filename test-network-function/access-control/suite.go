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

package container

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
	log "github.com/sirupsen/logrus"
	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/internal/api"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	utilspods "github.com/test-network-function/test-network-function/utilspods"
)

const (
	// The default test timeout.
	defaultTimeoutSeconds = 10
	// timeout for eventually call
	eventuallyTimeoutSeconds = 30
	// interval of time
	interval           = 1
	configuredTestFile = "testconfigure.yml"
	testSpecName       = "container"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	cnfsInTest     []configsections.Cnf
	certAPIClient  api.CertAPIClient
)

var _ = ginkgo.Describe(testSpecName, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testSpecName) {

		config := GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		containersUnderTest := createContainersUnderTest(config)
		partnerContainers := createPartnerContainers(config)
		testOrchestrator := partnerContainers[config.TestOrchestrator]
		fsDiffContainer := partnerContainers[config.FsDiffMasterContainer]
		log.Info(testOrchestrator)
		log.Info(containersUnderTest)

		for _, containerUnderTest := range containersUnderTest {
			testNodeSelector(getContext(), containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodNamespace())
		}

		ginkgo.Context("Container does not have additional packages installed", func() {
			// use this boolean to turn off tests that require OS packages
			if !isMinikube() {
				if fsDiffContainer != nil {
					for _, containerUnderTest := range containersUnderTest {
						testFsDiff(fsDiffContainer.oc, containerUnderTest.oc)
					}
				} else {
					log.Warn("no fs diff container is configured, cannot run fs diff test")
				}
			}
		})

		ginkgo.Context("Both Pods are on the Default network", func() {
			// for each container under test, ensure bidirectional ICMP traffic between the container and the orchestrator.
			for _, containerUnderTest := range containersUnderTest {
				if _, ok := containersToExcludeFromConnectivityTests[containerUnderTest.containerIdentifier]; !ok {
					testNetworkConnectivity(containerUnderTest.oc, testOrchestrator.oc, testOrchestrator.defaultNetworkIPAddress, defaultNumPings)
					testNetworkConnectivity(testOrchestrator.oc, containerUnderTest.oc, containerUnderTest.defaultNetworkIPAddress, defaultNumPings)
				}
			}
		})

		
		for _, containerUnderTest := range containersUnderTest {
			testNamespace(containerUnderTest.oc)
		}

		for _, containerUnderTest := range containersUnderTest {
			testRoles(containerUnderTest.oc.GetPodName(), containerUnderTest.oc.GetPodNamespace())
		}


		defer ginkgo.GinkgoRecover()
		ginkgo.When("a local shell is spawned", func() {
			goExpectSpawner := interactive.NewGoExpectSpawner()
			var spawner interactive.Spawner = goExpectSpawner
			context, err = interactive.SpawnShell(&spawner, defaultTimeout, interactive.Verbose(true))
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(context).ToNot(gomega.BeNil())
			gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
		})
		// Query API for certification status of listed containers
		ginkgo.When("getting certification status", func() {
			conf := config.GetConfigInstance()
			cnfsToQuery := conf.CertifiedContainerInfo
			if len(cnfsToQuery) > 0 {
				certAPIClient = api.NewHTTPClient()
				for _, cnfRequestInfo := range cnfsToQuery {
					cnf := cnfRequestInfo
					// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
					ginkgo.It(fmt.Sprintf("container %s/%s should eventually be verified as certified", cnf.Repository, cnf.Name), func() {
						defer results.RecordResult(identifiers.TestContainerIsCertifiedIdentifier)
						cnf := cnf // pin
						gomega.Eventually(func() bool {
							isCertified := certAPIClient.IsContainerCertified(cnf.Repository, cnf.Name)
							return isCertified
						}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
					})
				}
			}
		})
		// Run the tests that interact with the containers
		ginkgo.When("under test", func() {
			conf := config.GetConfigInstance()
			cnfsInTest = conf.CNFs
			gomega.Expect(cnfsInTest).ToNot(gomega.BeNil())
			for _, cnf := range cnfsInTest {
				cnf := cnf
				var containerFact = testcases.ContainerFact{Namespace: cnf.Namespace, Name: cnf.Name, ContainerCount: 0, HasClusterRole: false, Exists: true}
				// Gather facts for containers
				podFacts, err := testcases.LoadCnfTestCaseSpecs(testcases.GatherFacts)
				gomega.Expect(err).To(gomega.BeNil())
				// Collect container facts
				for _, factsTest := range podFacts.TestCase {
					args := strings.Split(fmt.Sprintf(factsTest.Command, cnf.Name, cnf.Namespace), " ")
					cnfInTest := container.NewPod(args, cnf.Name, cnf.Namespace, factsTest.ExpectedStatus, factsTest.ResultType, factsTest.Action, defaultTimeout)
					test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(test).ToNot(gomega.BeNil())
					_, err = test.Run()
					gomega.Expect(err).To(gomega.BeNil())
					if factsTest.Name == string(testcases.ContainerCount) {
						containerFact.ContainerCount, _ = strconv.Atoi(cnfInTest.Facts())
					} else if factsTest.Name == string(testcases.ServiceAccountName) {
						containerFact.ServiceAccount = cnfInTest.Facts()
					} else if factsTest.Name == string(testcases.Name) {
						containerFact.Name = cnfInTest.Facts()
						gomega.Expect(containerFact.Name).To(gomega.Equal(cnf.Name))
						if strings.Compare(containerFact.Name, cnf.Name) > 0 {
							containerFact.Exists = true
						}
					}
				}
				// loop through various cnfs test
				if !containerFact.Exists {
					ginkgo.It(fmt.Sprintf("is running test pod exists : %s/%s for test command :  %s", containerFact.Namespace, containerFact.Name, "POD EXISTS"), func() {
						gomega.Expect(containerFact.Exists).To(gomega.BeTrue())
					})
					continue
				}
				for _, testType := range cnf.Tests {
					testFile, err := testcases.LoadConfiguredTestFile(configuredTestFile)
					gomega.Expect(testFile).ToNot(gomega.BeNil())
					gomega.Expect(err).To(gomega.BeNil())
					testConfigure := testcases.ContainsConfiguredTest(testFile.CnfTest, testType)
					renderedTestCase, err := testConfigure.RenderTestCaseSpec(testcases.Cnf, testType)
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(renderedTestCase).ToNot(gomega.BeNil())
					for _, testCase := range renderedTestCase.TestCase {
						if !testCase.SkipTest {
							if testCase.ExpectedType == testcases.Function {
								for _, val := range testCase.ExpectedStatus {
									testCase.ExpectedStatusFn(cnf.Name, testcases.StatusFunctionType(val))
								}
							}
							if testCase.Loop > 0 {
								runTestsOnCNF(containerFact.ContainerCount, testCase, testType, containerFact)
							} else {
								runTestsOnCNF(testCase.Loop, testCase, testType, containerFact)
							}
						}
					}
				}
			}
		})
	}
})

//nolint:gocritic // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestsOnCNF(containerCount int, testCmd testcases.BaseTestCase,
	testType string, facts testcases.ContainerFact) {
	ginkgo.It(fmt.Sprintf("is running test for : %s/%s for test command :  %s", facts.Namespace, facts.Name, testCmd.Name), func() {
		defer results.RecordResult(identifiers.TestContainerBestPracticesIdentifier)
		containerCount := containerCount
		testType := testType
		facts := facts
		testCmd := testCmd
		var args []interface{}
		if testType == testcases.PrivilegedRoles {
			args = []interface{}{facts.Namespace, facts.Namespace, facts.ServiceAccount}
		} else {
			args = []interface{}{facts.Name, facts.Namespace}
		}
		if containerCount > 0 {
			count := 0
			for count < containerCount {
				argsCount := append(args, count)
				cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, argsCount...), " ")
				cnfInTest := container.NewPod(cmdArgs, facts.Name, facts.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
				gomega.Expect(cnfInTest).ToNot(gomega.BeNil())
				test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
				count++
			}
		} else {
			cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, args...), " ")
			cnfInTest := container.NewPod(cmdArgs, facts.Name, facts.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
			gomega.Expect(cnfInTest).ToNot(gomega.BeNil())
			test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		}
	})
}


func testNamespace(oc *interactive.Oc) {
	pod := oc.GetPodName()
	container := oc.GetPodContainerName()
	ginkgo.When(fmt.Sprintf("Reading namespace of %s/%s", pod, container), func() {
		ginkgo.It("Should not be 'default' and should not begin with 'openshift-'", func() {
			defer results.RecordResult(identifiers.TestNamespaceBestPracticesIdentifier)
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.Equal("default")))
			gomega.Expect(oc.GetPodNamespace()).To(gomega.Not(gomega.HavePrefix("openshift-")))
		})
	})
}
func testServiceAccount(podName, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should have a valid ServiceAccount name", func() {
		defer results.RecordResult(identifiers.TestPodServiceAccountBestPracticesIdentifier)
		context := getContext()
		tester := serviceaccount.NewServiceAccount(defaultTimeout, podName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
		*serviceAccountName = tester.GetServiceAccountName()
		gomega.Expect(*serviceAccountName).ToNot(gomega.BeEmpty())
	})
}

func testRoleBindings(podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have RoleBinding in other namespaces", func() {
		defer results.RecordResult(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		context := getContext()
		rbTester := rolebinding.NewRoleBinding(defaultTimeout, *serviceAccountName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), rbTester, []reel.Handler{rbTester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		if rbTester.Result() == tnf.FAILURE {
			log.Info("RoleBindings: ", rbTester.GetRoleBindings())
		}
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		gomega.Expect(err).To(gomega.BeNil())
	})
}

func testClusterRoleBindings(podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should not have ClusterRoleBindings", func() {
		defer results.RecordResult(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
		if *serviceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}
		context := getContext()
		crbTester := clusterrolebinding.NewClusterRoleBinding(defaultTimeout, *serviceAccountName, podNamespace)
		test, err := tnf.NewTest(context.GetExpecter(), crbTester, []reel.Handler{crbTester}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		testResult, err := test.Run()
		if crbTester.Result() == tnf.FAILURE {
			log.Info("ClusterRoleBindings: ", crbTester.GetClusterRoleBindings())
		}
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	})
}
func testRoles(podName, podNamespace string) {
	var serviceAccountName string
	ginkgo.When(fmt.Sprintf("Testing roles and privileges of %s/%s", podNamespace, podName), func() {
		testServiceAccount(podName, podNamespace, &serviceAccountName)
		testRoleBindings(podNamespace, &serviceAccountName)
		testClusterRoleBindings(podNamespace, &serviceAccountName)
	})
}

// Runs the "generic" CNF test cases.
var _ = ginkgo.Describe(testsKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testsKey) {
		

	}
})