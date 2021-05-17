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

package container

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/internal/api"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	// The default test timeout.
	defaultTimeoutSeconds = 10
	// timeout for eventually call
	eventuallyTimeoutSeconds = 30
	// interval of time
	interval           = 1
	configuredTestFile = "testconfigure.yml"
	testSpecName       = "container"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	cnfsInTest     []configsections.Cnf
	certAPIClient  api.CertAPIClient
)

var _ = ginkgo.Describe(testSpecName, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testSpecName) {
		defer ginkgo.GinkgoRecover()
		ginkgo.When("a local shell is spawned", func() {
			goExpectSpawner := interactive.NewGoExpectSpawner()
			var spawner interactive.Spawner = goExpectSpawner
			context, err = interactive.SpawnShell(&spawner, defaultTimeout, interactive.Verbose(true))
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(context).ToNot(gomega.BeNil())
			gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
		})
		// Query API for certification status of listed containers
		ginkgo.When("getting certification status", func() {
			conf := config.GetConfigInstance()
			cnfsToQuery := conf.CertifiedContainerInfo
			gomega.Expect(cnfsToQuery).ToNot(gomega.BeNil())
			certAPIClient = api.NewHTTPClient()
			for _, cnfRequestInfo := range cnfsToQuery {
				cnf := cnfRequestInfo
				// Care: this test takes some time to run, failures at later points while before this has finished may be reported as a failure here. Read the failure reason carefully.
				ginkgo.It(fmt.Sprintf("container %s/%s should eventually be verified as certified", cnf.Repository, cnf.Name), func() {
					cnf := cnf // pin
					gomega.Eventually(func() bool {
						isCertified := certAPIClient.IsContainerCertified(cnf.Repository, cnf.Name)
						return isCertified
					}, eventuallyTimeoutSeconds, interval).Should(gomega.BeTrue())
				})
			}
		})
		// Run the tests that interact with the containers
		ginkgo.When("under test", func() {
			conf := config.GetConfigInstance()
			cnfsInTest = conf.CNFs
			gomega.Expect(cnfsInTest).ToNot(gomega.BeNil())
			for _, cnf := range cnfsInTest {
				cnf := cnf
				var containerFact = testcases.ContainerFact{Namespace: cnf.Namespace, Name: cnf.Name, ContainerCount: 0, HasClusterRole: false, Exists: true}
				// Gather facts for containers
				podFacts, err := testcases.LoadCnfTestCaseSpecs(testcases.GatherFacts)
				gomega.Expect(err).To(gomega.BeNil())
				// Collect container facts
				for _, factsTest := range podFacts.TestCase {
					args := strings.Split(fmt.Sprintf(factsTest.Command, cnf.Name, cnf.Namespace), " ")
					cnfInTest := container.NewPod(args, cnf.Name, cnf.Namespace, factsTest.ExpectedStatus, factsTest.ResultType, factsTest.Action, defaultTimeout)
					test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
					gomega.Expect(err).To(gomega.BeNil())
					gomega.Expect(test).ToNot(gomega.BeNil())
					_, err = test.Run()
					gomega.Expect(err).To(gomega.BeNil())
					if factsTest.Name == string(testcases.ContainerCount) {
						containerFact.ContainerCount, _ = strconv.Atoi(cnfInTest.Facts())
					} else if factsTest.Name == string(testcases.ServiceAccountName) {
						containerFact.ServiceAccount = cnfInTest.Facts()
					} else if factsTest.Name == string(testcases.Name) {
						containerFact.Name = cnfInTest.Facts()
						gomega.Expect(containerFact.Name).To(gomega.Equal(cnf.Name))
						if strings.Compare(containerFact.Name, cnf.Name) > 0 {
							containerFact.Exists = true
						}
					}
				}
				// loop through various cnfs test
				if !containerFact.Exists {
					ginkgo.It(fmt.Sprintf("is running test pod exists : %s/%s for test command :  %s", containerFact.Namespace, containerFact.Name, "POD EXISTS"), func() {
						gomega.Expect(containerFact.Exists).To(gomega.BeTrue())
					})
					continue
				}
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
									testCase.ExpectedStatusFn(cnf.Name, testcases.StatusFunctionType(val))
								}
							}
							if testCase.Loop > 0 {
								runTestsOnCNF(containerFact.ContainerCount, testCase, testType, containerFact)
							} else {
								runTestsOnCNF(testCase.Loop, testCase, testType, containerFact)
							}
						}
					}
				}

			}
		})
	}
})

//nolint:gocritic // ignore hugeParam error. Pointers to loop iterator vars are bad and `testCmd` is likely to be such.
func runTestsOnCNF(containerCount int, testCmd testcases.BaseTestCase,
	testType string, facts testcases.ContainerFact) {
	ginkgo.It(fmt.Sprintf("is running test for : %s/%s for test command :  %s", facts.Namespace, facts.Name, testCmd.Name), func() {
		containerCount := containerCount
		testType := testType
		facts := facts
		testCmd := testCmd
		var args []interface{}
		if testType == testcases.PrivilegedRoles {
			args = []interface{}{facts.Namespace, facts.Namespace, facts.ServiceAccount}
		} else {
			args = []interface{}{facts.Name, facts.Namespace}
		}
		if containerCount > 0 {
			count := 0
			for count < containerCount {
				argsCount := append(args, count)
				cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, argsCount...), " ")
				cnfInTest := container.NewPod(cmdArgs, facts.Name, facts.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
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
			cnfInTest := container.NewPod(cmdArgs, facts.Name, facts.Namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
			gomega.Expect(cnfInTest).ToNot(gomega.BeNil())
			test, err := tnf.NewTest(context.GetExpecter(), cnfInTest, []reel.Handler{cnfInTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		}
	})
}
