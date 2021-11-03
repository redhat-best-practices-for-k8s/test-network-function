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

package clusterversion

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

// TestMetadata holds OCP version strings and test metadata
type TestMetadata struct {
	versions ClusterVersion
	result   int
	timeout  time.Duration
	args     []string
}

// ClusterVersion holds OCP version strings
type ClusterVersion struct {
	Ocp, Oc, K8s string
}

// NewClusterVersion creates a new TestMetadata tnf.Test.
// Just gets the ocp version for client and server
func NewClusterVersion(timeout time.Duration) *TestMetadata {
	args := []string{"oc", "version"}
	return &TestMetadata{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    args,
	}
}

// Args returns the command line args for the test.
func (ver *TestMetadata) Args() []string {
	return ver.args
}

// GetVersions returns OCP client version.
func (ver *TestMetadata) GetVersions() ClusterVersion {
	return ver.versions
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (ver *TestMetadata) GetIdentifier() identifier.Identifier {
	return identifier.ClusterVersionIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (ver *TestMetadata) Timeout() time.Duration {
	return ver.timeout
}

// Result returns the test result.
func (ver *TestMetadata) Result() int {
	return ver.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (ver *TestMetadata) ReelFirst() *reel.Step {
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
func (ver *TestMetadata) ReelMatch(_, _, match string, status int) *reel.Step {
	re := regexp.MustCompile("(Server Version: )|(Client Version: )|(Kubernetes Version: )|(\n)")
	versions := re.Split(match, -1)
	versions = deleteEmpty(versions)
	if len(versions) != numVersionsOcp && len(versions) != numVersionsMinikube {
		ver.result = tnf.FAILURE
		return nil
	}
	ver.result = tnf.SUCCESS

	if len(versions) == numVersionsOcp {
		ver.versions.Oc = versions[0]
		ver.versions.Ocp = versions[1]
		ver.versions.K8s = versions[2]
	} else {
		ver.versions.Oc = versions[0]
		ver.versions.Ocp = "n/a"
		ver.versions.K8s = versions[1]
	}
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (ver *TestMetadata) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (ver *TestMetadata) ReelEOF() {
}
