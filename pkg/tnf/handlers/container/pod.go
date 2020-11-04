package container

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container/testcases"
	"regexp"
	"strings"
	"time"
)

//Pod that is under test.
type Pod struct {
	result       int
	timeout      time.Duration
	args         []string
	status       string
	Command      string
	Name         string
	Namespace    string
	ExpectStatus []string
	Action       string
	ResultType   string
	FailOn       string
}

// Args returns the command line args for the test.
func (p *Pod) Args() []string {
	return p.args
}

// Timeout return the timeout for the test.
func (p *Pod) Timeout() time.Duration {
	return p.timeout
}

// Result returns the test result.
func (p *Pod) Result() int {
	return p.result
}

// ReelFirst returns a step which expects an pod status for the given pod.
func (p *Pod) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{testcases.GetOutRegExp("ALLOW_ALL")},
		Timeout: p.timeout,
	}
}

func contains(arr []string, str string) (found bool) {
	found = false
	for _, a := range arr {
		if a == str {
			found = true
			break
		}
	}
	return
}

// ReelMatch parses the pod status output and set the test result on match.
// Returns no step; the test is complete.
func (p *Pod) ReelMatch(_ string, _ string, match string) *reel.Step {
	//for type: array ,should match for any expected status or fail on any expected status
	//based on the action type allow (default)|deny
	if p.ResultType == "array" {
		re := regexp.MustCompile(testcases.GetOutRegExp("NULL")) //Not having capabilities is positive
		matched := re.MatchString(match)
		if matched {
			p.result = tnf.SUCCESS
			return nil
		}
		replacer := strings.NewReplacer(`[`, ``, "\"", ``, `]`, ``, `, `, `,`)
		match = replacer.Replace(match)
		matchSlice := strings.Split(match, ",")
		for _, status := range matchSlice {
			if contains(p.ExpectStatus, status) {
				if p.Action == "deny" { //Single deny match is failure.
					return nil
				}
			} else if p.Action == "allow" {
				return nil //should be in allowed list
			}
		}
	} else {
		for _, status := range p.ExpectStatus {
			re := regexp.MustCompile(testcases.GetOutRegExp(status))
			matched := re.MatchString(match)
			if !matched {
				return nil
			}
		}
	}

	p.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;
func (p *Pod) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing.
func (p *Pod) ReelEOF() {
}

// NewPod creates a `Container` test  on the configured test cases.
func NewPod(command, name, namespace string, expectedStatus []string, resultType string, action string, timeout time.Duration) *Pod {
	args := strings.Split(fmt.Sprintf(command, name, namespace), " ")
	return &Pod{
		Name:         name,
		Namespace:    namespace,
		ExpectStatus: expectedStatus,
		Action:       action,
		ResultType:   resultType,
		result:       tnf.ERROR,
		timeout:      timeout,
		args:         args,
	}
}
