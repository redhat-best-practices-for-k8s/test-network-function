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

package podsets

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

// PodSet holds information about a single Deployment/statefulsets
type PodSet struct {
	Replicas    int
	Ready       int
	UpToDate    int
	Available   int
	Unavailable int
	Current     int
}

// PodSetMap name to Deployment/statefulsets
type PodSetMap map[string]PodSet

// PodSets holds information derived from running "oc -n <namespace> get deployments/statefulsets" on the command line.
type PodSets struct {
	podsets   PodSetMap
	namespace string
	result    int
	timeout   time.Duration
	args      []string
}

// NewPodSets creates a new deployments/statefulsets tnf.Test.
func NewPodSets(timeout time.Duration, namespace, resourceType string) *PodSets {
	return &PodSets{
		timeout:   timeout,
		namespace: namespace,
		result:    tnf.ERROR,
		args: []string{"oc", "-n", namespace, "get", resourceType, "-o", "custom-columns=" +
			"NAME:.metadata.name," +
			"REPLICAS:.spec.replicas," +
			"READY:.status.readyReplicas," +
			"UPDATED:.status.updatedReplicas," +
			"AVAILABLE:.status.availableReplicas," +
			"UNAVAILABLE:.status.unavailableReplicas," +
			"CURRENT:.status.currentReplicas",
		},

		podsets: PodSetMap{},
	}
}

// GetPodSets returns deployments/statefulsets extracted from running the deployments/statefulsets tnf.Test.
func (ps *PodSets) GetPodSets() PodSetMap {
	return ps.podsets
}

// Args returns the command line args for the test.
func (ps *PodSets) Args() []string {
	return ps.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (ps *PodSets) GetIdentifier() identifier.Identifier {
	return identifier.PodSetsIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (ps *PodSets) Timeout() time.Duration {
	return ps.timeout
}

// Result returns the test result.
func (ps *PodSets) Result() int {
	return ps.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (ps *PodSets) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{dpRegex},
		Timeout: ps.timeout,
	}
}

// ReelMatch ensures that list of nodes is not empty and stores the names as []string
func (ps *PodSets) ReelMatch(_, _, match string) *reel.Step {
	const numExepctedFields = 7
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
		// we can have the same deployment in different namespaces
		// this ensures the uniqueness of the deployment in the test
		key := ps.namespace + ":" + fields[0]
		ps.podsets[key] = PodSet{atoi(fields[1]), atoi(fields[2]), atoi(fields[3]), atoi(fields[4]), atoi(fields[5]), atoi(fields[6])}
	}

	ps.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (ps *PodSets) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (ps *PodSets) ReelEOF() {
}

func atoi(s string) int {
	const noneStr = "<none>"
	var num int
	if s != noneStr {
		num, _ = strconv.Atoi(s)
	}
	return num
}
