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

package serviceaccount

import (
	"regexp"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	saRegex = " serviceAccountName: (.+)"
)

// ServiceAccount holds information from extracting Service Account information from a Pod definition.
type ServiceAccount struct {
	serviceAccountName string // Output variable for retrieving the result
	result             int
	timeout            time.Duration
	args               []string
}

// NewServiceAccount creates a new ServiceAccount tnf.Test.
func NewServiceAccount(timeout time.Duration, podName, podNamespace string) *ServiceAccount {
	return &ServiceAccount{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    []string{"oc", "-n", podNamespace, "get", "pods", podName, "-o", "yaml"},
	}
}

// Args returns the command line args for the test.
func (sa *ServiceAccount) Args() []string {
	return sa.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (sa *ServiceAccount) GetIdentifier() identifier.Identifier {
	return identifier.ServiceAccountIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (sa *ServiceAccount) Timeout() time.Duration {
	return sa.timeout
}

// Result returns the test result.
func (sa *ServiceAccount) Result() int {
	return sa.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (sa *ServiceAccount) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{saRegex},
		Timeout: sa.timeout,
	}
}

// ReelMatch ensures that the correct number of ServiceAccount annotations exist, and stores the correct SA within
// the ServiceAccount struct for later retrieval.
func (sa *ServiceAccount) ReelMatch(_, _, match string) *reel.Step {
	numExpectedMatches := 2
	saMatchIdx := 1
	re := regexp.MustCompile(saRegex)
	matched := re.FindStringSubmatch(match)
	if len(matched) < numExpectedMatches {
		return nil
	}

	sa.serviceAccountName = matched[saMatchIdx]
	sa.result = tnf.SUCCESS

	return nil
}

// ReelTimeout does nothing;  no action is needed upon timeout.
func (sa *ServiceAccount) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no aciton is needed upon EOF.
func (sa *ServiceAccount) ReelEOF() {
}

// GetServiceAccountName extracts the ServiceAccount (SA) for a Pod, if one exists.
func (sa *ServiceAccount) GetServiceAccountName() string {
	return sa.serviceAccountName
}
