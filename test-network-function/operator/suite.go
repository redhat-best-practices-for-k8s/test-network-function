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

package operator

import (
	"fmt"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/utils"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/operator"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	configuredTestFile = "testconfigure.yml"
	// The default test timeout.
	testSpecName = "operator"
)

var (

	// checkSubscriptionTestPath is the file location of the uncordon.json test case relative to the project root.
	checkSubscriptionTestPath = path.Join("pkg", "tnf", "handlers", "checksubscription", "check-subscription.json")

	// pathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// relativecheckSubscriptionTestPath is the relative path to the nodes.json test case.
	relativecheckSubscriptionTestPath = path.Join(pathRelativeToRoot, checkSubscriptionTestPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

var _ = ginkgo.Describe(testSpecName, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, testSpecName) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			if len(env.OperatorsUnderTest) == 0 {
				ginkgo.Skip("No Operator found.")
			}
		})
		ginkgo.ReportAfterEach(results.RecordResult)
		ginkgo.AfterEach(env.CloseLocalShellContext)
		defer ginkgo.GinkgoRecover()
		ginkgo.Context("Runs test on operators", func() {
			itRunsTestsOnOperator(env)
		})
		testOperatorsAreInstalledViaOLM(env)
	}
})

// testOperatorsAreInstalledViaOLM ensures all configured operators have a proper OLM subscription.
func testOperatorsAreInstalledViaOLM(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badOperators := []configsections.Operator{}
		for _, operatorInTest := range env.OperatorsUnderTest {
			ginkgo.By(fmt.Sprintf("%s in namespace %s Should have a valid subscription", operatorInTest.SubscriptionName, operatorInTest.Namespace))
			values := make(map[string]interface{})
			values["SUBSCRIPTION_NAME"] = operatorInTest.SubscriptionName
			values["SUBSCRIPTION_NAMESPACE"] = operatorInTest.Namespace
			tester, handlers := utils.NewGenericTesterAndValidate(relativecheckSubscriptionTestPath, relativeSchemaPath, values)
			context := env.GetLocalShellContext()
			test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())

			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("Operator %s doesn't have a proper OLM subscription.", operatorInTest.Name)
				badOperators = append(badOperators, operatorInTest)
			}, func(err error) {
				tnf.ClaimFilePrintf("Operator %s doesn't have a proper OLM subscription. Error: %v", operatorInTest.Name, err)
				badOperators = append(badOperators, operatorInTest)
			})
		}

		if n := len(badOperators); n > 0 {
			log.Warnf("Operators without proper OLM subscription: %+v", badOperators)
			ginkgo.Fail(fmt.Sprintf("%d operators found without proper OLM subscription.", n))
		}
	})
}

func itRunsTestsOnOperator(env *config.TestEnvironment) {
	for _, testType := range testcases.GetConfiguredOperatorTests() {
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
			runTestsOnOperator(env, testCase)
		}
	}
}

//nolint:gocritic // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestsOnOperator(env *config.TestEnvironment, testCase testcases.BaseTestCase) {
	testID := identifiers.XformToGinkgoItIdentifierExtended(identifiers.TestOperatorInstallStatusIdentifier, testCase.Name)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badOperators := []configsections.Operator{}
		for _, op := range env.OperatorsUnderTest {
			if testCase.ExpectedType == testcases.Function {
				for _, val := range testCase.ExpectedStatus {
					testCase.ExpectedStatusFn(op.Name, testcases.StatusFunctionType(val))
				}
			}
			name := op.Name
			args := []interface{}{name, op.Namespace}
			cmdArgs := strings.Split(fmt.Sprintf(testCase.Command, args...), " ")
			opInTest := operator.NewOperator(cmdArgs, name, op.Namespace, testCase.ExpectedStatus, testCase.ResultType, testCase.Action, common.DefaultTimeout)
			gomega.Expect(opInTest).ToNot(gomega.BeNil())
			context := env.GetLocalShellContext()
			test, err := tnf.NewTest(context.GetExpecter(), opInTest, []reel.Handler{opInTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())

			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("Operator %s failed TC: %s", name, testCase.Name)
				badOperators = append(badOperators, op)
			}, func(err error) {
				tnf.ClaimFilePrintf("Operator %s failed TC: %s. Error: %v", name, testCase.Name, err)
				badOperators = append(badOperators, op)
			})
		}

		if n := len(badOperators); n > 0 {
			log.Warnf("Operators that failed TC %s: %+v", testCase.Name, badOperators)
			ginkgo.Fail(fmt.Sprintf("%d operators failed TC %s", n, testCase.Name))
		}
	})
}
