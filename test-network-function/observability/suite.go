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

package observability

import (
	"fmt"
	"path"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var (
	// loggingTestPath is the file location of the logging.json test case relative to the project root.
	loggingTestPath = path.Join("pkg", "tnf", "handlers", "logging", "logging.json")
	// relativeLoggingTestPath is the relative path to the logging.json test case.
	relativeLoggingTestPath = path.Join(common.PathRelativeToRoot, loggingTestPath)

	// crdTestPath is the file location of the CRD status existence test case relative to the project root.
	crdTestPath = path.Join("pkg", "tnf", "handlers", "crdstatusexistence", "crdstatusexistence.json")
	// relativeCrdTestPath is the relatieve path to the crdstatusexistence.json test case.
	relativeCrdTestPath = path.Join(common.PathRelativeToRoot, crdTestPath)
	// testCrdsTimeout is the timeout in seconds for the CRDs TC.
	testCrdsTimeout = 10 * time.Second
	// retrieve the singleton instance of test environment
	env *config.TestEnvironment = config.GetTestEnvironment()
)
var _ = ginkgo.Describe(common.ObservabilityTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()

	if testcases.IsInFocus(conf.FocusStrings, common.ObservabilityTestKey) {
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})
		ginkgo.ReportAfterEach(results.RecordResult)
		testLogging()
		testCrds()
	}
})

func testLogging() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLoggingIdentifier)
	ginkgo.It(testID, func() {
		for _, cut := range env.ContainersUnderTest {
			ginkgo.By(fmt.Sprintf("Test container: %+v. should emit at least one line of log to stderr/stdout", cut.ContainerIdentifier))
			loggingTest(&cut.ContainerIdentifier)
		}
	})
}
func loggingTest(c *configsections.ContainerIdentifier) {
	context := common.GetContext()

	values := make(map[string]interface{})
	values["POD_NAMESPACE"] = c.Namespace
	values["POD_NAME"] = c.PodName
	values["CONTAINER_NAME"] = c.ContainerName
	tester, handlers := utils.NewGenericTestAndValidate(relativeLoggingTestPath, common.RelativeSchemaPath, values)
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())

	test.RunAndValidate()
}

func testCrds() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestCrdsStatusSubresourceIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("CRDs should have a status subresource")
		context := common.GetContext()
		for _, crdName := range env.CrdNames {
			ginkgo.By("Testing CRD " + crdName)

			values := make(map[string]interface{})
			values["CRD_NAME"] = crdName
			values["TIMEOUT"] = testCrdsTimeout.Nanoseconds()

			tester, handlers := utils.NewGenericTestAndValidate(relativeCrdTestPath, common.RelativeSchemaPath, values)
			test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
			gomega.Expect(test).ToNot(gomega.BeNil())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunAndValidate()
		}
	})
}
