// Copyright (C) 2020-2022 Red Hat, Inc.
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

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/automountservice"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterrolebinding"
	containerpkg "github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
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
		ginkgo.AfterEach(env.CloseLocalShellContext)

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

type failedTcInfo struct {
	tc           string
	containerIdx int
	ns           string
}

func addFailedTcInfo(failedTcs map[string][]failedTcInfo, tc, pod, ns string, containerIdx int) {
	if tcs, exists := failedTcs[pod]; exists {
		tcs = append(tcs, failedTcInfo{tc: tc, containerIdx: containerIdx, ns: ns})
		failedTcs[pod] = tcs
	} else {
		failedTcs[pod] = []failedTcInfo{{tc: tc, containerIdx: containerIdx, ns: ns}}
	}
}

//nolint:gocritic,funlen // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestOnPods(env *config.TestEnvironment, testCmd testcases.BaseTestCase, testType string) {
	const noContainerIdx = -1
	testID := identifiers.XformToGinkgoItIdentifierExtended(identifiers.TestHostResourceIdentifier, testCmd.Name)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		context := env.GetLocalShellContext()
		failedTcs := map[string][]failedTcInfo{} // maps a pod name to a slice of failed TCs
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
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
					ginkgo.By(fmt.Sprintf("Executing TC %s on pod %s (ns %s), container index %d", testCmd.Name, podNamespace, podName, count))
					argsCount := append(args, count)
					cmd := fmt.Sprintf(testCmd.Command, argsCount...)
					cmdArgs := strings.Split(cmd, " ")
					cnfInTest := containerpkg.NewPod(cmdArgs, podUnderTest.Name, podUnderTest.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
					gomega.Expect(cnfInTest).ToNot(gomega.BeNil())
					test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(test).ToNot(gomega.BeNil())
					test.RunWithCallbacks(nil, func() {
						tnf.ClaimFilePrintf("FAILURE: Command sent: %s, Expectations: %v", cmd, testCmd.ExpectedStatus)
						addFailedTcInfo(failedTcs, testCmd.Name, podName, podNamespace, count)
					}, func(e error) {
						tnf.ClaimFilePrintf("ERROR: Command sent: %s, Expectations: %v, Error: %v", cmd, testCmd.ExpectedStatus, e)
						addFailedTcInfo(failedTcs, testCmd.Name, podName, podNamespace, count)
					})
					count++
				}
			} else {
				ginkgo.By(fmt.Sprintf("Executing TC %s on pod %s (ns %s)", testCmd.Name, podNamespace, podName))
				cmd := fmt.Sprintf(testCmd.Command, args...)
				cmdArgs := strings.Split(cmd, " ")
				podTest := containerpkg.NewPod(cmdArgs, podUnderTest.Name, podUnderTest.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, common.DefaultTimeout)
				gomega.Expect(podTest).ToNot(gomega.BeNil())
				test, err := tnf.NewTest(context.GetExpecter(), podTest, []reel.Handler{podTest}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				test.RunWithCallbacks(nil, func() {
					tnf.ClaimFilePrintf("FAILURE: Command sent: %s, Expectations: %v", cmd, testCmd.ExpectedStatus)
					addFailedTcInfo(failedTcs, testCmd.Name, podName, podNamespace, noContainerIdx)
				}, func(e error) {
					tnf.ClaimFilePrintf("ERROR: Command sent: %s, Expectations: %v, Error: %v", cmd, testCmd.ExpectedStatus, e)
					addFailedTcInfo(failedTcs, testCmd.Name, podName, podNamespace, noContainerIdx)
				})
			}
		}

		if n := len(failedTcs); n > 0 {
			log.Debugf("Failed TCs: %+v", failedTcs)
			ginkgo.Fail(fmt.Sprintf("%d pods failed the test.", n))
		}
	})
}

