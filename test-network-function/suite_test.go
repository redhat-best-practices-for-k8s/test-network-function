package suite

import (
	j "encoding/json"
	"flag"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function-claim/pkg/claim"
	"github.com/redhat-nfvpe/test-network-function/pkg/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/junit"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/casa/cnf"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/cisco/kiknos"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/generic"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/generic"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/version"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"
	"path"
	"testing"
	"time"
)

const (
	claimFilePermissions          = 0644
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	JunitXMLFileName              = "cnf-certification-tests_junit.xml"
	reportFlagKey                 = "report"
)

var (
	claimDefaultOutputFile = path.Join("..", "claim.json")
	junitPath              *string
	reportPath             *string
)

func init() {
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
	reportPath = flag.String(reportFlagKey, defaultCliArgValue,
		"the path of the report file containing details for failed tests")
}

func createClaimRoot() *claim.Root {
	// Initialize the claim with the start time.
	startTime := time.Now()
	c := &claim.Claim{
		StartTime: startTime.String(),
	}
	return &claim.Root{
		Claim: c,
	}
}

// Invokes the CNF Certification Tests.
func TestTest(t *testing.T) {
	flag.Parse()
	gomega.RegisterFailHandler(ginkgo.Fail)

	// Extract the version, which should be placed by the build system.
	version, err := version.GetVersion()
	if err != nil {
		log.Fatalf("Couldn't determine the version: %v", err)
	}

	// Initialize the claim with the start time.
	claimRoot := createClaimRoot()
	claimData := claimRoot.Claim

	claimData.TestConfigurations = make(map[string]interface{})

	equipmentMap := make(map[string]interface{})
	for _, key := range generic.GetTestConfiguration().Hosts {
		// For now, just initialize the payload as empty.
		equipmentMap[key] = make(map[string]interface{})
	}
	claimData.Hosts = &claim.Hosts{LshwOutput: equipmentMap}
	claimData.TnfVersion = version.Tag

	var ginkgoReporters []ginkgo.Reporter
	if ginkgoreporters.Polarion.Run {
		ginkgoReporters = append(ginkgoReporters, &ginkgoreporters.Polarion)
	}

	if *junitPath != "" {
		junitFile := path.Join(*junitPath, JunitXMLFileName)
		ginkgoReporters = append(ginkgoReporters, reporters.NewJUnitReporter(junitFile))
	}

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, CnfCertificationTestSuiteName, ginkgoReporters)
	junitMap, err := junit.ExportJUnitAsJSON(JunitXMLFileName)

	endTime := time.Now()
	claimData.JunitResults = junitMap

	configurations, err := j.Marshal(config.GetInstance().GetConfigurations())
	if err != nil {
		log.Fatalf("error converting configurations to JSON: %v", err)
	}
	err = j.Unmarshal(configurations, &claimData.TestConfigurations)
	claimData.EndTime = endTime.String()

	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	err = ioutil.WriteFile(claimDefaultOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
}
