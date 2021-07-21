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

	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/junit"
	"github.com/test-network-function/test-network-function/pkg/tnf"

	_ "github.com/test-network-function/test-network-function/test-network-function/access-control"
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
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	dateTimeFormatDirective = "2006-01-02T15:04:05+00:00"
	extraInfoKey            = "testsExtraInfo"
)

var (
	claimPath *string
	junitPath *string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
}

// createClaimRoot creates the claim based on the model created in
// https://github.com/test-network-function/test-network-function-claim.
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

// loadJUnitXMLIntoMap converts junitFilename's XML-formatted JUnit test results into a Go map, and adds the result to
// the result Map.
func loadJUnitXMLIntoMap(result map[string]interface{}, junitFilename, key string) {
	var err error
	if key == "" {
		var extension = filepath.Ext(junitFilename)
		key = junitFilename[0 : len(junitFilename)-len(extension)]
	}
	result[key], err = junit.ExportJUnitAsMap(junitFilename)
	if err != nil {
		log.Fatalf("error reading JUnit XML file into JSON: %v", err)
	}
}

// TestTest invokes the CNF Certification Test Suite.
func TestTest(t *testing.T) {
	// set up input flags and register failure handlers.
	flag.Parse()
	gomega.RegisterFailHandler(ginkgo.Fail)

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot := createClaimRoot()
	claimData := claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})
	incorporateTNFVersion(claimData)

	// run the test suite
	ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)
	endTime := time.Now()

	// process the test results from this test suite, the cnf-features-deploy test suite, and any extra informational
	// messages.
	junitMap := make(map[string]interface{})
	cnfCertificationJUnitFilename := filepath.Join(*junitPath, TNFJunitXMLFileName)
	loadJUnitXMLIntoMap(junitMap, cnfCertificationJUnitFilename, TNFReportKey)
	appendCNFFeatureValidationReportResults(junitPath, junitMap)
	junitMap[extraInfoKey] = tnf.TestsExtraInfo

	// fill out the remaining claim information.
	claimData.RawResults = junitMap
	resultMap := generateResultMap(junitMap)
	claimData.Results = results.GetReconciledResults(resultMap)
	configurations := marshalConfigurations()
	claimData.Nodes = generateNodes()
	unmarshalConfigurations(configurations, claimData.Configurations)
	claimData.Metadata.EndTime = endTime.UTC().Format(dateTimeFormatDirective)

	// marshal the claim and output to file
	payload := marshalClaimOutput(claimRoot)
	claimOutputFile := filepath.Join(*claimPath, claimFileName)
	writeClaimOutput(claimOutputFile, payload)
}

// getTNFVersion gets the TNF version, or fatally fails.
func getTNFVersion() *version.Version {
	// Extract the version, which should be placed by the build system.
	tnfVersion, err := version.GetVersion()
	if err != nil {
		log.Fatalf("Couldn't determine the version: %v", err)
	}
	return tnfVersion
}

// incorporateTNFVersion adds the TNF version to the claim.
func incorporateTNFVersion(claimData *claim.Claim) {
	claimData.Versions = &claim.Versions{
		Tnf: getTNFVersion().Tag,
	}
}

// generateResultMap is a conversion utility to generate results.  If an error is encountered, than this method fails
// fatally.
func generateResultMap(junitMap map[string]interface{}) map[string]junit.TestResult {
	resultMap, err := junit.ExtractTestSuiteResults(junitMap, TNFReportKey)
	if err != nil {
		log.Fatalf("Could not extract the test suite results: %s", err)
	}
	return resultMap
}

// appendCNFFeatureValidationReportResults is a helper method to add the results of running the cnf-features-deploy
// test suite to the claim file.
func appendCNFFeatureValidationReportResults(junitPath *string, junitMap map[string]interface{}) {
	cnfFeaturesDeployJUnitFile := filepath.Join(*junitPath, CNFFeatureValidationJunitXMLFileName)
	if _, err := os.Stat(cnfFeaturesDeployJUnitFile); err == nil {
		loadJUnitXMLIntoMap(junitMap, cnfFeaturesDeployJUnitFile, CNFFeatureValidationReportKey)
	}
}

// marshalConfigurations creates a byte stream representation of the test configurations.  In the event of an error,
// this method fatally fails.
func marshalConfigurations() []byte {
	configurations, err := j.Marshal(config.GetConfigInstance())
	if err != nil {
		log.Fatalf("error converting configurations to JSON: %v", err)
	}
	return configurations
}

// unmarshalConfigurations creates a map from configurations byte stream.  In the event of an error, this method fatally
// fails.
func unmarshalConfigurations(configurations []byte, claimConfigurations map[string]interface{}) {
	err := j.Unmarshal(configurations, &claimConfigurations)
	if err != nil {
		log.Fatalf("error unmarshalling configurations: %v", err)
	}
}

// marshalClaimOutput is a helper function to serialize a claim as JSON for output.  In the event of an error, this
// method fatally fails.
func marshalClaimOutput(claimRoot *claim.Root) []byte {
	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	return payload
}

// writeClaimOutput writes the output payload to the claim file.  In the event of an error, this method fatally fails.
func writeClaimOutput(claimOutputFile string, payload []byte) {
	err := ioutil.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
}

func generateNodes() map[string]interface{} {
	const (
		nodeSummaryField = "nodeSummary"
		cniPluginsField  = "cniPlugins"
		nodesHwInfo      = "nodesHwInfo"
	)
	nodes := map[string]interface{}{}
	nodes[nodeSummaryField] = diagnostic.GetNodeSummary()
	nodes[cniPluginsField] = diagnostic.GetCniPlugins()
	nodes[nodesHwInfo] = diagnostic.GetNodesHwInfo()
	return nodes
}