func getCrsNamespaces(crdName, crdKind string, context *interactive.Context) (map[string]string, error) {
	const expectedNumFields = 2
	const crNameFieldIdx = 0
	const namespaceFieldIdx = 0

	gomega.Expect(crdKind).NotTo(gomega.BeEmpty())
	getCrNamespaceCommand := fmt.Sprintf(ocGetCrNamespaceFormat, crdKind)
	cmdOut := utils.ExecuteCommandAndValidate(getCrNamespaceCommand, common.DefaultTimeout, context, func() {
		common.TcClaimLogPrintf("CRD %s: Failed to get CRs (kind=%s)", crdName, crdKind)
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

func testCrsNamespaces(crNames, configNamespaces []string, context *interactive.Context) map[string][]string {
	invalidCrs := map[string][]string{}
	for _, crdName := range crNames {
		getCrPluralNameCommand := fmt.Sprintf(ocGetCrPluralNameFormat, crdName)
		crdPluralName := utils.ExecuteCommandAndValidate(getCrPluralNameCommand, common.DefaultTimeout, context, func() {
			common.TcClaimLogPrintf("CRD %s: Failed to get CR plural name.", crdName)
		})

		crNamespaces, err := getCrsNamespaces(crdName, crdPluralName, context)
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
				common.TcClaimLogPrintf("CRD: %s (kind:%s) - CR %s has an invalid namespace (%s)", crdName, crdPluralName, crName, namespace)
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
		ginkgo.It(testID, ginkgo.Label(testID), func() {
			ginkgo.By(fmt.Sprintf("CNF resources' namespaces should not have any of the following prefixes: %v", invalidNamespacePrefixes))
			var failedNamespaces []string
			for _, namespace := range env.NameSpacesUnderTest {
				ginkgo.By(fmt.Sprintf("Checking namespace %s", namespace))
				for _, invalidPrefix := range invalidNamespacePrefixes {
					if strings.HasPrefix(namespace, invalidPrefix) {
						common.TcClaimLogPrintf("Namespace %s has invalid prefix %s", namespace, invalidPrefix)
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
					common.TcClaimLogPrintf("Pod %s has invalid namespace %s", invalidPod.Name, invalidPod.Namespace)
				}

				ginkgo.Fail(fmt.Sprintf("Found %d pods under test belonging to invalid namespaces.", nonValidPodsNum))
			}

			ginkgo.By(fmt.Sprintf("CRs from autodiscovered CRDs should belong to the configured namespaces: %v", env.NameSpacesUnderTest))
			invalidCrs := testCrsNamespaces(env.CrdNames, env.NameSpacesUnderTest, env.GetLocalShellContext())

			if invalidCrsNum := len(invalidCrs); invalidCrsNum > 0 {
				for crdName, crs := range invalidCrs {
					for _, crName := range crs {
						common.TcClaimLogPrintf("CRD %s - CR %s has an invalid namespace.", crdName, crName)
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
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Should have a valid ServiceAccount name")
		failedPods := []*configsections.Pod{}
		for _, podUnderTest := range env.PodsUnderTest {
			ginkgo.By(fmt.Sprintf("Testing service account for pod %s (ns: %s)", podUnderTest.Name, podUnderTest.Namespace))
			if podUnderTest.ServiceAccount == "" {
				tnf.ClaimFilePrintf("Pod %s (ns: %s) doesn't have a service account name.", podUnderTest.Name, podUnderTest.Namespace)
				failedPods = append(failedPods, podUnderTest)
			}
		}
		if n := len(failedPods); n > 0 {
			log.Debugf("Pods without service account: %+v", failedPods)
			ginkgo.Fail(fmt.Sprintf("%d pods don't have a service account name.", n))
		}
	})
}

//nolint:funlen
func testAutomountService(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodAutomountServiceAccountIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Should have automountServiceAccountToken set to false")
		msg := []string{}
		for _, podUnderTest := range env.PodsUnderTest {
			ginkgo.By(fmt.Sprintf("check the existence of pod service account %s (ns= %s )", podUnderTest.Namespace, podUnderTest.Name))
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			gomega.Expect(serviceAccountName).ToNot(gomega.BeEmpty())
			context := env.GetLocalShellContext()
			tester := automountservice.NewAutomountService(automountservice.WithNamespace(podNamespace), automountservice.WithServiceAccount(serviceAccountName))
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			serviceAccountToken := tester.Token()
			tester = automountservice.NewAutomountService(automountservice.WithNamespace(podNamespace), automountservice.WithPodname(podName))
			test, err = tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
			podToken := tester.Token()
			// The token can be specified in the pod directly
			// or it can be specified in the service account of the pod
			// if no service account is configured, then the pod will use the configuration
			// of the default service account in that namespace
			// the token defined in the pod has takes precedence
			// the test would pass iif token is explicitly set to false
			// if the token is set to true in the pod, the test would fail right away
			if podToken == automountservice.TokenIsTrue {
				msg = append(msg, fmt.Sprintf("Pod %s:%s is configured with automountServiceAccountToken set to true ", podNamespace, podName))
				continue
			}
			// The pod token is false means the pod is configured properly
			// The pod is not configured and the service account is configured with false means
			// the pod will inherit the behavior `false` and the test would pass
			if podToken == automountservice.TokenIsFalse || serviceAccountToken == automountservice.TokenIsFalse {
				continue
			}
			// the service account is configured with true means all the pods
			// using this service account are not configured properly, register the error
			// message and fail
			if serviceAccountToken == automountservice.TokenIsTrue {
				msg = append(msg, fmt.Sprintf("serviceaccount %s:%s is configured with automountServiceAccountToken set to true, impacting pod %s ", podNamespace, serviceAccountName, podName))
			}
			// the token should be set explicitly to false, otherwise, it's a failure
			// register the error message and check the next pod
			if serviceAccountToken == automountservice.TokenNotSet {
				msg = append(msg, fmt.Sprintf("serviceaccount %s:%s is not configured with automountServiceAccountToken set to false, impacting pod %s ", podNamespace, serviceAccountName, podName))
			}
		}
		if len(msg) > 0 {
			tnf.ClaimFilePrintf(strings.Join(msg, ""))
		}
		gomega.Expect(msg).To(gomega.BeEmpty())
	})
}

func testRoleBindings(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		failedPods := []*configsections.Pod{}
		ginkgo.By("Should not have RoleBinding in other namespaces")
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			context := env.GetLocalShellContext()
			ginkgo.By(fmt.Sprintf("Testing role binding  %s %s", podNamespace, podName))
			if serviceAccountName == "" {
				ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
			}
			rbTester := rolebinding.NewRoleBinding(common.DefaultTimeout, serviceAccountName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), rbTester, []reel.Handler{rbTester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("FAILURE: Pod %s (ns: %s) roleBindings: %v", podName, podNamespace, rbTester.GetRoleBindings())
				failedPods = append(failedPods, podUnderTest)
			}, func(err error) {
				tnf.ClaimFilePrintf("ERROR: Pod %s (ns: %s) roleBindings: %v, error: %v", podName, podNamespace, rbTester.GetRoleBindings(), err)
				failedPods = append(failedPods, podUnderTest)
			})
		}
		if n := len(failedPods); n > 0 {
			log.Debugf("Pods with role bindings: %+v", failedPods)
			ginkgo.Fail(fmt.Sprintf("%d pods have role bindings in other namespaces.", n))
		}
	})
}

func testClusterRoleBindings(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Should not have ClusterRoleBindings")
		failedPods := []*configsections.Pod{}
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNamespace := podUnderTest.Namespace
			serviceAccountName := podUnderTest.ServiceAccount
			context := env.GetLocalShellContext()
			ginkgo.By(fmt.Sprintf("Testing cluster role binding  %s %s", podNamespace, podName))
			if serviceAccountName == "" {
				ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
			}
			crbTester := clusterrolebinding.NewClusterRoleBinding(common.DefaultTimeout, serviceAccountName, podNamespace)
			test, err := tnf.NewTest(context.GetExpecter(), crbTester, []reel.Handler{crbTester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("FAILURE: Pod: %s (ns: %s) SA: %s clusterRoleBindings: %v", podName, podNamespace, serviceAccountName, crbTester.GetClusterRoleBindings())
				failedPods = append(failedPods, podUnderTest)
			}, func(err error) {
				tnf.ClaimFilePrintf("ERROR: Pod: %s (ns: %s) SA: %s clusterRoleBindings: %v, error: %v", podName, podNamespace, serviceAccountName, crbTester.GetClusterRoleBindings(), err)
				failedPods = append(failedPods, podUnderTest)
			})
		}
		if n := len(failedPods); n > 0 {
			log.Debugf("Pods with cluster role bindings: %+v", failedPods)
			ginkgo.Fail(fmt.Sprintf("%d pods have cluster role bindings.", n))
		}
	})
}
