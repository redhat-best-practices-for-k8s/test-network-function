package nrf_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/casa/cnf/nrf"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testNamespace = "default"
	testTimeout   = time.Second * 2
)

// TestNewCheckRegistration also tests Timeout and Result.
func TestNewCheckRegistration(t *testing.T) {
	cr, err := nrf.NewCheckRegistration(testNamespace, testTimeout, &nrf.NRFID{})
	assert.Nil(t, err)
	assert.NotNil(t, cr)
	assert.Equal(t, testTimeout, cr.Timeout())
	assert.Equal(t, tnf.ERROR, cr.Result())
}

func TestCheckRegistration_Args(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr, err := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.Nil(t, err)
	assert.NotNil(t, cr)

	expected := []string{"oc", "get", "-n", "default", "nfregistrations.mgmt.casa.io", "nrf123", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "-o", "jsonpath='{.items[*].spec.data}'", "|", "jq", "'.nfStatus'"}
	actualCommand, err := nrf.FormCheckRegistrationCmd(testNamespace, nrfID)
	assert.Nil(t, actualCommand)
	assert.Equal(t, expected, actualCommand)
	assert.Equal(t, expected, cr.Args())
}

func TestCheckRegistration_ReelFirst(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr, err := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.Nil(t, err)
	assert.NotNil(t, cr)

	step := cr.ReelFirst()
	assert.NotNil(t, step)
	assert.Equal(t, testTimeout, step.Timeout)
	assert.Equal(t, "", step.Execute)
	assert.Contains(t, step.Expect, nrf.SuccessfulRegistrationOutputRegexString)
	assert.Contains(t, step.Expect, nrf.UnsuccessfulRegistrationOutputRegexString)
}

func TestCheckRegistration_ReelMatch(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr, err := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.Nil(t, err)
	assert.NotNil(t, cr)

	step := cr.ReelMatch(nrf.SuccessfulRegistrationOutputRegexString, "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, cr.Result())

	step = cr.ReelMatch(nrf.UnsuccessfulRegistrationOutputRegexString, "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, cr.Result())

	step = cr.ReelMatch("", "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, cr.Result())
}

func TestCheckRegistration_ReelTimeout(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr, err := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.Nil(t, err)
	assert.NotNil(t, cr)

	step := cr.ReelTimeout()
	assert.Nil(t, step)
}

func TestCheckRegistration_ReelEof(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr, err := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.Nil(t, err)
	assert.NotNil(t, cr)

	// just ensures no panics.
	cr.ReelEof()
}
