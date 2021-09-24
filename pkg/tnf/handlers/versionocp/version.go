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

package versionocp

import (
	"regexp"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	verRegex            = "(?s).+"
	numVersionsOcp      = 3
	numVersionsMinikube = 2
)

// VersionOCP holds OCP version strings
type VersionOCP struct {
	versions []string
	result   int
	timeout  time.Duration
	args     []string
}

// NewVersionOCP creates a new VersionOCP tnf.Test.
// Just gets the ocp version for client and server
func NewVersionOCP(timeout time.Duration) *VersionOCP {
	args := []string{"oc", "version"}
	return &VersionOCP{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    args,
	}
}

// Args returns the command line args for the test.
func (ver *VersionOCP) Args() []string {
	return ver.args
}

// GetVersions returns OCP client version.
func (ver *VersionOCP) GetVersions() []string {
	return ver.versions
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (ver *VersionOCP) GetIdentifier() identifier.Identifier {
	return identifier.VersionOcpIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (ver *VersionOCP) Timeout() time.Duration {
	return ver.timeout
}

// Result returns the test result.
func (ver *VersionOCP) Result() int {
	return ver.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (ver *VersionOCP) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{verRegex},
		Timeout: ver.timeout,
	}
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// ReelMatch ensures that list of nodes is not empty and stores the names as []string
func (ver *VersionOCP) ReelMatch(_, _, match string) *reel.Step {
	re := regexp.MustCompile("(Server Version: )|(Client Version: )|(Kubernetes Version: )|(\n)")
	ver.versions = re.Split(match, -1)
	ver.versions = deleteEmpty(ver.versions)
	if len(ver.versions) != numVersionsOcp && len(ver.versions) != numVersionsMinikube {
		ver.result = tnf.FAILURE
	} else {
		ver.result = tnf.SUCCESS
	}
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (ver *VersionOCP) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (ver *VersionOCP) ReelEOF() {
}
