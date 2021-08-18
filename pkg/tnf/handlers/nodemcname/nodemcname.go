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

package nodemcname

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `.+`
)

// NodeMcName holds information derived from getting the node's spec and extracting the current MC out of it.
type NodeMcName struct {
	McName  string // Output variable that stores the name of the node
	result  int
	timeout time.Duration
	args    []string
}

// NewNodeMcName creates a NodeMcName tnf.Test.
func NewNodeMcName(timeout time.Duration, nodeName string) *NodeMcName {
	return &NodeMcName{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"oc get node " + nodeName + ` -o json | jq '.metadata.annotations."machineconfiguration.openshift.io/currentConfig"'`,
		},
	}
}

// GetMcName returns the name of the mc extracted while running the NodeMcName tnf.Test.
func (nmn *NodeMcName) GetMcName() string {
	return nmn.McName
}

// Args returns the command line args for the test.
func (nmn *NodeMcName) Args() []string {
	return nmn.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (nmn *NodeMcName) GetIdentifier() identifier.Identifier {
	return identifier.NodeMcNameIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (nmn *NodeMcName) Timeout() time.Duration {
	return nmn.timeout
}

// Result returns the test result.
func (nmn *NodeMcName) Result() int {
	return nmn.result
}

// ReelFirst returns a step which expects the node mc name within the test timeout.
func (nmn *NodeMcName) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: nmn.timeout,
	}
}

// ReelMatch ensures that there are no NodeMcName matched in the command output.
func (nmn *NodeMcName) ReelMatch(_, _, match string) *reel.Step {
	nmn.McName = match
	nmn.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (nmn *NodeMcName) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (nmn *NodeMcName) ReelEOF() {
}
