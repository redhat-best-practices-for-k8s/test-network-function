// Copyright (C) 2020 Red Hat, Inc.
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

package redhat

import (
	"fmt"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// NotRedHatBasedRegex is the expected output for a container that is not based on Red Hat technologies.
	NotRedHatBasedRegex = `(?m)Unknown Base Image`
	// VersionRegex is regular expression expected for a container based on Red Hat technologies.
	VersionRegex = `(?m)Red Hat Enterprise Linux( Server)? release (\d+\.\d+) \(\w+\)`
)

var (
	// ReleaseCommand is the Unix command used to check whether a container is based on Red Hat technologies.
	ReleaseCommand = fmt.Sprintf("if [ -e /etc/redhat-release ]; then %s /etc/redhat-release; else echo \"Unknown Base Image\"; fi", dependencies.CatBinaryName)
)

// Release is an implementation of tnf.Test used to determine whether a container is based on Red Hat technologies.
type Release struct {
	// result is the result of the test.
	result int
	// timeout is the timeout duration for the test.
	timeout time.Duration
	// args stores the command and arguments.
	args []string
	// release contains the contents of /etc/redhat-release if it exists, or "NOT Red Hat Based" if it does not exist.
	release string
	// isRedHatBased contains whether the container is based on Red Hat technologies.
	isRedHatBased bool
}

// Args returns the command line arguments for the test.
func (r *Release) Args() []string {
	return r.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (r *Release) GetIdentifier() identifier.Identifier {
	return identifier.VersionIdentifier
}

// Timeout returns the timeout for the test.
func (r *Release) Timeout() time.Duration {
	return r.timeout
}

// Result returns the test result.
func (r *Release) Result() int {
	return r.result
}

// ReelFirst returns a reel.Step which expects output from running the Args command.
func (r *Release) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{VersionRegex, NotRedHatBasedRegex},
		Timeout: r.timeout,
	}
}

// ReelMatch determines whether the container is based on Red Hat technologies through pattern matching logic.
func (r *Release) ReelMatch(pattern, _, _ string) *reel.Step {
	if pattern == NotRedHatBasedRegex {
		r.result = tnf.FAILURE
		r.isRedHatBased = false
	} else if pattern == VersionRegex {
		// If the above conditional is not triggered, it can be deduced that we have matched the VersionRegex.
		r.result = tnf.SUCCESS
		r.isRedHatBased = true
	} else {
		r.result = tnf.ERROR
		r.isRedHatBased = false
	}
	return nil
}

// ReelTimeout does nothing;  no intervention is needed for a timeout.
func (r *Release) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no intervention is needed for EOF.
func (r *Release) ReelEOF() {
}

// NewRelease create a new Release tnf.Test.
func NewRelease(timeout time.Duration) *Release {
	return &Release{result: tnf.ERROR, timeout: timeout, args: strings.Split(ReleaseCommand, " ")}
}
