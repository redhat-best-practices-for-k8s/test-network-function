package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
)

// A SSH session test implemented using command line tool `ssh`.
type Ssh struct {
	result  int
	timeout int
	prompt  string
	args    []string
}

const areyousure = `Are you sure you want to continue connecting \(yes/no\)\?`
const yesorno = `Please type 'yes' or 'no': `
const closed = `Connection to .+ closed\..*$`

// Return the command line args for the test.
func (ssh *Ssh) Args() []string {
	return ssh.args
}

// Return the timeout in seconds for the test.
func (ssh *Ssh) Timeout() int {
	return ssh.timeout
}

// Return the test result.
func (ssh *Ssh) Result() int {
	return ssh.result
}

// Return a step which expects a SSH prompt.
func (ssh *Ssh) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{areyousure, yesorno, ssh.prompt},
		Timeout: ssh.timeout,
	}
}

// On match, if the session closed cleanly then set the test result to success.
// Otherwise, return a step which closes the session by sending it ^D.
func (ssh *Ssh) ReelMatch(pattern string, before string, match string) *reel.Step {
	if pattern == closed {
		ssh.result = SUCCESS
		return nil
	}
	return &reel.Step{
		Execute: reel.CTRL_D,
		Expect:  []string{closed},
		Timeout: ssh.timeout,
	}
}

// On timeout, return a step which closes the session by sending it ^D.
func (ssh *Ssh) ReelTimeout() *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_D,
		Expect:  []string{closed},
		Timeout: ssh.timeout,
	}
}

// On eof, take no action.
func (ssh *Ssh) ReelEof() {
	// empty
}

// Return command line args for establishing a SSH session with `host` using
// ssh command line options `sshopts`.
func SshCmd(host string, sshopts []string) []string {
	args := []string{"ssh"}
	if len(sshopts) > 0 {
		args = append(args, sshopts...)
		args = append(args, "--")
	}
	return append(args, host)
}

// Create a new `Ssh` test session with `host` using ssh command line options
// `sshopts`, expecting `prompt` string and requires steps to execute in
// `timeout` seconds.
func NewSsh(timeout int, prompt string, host string, sshopts []string) *Ssh {
	return &Ssh{
		result:  ERROR,
		timeout: timeout,
		prompt:  prompt,
		args:    SshCmd(host, sshopts),
	}
}
