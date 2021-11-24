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
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/automountservice"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	containerpkg "github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	// ocGetCrPluralNameFormat is the CR name to use with "oc get <resource_name>".
	ocGetCrPluralNameFormat = "oc get crd %s -o jsonpath='{.spec.names.plural}'"

	// ocGetCrNamespaceFormat is the "oc get" format string to get the namespaced-only resources created for a given CRD.
	ocGetCrNamespaceFormat = "oc get %s -A -o go-template='{{range .items}}{{if .metadata.namespace}}{{.metadata.name}},{{.metadata.namespace}}{{\"\n\"}}{{end}}{{end}}'"
)

var (
	invalidNamespacePrefixes = []string{
		"default",
		"openshift-",
		"istio-",
		"aspenmesh-",
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

func getCrsNamespaces(crdName, crdKind string) (map[string]string, error) {
	const expectedNumFields = 2
	const crNameFieldIdx = 0
	const namespaceFieldIdx = 0

	gomega.Expect(crdKind).NotTo(gomega.BeEmpty())
	getCrNamespaceCommand := fmt.Sprintf(ocGetCrNamespaceFormat, crdKind)
	cmdOut := utils.ExecuteCommand(getCrNamespaceCommand, common.DefaultTimeout, common.GetContext(), func() {
		tnf.ClaimFilePrintf("CRD %s: Failed to get CRs (kind=%s)", crdName, crdKind)
	})

	crNamespaces := map[string]string{}

	if cmdOut == "" {
		// Filter out empty (0 CRs) output.
		return crNamespaces, nil
	}

	lines := strings.Split(cmdOut, "\n")
	for _, line := range lines {
		lineFields := strings.Split(line, ",")
		if len(lineFields) != expectedNumFields {
			return crNamespaces, fmt.Errorf("failed to parse output line %s", line)
		}
		crNamespaces[lineFields[crNameFieldIdx]] = lineFields[namespaceFieldIdx]
	}

	return crNamespaces, nil
}

func testCrsNamespaces(crNames, configNamespaces []string) (invalidCrs map[string][]string) {
	invalidCrs = map[string][]string{}
	for _, crdName := range crNames {
		getCrPluralNameCommand := fmt.Sprintf(ocGetCrPluralNameFormat, crdName)
		crdPluralName := utils.ExecuteCommand(getCrPluralNameCommand, common.DefaultTimeout, common.GetContext(), func() {
			tnf.ClaimFilePrintf("CRD %s: Failed to get CR plural name.", crdName)
		})

		crNamespaces, err := getCrsNamespaces(crdName, crdPluralName)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Failed to get CRs for CRD %s - Error: %v", crdName, err))
		}

		ginkgo.By(fmt.Sprintf("CRD %s has %d CRs (plural name: %s).", crdName, len(crNamespaces), crdPluralName))
		for crName, namespace := range crNamespaces {
			ginkgo.By(fmt.Sprintf("Checking CR %s - Namespace %s", crName, namespace))
			found := false
			for _, configNamespace := range configNamespaces {
				if namespace == configNamespace {
					found = true
					break
				}
			}

			if !found {
				tnf.ClaimFilePrintf("CRD: %s (kind:%s) - CR %s has an invalid namespace (%s)", crdName, crdPluralName, crName, namespace)
				if crNames, exists := invalidCrs[crdName]; exists {
					invalidCrs[crdName] = append(crNames, crName)
				} else {
					invalidCrs[crdName] = []string{crName}
				}
			}
		}
	}
	return invalidCrs
}

func testNamespace(env *config.TestEnvironment) {
	ginkgo.When("test CNF namespaces", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestNamespaceBestPracticesIdentifier)
		ginkgo.It(testID, func() {
			ginkgo.By(fmt.Sprintf("CNF resources' namespaces should not have any of the following prefixes: %v", invalidNamespacePrefixes))
			var failedNamespaces []string
			for _, namespace := range env.NameSpacesUnderTest {
				ginkgo.By(fmt.Sprintf("Checking namespace %s", namespace))
				for _, invalidPrefix := range invalidNamespacePrefixes {
					if strings.HasPrefix(namespace, invalidPrefix) {
						tnf.ClaimFilePrintf("Namespace %s has invalid prefix %s", namespace, invalidPrefix)
						failedNamespaces = append(failedNamespaces, namespace)
					}
				}
			}

			if failedNamespacesNum := len(failedNamespaces); failedNamespacesNum > 0 {
				ginkgo.Fail(fmt.Sprintf("Found %d namespaces with an invalid prefix.", failedNamespacesNum))
			}

			ginkgo.By(fmt.Sprintf("CNF pods' should belong to any of the configured namespaces: %v", env.NameSpacesUnderTest))

			if nonValidPodsNum := len(env.Config.NonValidPods); nonValidPodsNum > 0 {
				for _, invalidPod := range env.Config.NonValidPods {
					tnf.ClaimFilePrintf("Pod %s has invalid namespace %s", invalidPod.Name, invalidPod.Namespace)
				}

				ginkgo.Fail(fmt.Sprintf("Found %d pods under test belonging to invalid namespaces.", nonValidPodsNum))
			}

			ginkgo.By(fmt.Sprintf("CRs from autodiscovered CRDs should belong to the configured namespaces: %v", env.NameSpacesUnderTest))
			invalidCrs := testCrsNamespaces(env.CrdNames, env.NameSpacesUnderTest)

			if invalidCrsNum := len(invalidCrs); invalidCrsNum > 0 {
				for crdName, crs := range invalidCrs {
					for _, crName := range crs {
						tnf.ClaimFilePrintf("CRD %s - CR %s has an invalid namespace.", crdName, crName)
					}
				}
				ginkgo.Fail(fmt.Sprintf("Found %d CRs belonging to invalid namespaces.", invalidCrsNum))
			}
		})
	})
}

