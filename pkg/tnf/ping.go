package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"regexp"
	"strconv"
)

// A ping test implemented using command line tool `ping`.
type Ping struct {
	result  int
	timeout int
	args    []string
}

const stat string = `(?m)^\D(\d+) packets transmitted, (\d+) received, (?:\+(\d+) errors)?.*$`
const done string = `\D\d+ packets transmitted.*\r\n(?:rtt )?.*$`

// Return the command line args for the test.
func (ping *Ping) Args() []string {
	return ping.args
}

// Return the timeout in seconds for the test.
func (ping *Ping) Timeout() int {
	return ping.timeout
}

// Return the test result.
func (ping *Ping) Result() int {
	return ping.result
}

// Return a step which expects the ping statistics within the test timeout.
func (ping *Ping) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{done},
		Timeout: ping.timeout,
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
func (ping *Ping) ReelMatch(pattern string, before string, match string) *reel.Step {
	re := regexp.MustCompile(stat)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		var txd, rxd, ers int
		// Ignore errors in converting matches to decimal integers.
		// Regular expression `stat` is required to underwrite this assumption.
		txd, _ = strconv.Atoi(matched[1])
		rxd, _ = strconv.Atoi(matched[2])
		ers, _ = strconv.Atoi(matched[3])
		switch {
		case txd == 0 || ers > 0:
			ping.result = ERROR
		case rxd > 0 && txd-rxd <= 1:
			ping.result = SUCCESS
		default:
			ping.result = FAILURE
		}
	}
	return nil
}

// On timeout, return a step which kills the ping test by sending it ^C.
func (ping *Ping) ReelTimeout() *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_C,
		Expect:  []string{done},
	}
}

// On eof, take no action.
func (ping *Ping) ReelEof() {
	// empty
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
func NewPing(timeout int, host string, count int) *Ping {
	return &Ping{
		result:  ERROR,
		timeout: timeout,
		args:    PingCmd(host, count),
	}
}
