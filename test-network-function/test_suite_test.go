package cnftests_test

import (
	"flag"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	ginkgoreporters "kubevirt.io/qe-tools/pkg/ginkgo-reporters"
	"path"
	"testing"
)

const (
	JunitXmlFileName = "test-network-function_junit.xml"
)

var junitPath *string
var reportPath *string

func init() {
	junitPath = flag.String("junit", "", "the path for the junit format report")
	reportPath = flag.String("report", "", "the path of the report file containing details for failed tests")
}

// Invokes all Ginkgo Run Specs.  For now, this includes just generic_cnf_tests.go.
func TestTest(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)

	var rr []ginkgo.Reporter
	if ginkgoreporters.Polarion.Run {
		rr = append(rr, &ginkgoreporters.Polarion)
	}

	if *junitPath != "" {
		junitFile := path.Join(*junitPath, JunitXmlFileName)
		rr = append(rr, reporters.NewJUnitReporter(junitFile))
	}

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "setup tests", rr)
}
