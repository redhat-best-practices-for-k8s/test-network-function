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
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/internal/api"
	"github.com/test-network-function/test-network-function/pkg/config"
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
	}
})

func getConfig() ([]config.CertifiedOperatorRequestInfo, []config.Operator) {
	conf := config.GetConfigInstance()
	operatorsToQuery := conf.CertifiedOperatorInfo
	operatorsInTest := conf.Operators
	return operatorsToQuery, operatorsInTest
}

func itRunsTestsOnOperator() {
	operatorsToQuery, operatorsInTest := getConfig()
	gomega.Expect(operatorsToQuery).ToNot(gomega.BeNil())
	certAPIClient := api.NewHTTPClient()
	for _, certified := range operatorsToQuery {
		ginkgo.It(fmt.Sprintf("should eventually be verified as certified (operator %s/%s)", certified.Organization, certified.Name), func() {
			certified := certified // pin
			gomega.Eventually(func() bool {
				isCertified := certAPIClient.IsOperatorCertified(certified.Organization, certified.Name)
				return isCertified
			}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
		})
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
	ginkgo.When(fmt.Sprintf("under test is: %s/%s ", namespace, name), func() {
		ginkgo.It(fmt.Sprintf("tests for: %s", testCmd.Name), func() {
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
	})
}
