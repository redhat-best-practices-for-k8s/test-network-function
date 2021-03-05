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

package suite

import (
	j "encoding/json"
	"flag"
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function-claim/pkg/claim"
	"github.com/redhat-nfvpe/test-network-function/pkg/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/junit"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/container"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/diagnostic"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/generic"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/operator"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/turnium"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/version"
	log "github.com/sirupsen/logrus"
	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"
)

const (
	claimFileName                 = "claim.json"
	claimFilePermissions          = 0644
	claimPathFlagKey              = "claimloc"
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultClaimPath              = ".."
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	JunitXMLFileName              = "cnf-certification-tests_junit.xml"
	reportFlagKey                 = "report"
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	dateTimeFormatDirective = "2006-01-02T15:04:05+00:00"
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

var _ = ginkgo.BeforeSuite(func() {
	tnfConfig, cfgError := config.GetConfig()
	if cfgError != nil || tnfConfig.CNFs == nil {
		ginkgo.Fail("Unable to load the configuration required for the test. Test aborted")
	}
})

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

// Invokes the CNF Certification Tests.
//nolint:funlen // Function is long but core entrypoint and linear. Consider revisiting later.
func TestTest(t *testing.T) {
	flag.Parse()
	claimOutputFile := path.Join(*claimPath, claimFileName)

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

	var ginkgoReporters []ginkgo.Reporter
	if ginkgoreporters.Polarion.Run {
		ginkgoReporters = append(ginkgoReporters, &ginkgoreporters.Polarion)
	}

	if *junitPath != "" {
		junitFile := path.Join(*junitPath, JunitXMLFileName)
		ginkgoReporters = append(ginkgoReporters, reporters.NewJUnitReporter(junitFile))
	}

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, CnfCertificationTestSuiteName, ginkgoReporters)
	junitMap := make(map[string]interface{})
	var extension = filepath.Ext(JunitXMLFileName)
	reportKeyName := JunitXMLFileName[0 : len(JunitXMLFileName)-len(extension)]
	junitMap[reportKeyName], err = junit.ExportJUnitAsJSON(JunitXMLFileName)
	if err != nil {
		log.Fatalf("error reading JUnit XML file into JSON: %v", err)
	}

	endTime := time.Now()
	claimData.Results = junitMap

	configurations, err := j.Marshal(config.GetInstance().GetConfigurations())
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
