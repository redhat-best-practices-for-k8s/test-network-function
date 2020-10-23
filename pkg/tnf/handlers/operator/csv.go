package operator

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"regexp"
	"strings"
	"time"
)

const (
	// CheckCSVCommand is the OC command for checking for CSV.
	CheckCSVCommand = "oc get csv %s -n %s -o json | jq -r '.status.phase'"
)

//Csv Cluster service version , manifests of the operator.
type Csv struct {
	result       int
	timeout      time.Duration
	args         []string
	Name         string
	Status       string
	Namespace    string
	ExpectStatus string
}

// Args returns the command line args for the test.
func (c *Csv) Args() []string {
	return c.args
}

// Timeout return the timeout for the test.
func (c *Csv) Timeout() time.Duration {
	return c.timeout
}

// Result returns the test result.
func (c *Csv) Result() int {
	return c.result
}

// ReelFirst returns a step which expects an csv status for the given csv.
func (c *Csv) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{c.ExpectStatus},
		Timeout: c.timeout,
	}
}

// ReelMatch parses the csv status output and set the test result on match.
// Returns no step; the test is complete.
func (c *Csv) ReelMatch(_ string, _ string, match string) *reel.Step {
	re := regexp.MustCompile(c.ExpectStatus)
	matched := re.MatchString(match)
	if matched {
		c.result = tnf.SUCCESS
	}
	return nil
}

// ReelTimeout does nothing;
func (c *Csv) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing.
func (c *Csv) ReelEOF() {
}

// NewCsv creates a `Operator Csv` test which determines the "csv" status.
func NewCsv(name, namespace string, expectedStatus string, timeout time.Duration) *Csv {
	args := strings.Split(fmt.Sprintf(CheckCSVCommand, name, namespace), " ")
	return &Csv{
		Name:         name,
		Namespace:    namespace,
		ExpectStatus: expectedStatus,
		result:       tnf.ERROR,
		timeout:      timeout,
		args:         args,
	}
}
