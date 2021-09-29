// Copyright (C) 2020-2021 Red Hat, Inc.
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

package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// Command provides a Command test implemented using command line tool Command.
type Command struct {
	result  int
	timeout time.Duration
	args    []string
	Output  string
	// adding special parameters
}

// NewCommand creates a new command handler.
func NewCommand(timeout time.Duration, resourceType, labelQuery, namespace, ocCommand string) *Command {
	command := fmt.Sprintf(ocCommand, resourceType, namespace, labelQuery)
	return &Command{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    strings.Fields(command),
	}
}

// Args function
func (h *Command) Args() []string {
	return h.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (h *Command) GetIdentifier() identifier.Identifier {
	return identifier.CommandIdentifier
	// Create identifier CommandIdentifier.
}

// Timeout return the timeout for the test.
func (h *Command) Timeout() time.Duration {
	return h.timeout
}

// Result returns the test result.
func (h *Command) Result() int {
	return h.result
}

// ReelFirst returns a step which expects an Command summary for the given device.
func (h *Command) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{".*"}, // TODO : pass the list of possible regex in here
		Timeout: h.timeout,
	}
}

// ReelMatch parses the Command output and set the test result on match.
func (h *Command) ReelMatch(_, _, match string) *reel.Step {
	h.result = tnf.SUCCESS
	h.Output = match

	return nil
}

// ReelTimeout does nothing, Command requires no explicit intervention for a timeout.
func (h *Command) ReelTimeout() *reel.Step {
	return nil
	// TODO : fill the stub
}

// ReelEOF does nothing, Command requires no explicit intervention for EOF.
func (h *Command) ReelEOF() {
	// TODO : fill the stub
}
