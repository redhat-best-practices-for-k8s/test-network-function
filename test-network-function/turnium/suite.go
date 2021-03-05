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

package turnium

import (
	"path"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	tnfConfig "github.com/redhat-nfvpe/test-network-function/pkg/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	expect "github.com/ryandgoulding/goexpect"
)

const (
	// testSuiteSpec contains the name of the Ginkgo test specification.
	testSuiteSpec = "turnium"
	// defaultTimeoutSeconds contains the default timeout in seconds.
	defaultTimeoutSeconds = 10
)

var (
	// defaultTestTimeout is the timeout for the test.
	defaultTestTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

	// testPath is the file location of the turnium.json test case relative to the project root.
	testPath = path.Join("test-network-function", "turnium", "turnium.json")

	// testPath is the file location of the Turnium config file relative to the project root.
	configPath = path.Join("test-network-function", "turnium-test-configuration.yml")

	// pathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// relativeTestPath is the relative path to the nodes.json test case.
	relativeTestPath = path.Join(pathRelativeToRoot, testPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeConfigPath = path.Join(pathRelativeToRoot, configPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

// createOc sets up an OC expect.Expecter, checking errors along the way.
func createOc(pod, container, namespace string) *interactive.Oc {
	oc, _, err := interactive.SpawnOc(interactive.CreateGoExpectSpawner(), pod, container, namespace, defaultTestTimeout, expect.Verbose(true))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(oc).ToNot(gomega.BeNil())
	return oc
}

var _ = ginkgo.Describe(testSuiteSpec, func() {
	ginkgo.When("the link state on the bonder is queried", func() {
		ginkgo.It("should report state is 'up'", func() {
			config := GetTurniumConfig()
			oc := createOc(config.TurniumBonder.PodName, config.TurniumBonder.ContainerName, config.TurniumBonder.Namespace)

			test, handlers, jsonParseResult, err := generic.NewGenericFromJSONFile(relativeTestPath, relativeSchemaPath)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(jsonParseResult).ToNot(gomega.BeNil())
			gomega.Expect(jsonParseResult.Valid()).To(gomega.BeTrue())
			gomega.Expect(handlers).ToNot(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())

			tester, err := tnf.NewTest(oc.GetExpecter(), *test, handlers, oc.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(tester).ToNot(gomega.BeNil())

			result, err := tester.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))
		})
	})
})

// GetTurniumConfig returns the cnf-certification-generic-tests test configuration.
func GetTurniumConfig() *tnfConfig.TurniumConfiguration {
	config, err := tnfConfig.GetTurniumConfiguration(relativeConfigPath)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(config).ToNot(gomega.BeNil())
	return config
}
