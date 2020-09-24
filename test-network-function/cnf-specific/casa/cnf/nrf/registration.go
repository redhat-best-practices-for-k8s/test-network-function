package nrf

import (
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/utils"
	"strings"
	"time"
)

const (
	// GetRegistrationNFStatusCmd gets the "nfStatus" for a particular CNF.
	GetRegistrationNFStatusCmd = "oc get -n %s nfregistrations.mgmt.casa.io %s %s -o jsonpath='{.items[*].spec.data}' | jq '.nfStatus'"
	// SuccessfulRegistrationOutputRegexString is the output regular expression expected when a CNF has successfully registered.
	SuccessfulRegistrationOutputRegexString = "(?m)\"REGISTERED\""
	// UnsuccessfulRegistrationOutputRegexString is the output regular expression expected when a CNF has not successfully registered.
	UnsuccessfulRegistrationOutputRegexString = "(?m)\"\""
)

// CheckRegistration checks whether a Casa CNF is registered.
type CheckRegistration struct {
	// nrf represents the underlying CNF information.
	nrf *NRFID
	// command is the Unix command to run to check the registration
	command []string
	// result is the result of the test.
	result int
	// timeout is the timeout of the test.
	timeout time.Duration
}

// Args returns the command to test that a CNF is registered.
func (c *CheckRegistration) Args() []string {
	return c.command
}

// Timeout returns the timeout of the test.
func (c *CheckRegistration) Timeout() time.Duration {
	return c.timeout
}

// Result returns the result of the test.
func (c *CheckRegistration) Result() int {
	return c.result
}

// ReelFirst returns the step that looks for whether a CNF is registered.
func (c *CheckRegistration) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{SuccessfulRegistrationOutputRegexString, UnsuccessfulRegistrationOutputRegexString},
		Timeout: c.timeout,
	}
}

// ReelMatch determines whether a CNF was successfully registered.  The returned result is nil since no further action
// is needed.
func (c *CheckRegistration) ReelMatch(pattern string, _ string, _ string) *reel.Step {
	if pattern == UnsuccessfulRegistrationOutputRegexString {
		c.result = tnf.FAILURE
	} else if pattern == SuccessfulRegistrationOutputRegexString {
		c.result = tnf.SUCCESS
	} else {
		c.result = tnf.ERROR
	}
	return nil
}

// ReelTimeout returns nil;  no further steps are required for a timeout.
func (c *CheckRegistration) ReelTimeout() *reel.Step {
	return nil
}

// ReelEof does nothing;  no further steps are required for EOF.
func (c *CheckRegistration) ReelEof() {
	// do nothing
}

// FormCheckRegistrationCmd forms the command to check that a CNF is registered.
func FormCheckRegistrationCmd(namespace string, nrfID *NRFID) ([]string, error) {
	command, err := utils.PrepareString(GetRegistrationNFStatusCmd, namespace, nrfID.nrf, nrfID.instID)
	if err != nil {
		return nil, err
	}
	return strings.Split(command, " "), nil
}

// NewCheckRegistration Creates a CheckRegistration tnf.Test.
func NewCheckRegistration(namespace string, timeout time.Duration, nrf *NRFID) (*CheckRegistration, error) {
	command, err := FormCheckRegistrationCmd(namespace, nrf)
	if err != nil {
		return nil, err
	}
	return &CheckRegistration{nrf: nrf, command: command, timeout: timeout, result: tnf.ERROR}, nil
}