func testRoles(env *config.TestEnvironment) {
	testServiceAccount(env)
	testRoleBindings(env)
	testClusterRoleBindings(env)
	testAutomountService(env)
}

func testServiceAccount(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Should have a valid ServiceAccount name")
		for _, podUnderTest := range env.PodsUnderTest {
			ginkgo.By(fmt.Sprintf("Testing pod service account %s %s", podUnderTest.Namespace, podUnderTest.Name))
			serviceAccountName := podUnderTest.ServiceAccount
			gomega.Expect(serviceAccountName).ToNot(gomega.BeEmpty())
		}
	})
}
func testAutomountService(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodAutomountServiceAccountIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Should have automountServiceAccountToken set to false")
		msg := []string{}
		for _, podUnderTest := range env.PodsUnderTest {
			ginkgo.By(fmt.Sprintf("check the existence of pod service account %s %s", podUnderTest.Namespace, podUnderTest.Name))
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			gomega.Expect(serviceAccountName).ToNot(gomega.BeEmpty())
			context := common.GetContext()
			tester := automountservice.NewAutomountservice(automountservice.WithNamespace(podNamespace), automountservice.WithServiceAccount(serviceAccountName))
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			serviceAccountToken := tester.Token()
			tester = automountservice.NewAutomountservice(automountservice.WithNamespace(podNamespace), automountservice.WithPodname(podName))
			test, err = tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			podToken := tester.Token()
			// the test would pass iif token is explicitly set to false
			if podToken == automountservice.TokenIsTrue {
				msg = append(msg, fmt.Sprintf("Pod %s:%s is configured with automountServiceAccountToken set to true ", podNamespace, podName))
				continue
			}
			if podToken == automountservice.TokenIsFalse || serviceAccountToken == automountservice.TokenIsFalse {
				// properly configured
				continue
			}
			if serviceAccountToken == automountservice.TokenIsTrue {
				msg = append(msg, fmt.Sprintf("serviceaccount %s:%s is configured with automountServiceAccountToken set to true, impacting pod %s ", podNamespace, serviceAccountName, podName))
			}
			if serviceAccountToken == automountservice.TokenNotSet {
				msg = append(msg, fmt.Sprintf("serviceaccount %s:%s is not configured with automountServiceAccountToken set to false, impacting pod %s ", podNamespace, serviceAccountName, podName))
			}
		}

		if len(msg) > 0 {
			_, err := ginkgo.GinkgoWriter.Write([]byte(strings.Join(msg, "")))
			if err != nil {
				log.Errorf("Ginkgo writer could not write because: %s", err)
			}
		}
		gomega.Expect(msg).To(gomega.BeEmpty())
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
			ginkgo.By(fmt.Sprintf("Testing cluster role  binding  %s %s", podNamespace, podName))
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
