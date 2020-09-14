package ping

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"regexp"
	"strconv"
	"time"
)

// A ping test implemented using command line tool `ping`.
type Ping struct {
	result      int
	timeout     time.Duration
	args        []string
	transmitted int
	received    int
	errors      int
}

const (
	ConnectInvalidArgument = `(?m)connect: Invalid argument$`
	SuccessfulOutputRegex  = `(?m)(\d+) packets transmitted, (\d+)( packets){0,1} received, (?:\+(\d+) errors)?.*$`
)

// Return the command line args for the test.
func (p *Ping) Args() []string {
	return p.args
}

// Return the timeout in seconds for the test.
func (p *Ping) Timeout() time.Duration {
	return p.timeout
}

// Return the test result.
func (p *Ping) Result() int {
	return p.result
}

// Return a step which expects the ping statistics within the test timeout.
func (p *Ping) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  p.GetReelFirstRegularExpressions(),
		Timeout: p.timeout,
	}
}

// On match, parse the ping statistics and set the test result.
// The result is success if at least one response was received and the number of
// responses received is at most one less than the number received (the "missing"
// response may be in flight).
// The result is error if ping reported a protocol error (e.g. destination host
// unreachable), no requests were sent or there was some test execution error.
// Otherwise the result is failure.
// Returns no step; the test is complete.
func (p *Ping) ReelMatch(_ string, _ string, match string) *reel.Step {
	re := regexp.MustCompile(ConnectInvalidArgument)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		p.result = tnf.ERROR
	}
	re = regexp.MustCompile(SuccessfulOutputRegex)
	matched = re.FindStringSubmatch(match)
	if matched != nil {
		// Ignore errors in converting matches to decimal integers.
		// Regular expression `stat` is required to underwrite this assumption.
		p.transmitted, _ = strconv.Atoi(matched[1])
		p.received, _ = strconv.Atoi(matched[2])
		p.errors, _ = strconv.Atoi(matched[4])
		switch {
		case p.transmitted == 0 || p.errors > 0:
			p.result = tnf.ERROR
		case p.received > 0 && (p.transmitted-p.received) <= 1:
			p.result = tnf.SUCCESS
		default:
			p.result = tnf.FAILURE
		}
	}
	return nil
}

// On timeout, return a step which kills the ping test by sending it ^C.
func (p *Ping) ReelTimeout() *reel.Step {
	return &reel.Step{Execute: reel.CtrlC}
}

// On eof, take no action.
func (p *Ping) ReelEof() {
	// empty
}

func (p *Ping) GetStats() (int, int, int) {
	return p.transmitted, p.received, p.errors
}

// Return command line args for pinging `host` with `count` requests, or
// indefinitely if `count` is not positive.
func PingCmd(host string, count int) []string {
	if count > 0 {
		return []string{"ping", "-c", strconv.Itoa(count), host}
	} else {
		return []string{"ping", host}
	}
}

// Create a new `Ping` test which pings `hosts` with `count` requests, or
// indefinitely if `count` is not positive, and executes within `timeout`
// seconds.
func NewPing(timeout time.Duration, host string, count int) *Ping {
	return &Ping{
		result:  tnf.ERROR,
		timeout: timeout,
		args:    PingCmd(host, count),
	}
}

// Utility method to get the regular expressions used for matching in ReelFirst(...).
func (p *Ping) GetReelFirstRegularExpressions() []string {
	return []string{ConnectInvalidArgument, SuccessfulOutputRegex}
}
