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

package container

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/api"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	containerTestConfig "github.com/redhat-nfvpe/test-network-function/pkg/tnf/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/testcases"
)

const (
	// The default test timeout.
	defaultTimeoutSeconds = 10
	configuredTestFile    = "testconfigure.yml"
	testSpecName          = "container"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	cnfInTest      *containerTestConfig.TnfContainerOperatorTestConfig
)

var _ = ginkgo.Describe(testSpecName, func() {

	ginkgo.When("a local shell is spawned", func() {
		goExpectSpawner := interactive.NewGoExpectSpawner()
		var spawner interactive.Spawner = goExpectSpawner
		context, err = interactive.SpawnShell(&spawner, defaultTimeout, expect.Verbose(true))
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(context).ToNot(gomega.BeNil())
		gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
	})
	defer ginkgo.GinkgoRecover()
	cnfInTest, err = containerTestConfig.GetConfig()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(cnfInTest).ToNot(gomega.BeNil())
	// Test for CNF certificates
	certAPIClient := api.NewHTTPClient()
	for _, cnf := range cnfInTest.CNFs {
		for _, certified := range cnf.CertifiedContainerRequestInfos {
			ginkgo.When(fmt.Sprintf("cnf certification test for: %s/%s ", certified.Repository, certified.Name), func() {
				ginkgo.It("tests for Container Certification Status", func() {
					certified := certified // pin
					isCertified := certAPIClient.IsContainerCertified(certified.Repository, certified.Name)
					gomega.Expect(isCertified).To(gomega.BeTrue())
				})
			})
		}
		// Gather facts for containers
		podFacts, err := testcases.LoadCnfTestCaseSpecs(testcases.GatherFacts)
		gomega.Expect(err).To(gomega.BeNil())
		for _, factsTest := range podFacts.TestCase {
			args := strings.Split(fmt.Sprintf(factsTest.Command, cnf.Name, cnf.Namespace), " ")
			cnfInTest := container.NewPod(args, cnf.Name, cnf.Namespace, factsTest.ExpectedStatus, factsTest.ResultType, factsTest.Action, defaultTimeout)
			test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			_, err = test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			if factsTest.Name == string(testcases.ContainerCount) {
				testcases.ContainerFacts[testcases.ContainerCount] = cnfInTest.Facts()
			} else if factsTest.Name == string(testcases.ServiceAccountName) {
				testcases.ContainerFacts[testcases.ServiceAccountName] = cnfInTest.Facts()
			}
		}
		// loop through various cnfs test
		for _, testType := range cnf.Tests {
			testFile, err := testcases.LoadConfiguredTestFile(configuredTestFile)
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
							testCase.ExpectedStatusFn(testcases.StatusFunctionType(val))
						}
					}
					var args []interface{}
					if testType == testcases.PrivilegedRoles {
						args = []interface{}{cnf.Namespace, cnf.Namespace, testcases.ContainerFacts[testcases.ServiceAccountName]}
					} else {
						args = []interface{}{cnf.Name, cnf.Namespace}
					}
					if testCase.Loop > 0 {
						containersCount, _ := strconv.Atoi(testcases.ContainerFacts[testcases.ContainerCount])
						runTestsOnCNF(args, cnf.Name, cnf.Namespace, containersCount, testCase)
					} else {
						runTestsOnCNF(args, cnf.Name, cnf.Namespace, testCase.Loop, testCase)
					}
				}
			}
		}

	}
})

//nolint:gocritic // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestsOnCNF(args []interface{}, name, namespace string, containerCount int, testCmd testcases.BaseTestCase) {
	ginkgo.When(fmt.Sprintf("cnf under test is: %s/%s ", namespace, name), func() {
		ginkgo.It(fmt.Sprintf("tests for: %s", testCmd.Name), func() {
			if containerCount > 0 {
				count := 0
				for count < containerCount {
					argsCount := append(args, count)
					cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, argsCount...), " ")
					cnfInTest := container.NewPod(cmdArgs, name, namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
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
				cnfInTest := container.NewPod(cmdArgs, name, namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
				gomega.Expect(cnfInTest).ToNot(gomega.BeNil())
				test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(test).ToNot(gomega.BeNil())
				testResult, err := test.Run()
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			}
		})
	})
}
