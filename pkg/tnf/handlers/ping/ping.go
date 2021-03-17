// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package ping

import (
	"regexp"
	"strconv"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// Ping provides a ping test implemented using command line tool `ping`.
type Ping struct {
	result      int
	timeout     time.Duration
	args        []string
	transmitted int
	received    int
	errors      int
}

const (
	// ConnectInvalidArgumentRegex is a regex which matches when an invalid IP address or hostname is provided as input.
	ConnectInvalidArgumentRegex = `(?m)connect: Invalid argument$`
	// SuccessfulOutputRegex matches a successfully run "ping" command.  That does not mean that no errors or drops
	// occurred during the test.
	SuccessfulOutputRegex = `(?m)(\d+) packets transmitted, (\d+)( packets){0,1} received, (?:\+(\d+) errors)?.*$`
)

// Args returns the command line args for the test.
func (p *Ping) Args() []string {
	return p.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (p *Ping) GetIdentifier() identifier.Identifier {
	return identifier.PingIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (p *Ping) Timeout() time.Duration {
	return p.timeout
}

// Result returns the test result.
func (p *Ping) Result() int {
	return p.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (p *Ping) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  p.GetReelFirstRegularExpressions(),
		Timeout: p.timeout,
	}
}

// ReelMatch parses the ping statistics and set the test result on match.
// The result is success if at least one response was received and the number of
// responses received is at most one less than the number received (the "missing"
// response may be in flight).
// The result is error if ping reported a protocol error (e.g. destination host
// unreachable), no requests were sent or there was some test execution error.
// Otherwise the result is failure.
// Returns no step; the test is complete.
func (p *Ping) ReelMatch(_, _, match string) *reel.Step {
	re := regexp.MustCompile(ConnectInvalidArgumentRegex)
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

// ReelTimeout returns a step which kills the ping test by sending it ^C.
func (p *Ping) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  ping requires no intervention on eof.
func (p *Ping) ReelEOF() {
}

// GetStats returns the transmitted, received and error counts.
func (p *Ping) GetStats() (transmitted, received, errors int) {
	return p.transmitted, p.received, p.errors
}

// Command returns command line args for pinging `host` with `count` requests, or indefinitely if `count` is not
// positive.
func Command(host string, count int) []string {
	if count > 0 {
		return []string{dependencies.PingBinaryName, "-c", strconv.Itoa(count), host}
	}
	return []string{dependencies.PingBinaryName, host}
}

// NewPing creates a new `Ping` test which pings `hosts` with `count` requests, or indefinitely if `count` is not
// positive, and executes within `timeout` seconds.
func NewPing(timeout time.Duration, host string, count int) *Ping {
	return &Ping{
		result:  tnf.ERROR,
		timeout: timeout,
		args:    Command(host, count),
	}
}

// GetReelFirstRegularExpressions returns the regular expressions used for matching in ReelFirst.
func (p *Ping) GetReelFirstRegularExpressions() []string {
	return []string{ConnectInvalidArgumentRegex, SuccessfulOutputRegex}
}
