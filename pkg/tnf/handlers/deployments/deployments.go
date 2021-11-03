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

package deployments

import (
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	dpRegex = "(?s).+"
)

// Deployment holds information about a single deployment
type Deployment struct {
	Replicas    int
	Ready       int
	UpToDate    int
	Available   int
	Unavailable int
}

// DeploymentMap name to Deployment
type DeploymentMap map[string]Deployment

// Deployments holds information derived from running "oc -n <namespace> get deployments" on the command line.
type Deployments struct {
	deployments DeploymentMap
	result      int
	timeout     time.Duration
	args        []string
}

// NewDeployments creates a new Deployments tnf.Test.
func NewDeployments(timeout time.Duration, namespace string) *Deployments {
	return &Deployments{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{"oc", "-n", namespace, "get", "deployments", "-o", "custom-columns=" +
			"NAME:.metadata.name," +
			"REPLICAS:.spec.replicas," +
			"READY:.status.readyReplicas," +
			"UPDATED:.status.updatedReplicas," +
			"AVAILABLE:.status.availableReplicas," +
			"UNAVAILABLE:.status.unavailableReplicas",
		},
		deployments: DeploymentMap{},
	}
}

// GetDeployments returns deployments extracted from running the Deployments tnf.Test.
func (dp *Deployments) GetDeployments() DeploymentMap {
	return dp.deployments
}

// Args returns the command line args for the test.
func (dp *Deployments) Args() []string {
	return dp.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (dp *Deployments) GetIdentifier() identifier.Identifier {
	return identifier.DeploymentsIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (dp *Deployments) Timeout() time.Duration {
	return dp.timeout
}

// Result returns the test result.
func (dp *Deployments) Result() int {
	return dp.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (dp *Deployments) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{dpRegex},
		Timeout: dp.timeout,
	}
}

// ReelMatch ensures that list of nodes is not empty and stores the names as []string
func (dp *Deployments) ReelMatch(_, _, match string, status int) *reel.Step {
	const numExepctedFields = 6
	trimmedMatch := strings.Trim(match, "\n")
	lines := strings.Split(trimmedMatch, "\n")[1:] // First line is the headers/titles line

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != numExepctedFields {
			return nil
		}
		dp.deployments[fields[0]] = Deployment{atoi(fields[1]), atoi(fields[2]), atoi(fields[3]), atoi(fields[4]), atoi(fields[5])}
	}

	dp.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (dp *Deployments) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (dp *Deployments) ReelEOF() {
}

func atoi(s string) int {
	const noneStr = "<none>"
	var num int
	if s != noneStr {
		num, _ = strconv.Atoi(s)
	}
	return num
}
