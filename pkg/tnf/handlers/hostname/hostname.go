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

package hostname

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// Hostname provides a hostname test implemented using command line tool "hostname".
type Hostname struct {
	result  int
	timeout time.Duration
	args    []string
	// The hostname
	hostname string
}

const (
	// Command is the command name for the unix "hostname" command.
	Command = dependencies.HostnameBinaryName
	// SuccessfulOutputRegex is the regular expression match for hostname output.
	SuccessfulOutputRegex = `.+`
)

// Args returns the command line args for the test.
func (h *Hostname) Args() []string {
	return h.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (h *Hostname) GetIdentifier() identifier.Identifier {
	return identifier.HostnameIdentifier
}

// Timeout return the timeout for the test.
func (h *Hostname) Timeout() time.Duration {
	return h.timeout
}

// Result returns the test result.
func (h *Hostname) Result() int {
	return h.result
}

// ReelFirst returns a step which expects an hostname summary for the given device.
func (h *Hostname) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{SuccessfulOutputRegex},
		Timeout: h.timeout,
	}
}

// ReelMatch parses the hostname output and set the test result on match.
// Returns no step; the test is complete.
func (h *Hostname) ReelMatch(_, _, match string) *reel.Step {
	h.hostname = match
	h.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  hostname requires no explicit intervention for a timeout.
func (h *Hostname) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  hostname requires no explicit intervention for EOF.
func (h *Hostname) ReelEOF() {
}

// GetHostname returns the extracted hostname, if one is extracted.
func (h *Hostname) GetHostname() string {
	return h.hostname
}

// NewHostname creates a new `Hostname` test which runs the "hostname" command.
func NewHostname(timeout time.Duration) *Hostname {
	return &Hostname{
		result:  tnf.ERROR,
		timeout: timeout,
		args:    []string{Command},
	}
}
