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

package suite

import (
	j "encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/test-network-function/test-network-function/test-network-function/diagnostic"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/junit"
	"github.com/test-network-function/test-network-function/pkg/tnf"

	utils "github.com/test-network-function/test-network-function/pkg/utils"
	_ "github.com/test-network-function/test-network-function/test-network-function/accesscontrol"
	_ "github.com/test-network-function/test-network-function/test-network-function/certification"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	_ "github.com/test-network-function/test-network-function/test-network-function/generic"
	_ "github.com/test-network-function/test-network-function/test-network-function/lifecycle"
	_ "github.com/test-network-function/test-network-function/test-network-function/networking"
	_ "github.com/test-network-function/test-network-function/test-network-function/observability"
	_ "github.com/test-network-function/test-network-function/test-network-function/operator"
	_ "github.com/test-network-function/test-network-function/test-network-function/platform"
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
	claimPath      *string
	junitPath      *string
	diagnosticFlag *bool
	// GitCommit is the latest commit in the current git branch
	GitCommit string
	// GitRelease is the list of tags (if any) applied to the latest commit
	// in the current branch
	GitRelease string
	// GitPreviousRelease is the last release at the date of the latest commit
	// in the current branch
	GitPreviousRelease string
	// gitDisplayRelease is a string used to hold the text to display
	// the version on screen and in the claim file
	gitDisplayRelease string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")

	// diagnosticFlag has precedence over the ginkgo -f (focus) and it
	// should be set by launcher scripts only in case no focus test suites were provided.
	diagnosticFlag = flag.Bool("diagnostic", false, "launch diagnostic mode only")
}

func isDiagnosticMode() bool {
	return *diagnosticFlag
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

//nolint:funlen // TestTest invokes the CNF Certification Test Suite.
func TestTest(t *testing.T) {
	// Set up input flags and register failure handlers.
	flag.Parse()

	// Checking if output directories exist
	utils.CheckFileExists(*claimPath, "claim")
	utils.CheckFileExists(*junitPath, "junit")

	gomega.RegisterFailHandler(ginkgo.Fail)
	common.SetLogFormat()
	common.SetLogLevel()
	if common.LogLevelTraceEnabled {
		config.EnableExpectersVerboseMode()
	}
	// Display GinkGo Version
	log.Info("Ginkgo Version: ", ginkgo.GINKGO_VERSION)
	// Display the latest previously released build in case this build is not released
	// Otherwise display the build version
	if GitRelease == "" {
		gitDisplayRelease = "Unreleased build post " + GitPreviousRelease
	} else {
		gitDisplayRelease = GitRelease
	}
	log.Info("Version: ", gitDisplayRelease, " ( ", GitCommit, " )")

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot := createClaimRoot()
	claimData := claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})

	if isDiagnosticMode() {
		log.Warn("No test suites selected to run. Diagnostic mode enabled.")
		// In diagnostic mode, we need to remove labels explicitly before exiting tnf.
		defer common.RemoveLabelsFromAllNodes()
	}

	// Make sure cluster nodes don't have the debug pod label from previous runs,
	// which might happen in case of some aborted/failed tnf runs.
	common.RemoveLabelsFromAllNodes()

	// Run first autodiscovery.
	config.GetTestEnvironment().LoadAndRefresh()

	// Collect diagnostic data
	errs := diagnostic.GetDiagnosticData()
	if len(errs) > 0 {
		// Should we abort here?
		log.Errorf("Errors found while getting diagnostic information from cluster: %v", errs)
	}

	incorporateVersions(claimData)

	configurations := marshalConfigurations()
	claimData.Nodes = generateNodes()
	unmarshalConfigurations(configurations, claimData.Configurations)

	// Run tests specs only if not in diagnostic mode.
	if !isDiagnosticMode() {
		// Run the test suite/s
		ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)
		// Process the test results from the suites, the cnf-features-deploy test suite,
		// and any extra informational messages.
		junitMap := make(map[string]interface{})
		cnfCertificationJUnitFilename := filepath.Join(*junitPath, TNFJunitXMLFileName)
		loadJUnitXMLIntoMap(junitMap, cnfCertificationJUnitFilename, TNFReportKey)
		appendCNFFeatureValidationReportResults(junitPath, junitMap)
		junitMap[extraInfoKey] = tnf.TestsExtraInfo

		// Append results to claim file data.
		claimData.RawResults = junitMap
		claimData.Results = results.GetReconciledResults()
	}

	endTime := time.Now()
	claimData.Metadata.EndTime = endTime.UTC().Format(dateTimeFormatDirective)

	// Marshal the claim and output to file
	payload := marshalClaimOutput(claimRoot)
	claimOutputFile := filepath.Join(*claimPath, claimFileName)
	writeClaimOutput(claimOutputFile, payload)
}

// incorporateTNFVersion adds the TNF version to the claim.
func incorporateVersions(claimData *claim.Claim) {
	claimData.Versions = &claim.Versions{
		Tnf:          gitDisplayRelease,
		TnfGitCommit: GitCommit,
		OcClient:     diagnostic.GetVersionsOcp().Oc,
		Ocp:          diagnostic.GetVersionsOcp().Ocp,
		K8s:          diagnostic.GetVersionsOcp().K8s,
	}
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
	configurations, err := j.Marshal(config.GetTestEnvironment().Config)
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
	err := os.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
}

func generateNodes() map[string]interface{} {
	const (
		nodeSummaryField = "nodeSummary"
		cniPluginsField  = "cniPlugins"
		nodesHwInfo      = "nodesHwInfo"
		csiDriverInfo    = "csiDriver"
	)
	nodes := map[string]interface{}{}
	nodes[nodeSummaryField] = diagnostic.GetNodeSummary()
	nodes[cniPluginsField] = diagnostic.GetCniPlugins()
	nodes[nodesHwInfo] = diagnostic.GetNodesHwInfo()
	nodes[csiDriverInfo] = diagnostic.GetCsiDriverInfo()
	return nodes
}
