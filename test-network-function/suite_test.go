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

package suite

import (
	j "encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/junit"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	_ "github.com/test-network-function/test-network-function/test-network-function/container"
	"github.com/test-network-function/test-network-function/test-network-function/diagnostic"
	_ "github.com/test-network-function/test-network-function/test-network-function/generic"
	_ "github.com/test-network-function/test-network-function/test-network-function/operator"
	"github.com/test-network-function/test-network-function/test-network-function/version"
)

const (
	claimFileName                        = "claim.json"
	claimFilePermissions                 = 0644
	claimPathFlagKey                     = "claimloc"
	CnfCertificationTestSuiteName        = "CNF Certification Test Suite"
	defaultClaimPath                     = ".."
	defaultCliArgValue                   = ""
	junitFlagKey                         = "junit"
	TNFJunitXMLFileName                  = "cnf-certification-tests_junit.xml"
	TNFReportKey                         = "cnf-certification-test"
	CNFFeatureValidationJunitXMLFileName = "validation_junit.xml"
	CNFFeatureValidationReportKey        = "cnf-feature-validation"
	reportFlagKey                        = "report"
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	dateTimeFormatDirective = "2006-01-02T15:04:05+00:00"
	extraInfoKey            = "testsExtraInfo"
)

var (
	claimPath  *string
	junitPath  *string
	reportPath *string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
	reportPath = flag.String(reportFlagKey, defaultCliArgValue,
		"the path of the report file containing details for failed tests")
}

func createClaimRoot() *claim.Root {
	// Initialize the claim with the start time.
	startTime := time.Now()
	c := &claim.Claim{
		Metadata: &claim.Metadata{
			StartTime: startTime.UTC().Format(dateTimeFormatDirective),
		},
	}
	return &claim.Root{
		Claim: c,
	}
}

func loadJunitXMLIntoMap(result map[string]interface{}, junitFilepath, key string) {
	var err error
	if key == "" {
		var extension = filepath.Ext(junitFilepath)
		key = junitFilepath[0 : len(junitFilepath)-len(extension)]
	}
	result[key], err = junit.ExportJUnitAsJSON(junitFilepath)
	if err != nil {
		log.Fatalf("error reading JUnit XML file into JSON: %v", err)
	}
}

// Invokes the CNF Certification Tests.
//nolint:funlen // Function is long but core entrypoint and linear. Consider revisiting later.
func TestTest(t *testing.T) {
	flag.Parse()
	claimOutputFile := filepath.Join(*claimPath, claimFileName)

	gomega.RegisterFailHandler(ginkgo.Fail)

	// Extract the version, which should be placed by the build system.
	tnfVersion, err := version.GetVersion()
	if err != nil {
		log.Fatalf("Couldn't determine the version: %v", err)
	}

	// Initialize the claim with the start time.
	claimRoot := createClaimRoot()
	claimData := claimRoot.Claim

	claimData.Configurations = make(map[string]interface{})

	equipmentMap := make(map[string]interface{})

	claimData.Nodes = equipmentMap
	claimData.Versions = &claim.Versions{
		Tnf: tnfVersion.Tag,
	}
	junitFile := filepath.Join(*junitPath, TNFJunitXMLFileName)
	ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)

	endTime := time.Now()
	junitMap := make(map[string]interface{})
	loadJunitXMLIntoMap(junitMap, junitFile, TNFReportKey)

	junitFile = filepath.Join(*junitPath, CNFFeatureValidationJunitXMLFileName)
	if _, err = os.Stat(junitFile); err == nil {
		loadJunitXMLIntoMap(junitMap, junitFile, CNFFeatureValidationReportKey)
	}
	junitMap[extraInfoKey] = tnf.TestsExtraInfo

	claimData.RawResults = junitMap
	resultMap, err := junit.ExtractTestSuiteResults(junitMap, TNFReportKey)
	assert.Nil(t, err)
	claimData.Results = results.GetReconciledResults(resultMap)

	conf := config.GetConfigInstance()
	configurations, err := j.Marshal(conf)
	if err != nil {
		log.Fatalf("error converting configurations to JSON: %v", err)
	}

	claimData.Nodes = diagnostic.GetNodeSummary()

	err = j.Unmarshal(configurations, &claimData.Configurations)
	if err != nil {
		log.Fatalf("error unmarshalling configurations: %v", err)
	}
	claimData.Metadata.EndTime = endTime.UTC().Format(dateTimeFormatDirective)

	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	err = ioutil.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
}
