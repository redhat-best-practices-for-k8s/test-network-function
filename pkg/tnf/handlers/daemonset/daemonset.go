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

package daemonset

import (
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

type Status struct {
	Name         string
	Desired      int
	Current      int
	Ready        int
	Available    int
	Misscheduled int
}

// DaemonSet is the reel handler struct.
type DaemonSet struct {
	result  int
	timeout time.Duration
	args    []string
	status  Status
}

const (
	dsRegex = "(?s).+"
)

// NewDaemonSet returns a new DaemonSet handler struct.
func NewDaemonSet(timeout time.Duration, daemonset, namespace string) *DaemonSet {
	return &DaemonSet{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{"oc", "-n", namespace, "get", "ds", daemonset, "-o",
			"go-template='{{ .spec.template.metadata.name }} ",
			"{{ .status.desiredNumberScheduled }}",
			"{{ .status.currentNumberScheduled  }}",
			"{{ .status.numberAvailable }}",
			"{{ .status.numberReady }}",
			"{{ .status.numberMisscheduled }} {{ printf \"\\n\" }}'",
		},
		status: Status{},
	}
}

// Args returns the initial execution/send command strings for handler DaemonSet.
func (ds *DaemonSet) Args() []string {
	return ds.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (ds *DaemonSet) GetIdentifier() identifier.Identifier {
	// Return the DaemonSet handler identifier.
	return identifier.DaemonSetIdentifier
}

// Timeout returns the timeout for the test.
func (ds *DaemonSet) Timeout() time.Duration {
	return ds.timeout
}

// Result returns the test result.
func (ds *DaemonSet) Result() int {
	return ds.result
}

// ReelFirst returns a reel step for handler DaemonSet.
func (ds *DaemonSet) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{dsRegex},
		Timeout: ds.timeout,
	}
}

// ReelMatch parses the DaemonSet output and set the test result on match.
func (ds *DaemonSet) ReelMatch(_, _, match string) *reel.Step {
	const numExpectedFields = 6
	trimmedMatch := strings.Trim(match, "\n")
	lines := strings.Split(trimmedMatch, "\n")[0:] // Keep First line only

	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != numExpectedFields {
			return nil
		}
		err := processResult(&ds.status, fields)
		if err != nil {
			log.Error("Error processing output ", err)
			ds.status = Status{}
			return nil
		}
		ds.result = tnf.SUCCESS
		return nil
	}
	return nil
}
func processResult(status *Status, fields []string) error {
	var err error
	status.Name = fields[0]
	status.Desired, err = strconv.Atoi(fields[1])
	if err != nil {
		return err
	}
	status.Current, err = strconv.Atoi(fields[2])
	if err != nil {
		return nil
	}
	status.Available, err = strconv.Atoi(fields[3])
	if err != nil {
		return nil
	}
	status.Ready, err = strconv.Atoi(fields[4])
	if err != nil {
		return nil
	}
	status.Misscheduled, err = strconv.Atoi(fields[5])
	if err != nil {
		return nil
	}
	return nil
}

// ReelTimeout function for DaemonSet will be called by the reel FSM when a expect timeout occurs.
func (ds *DaemonSet) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF function for DaemonSet will be called by the reel FSM when a EOF is read.
func (ds *DaemonSet) ReelEOF() {
}

func (ds *DaemonSet) GetStatus() Status {
	return ds.status
}
