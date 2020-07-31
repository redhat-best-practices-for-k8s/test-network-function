package tnf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"regexp"
	"strconv"
)

// A oc ping test.
type OcPing struct {
	result  int
	timeout int
	args    []string
}

const ocstat string = `(?m)^\D(\d+) packets transmitted, (\d+) received, (?:\+(\d+) errors)?.*$`
const ocdone string = `\D\d+ packets transmitted.*\r\n(?:rtt )?.*$`

func (ping *OcPing) Args() []string {
	return ping.args
}

func (ping *OcPing) Timeout() int {
	return ping.timeout
}

func (ping *OcPing) Result() int {
	return ping.result
}

// Return a step which expects the ping statistics within the test timeout.
func (ping *OcPing) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{ocdone},
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
func (ping *OcPing) ReelMatch(pattern string, before string, match string) *reel.Step {
	re := regexp.MustCompile(ocstat)
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
func (ping *OcPing) ReelTimeout() *reel.Step {
	return &reel.Step{
		Execute: reel.CTRL_C,
		Expect:  []string{ocdone},
	}
}

// On eof, take no action.
func (ping *OcPing) ReelEof() {
	// empty
}

// Return command line args for pinging `host` with `count` requests, or
// indefinitely if `count` is not positive.
func OcPingCmd(pod, container, host string, opts []string, count int) []string {
	args := []string{"oc", "exec", pod, "-c", container}
	if len(opts) > 0 {
		args = append(args, opts...)
	}
	if count > 0 {
		return append(args, "--", "ping", "-c", strconv.Itoa(count), host)
	} else {
		return append(args, "--", "ping", host)
	}
}

// Create a new `OcPing` test which pings `hosts` with `count` requests, or
// indefinitely if `count` is not positive, and executes within `timeout`
// seconds.
func NewOcPing(timeout int, pod, container, host string, opts []string, count int) *OcPing {
	return &OcPing{
		result:  ERROR,
		timeout: timeout,
		args:    OcPingCmd(pod, container, host, opts, count),
	}
}
