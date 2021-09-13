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

package nodehugepages

import (
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/common"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	nhRegex = "(?s).+"
)

// NodeHugepages test
type NodeHugepages struct {
	hugepagesz int // The cluster's configuration
	hugepages  int //
	result     int
	timeout    time.Duration
	args       []string
}

// NewNodeHugepages creates a new NodeHugepages tnf.Test.
func NewNodeHugepages(timeout time.Duration, node string, hugepagesz, hugepages int) *NodeHugepages {
	return &NodeHugepages{
		hugepagesz: hugepagesz,
		hugepages:  hugepages,
		timeout:    timeout,
		result:     tnf.ERROR,
		args: []string{
			"echo", "\"grep -E 'HugePages_Total:|Hugepagesize:' /proc/meminfo\"", "|", common.GetOcDebugCommand(), "node/" + node,
		},
	}
}

// Args returns the command line args for the test.
func (nh *NodeHugepages) Args() []string {
	return nh.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (nh *NodeHugepages) GetIdentifier() identifier.Identifier {
	return identifier.NodeHugepagesIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (nh *NodeHugepages) Timeout() time.Duration {
	return nh.timeout
}

// Result returns the test result.
func (nh *NodeHugepages) Result() int {
	return nh.result
}

// ReelFirst returns a step which expects the output within the test timeout.
func (nh *NodeHugepages) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{nhRegex},
		Timeout: nh.timeout,
	}
}

// ReelMatch tests the node's hugepages configuration
func (nh *NodeHugepages) ReelMatch(_, _, match string) *reel.Step {
	trimmedMatch := strings.Trim(match, "\n")
	lines := strings.Split(trimmedMatch, "\n")

	const numExpectedLines = 2

	if len(lines) != numExpectedLines {
		return nil
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		name := fields[0][:len(fields[0])-1]
		value := fields[1]
		if !nh.validateParameter(name, value) {
			nh.result = tnf.FAILURE
		}
	}

	if nh.result != tnf.FAILURE {
		nh.result = tnf.SUCCESS
	}
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (nh *NodeHugepages) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (nh *NodeHugepages) ReelEOF() {
}

func (nh *NodeHugepages) validateParameter(name, value string) bool {
	const (
		hugePagesTotal = "HugePages_Total"
		hugepagesize   = "Hugepagesize"
	)
	num, _ := strconv.Atoi(value)
	switch name {
	case hugePagesTotal:
		return num == nh.hugepages
	case hugepagesize:
		return num == nh.hugepagesz
	}
	return false
}
