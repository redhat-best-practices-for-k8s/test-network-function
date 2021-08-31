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

package deploymentsdrain

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	ddRegex                = "SUCCESS"
	drainTimeoutPercentage = 90 // drain timeout is a percentage of test timeout
)

// DeploymentsDrain holds information derived from running "oc adm drain" on the command line.
type DeploymentsDrain struct {
	result  int
	timeout time.Duration
	args    []string
	node    string
}

// NewDeploymentsDrain creates a new DeploymentsDrain tnf.Test.
func NewDeploymentsDrain(timeout time.Duration, nodeName string) *DeploymentsDrain {
	drainTimeout := timeout * drainTimeoutPercentage / 100
	drainTimeoutString := drainTimeout.String()
	return &DeploymentsDrain{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"oc", "adm", "drain", nodeName, "--pod-selector=pod-template-hash", "--disable-eviction=true",
			"--delete-local-data=true", "--ignore-daemonsets=true", "--timeout=" + drainTimeoutString,
			"&&", "echo", "SUCCESS",
		},
		node: nodeName,
	}
}

// Args returns the command line args for the test.
func (dd *DeploymentsDrain) Args() []string {
	return dd.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (dd *DeploymentsDrain) GetIdentifier() identifier.Identifier {
	return identifier.DeploymentsDrainIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (dd *DeploymentsDrain) Timeout() time.Duration {
	return dd.timeout
}

// Result returns the test result.
func (dd *DeploymentsDrain) Result() int {
	return dd.result
}

// ReelFirst returns a step which expects the output within the test timeout.
func (dd *DeploymentsDrain) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{ddRegex},
		Timeout: dd.timeout,
	}
}

// ReelMatch sets result
func (dd *DeploymentsDrain) ReelMatch(_, _, _ string) *reel.Step {
	dd.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (dd *DeploymentsDrain) ReelTimeout() *reel.Step {
	str := []string{"oc", "adm", "uncordon", dd.node}
	return &reel.Step{
		Expect:  []string{"(?m).*uncordoned"},
		Timeout: dd.timeout,
		Execute: strings.Join(str, " "),
	}
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (dd *DeploymentsDrain) ReelEOF() {
}
