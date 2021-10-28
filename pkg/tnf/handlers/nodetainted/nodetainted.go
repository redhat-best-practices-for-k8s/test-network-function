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

package nodetainted

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	ntRegex = "^\\d"
)

// NodeTainted holds information about tainted nodes.
type NodeTainted struct {
	result  int
	timeout time.Duration
	args    []string
	Match   string
}

// NewNodeTainted creates a new NodeTainted tnf.Test.
func NewNodeTainted(timeout time.Duration) *NodeTainted {
	return &NodeTainted{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"cat", "/proc/sys/kernel/tainted",
		},
	}
}

// Args returns the command line args for the test.
func (nt *NodeTainted) Args() []string {
	return nt.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (nt *NodeTainted) GetIdentifier() identifier.Identifier {
	return identifier.NodeTaintedIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (nt *NodeTainted) Timeout() time.Duration {
	return nt.timeout
}

// Result returns the test result.
func (nt *NodeTainted) Result() int {
	return nt.result
}

// ReelFirst returns a step which expects the output within the test timeout.
func (nt *NodeTainted) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{ntRegex},
		Timeout: nt.timeout,
	}
}

// ReelMatch tests whether node is tainted or not
func (nt *NodeTainted) ReelMatch(_, _, match string) *reel.Step {
	nt.Match = match
	if match == "0" {
		nt.result = tnf.SUCCESS
	} else {
		nt.result = tnf.FAILURE
	}

	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (nt *NodeTainted) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (nt *NodeTainted) ReelEOF() {
}
