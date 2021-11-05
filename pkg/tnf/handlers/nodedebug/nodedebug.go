// Copyright (C) 2021 Red Hat, Inc.
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

package nodedebug

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	ndRegex = "(?s).+"
)

// NodeDebug contains the output of the command
type NodeDebug struct {
	result    int
	timeout   time.Duration
	args      []string
	Trim      bool     // trim leading and trailing new lines
	Split     bool     // split lines of text into slice
	Raw       string   // output of executed command
	Processed []string // output after splitting and trimming. nil if !split&&!trim
}

// NewNodeDebug creates a new NodeDebug tnf.Test.
// command: caller handles escaping and preparing for shell execution
//          empty output indicates error - make sure you always print something
// trim: trim leading and trailing new lines
// split: split lines of text into slice
func NewNodeDebug(timeout time.Duration, nodeName, command string, trim, split bool) *NodeDebug {
	return &NodeDebug{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			command,
		},
		Trim:  trim,
		Split: split,
	}
}

// Args returns the command line args for the test.
func (nd *NodeDebug) Args() []string {
	return nd.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (nd *NodeDebug) GetIdentifier() identifier.Identifier {
	return identifier.NodeDebugIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (nd *NodeDebug) Timeout() time.Duration {
	return nd.timeout
}

// Result returns the test result.
func (nd *NodeDebug) Result() int {
	return nd.result
}

// ReelFirst returns a step which expects the output within the test timeout.
func (nd *NodeDebug) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{ndRegex},
		Timeout: nd.timeout,
	}
}

// ReelMatch executs the command and parses the output
func (nd *NodeDebug) ReelMatch(_, _, match string) *reel.Step {
	nd.result = tnf.SUCCESS
	nd.Raw = match
	if !nd.Split && !nd.Trim {
		return nil
	}

	trimmedMatch := match
	if nd.Trim {
		trimmedMatch = strings.Trim(match, "\n")
	}
	if nd.Split {
		nd.Processed = strings.Split(trimmedMatch, "\n")
	} else {
		nd.Processed = []string{trimmedMatch}
	}

	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (nd *NodeDebug) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (nd *NodeDebug) ReelEOF() {
}
