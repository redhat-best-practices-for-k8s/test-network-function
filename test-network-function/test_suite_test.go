package test_network_function

import (
	"flag"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/casa/cnf"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/cisco/kiknos"
	_ "github.com/redhat-nfvpe/test-network-function/test-network-function/generic"
	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"
	"path"
	"testing"
)

const (
	defaultCliArgValue            = ""
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	junitFlagKey                  = "junit"
	JunitXmlFileName              = "cnf-certification-tests_junit.xml"
	reportFlagKey                 = "report"
)

var junitPath *string
var reportPath *string

func init() {
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
	reportPath = flag.String(reportFlagKey, defaultCliArgValue,
		"the path of the report file containing details for failed tests")
}

// Invokes the CNF Certification Tests.
func TestTest(t *testing.T) {
	flag.Parse()
	gomega.RegisterFailHandler(ginkgo.Fail)

	var ginkgoReporters []ginkgo.Reporter
	if ginkgoreporters.Polarion.Run {
		ginkgoReporters = append(ginkgoReporters, &ginkgoreporters.Polarion)
	}

	if *junitPath != "" {
		junitFile := path.Join(*junitPath, JunitXmlFileName)
		ginkgoReporters = append(ginkgoReporters, reporters.NewJUnitReporter(junitFile))
	}

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, CnfCertificationTestSuiteName, ginkgoReporters)
}
