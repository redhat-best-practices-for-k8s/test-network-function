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

package podnodename

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `.+`
)

// PodNodeName holds information derived from running "oc get pod -o jsonpath=\"{.spec.nodeName}\"" on the command line.
type PodNodeName struct {
	NodeName string // Output variable that stores the name of the node
	result   int
	timeout  time.Duration
	args     []string
}

// NewPodNodeName creates a PodNodeName tnf.Test.
func NewPodNodeName(timeout time.Duration, podName, podNamespace string) *PodNodeName {
	return &PodNodeName{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"oc get pod -n " + podNamespace + " " + podName + " -o jsonpath=\"{.spec.nodeName}\"",
		},
	}
}

// GetNodeName returns the name of the node extracted while running the PodNodeName tnf.Test.
func (pnn *PodNodeName) GetNodeName() string {
	return pnn.NodeName
}

// Args returns the command line args for the test.
func (pnn *PodNodeName) Args() []string {
	return pnn.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (pnn *PodNodeName) GetIdentifier() identifier.Identifier {
	return identifier.PodNodeNameIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (pnn *PodNodeName) Timeout() time.Duration {
	return pnn.timeout
}

// Result returns the test result.
func (pnn *PodNodeName) Result() int {
	return pnn.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (pnn *PodNodeName) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: pnn.timeout,
	}
}

// ReelMatch ensures that there are no PodNodeName matched in the command output.
func (pnn *PodNodeName) ReelMatch(_, _, match string) *reel.Step {
	pnn.NodeName = match
	pnn.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (pnn *PodNodeName) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (pnn *PodNodeName) ReelEOF() {
}
