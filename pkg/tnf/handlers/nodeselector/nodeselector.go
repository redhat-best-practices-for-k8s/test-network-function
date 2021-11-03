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

package nodeselector

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	nsRegex = "<none>\\s*<none>"
)

// NodeSelector holds information from extracting NodeSelector and NodeAffinity information from a Pod definition.
type NodeSelector struct {
	result  int
	timeout time.Duration
	args    []string
}

// NewNodeSelector creates a new NodeSelector tnf.Test.
func NewNodeSelector(timeout time.Duration, podName, podNamespace string) *NodeSelector {
	return &NodeSelector{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    []string{"oc", "-n", podNamespace, "get", "pods", podName, "-o", "custom-columns=nodeselector:.spec.nodeSelector,nodeaffinity:.spec.nodeAffinity"},
	}
}

// Args returns the command line args for the test.
func (ns *NodeSelector) Args() []string {
	return ns.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (ns *NodeSelector) GetIdentifier() identifier.Identifier {
	return identifier.NodeSelectorIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (ns *NodeSelector) Timeout() time.Duration {
	return ns.timeout
}

// Result returns the test result.
func (ns *NodeSelector) Result() int {
	return ns.result
}

// ReelFirst returns a step which expects the nodeSelector within the test timeout.
func (ns *NodeSelector) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{nsRegex},
		Timeout: ns.timeout,
	}
}

// ReelMatch ensures that there is no nodeSelector or nodeAffinity on the pod spec
func (ns *NodeSelector) ReelMatch(_, _, match string, status int) *reel.Step {
	ns.result = tnf.SUCCESS

	return nil
}

// ReelTimeout does nothing;  no action is needed upon timeout.
func (ns *NodeSelector) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no aciton is needed upon EOF.
func (ns *NodeSelector) ReelEOF() {
}
