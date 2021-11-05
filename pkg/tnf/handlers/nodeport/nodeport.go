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

package nodeport

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	npRegex = "(?s).+"
)

// NodePort holds information derived from inspecting NodePort services.
type NodePort struct {
	result  int
	timeout time.Duration
	args    []string
}

// NewNodePort creates a new NodePort tnf.Test.
func NewNodePort(timeout time.Duration, podNamespace string) *NodePort {
	return &NodePort{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{"oc", "-n", podNamespace, "get", "services", "-o",
			"custom-columns=TYPE:.spec.type", "|", "grep", "-E", "'NodePort|TYPE'"},
	}
}

// Args returns the command line args for the test.
func (np *NodePort) Args() []string {
	return np.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (np *NodePort) GetIdentifier() identifier.Identifier {
	return identifier.NodePortIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (np *NodePort) Timeout() time.Duration {
	return np.timeout
}

// Result returns the test result.
func (np *NodePort) Result() int {
	return np.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (np *NodePort) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{npRegex},
		Timeout: np.timeout,
	}
}

// ReelMatch ensures that no services utilize NodePort(s).
func (np *NodePort) ReelMatch(_, _, match string) *reel.Step {
	numExpectedLines := 1 // We want to have just the headers/titles line. Any other line is a NodePort line.

	trimmedMatch := strings.Trim(match, "\n")

	lines := strings.Split(trimmedMatch, "\n")
	numLines := len(lines)

	if numLines == numExpectedLines {
		np.result = tnf.SUCCESS
	}

	if numLines > numExpectedLines {
		np.result = tnf.FAILURE
	}

	return nil
}

// ReelTimeout does nothing;  no additional action is required upon timeout.
func (np *NodePort) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no additional action is required upon EOF.
func (np *NodePort) ReelEOF() {
}
