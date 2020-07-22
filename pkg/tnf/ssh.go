package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
)

type Ssh struct {
	result  int
	timeout int
	prompt  string
	args    []string
}

const areyousure = `Are you sure you want to continue connecting \(yes/no\)\?`
const yesorno = `Please type 'yes' or 'no': `
const closed = `Connection to .+ closed\..*$`

func (ssh *Ssh) Args() []string {
	return ssh.args
}
func (ssh *Ssh) Timeout() int {
	return ssh.timeout
}
func (ssh *Ssh) Result() int {
	return ssh.result
}
func (ssh *Ssh) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{areyousure, yesorno, ssh.prompt},
		Timeout: ssh.timeout,
	}
}
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
func (ssh *Ssh) ReelTimeout() *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_D,
		Expect:  []string{closed},
		Timeout: ssh.timeout,
	}
}
func (ssh *Ssh) ReelEof() *reel.Step {
	return nil
}

func SshCmd(host string, sshopts []string) []string {
	args := []string{"ssh"}
	if len(sshopts) > 0 {
		args = append(args, sshopts...)
		args = append(args, "--")
	}
	return append(args, host)
}

func NewSsh(timeout int, prompt string, host string, sshopts []string) *Ssh {
	return &Ssh{
		result:  ERROR,
		timeout: timeout,
		prompt:  prompt,
		args:    SshCmd(host, sshopts),
	}
}
