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

package crdstatus

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	ocLabelQueryFlag = " -l "
	ocCommand        = "%s get crds -o json"
	jqCommand        = "'[ .items[] | {\"name\": .metadata.name, \"kind\" : .spec.names.kind, \"status\" : .spec.versions[].schema.openAPIV3Schema.properties.status}]'"
)

// Crd maps to a custom json object to get the Status config of a CRD.
type Crd struct {
	Name   string                 `json:"name"`
	Kind   string                 `json:"kind"`
	Status map[string]interface{} `json:"status"`
}

// CrdStatus provides a CrdStatus test implemented using command line tool CrdStatus.
type CrdStatus struct {
	result  int
	timeout time.Duration
	args    []string

	CrdItems []Crd
	// adding special parameters
}

func getCommand(labels []string) string {
	command := fmt.Sprintf(ocCommand, dependencies.OcBinaryName)

	if len(labels) > 0 {
		command += ocLabelQueryFlag
		for _, label := range labels {
			command += label + ","
		}
		// Remove last ","
		command = command[:len(command)-1]
	}

	// Add jq filter.
	command += " | " + dependencies.JqBinaryName + " " + jqCommand
	return command
}

// NewCrdStatus returns a new CrdStatus test handler.
func NewCrdStatus(timeout time.Duration, labels []string) *CrdStatus {
	return &CrdStatus{
		result:   tnf.ERROR,
		timeout:  timeout,
		args:     strings.Split(getCommand(labels), " "),
		CrdItems: []Crd{},
	}
}

// Args returns the handler arguments
func (h *CrdStatus) Args() []string {
	return h.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (h *CrdStatus) GetIdentifier() identifier.Identifier {
	return identifier.DeploymentsIdentifier
}

// Timeout return the timeout for the test.
func (h *CrdStatus) Timeout() time.Duration {
	return h.timeout
}

// Result returns the test result.
func (h *CrdStatus) Result() int {
	return h.result
}

// ReelFirst returns a step which expects an CrdStatus summary for the given device.
func (h *CrdStatus) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{".+"}, // TODO : pass the list of possible regex in here
		Timeout: h.timeout,
	}
}

// ReelMatch parses the CrdStatus output and set the test result on match.
func (h *CrdStatus) ReelMatch(_, _, match string) *reel.Step {
	err := json.Unmarshal([]byte(match), &h.CrdItems)
	if err != nil {
		logrus.Error("Unable to unmarshall json CRDs. Error: ", err)
		h.result = tnf.FAILURE
	} else {
		h.result = tnf.SUCCESS
	}
	return nil
}

// ReelTimeout does nothing, CrdStatus requires no explicit intervention for a timeout.
func (h *CrdStatus) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing, CrdStatus requires no explicit intervention for EOF.
func (h *CrdStatus) ReelEOF() {
}
