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

package scaling

import (
	"fmt"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	ocCommand = "oc scale --replicas=%d deployment %s -n %s"
	regex     = "^deployment.*/%s scaled"
)

// Scaling holds information derived from running "oc -n <namespace> get deployments" on the command line.
type Scaling struct {
	result  int
	timeout time.Duration
	args    []string
	regex   string
}

// NewScaling creates a new Scaling handler
func NewScaling(timeout time.Duration, namespace, deploymentName string, replicaCount int) *Scaling {
	command := fmt.Sprintf(ocCommand, replicaCount, deploymentName, namespace)
	return &Scaling{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    strings.Fields(command),
		regex:   fmt.Sprintf(regex, deploymentName),
	}
}

// Args returns the command line args for the test.
func (scaling *Scaling) Args() []string {
	return scaling.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (scaling *Scaling) GetIdentifier() identifier.Identifier {
	return identifier.ScalingIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (scaling *Scaling) Timeout() time.Duration {
	return scaling.timeout
}

// Result returns the test result.
func (scaling *Scaling) Result() int {
	return scaling.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (scaling *Scaling) ReelFirst() *reel.Step {
	return &reel.Step{
		Execute: "",
		Expect:  []string{scaling.regex},
		Timeout: scaling.timeout,
	}
}

// ReelMatch does nothing, just set the test result as success.
func (scaling *Scaling) ReelMatch(_, _, match string) *reel.Step {
	scaling.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (scaling *Scaling) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (scaling *Scaling) ReelEOF() {
}
