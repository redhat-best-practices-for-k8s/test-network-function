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

package observability

import (
	"fmt"
	"path"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
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
	// relativeCrdTestPath is the relative path to the crdstatusexistence.json test case.
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
		ginkgo.AfterEach(env.CloseLocalShellContext)
		testLogging()
		testCrds()
	}
})

func testLogging() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLoggingIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		failedCutIds := []*configsections.ContainerIdentifier{}
		for _, cut := range env.ContainersUnderTest {
			cutIdentifier := &cut.ContainerIdentifier
			ginkgo.By(fmt.Sprintf("Test container: %+v. should emit at least one line of log to stderr/stdout", cutIdentifier))

			context := env.GetLocalShellContext()

			values := make(map[string]interface{})
			values["POD_NAMESPACE"] = cutIdentifier.Namespace
			values["POD_NAME"] = cutIdentifier.PodName
			values["CONTAINER_NAME"] = cutIdentifier.ContainerName
			tester, handlers := utils.NewGenericTesterAndValidate(relativeLoggingTestPath, common.RelativeSchemaPath, values)
			test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())

			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("FAILURE: Container: %s (Pod %s ns %s) does not have any line of log to stderr/stdout",
					cutIdentifier.ContainerName, cutIdentifier.PodName, cutIdentifier.Namespace)
				failedCutIds = append(failedCutIds, cutIdentifier)
			}, func(err error) {
				tnf.ClaimFilePrintf("ERROR: Container: %s (Pod %s) does not have any line of log to stderr/stdout. Error: %v",
					cutIdentifier.ContainerName, cutIdentifier.PodName, cutIdentifier.Namespace, err)
				failedCutIds = append(failedCutIds, cutIdentifier)
			})
		}

		if n := len(failedCutIds); n > 0 {
			log.Debugf("Containers without logging: %+v", failedCutIds)
			ginkgo.Fail(fmt.Sprintf("%d containers don't have any log to stdout/stderr.", n))
		}
	})
}

func testCrds() {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestCrdsStatusSubresourceIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("CRDs should have a status subresource")
		context := env.GetLocalShellContext()
		failedCrds := []string{}
		for _, crdName := range env.CrdNames {
			ginkgo.By("Testing CRD " + crdName)

			values := make(map[string]interface{})
			values["CRD_NAME"] = crdName
			values["TIMEOUT"] = testCrdsTimeout.Nanoseconds()

			tester, handlers := utils.NewGenericTesterAndValidate(relativeCrdTestPath, common.RelativeSchemaPath, values)
			test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
			gomega.Expect(test).ToNot(gomega.BeNil())
			gomega.Expect(err).To(gomega.BeNil())

			test.RunWithCallbacks(nil, func() {
				tnf.ClaimFilePrintf("FAILURE: CRD %s does not have a status subresource.", crdName)
				failedCrds = append(failedCrds, crdName)
			}, func(err error) {
				tnf.ClaimFilePrintf("FAILURE: CRD %s does not have a status subresource.", crdName)
				failedCrds = append(failedCrds, crdName)
			})
		}

		if n := len(failedCrds); n > 0 {
			log.Debugf("CRDs without status subresource: %+v", failedCrds)
			ginkgo.Fail(fmt.Sprintf("%d CRDs don't have status subresource", n))
		}
	})
}
