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

package hugepages

import (
	"strconv"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	hpRegex = "(?s).+"
	// RhelDefaultHugepages const
	RhelDefaultHugepages = 0
	// RhelDefaultHugepagesz const
	RhelDefaultHugepagesz = 2048 // kB
)

// Hugepages holds information derived from running "oc get MachineConfig" on the command line.
type Hugepages struct {
	hugepagesz int
	hugepages  int
	result     int
	timeout    time.Duration
	args       []string
}

// NewHugepages creates a new Hugepages tnf.Test.
func NewHugepages(timeout time.Duration) *Hugepages {
	return &Hugepages{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"oc", "get", "machineconfigs", "-l", "machineconfiguration.openshift.io/role=worker",
			"-o", "custom-columns=KARGS:.spec.kernelArguments",
			"|", "grep", "-v", "nil", "|", "grep", "-E", "'hugepage|KARGS'",
		},
	}
}

// Args returns the command line args for the test.
func (hp *Hugepages) Args() []string {
	return hp.args
}

// GetIdentifier returns the tnf.Test specific identifiesa.
func (hp *Hugepages) GetIdentifier() identifier.Identifier {
	return identifier.HugepagesIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (hp *Hugepages) Timeout() time.Duration {
	return hp.timeout
}

// Result returns the test result.
func (hp *Hugepages) Result() int {
	return hp.result
}

// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (hp *Hugepages) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{hpRegex},
		Timeout: hp.timeout,
	}
}

// ReelMatch sets the hugepages parameters based on cluster configuration and RHEL defaults
func (hp *Hugepages) ReelMatch(_, _, match string) *reel.Step {
	trimmedMatch := strings.Trim(match, "\n")
	lines := strings.Split(trimmedMatch, "\n")[1:] // First line is the headers/titles line

	params := map[string]string{}

	// Each line is of the form [name=value name=value ...]
	// Find the relevant parameters and store in params
	for _, line := range lines {
		line = line[1 : len(line)-1] // trim '[' and ']'
		fields := strings.Fields(line)
		for _, field := range fields {
			nameValue := strings.Split(field, "=")
			name := nameValue[0]
			value := nameValue[1]
			if isHugepagesParam(name) {
				params[name] = value
			}
		}
	}

	// Use params for determining the hugepages settings
	hugepages, ok := params["hugepages"]
	if ok {
		hp.hugepages, _ = strconv.Atoi(hugepages)
	} else {
		hp.hugepages = RhelDefaultHugepages
	}

	hugepagesz, ok := params["hugepagesz"]
	if ok {
		hp.hugepagesz = atoi(hugepagesz)
	} else {
		hugepagesz, ok := params["default_hugepagesz"]
		if ok {
			hp.hugepagesz = atoi(hugepagesz)
		} else {
			hp.hugepagesz = RhelDefaultHugepagesz
		}
	}

	hp.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (hp *Hugepages) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary upon EOF.
func (hp *Hugepages) ReelEOF() {
}

// GetHugepages func
func (hp *Hugepages) GetHugepages() int {
	return hp.hugepages
}

// GetHugepagesz func
func (hp *Hugepages) GetHugepagesz() int {
	return hp.hugepagesz
}

func isHugepagesParam(param string) bool {
	const (
		hugepagesz        = "hugepagesz"
		defaultHugepagesz = "default_hugepagesz"
		hugepages         = "hugepages"
	)
	return param == hugepagesz || param == defaultHugepagesz || param == hugepages
}

// atoi takes string in the format size[KMG] and returns an int in KB units
func atoi(s string) int {
	num, _ := strconv.Atoi(s[:len(s)-1])
	unit := s[len(s)-1]
	switch unit {
	case 'M':
		num *= 1024
	case 'G':
		num *= 1024 * 1024
	}

	return num
}
