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

package operator

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"

	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/internal/api"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/operator"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	configuredTestFile = "testconfigure.yml"
	// The default test timeout.
	defaultTimeoutSeconds = 10
	// timeout for eventually call
	eventuallyTimeoutSeconds = 30
	// interval of time
	interval         = 1
	testSpecName     = "operator"
	subscriptionTest = "SUBSCRIPTION_INSTALLED"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error

	// checkSubscriptionTestPath is the file location of the uncordon.json test case relative to the project root.
	checkSubscriptionTestPath = path.Join("pkg", "tnf", "handlers", "checksubscription", "check-subscription.json")

	// pathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// relativeNodesTestPath is the relative path to the nodes.json test case.
	relativeNodesTestPath = path.Join(pathRelativeToRoot, checkSubscriptionTestPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

var _ = ginkgo.Describe(testSpecName, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testSpecName) {
		defer ginkgo.GinkgoRecover()
		ginkgo.When("a local shell is spawned", func() {
			goExpectSpawner := interactive.NewGoExpectSpawner()
			var spawner interactive.Spawner = goExpectSpawner
			context, err = interactive.SpawnShell(&spawner, defaultTimeout, interactive.Verbose(true))
			ginkgo.It("should be created without error", func() {
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(context).ToNot(gomega.BeNil())
				gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
			})
		})
		ginkgo.Context("Runs test on operators", func() {
			itRunsTestsOnOperator()
		})
		testOperatorsAreInstalledViaOLM()
	}
})

// testOperatorsAreInstalledViaOLM ensures all configured operators have a proper OLM subscription.
func testOperatorsAreInstalledViaOLM() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	ginkgo.It(testID, func() {
		_, operatorsInTest := getConfig()
		for _, operatorInTest := range operatorsInTest {
			defer results.RecordResult(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
			ginkgo.By(fmt.Sprintf("%s in namespace %s Should have a valid subscription", operatorInTest.SubscriptionName, operatorInTest.Namespace))
			testOperatorIsInstalledViaOLM(operatorInTest.SubscriptionName, operatorInTest.Namespace)
		}
	})
}

// testOperatorIsInstalledViaOLM tests that an operator is installed via OLM.
func testOperatorIsInstalledViaOLM(subscriptionName, subscriptionNamespace string) {
	values := make(map[string]interface{})
	values["SUBSCRIPTION_NAME"] = subscriptionName
	values["SUBSCRIPTION_NAMESPACE"] = subscriptionNamespace
	test, handlers, result, err := generic.NewGenericFromMap(relativeNodesTestPath, relativeSchemaPath, values)
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

func getConfig() ([]configsections.CertifiedOperatorRequestInfo, []configsections.Operator) {
	conf := config.GetTestEnvironment().Config
	operatorsToQuery := conf.CertifiedOperatorInfo
	operatorsInTest := conf.Operators
	return operatorsToQuery, operatorsInTest
}

func itRunsTestsOnOperator() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, func() {
		operatorsToQuery, operatorsInTest := getConfig()
		if len(operatorsToQuery) > 0 {
			certAPIClient := api.NewHTTPClient()
			for _, certified := range operatorsToQuery {
				// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
				ginkgo.By(fmt.Sprintf("should eventually be verified as certified (operator %s/%s)", certified.Organization, certified.Name))
				defer results.RecordResult(identifiers.TestOperatorIsCertifiedIdentifier)
				certified := certified // pin
				gomega.Eventually(func() bool {
					isCertified := certAPIClient.IsOperatorCertified(certified.Organization, certified.Name)
					return isCertified
				}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
			}
		}
		gomega.Expect(operatorsInTest).ToNot(gomega.BeNil())
		for _, op := range operatorsInTest {
			// TODO: Gather facts for operator
			for _, testType := range op.Tests {
				testFile, err := testcases.LoadConfiguredTestFile(configuredTestFile)
				gomega.Expect(testFile).ToNot(gomega.BeNil())
				gomega.Expect(err).To(gomega.BeNil())
				testConfigure := testcases.ContainsConfiguredTest(testFile.OperatorTest, testType)
				renderedTestCase, err := testConfigure.RenderTestCaseSpec(testcases.Operator, testType)
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(renderedTestCase).ToNot(gomega.BeNil())
				for _, testCase := range renderedTestCase.TestCase {
					if testCase.SkipTest {
						continue
					}
					if testCase.ExpectedType == testcases.Function {
						for _, val := range testCase.ExpectedStatus {
							testCase.ExpectedStatusFn(op.Name, testcases.StatusFunctionType(val))
						}
					}
					name := agrName(op.Name, op.SubscriptionName, testCase.Name)
					args := []interface{}{name, op.Namespace}
					runTestsOnOperator(args, name, op.Namespace, testCase)
				}
			}
		}
	})
}

func agrName(operatorName, subName, testName string) string {
	name := operatorName
	if testName == subscriptionTest {
		name = subName
	}
	return name
}

//nolint:gocritic // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestsOnOperator(args []interface{}, name, namespace string, testCmd testcases.BaseTestCase) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorInstallStatusIdentifier) + "-" + testCmd.Name
	ginkgo.It(testID, func() {
		defer results.RecordResult(identifiers.TestOperatorInstallStatusIdentifier)
		cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, args...), " ")
		opInTest := operator.NewOperator(cmdArgs, name, namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
		gomega.Expect(opInTest).ToNot(gomega.BeNil())
		test, err := tnf.NewTest(context.GetExpecter(), opInTest, []reel.Handler{opInTest}, context.GetErrorChannel())
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(test).ToNot(gomega.BeNil())
		testResult, err := test.Run()
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	})
}
