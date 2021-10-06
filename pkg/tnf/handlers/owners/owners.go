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

package owners

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	owRegex     = "(?s)OWNERKIND\n.+"
	StatefulSet = "StatefulSet"
	ReplicaSet  = "ReplicaSet"
	DaemonSet   = "DaemonSet"
)

// Owners tests pod owners
type Owners struct {
	result  int
	timeout time.Duration
	args    []string
}

// NewOwners creates a new Owners tnf.Test.
func NewOwners(timeout time.Duration, podNamespace, podName string) *Owners {
	return &Owners{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{"oc", "-n", podNamespace, "get", "pods", podName,
			"-o", `custom-columns=OWNERKIND:.metadata.ownerReferences\[\*\].kind`},
	}
}

// Args returns the command line args for the test.
func (ow *Owners) Args() []string {
	return ow.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (ow *Owners) GetIdentifier() identifier.Identifier {
	return identifier.OwnersIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (ow *Owners) Timeout() time.Duration {
	return ow.timeout
}

// Result returns the test result.
func (ow *Owners) Result() int {
	return ow.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (ow *Owners) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{owRegex},
		Timeout: ow.timeout,
	}
}

// ReelMatch ensures that list of nodes is not empty and stores the names as []string
func (ow *Owners) ReelMatch(_, _, match string) *reel.Step {
	if (strings.Contains(match, StatefulSet) || strings.Contains(match, ReplicaSet)) &&
		!strings.Contains(match, DaemonSet) {
		ow.result = tnf.SUCCESS
	} else {
		ow.result = tnf.FAILURE
	}
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (ow *Owners) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (ow *Owners) ReelEOF() {
}
