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
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	containerpkg "github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	ocGetCrKindFormat      = "oc get crd %s -o jsonpath='{.spec.names.kind}'"
	ocGetCrNamespaceFormat = "oc get %s -A -o go-template='{{range .items}}{{.metadata.name}},{{.metadata.namespace}}{{\"\n\"}}{{end}}'"
)

var (
	invalidCrNamespacePrefixes = []string{
		"istio-",
		"aspenmesh-",
	}

	tcClaimLogPrintf = func(format string, args ...interface{}) {
		message := fmt.Sprintf(format+"\n", args...)
		_, err := ginkgo.GinkgoWriter.Write([]byte(message))
		if err != nil {
			log.Errorf("Ginkgo writer could not write msg '%s' because: %s", message, err)
		}
	}
)

var _ = ginkgo.Describe(common.AccessControlTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.AccessControlTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})

		ginkgo.ReportAfterEach(results.RecordResult)

		testNamespace(env)

		testRoles(env)

		defer ginkgo.GinkgoRecover()

		// Run the tests that interact with the pods
		ginkgo.When("under test", func() {
			allTests := testcases.GetConfiguredPodTests()
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
						runTestOnPods(env, testCase, testType)
					}
				}
			}
		})
	}
})

//nolint:gocritic,funlen // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestOnPods(env *config.TestEnvironment, testCmd testcases.BaseTestCase, testType string) {
	testID := identifiers.XformToGinkgoItIdentifierExtended(identifiers.TestHostResourceIdentifier, testCmd.Name)
	ginkgo.It(testID, func() {
		context := common.GetContext()
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			ginkgo.By(fmt.Sprintf("Reading namespace of podnamespace= %s podname= %s", podNamespace, podName))
			if testCmd.ExpectedType == testcases.Function {
				for _, val := range testCmd.ExpectedStatus {
					testCmd.ExpectedStatusFn(podName, testcases.StatusFunctionType(val))
				}
			}
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
					test.RunAndValidate()
					count++
				}
			} else {
				cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, args...), " ")
				podTest := containerpkg.NewPod(cmdArgs, podUnderTest.Name, podUnderTest.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
				gomega.Expect(podTest).ToNot(gomega.BeNil())
				test, err := tnf.NewTest(context.GetExpecter(), podTest, []reel.Handler{podTest}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				test.RunAndValidate()
			}
		}
	})
}

func testCrsNamespaces(crNames []string) (invalidCrs map[string][]string) {
	invalidCrs = map[string][]string{}
	context := common.GetContext()
	for _, crdName := range crNames {
		getCrKindCommand := fmt.Sprintf(ocGetCrKindFormat, crdName)
		crdKind := utils.ExecuteCommand(getCrKindCommand, common.DefaultTimeout, context, func() {
			tcClaimLogPrintf("CRD %s: Failed to get CR kind.", crdName)
		})

		gomega.Expect(crdKind).NotTo(gomega.BeEmpty())
		getCrNamespaceCommand := fmt.Sprintf(ocGetCrNamespaceFormat, crdKind)
		cmdOut := utils.ExecuteCommand(getCrNamespaceCommand, common.DefaultTimeout, context, func() {
			tcClaimLogPrintf("CRD %s: Failed to get CRs (kind=%s)", crdName, crdKind)
		})
		// No CRs created for this CRD yet.
		if cmdOut == "" {
			continue
		}
		lines := strings.Split(cmdOut, "\n")

		if len(lines) == 0 {
			log.Infof("No CRs found for CRD: %s", crdName)
			continue
		}

		ginkgo.By(fmt.Sprintf("CRD %s has %d CRs. Checking their namespaces do not have the following prefixes: %v", crdName, len(lines)-1, invalidCrNamespacePrefixes))
		for _, line := range lines {
			if line == "" {
				continue
			}
			log.Debugf("CRD %s: Name|Namespace: %s", crdName, line)
			lineFields := strings.Split(line, ",")
			crName := lineFields[0]
			namespace := lineFields[1]

			for _, invalidPrefix := range invalidCrNamespacePrefixes {
				if strings.HasPrefix(namespace, invalidPrefix) {
					tcClaimLogPrintf("CR %s (kind=%s, crd=%s) has an invalid namespace (%s)", crName, crdKind, crdName, namespace)
					if crNames, exists := invalidCrs[crdName]; exists {
						invalidCrs[crdName] = append(crNames, crName)
					} else {
						invalidCrs[crdName] = []string{crName}
					}
				}
			}
		}
	}
	return invalidCrs
}

func testNamespace(env *config.TestEnvironment) {
	ginkgo.When("test deployment namespace", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestNamespaceBestPracticesIdentifier)
		ginkgo.It(testID, func() {
			for _, podUnderTest := range env.PodsUnderTest {
				podName := podUnderTest.Name
				podNamespace := podUnderTest.Namespace
				ginkgo.By(fmt.Sprintf("Reading namespace of podnamespace= %s podname= %s, should not be 'default' or begin with openshift-", podNamespace, podName))
				gomega.Expect(podNamespace).To(gomega.Not(gomega.Equal("default")))
				gomega.Expect(podNamespace).To(gomega.Not(gomega.HavePrefix("openshift-")))
			}

			invalidCrs := testCrsNamespaces(env.CrdNames)
			if len(invalidCrs) > 0 {
				for crdName, crs := range invalidCrs {
					for _, crName := range crs {
						tcClaimLogPrintf("CRD %s - CR %s has an invalid namespace.", crdName, crName)
					}
				}
				ginkgo.Fail("Found CRs with invalid namespaces.")
			}
		})
	})
}

func testRoles(env *config.TestEnvironment) {
	testServiceAccount(env)
	testRoleBindings(env)
	testClusterRoleBindings(env)
}

func testServiceAccount(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Should have a valid ServiceAccount name")
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			context := common.GetContext()
			ginkgo.By(fmt.Sprintf("Testing pod service account %s %s", podNamespace, podName))
			tester := serviceaccount.NewServiceAccount(common.DefaultTimeout, podName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			serviceAccountName := tester.GetServiceAccountName()
			gomega.Expect(serviceAccountName).ToNot(gomega.BeEmpty())
		}
	})
}

//nolint:dupl
func testRoleBindings(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Should not have RoleBinding in other namespaces")
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			context := common.GetContext()
			ginkgo.By(fmt.Sprintf("Testing role  bidning  %s %s", podNamespace, podName))
			if serviceAccountName == "" {
				ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
			}
			rbTester := rolebinding.NewRoleBinding(common.DefaultTimeout, serviceAccountName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), rbTester, []reel.Handler{rbTester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidateWithFailureCallback(func() { log.Info("RoleBindings: ", rbTester.GetRoleBindings()) })
		}
	})
}

//nolint:dupl
func testClusterRoleBindings(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Should not have ClusterRoleBindings")
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			context := common.GetContext()
			ginkgo.By(fmt.Sprintf("Testing cluster role  bidning  %s %s", podNamespace, podName))
			if serviceAccountName == "" {
				ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
			}
			crbTester := clusterrolebinding.NewClusterRoleBinding(common.DefaultTimeout, serviceAccountName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), crbTester, []reel.Handler{crbTester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidateWithFailureCallback(func() { log.Info("ClusterRoleBindings: ", crbTester.GetClusterRoleBindings()) })
		}
	})
}
