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

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
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
)

var _ = ginkgo.Describe(common.ObservabilityTestKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.ObservabilityTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})
		testLogging(env)
	}
})

func testLogging(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLoggingIdentifier)
	ginkgo.It(testID, func() {
		for _, cut := range env.ContainersUnderTest {
			ginkgo.By(fmt.Sprintf("Test container: %+v. should emit at least one line of log to stderr/stdout", cut.ContainerIdentifier))
			defer results.RecordResult(identifiers.TestLoggingIdentifier)
			loggingTest(cut.ContainerIdentifier)
		}
	})
}
func loggingTest(c configsections.ContainerIdentifier) {
	context := common.GetContext()

	values := make(map[string]interface{})
	values["POD_NAMESPACE"] = c.Namespace
	values["POD_NAME"] = c.PodName
	values["CONTAINER_NAME"] = c.ContainerName
	test, handlers, result, err := generic.NewGenericFromMap(relativeLoggingTestPath, common.RelativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())
	tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	testResult, err := tester.Run()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
}
