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

package graceperiod

import (
	"regexp"
	"strconv"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	gpRegex = "(?s).+"
)

// GracePeriod holds information from extracting terminationGracePeriod from a Pod definition.
type GracePeriod struct {
	gracePeriod int // Output variable for retrieving the result
	result      int
	timeout     time.Duration
	args        []string
}

// NewGracePeriod creates a new GracePeriod tnf.Test.
func NewGracePeriod(timeout time.Duration, podName, podNamespace string) *GracePeriod {
	return &GracePeriod{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    []string{"oc", "-n", podNamespace, "get", "pod", podName, "-o", "jsonpath=\"{.spec.terminationGracePeriodSeconds}\""},
	}
}

// Args returns the command line args for the test.
func (gp *GracePeriod) Args() []string {
	return gp.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (gp *GracePeriod) GetIdentifier() identifier.Identifier {
	return identifier.GracePeriodIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (gp *GracePeriod) Timeout() time.Duration {
	return gp.timeout
}

// Result returns the test result.
func (gp *GracePeriod) Result() int {
	return gp.result
}

// ReelFirst returns a step which expects the pod's grace period within the test timeout.
func (gp *GracePeriod) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{gpRegex},
		Timeout: gp.timeout,
	}
}

// ReelMatch ensures that the terminationGracePeriod exist, and stores the correct grace period within
// the GracePeriod struct for later retrieval.
func (gp *GracePeriod) ReelMatch(_, _, match string, status int) *reel.Step {
	re := regexp.MustCompile(gpRegex)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		if len(matched) != 1 {
			gp.result = tnf.FAILURE
			return nil
		}
		gracePeriod, err := strconv.Atoi(matched[0])
		if err != nil {
			gp.result = tnf.FAILURE
			return nil
		}
		gp.result = tnf.SUCCESS
		gp.gracePeriod = gracePeriod
	} else {
		gp.result = tnf.FAILURE
	}
	return nil
}

// ReelTimeout does nothing;  no action is needed upon timeout.
func (gp *GracePeriod) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no aciton is needed upon EOF.
func (gp *GracePeriod) ReelEOF() {
}

// GetGracePeriod extracts the terminationGracePeriod from a Pod.
func (gp *GracePeriod) GetGracePeriod() int {
	return gp.gracePeriod
}
