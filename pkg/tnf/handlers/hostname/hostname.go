package hostname

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"time"
)

// A hostname test implemented using command line tool "hostname".
type Hostname struct {
	result  int
	timeout time.Duration
	args    []string
	// The hostname
	hostname string
}

const (
	Command = "hostname"
	// Anything other than the empty string is considered good output.
	SuccessfulOutputRegex = `.+`
)

// Return the command line args for the test.
func (h *Hostname) Args() []string {
	return h.args
}

// Return the timeout in seconds for the test.
func (h *Hostname) Timeout() time.Duration {
	return h.timeout
}

// Return the test result.
func (h *Hostname) Result() int {
	return h.result
}

// Return a step which expects an ip summary for the given device.
func (h *Hostname) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{SuccessfulOutputRegex},
		Timeout: h.timeout,
	}
}

// On match, parse the hostname output and set the test result.
// Returns no step; the test is complete.
func (h *Hostname) ReelMatch(_ string, _ string, match string) *reel.Step {
	h.hostname = match
	h.result = tnf.SUCCESS
	return nil
}

// On timeout, return a step which kills the ping test by sending it ^C.
func (h *Hostname) ReelTimeout() *reel.Step {
	return nil
}

// On eof, take no action.
func (h *Hostname) ReelEof() {
}

func (h *Hostname) GetHostname() string {
	return h.hostname
}

// Create a new `Ping` test which pings `hosts` with `count` requests, or
// indefinitely if `count` is not positive, and executes within `timeout`
// seconds.
func NewHostname(timeout time.Duration) *Hostname {
	return &Hostname{
		result:  tnf.ERROR,
		timeout: timeout,
		args:    []string{Command},
	}
}
