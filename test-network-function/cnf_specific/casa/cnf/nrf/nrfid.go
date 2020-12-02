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

package nrf

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"regexp"
	"strings"
	"time"
)

const (
	// CheckRegistrationCommand is the Unix command for checking NRF registrations.
	CheckRegistrationCommand = "oc -n %s get nfregistrations.mgmt.casa.io $(oc -n %s get nfregistrations.mgmt.casa.io | awk {\"print $2\"} | xargs -n 1)"
	// CommandCompleteRegexString is the regular expression indicating the command has completed.
	CommandCompleteRegexString = `(?m)NRF\s+TYPE\s+INSTID\s+STATUS\s+`
	// OutputRegexString is the regular expression capturing all CNFs in the NRF registration output.
	OutputRegexString = `(?m)((\w+)\s+(\w+)\s+([0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12})\s+(\w+)\n)+`
	// SingleEntryRegexString is the regular expression capturing one CNF in the NRF registration output.
	SingleEntryRegexString = `(\w+)\s+(\w+)\s+([0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12})\s+(\w+)`
)

// OutputRegex matches the output from inspecting ID registrations.
var OutputRegex = regexp.MustCompile(OutputRegexString)

// SingleEntryRegex matches a single matching ID.
var SingleEntryRegex = regexp.MustCompile(SingleEntryRegexString)

// ID follows the container design pattern, and stores Network Registration information.
type ID struct {
	nrf    string
	typ    string
	instID string
	status string
}

// GetType extracts the type of CNF.
func (n *ID) GetType() string {
	return n.typ
}

// NewNRFID creates a new ID.
func NewNRFID(nrf, typ, instID, status string) *ID {
	return &ID{nrf: nrf, typ: typ, instID: instID, status: status}
}

// fromString creates an ID from a string.
func fromString(nrf string) (*ID, error) {
	match := SingleEntryRegex.FindStringSubmatch(nrf)
	if match != nil {
		nrf := match[1] //nolint:govet // acceptable shadowing, uses outside `if` are unreachable.
		typ := match[2]
		instID := match[3]
		status := match[4]
		return NewNRFID(nrf, typ, instID, status), nil
	}
	// Untestable code below.  No current clients will ever call the code below.
	return nil, fmt.Errorf("nrf string is not parsable: %s", nrf)
}

// Registration is a tnf.Test that dumps the output of all CNF registrations in the CR.
type Registration struct {
	// args represents the Unix command.
	args []string
	// namespace represents the namespace the CNF operators are deployed within.
	namespace string
	// result is the result of the tnf.Test.
	result int
	// timeout is the tnf.Test timeout.
	timeout time.Duration
	// registeredNRFs is a mapping of uuid to other NRF facets (ID).
	registeredNRFs map[string]*ID
}

// Args returns the command line args for the test.
func (r *Registration) Args() []string {
	return r.args
}

// Timeout returns the timeout in seconds for the test.
func (r *Registration) Timeout() time.Duration {
	return r.timeout
}

// Result returns the test result.
func (r *Registration) Result() int {
	return r.result
}

// ReelFirst returns a step which expects an ip summary for the given device.
func (r *Registration) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{OutputRegexString, CommandCompleteRegexString},
		Timeout: r.timeout,
	}
}

// ReelMatch parses the Registration.  Returns no step; the test is complete.
func (r *Registration) ReelMatch(pattern, _, match string) *reel.Step {
	if pattern == CommandCompleteRegexString {
		// Indicates that the command was successfully run, but there were no registered NRFs.
		r.result = tnf.FAILURE
		return nil
	} else if pattern == OutputRegexString {
		matches := SingleEntryRegex.FindAllString(match, -1)
		r.result = tnf.SUCCESS
		for _, m := range matches {
			nrf, err := fromString(m)
			// defensive;  should have already matched. Untestable, as we ensure a match prior to calling fromString()
			// fromString() is also defensive in case there is a future client that doesn't perform such a check.
			if err != nil {
				r.result = tnf.ERROR
				return nil
			}
			nrfInstID := nrf.instID
			r.registeredNRFs[nrfInstID] = nrf
		}
		return nil
	}
	r.result = tnf.ERROR
	return nil
}

// ReelTimeout does nothing;  no intervention is needed.
func (r *Registration) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing.  On EOF, take no action.
func (r *Registration) ReelEOF() {
}

// GetRegisteredNRFs returns the map of registered ID instances.
func (r *Registration) GetRegisteredNRFs() map[string]*ID {
	return r.registeredNRFs
}

// registrationCommand creates the Unix command to check for registration.
func registrationCommand(namespace string) []string {
	command := fmt.Sprintf(CheckRegistrationCommand, namespace, namespace)
	return strings.Split(command, " ")
}

// NewRegistration creates a Registration instance.
func NewRegistration(timeout time.Duration, namespace string) *Registration {
	return &Registration{result: tnf.ERROR, timeout: timeout, args: registrationCommand(namespace), namespace: namespace, registeredNRFs: map[string]*ID{}}
}
