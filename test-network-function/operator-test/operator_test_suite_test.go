package operator

import (
	"flag"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"
	"path"
	"testing"
)

const (
	defaultCliArgValue            = ""
	CnfCertificationTestSuiteName = "CNF Certification Operator Test Suite"
	junitFlagKey                  = "junit"
	JunitXMLFileName              = "cnf-operator-certification-tests_junit.xml"
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

//TestOperatorTest Entry function to run k8s operator test  cases
func TestOperatorTest(t *testing.T) {
	flag.Parse()
	RegisterFailHandler(Fail)
	var ginkgoReporters []Reporter
	if ginkgoreporters.Polarion.Run {
		ginkgoReporters = append(ginkgoReporters, &ginkgoreporters.Polarion)
	}
	if *junitPath != "" {
		junitFile := path.Join(*junitPath, JunitXMLFileName)
		ginkgoReporters = append(ginkgoReporters, reporters.NewJUnitReporter(junitFile))
	}
	RunSpecsWithDefaultAndCustomReporters(t, CnfCertificationTestSuiteName, ginkgoReporters)
}
