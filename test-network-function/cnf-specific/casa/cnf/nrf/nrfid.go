package nrf

import (
	"errors"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/utils"
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

// OutputRegex matches the output from inspecting NRFID registrations.
var OutputRegex = regexp.MustCompile(OutputRegexString)

// SingleEntryRegex matches a single matching NRFID.
var SingleEntryRegex = regexp.MustCompile(SingleEntryRegexString)

// NRFID follows the container design pattern, and stores Network Registration information.
type NRFID struct {
	nrf    string
	typ    string
	instID string
	status string
}

// GetType extracts the type of CNF.
func (n *NRFID) GetType() string {
	return n.typ
}

// NewNRFID creates a new NRFID.
func NewNRFID(nrf, typ, instID, status string) *NRFID {
	return &NRFID{nrf: nrf, typ: typ, instID: instID, status: status}
}

// fromString creates an NRFID from a string.
func fromString(nrf string) (*NRFID, error) {
	match := SingleEntryRegex.FindStringSubmatch(nrf)
	if match != nil {
		nrf := match[1]
		typ := match[2]
		instID := match[3]
		status := match[4]
		return NewNRFID(nrf, typ, instID, status), nil
	}
	// Untestable code below.  No current clients will ever call the code below.
	return nil, errors.New(fmt.Sprintf("nrf string is not parsable: %s", nrf))
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
	// registeredNRFs is a mapping of uuid to other NRF facets (NRFID).
	registeredNRFs map[string]*NRFID
}

// Return the command line args for the test.
func (r *Registration) Args() []string {
	return r.args
}

// Return the timeout in seconds for the test.
func (r *Registration) Timeout() time.Duration {
	return r.timeout
}

// Return the test result.
func (r *Registration) Result() int {
	return r.result
}

// Return a step which expects an ip summary for the given device.
func (r *Registration) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{OutputRegexString, CommandCompleteRegexString},
		Timeout: r.timeout,
	}
}

// On match, parse the ip addr output and set the test result.
// Returns no step; the test is complete.
func (r *Registration) ReelMatch(pattern string, _ string, match string) *reel.Step {
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

// On timeout, do nothing;  no intervention is needed.
func (r *Registration) ReelTimeout() *reel.Step {
	return nil
}

// On eof, take no action.
func (r *Registration) ReelEof() {
}

// GetRegisteredNRFs returns the map of registered NRFID instances.
func (r *Registration) GetRegisteredNRFs() map[string]*NRFID {
	return r.registeredNRFs
}

// registrationCommand creates the Unix command to check for registration.
func registrationCommand(namespace string) ([]string, error) {
	command, err := utils.PrepareString(CheckRegistrationCommand, namespace, namespace)
	if err != nil {
		return nil, err
	}
	return strings.Split(command, " "), nil
}

// NewRegistration creates a Registration instance.
func NewRegistration(timeout time.Duration, namespace string) (*Registration, error) {
	preparedCommand, err := registrationCommand(namespace)
	if err != nil {
		return nil, err
	}
	return &Registration{result: tnf.ERROR, timeout: timeout, args: preparedCommand, namespace: namespace, registeredNRFs: map[string]*NRFID{}}, nil
}
