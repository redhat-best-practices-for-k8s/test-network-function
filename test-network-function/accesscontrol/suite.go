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
	"strings"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
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
		configData := common.ConfigurationData{}
		configData.SetNeedsRefresh()
		ginkgo.BeforeEach(func() {
			common.ReloadConfiguration(&configData)
		})
		common.ReloadConfiguration(&configData)
		testNamespace(&configData)

		testRoles(&configData)

		defer ginkgo.GinkgoRecover()

		// Run the tests that interact with the pods
		ginkgo.When("under test", func() {
			podUnderTest := configData.PodsUnderTest
			gomega.Expect(podUnderTest).ToNot(gomega.BeNil())
			context := common.GetContext()
			var allTests []string = testcases.GetConfiguredPodTests()
			for _, testType := range allTests {
				testFile, err := testcases.LoadConfiguredTestFile(common.ConfiguredTestFile)
				gomega.Expect(testFile).ToNot(gomega.BeNil())
				gomega.Expect(err).To(gomega.BeNil())
				testConfigure := testcases.ContainsConfiguredTest(testFile.CnfTest, testType)
				renderedTestCase, err := testConfigure.RenderTestCaseSpec(testcases.Cnf, testType)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(renderedTestCase).ToNot(gomega.BeNil())
				for _, testCase := range renderedTestCase.TestCase {
					if !testCase.SkipTest {
						runTestOnPods(&configData, testCase, testType, context)
					}
				}
			}
		})
	}
})

//nolint:gocritic,funlen // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestOnPods(configData *common.ConfigurationData, testCmd testcases.BaseTestCase, testType string, context *interactive.Context) {
	ginkgo.It(fmt.Sprintf("Running Pod test : %s", testCmd.Name), func() {
		for _, podUnderTest := range configData.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			ginkgo.By(fmt.Sprintf("Reading namespace of podnamespace= %s podname= %s", podNamespace, podName))
			log.Info(fmt.Sprintf("Reading namespace of podnamespace= %s podname= %s", podNamespace, podName))
			if testCmd.ExpectedType == testcases.Function {
				for _, val := range testCmd.ExpectedStatus {
					testCmd.ExpectedStatusFn(podName, testcases.StatusFunctionType(val))
				}
			}
			defer results.RecordResult(identifiers.TestHostResourceIdentifier)
			testType := testType
			testCmd := testCmd
			var args []interface{}
			if testType == testcases.PrivilegedRoles {
				args = []interface{}{podUnderTest.Namespace, podUnderTest.Namespace, podUnderTest.ServiceAccount}
			} else {
				args = []interface{}{podUnderTest.Name, podUnderTest.Namespace}
			}
			var count int
			if testCmd.Loop > 0 {
				count = podUnderTest.ContainerCount
			} else {
				count = testCmd.Loop
			}

			if count > 0 {
				count := 0
				for count < podUnderTest.ContainerCount {
					argsCount := append(args, count)
					cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, argsCount...), " ")
					cnfInTest := containerpkg.NewPod(cmdArgs, podUnderTest.Name, podUnderTest.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
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
				podTest := containerpkg.NewPod(cmdArgs, podUnderTest.Name, podUnderTest.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
				gomega.Expect(podTest).ToNot(gomega.BeNil())
				test, err := tnf.NewTest(context.GetExpecter(), podTest, []reel.Handler{podTest}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			}
		}
	})
}

func testNamespace(configData *common.ConfigurationData) {
	ginkgo.When("test deployment namespace", func() {
		ginkgo.It("Should not be 'default' and should not begin with 'openshift-'", func() {
			for _, podUnderTest := range configData.PodsUnderTest {
				podName := podUnderTest.Name
				podNamespace := podUnderTest.Namespace
				ginkgo.By(fmt.Sprintf("Reading namespace of podnamespace= %s podname= %s", podNamespace, podName))
				log.Info(fmt.Sprintf("Reading namespace of podnamespace= %s podname= %s", podNamespace, podName))
				defer results.RecordResult(identifiers.TestNamespaceBestPracticesIdentifier)
				gomega.Expect(podNamespace).To(gomega.Not(gomega.Equal("default")))
				gomega.Expect(podNamespace).To(gomega.Not(gomega.HavePrefix("openshift-")))
			}
		})
	})
}

func testRoles(configData *common.ConfigurationData) {
	testServiceAccount(configData)
	testRoleBindings(configData)
	testClusterRoleBindings(configData)
}

func testServiceAccount(configData *common.ConfigurationData) {
	ginkgo.It("Should have a valid ServiceAccount name", func() {
		for _, podUnderTest := range configData.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			context := common.GetContext()
			ginkgo.By(fmt.Sprintf("Testing pod service account %s %s", podNamespace, podName))
			defer results.RecordResult(identifiers.TestPodServiceAccountBestPracticesIdentifier)
			tester := serviceaccount.NewServiceAccount(common.DefaultTimeout, podName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
			serviceAccountName := tester.GetServiceAccountName()
			gomega.Expect(serviceAccountName).ToNot(gomega.BeEmpty())
		}
	})
}

func testRoleBindings(configData *common.ConfigurationData) {
	ginkgo.It("Should not have RoleBinding in other namespaces", func() {
		for _, podUnderTest := range configData.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			context := common.GetContext()
			defer results.RecordResult(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
			ginkgo.By(fmt.Sprintf("Testing role  bidning  %s %s", podNamespace, podName))
			if serviceAccountName == "" {
				ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
			}
			rbTester := rolebinding.NewRoleBinding(common.DefaultTimeout, serviceAccountName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), rbTester, []reel.Handler{rbTester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			if rbTester.Result() == tnf.FAILURE {
				log.Info("RoleBindings: ", rbTester.GetRoleBindings())
			}
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		}
	})
}

func testClusterRoleBindings(configData *common.ConfigurationData) {
	ginkgo.It("Should not have ClusterRoleBindings", func() {
		for _, podUnderTest := range configData.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			context := common.GetContext()
			defer results.RecordResult(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
			ginkgo.By(fmt.Sprintf("Testing cluster role  bidning  %s %s", podNamespace, podName))
			if serviceAccountName == "" {
				ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
			}
			crbTester := clusterrolebinding.NewClusterRoleBinding(common.DefaultTimeout, serviceAccountName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), crbTester, []reel.Handler{crbTester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			if crbTester.Result() == tnf.FAILURE {
				log.Info("ClusterRoleBindings: ", crbTester.GetClusterRoleBindings())
			}
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		}
	})
}
