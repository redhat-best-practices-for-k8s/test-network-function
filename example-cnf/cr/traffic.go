// Copyright (C) 2020-2021 Red Hat, Inc.
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

package cr

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// trafficPassedSuccessfullyRegex is regular expression expected for a container based on Red Hat technologies.
	trafficPassedSuccessfullyRegex = `.+`
)

// Traffic is an implementation of tnf.Test used to determine whether a container is based on Red Hat technologies.
type Traffic struct {
	// result is the result of the test.
	result int

	// timeout is the timeout duration for the test.
	timeout time.Duration

	// args stores the command and arguments.
	args []string

	// crFile is the path to the Custom Resource file.
	crFile string
}

// Args returns the command line arguments for the test.
func (r *Traffic) Args() []string {
	return r.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (r *Traffic) GetIdentifier() identifier.Identifier {
	return identifier.Identifier{
		URL:             "http://test-network-function.com/test-network-function/cr/create",
		SemanticVersion: identifier.VersionOne,
	}
}

// Timeout returns the timeout for the test.
func (r *Traffic) Timeout() time.Duration {
	return r.timeout
}

// Result returns the test result.
func (r *Traffic) Result() int {
	return r.result
}

// ReelFirst returns a reel.Step which expects output from running the Args command.
func (r *Traffic) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{trafficPassedSuccessfullyRegex},
		Timeout: r.timeout,
	}
}

// ReelMatch determines whether the container is based on Red Hat technologies through pattern matching logic.
func (r *Traffic) ReelMatch(_, _, _ string) *reel.Step {
	r.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no intervention is needed for a timeout.
func (r *Traffic) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no intervention is needed for EOF.
func (r *Traffic) ReelEOF() {
}

// NewTraffic create a new Traffic tnf.Test.
func NewTraffic(namespace, crType, crName string, timeout time.Duration) *Traffic {
	return &Traffic{
		result:  tnf.ERROR,
		timeout: timeout,
		args:    []string{dependencies.OcBinaryName, "describe", "-n", namespace, crType, crName, "|", dependencies.GrepBinaryName, "TestPassed"},
	}
}
