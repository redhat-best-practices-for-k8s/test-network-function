// Copyright (C) 2021 Red Hat, Inc.
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

package autodiscover

import (
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	cnfLabelName       = "container"
	configuredTestFile = "testconfigure.yml"
)

var (
	cnfTestsAnnotationName = buildAnnotationName("container_tests")
)

// BuildCNFsConfig builds a `[]configsections.Cnf` from the current state of the cluster,
// using labels and annotations to populate the data.
func BuildCNFsConfig() (cnfs []configsections.Cnf) {
	pods, err := GetPodsByLabel(labelNamespace, cnfLabelName, anyLabelValue)
	if err != nil {
		log.Fatalf("found no CNFs to test while 'container' spec enabled: %s", err)
	}
	for i := range pods.Items {
		cnfs = append(cnfs, BuildCnfFromPodResource(&pods.Items[i]))
	}
	return cnfs
}

// BuildCnfFromPodResource builds a single `configsections.Cnf` from a PodResource
func BuildCnfFromPodResource(pr *PodResource) (cnf configsections.Cnf) {
	var err error
	cnf.Namespace = pr.Metadata.Namespace
	cnf.Name = pr.Metadata.Name

	var tests []string
	err = pr.GetAnnotationValue(cnfTestsAnnotationName, &tests)
	if err != nil {
		log.Warnf("unable to extract tests from annotation on '%s/%s' (error: %s). Attempting to fallback to all tests", cnf.Namespace, cnf.Name, err)
		cnf.Tests = getConfiguredCNFTests()
	} else {
		cnf.Tests = tests
	}
	return
}

// getConfiguredCNFTests loads the `configuredTestFile` used by the `operator` and `container` specs, and extracts
// the names of test groups from it.
func getConfiguredCNFTests() (cnfTests []string) {
	configuredTests, err := testcases.LoadConfiguredTestFile(configuredTestFile)
	if err != nil {
		log.Errorf("failed to load %s, continuing with no tests", configuredTestFile)
		return []string{}
	}
	for _, configuredTest := range configuredTests.CnfTest {
		cnfTests = append(cnfTests, configuredTest.Name)
	}
	log.WithField("cnfTests", cnfTests).Infof("got all tests from %s.", configuredTestFile)
	return cnfTests
}
