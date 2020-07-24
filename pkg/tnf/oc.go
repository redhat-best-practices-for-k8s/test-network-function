package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
)

// An OpenShift test implemented using command line `oc exec -it ... -- sh`.
type Oc struct {
	result  int
	timeout int
	args    []string
}

const OcPrompt string = "# "

// Return the command line args for the test.
func (oc *Oc) Args() []string {
	return oc.args
}

// Return the timeout in seconds for the test.
func (oc *Oc) Timeout() int {
	return oc.timeout
}

// Return the test result.
func (oc *Oc) Result() int {
	return oc.result
}

// Return a step which expects a shell prompt.
func (oc *Oc) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{OcPrompt},
		Timeout: oc.timeout,
	}
}

// On match, return a step which closes the session by sending it ^D.
func (oc *Oc) ReelMatch(pattern string, before string, match string) *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_D,
		Expect:  []string{OcPrompt},
		Timeout: oc.timeout,
	}
}

// On timeout, return a step which closes the session by sending it ^D.
func (oc *Oc) ReelTimeout() *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_D,
		Expect:  []string{OcPrompt},
		Timeout: oc.timeout,
	}
}

// On eof, set the test result to success.
func (oc *Oc) ReelEof() {
	oc.result = SUCCESS
}

// Return command line args for establishing an OpenShift shell using oc exec
// command line options `opts`.
func OcCmd(pod string, opts []string) []string {
	args := []string{"oc", "exec", "-it"}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	return append(args, pod, "--", "sh")
}

// Create a new `Oc` test session with `pod` using oc command line options `opts`
// and requiring steps to execute within `timeout` seconds.
func NewOc(timeout int, pod string, opts []string) *Oc {
	return &Oc{
		result:  ERROR,
		timeout: timeout,
		args:    OcCmd(pod, opts),
	}
}
