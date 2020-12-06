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

	expect "github.com/google/goexpect"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/api"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	operatorTestConfig "github.com/redhat-nfvpe/test-network-function/pkg/tnf/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/operator"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/testcases"
)

const (
	// The default test timeout.
	defaultTimeoutSeconds = 10
	configuredTestFile    = "testconfigure.yml"
	testSpecName          = "operator"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	operatorInTest *operatorTestConfig.TnfContainerOperatorTestConfig
)

var _ = ginkgo.Describe(testSpecName, func() {

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
	defer ginkgo.GinkgoRecover()
	operatorInTest, err = operatorTestConfig.GetConfig()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(operatorInTest).ToNot(gomega.BeNil())
	certAPIClient := api.NewHTTPClient()
	for _, operator := range operatorInTest.Operator {
		for _, certified := range operator.CertifiedOperatorRequestInfos {
			ginkgo.When(fmt.Sprintf("cnf certification test for: %s/%s ", certified.Organization, certified.Name), func() {
				ginkgo.It("tests for Operator Certification Status", func() {
					certified := certified // pin
					isCertified := certAPIClient.IsOperatorCertified(certified.Organization, certified.Name)
					gomega.Expect(isCertified).To(gomega.BeTrue())
				})
			})
		}
		// TODO: Gather facts for operator
		for _, testType := range operator.Tests {
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
							testCase.ExpectedStatusFn(testcases.StatusFunctionType(val))
						}
					}
					args := []interface{}{operator.Name, operator.Namespace}
					runTestsOnOperator(args, operator.Name, operator.Namespace, testCase)
				}
			}
		}
	}
})

//nolint:gocritic // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestsOnOperator(args []interface{}, name, namespace string, testCmd testcases.BaseTestCase) {
	ginkgo.When(fmt.Sprintf("operator under test is: %s/%s ", namespace, name), func() {
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
