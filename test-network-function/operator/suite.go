// Copyright (C) 2020 Red Hat, Inc.
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
	expect "github.com/ryandgoulding/goexpect"
	"github.com/test-network-function/test-network-function/internal/api"
	configpool "github.com/test-network-function/test-network-function/pkg/config"
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
	interval     = 1
	testSpecName = "operator"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	operatorInTest *configpool.TnfContainerOperatorTestConfig
)

var _ = ginkgo.Describe(testSpecName, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testSpecName) {
		defer ginkgo.GinkgoRecover()
		ginkgo.When("a local shell is spawned", func() {
			goExpectSpawner := interactive.NewGoExpectSpawner()
			var spawner interactive.Spawner = goExpectSpawner
			context, err = interactive.SpawnShell(&spawner, defaultTimeout, expect.Verbose(true))
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

func itRunsTestsOnOperator() {
	operatorInTest, err = configpool.GetConfig()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(operatorInTest).ToNot(gomega.BeNil())
	//nolint:errcheck // Even if not run, each of the suites attempts to initialise the config. This results in
	// RegisterConfigurations erroring due to duplicate keys.
	(*configpool.GetInstance()).RegisterConfiguration(configpool.CNFConfigName, operatorInTest)
	certAPIClient := api.NewHTTPClient()
	for index := range operatorInTest.Operator {
		for _, certified := range operatorInTest.Operator[index].CertifiedOperatorRequestInfos {
			ginkgo.It(fmt.Sprintf("cnf certification test for: %s/%s ", certified.Organization, certified.Name), func() {
				certified := certified // pin
				gomega.Eventually(func() bool {
					isCertified := certAPIClient.IsOperatorCertified(certified.Organization, certified.Name)
					return isCertified
				}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
			})
		}
		// TODO: Gather facts for operator
		for _, testType := range operatorInTest.Operator[index].Tests {
			testFile, err := testcases.LoadConfiguredTestFile(configuredTestFile)
			gomega.Expect(testFile).ToNot(gomega.BeNil())
			gomega.Expect(err).To(gomega.BeNil())
			testConfigure := testcases.ContainsConfiguredTest(testFile.OperatorTest, testType)
			renderedTestCase, err := testConfigure.RenderTestCaseSpec(testcases.Operator, testType)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(renderedTestCase).ToNot(gomega.BeNil())
			for _, testCase := range renderedTestCase.TestCase {
				if !testCase.SkipTest {
					if testCase.ExpectedType == testcases.Function {
						for _, val := range testCase.ExpectedStatus {
							testCase.ExpectedStatusFn(operatorInTest.Operator[index].Name, testcases.StatusFunctionType(val))
						}
					}
					args := []interface{}{operatorInTest.Operator[index].Name, operatorInTest.Operator[index].Namespace}
					runTestsOnOperator(args, operatorInTest.Operator[index].Name, operatorInTest.Operator[index].Namespace, testCase)
				}
			}
		}
	}
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
