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

package accesscontrol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	configpkg "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	containerpkg "github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

var _ = ginkgo.Describe(common.AccessControlTestKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.AccessControlTestKey) {
		config := common.GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		containersUnderTest := common.CreateContainersUnderTest(config)
		log.Info(containersUnderTest)

		for _, containerUnderTest := range containersUnderTest {
			testNamespace(containerUnderTest.Oc)
		}

		for _, containerUnderTest := range containersUnderTest {
			testRoles(containerUnderTest.Oc.GetPodName(), containerUnderTest.Oc.GetPodNamespace())
		}

		// Former "container" tests
		defer ginkgo.GinkgoRecover()

		// Run the tests that interact with the containers
		ginkgo.When("under test", func() {
			conf := configpkg.GetConfigInstance()
			cnfsInTest := conf.CNFs
			gomega.Expect(cnfsInTest).ToNot(gomega.BeNil())
			for _, cnf := range cnfsInTest {
				cnf := cnf
				var containerFact = testcases.ContainerFact{Namespace: cnf.Namespace, Name: cnf.Name, ContainerCount: 0, HasClusterRole: false, Exists: true}
				// Gather facts for containers
				podFacts, err := testcases.LoadCnfTestCaseSpecs(testcases.GatherFacts)
				gomega.Expect(err).To(gomega.BeNil())
				context := common.GetContext()
				// Collect container facts
				for _, factsTest := range podFacts.TestCase {
					args := strings.Split(fmt.Sprintf(factsTest.Command, cnf.Name, cnf.Namespace), " ")
					cnfInTest := containerpkg.NewPod(args, cnf.Name, cnf.Namespace, factsTest.ExpectedStatus, factsTest.ResultType, factsTest.Action, common.DefaultTimeout)
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
					testFile, err := testcases.LoadConfiguredTestFile(common.ConfiguredTestFile)
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
								runTestsOnCNF(containerFact.ContainerCount, testCase, testType, containerFact, context)
							} else {
								runTestsOnCNF(testCase.Loop, testCase, testType, containerFact, context)
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
	testType string, facts testcases.ContainerFact, context *interactive.Context) {
	ginkgo.It(fmt.Sprintf("is running test for : %s/%s for test command :  %s", facts.Namespace, facts.Name, testCmd.Name), func() {
		defer results.RecordResult(identifiers.TestHostResourceIdentifier)
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
				cnfInTest := containerpkg.NewPod(cmdArgs, facts.Name, facts.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
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
			cnfInTest := containerpkg.NewPod(cmdArgs, facts.Name, facts.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
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

func testRoles(podName, podNamespace string) {
	var serviceAccountName string
	ginkgo.When(fmt.Sprintf("Testing roles and privileges of %s/%s", podNamespace, podName), func() {
		testServiceAccount(podName, podNamespace, &serviceAccountName)
		testRoleBindings(podNamespace, &serviceAccountName)
		testClusterRoleBindings(podNamespace, &serviceAccountName)
	})
}

func testServiceAccount(podName, podNamespace string, serviceAccountName *string) {
	ginkgo.It("Should have a valid ServiceAccount name", func() {
		defer results.RecordResult(identifiers.TestPodServiceAccountBestPracticesIdentifier)
		context := common.GetContext()
		tester := serviceaccount.NewServiceAccount(common.DefaultTimeout, podName, podNamespace)
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
		context := common.GetContext()
		rbTester := rolebinding.NewRoleBinding(common.DefaultTimeout, *serviceAccountName, podNamespace)
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
		context := common.GetContext()
		crbTester := clusterrolebinding.NewClusterRoleBinding(common.DefaultTimeout, *serviceAccountName, podNamespace)
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
