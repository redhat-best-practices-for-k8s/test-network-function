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

package nrf_test

import (
	"testing"
	"time"

	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf_specific/casa/cnf/nrf"
	"github.com/stretchr/testify/assert"
)

const (
	testNamespace = "default"
	testTimeout   = time.Second * 2
)

// TestNewCheckRegistration also tests Timeout and Result.
func TestNewCheckRegistration(t *testing.T) {
	cr := nrf.NewCheckRegistration(testNamespace, testTimeout, &nrf.ID{})
	assert.NotNil(t, cr)
	assert.Equal(t, testTimeout, cr.Timeout())
	assert.Equal(t, tnf.ERROR, cr.Result())
}

func TestCheckRegistration_Args(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.NotNil(t, cr)

	expected := []string{"oc", "get", "-n", "default", "nfregistrations.mgmt.casa.io", "nrf123",
		"0a0a2ede-be2e-40f0-8145-5ea6c565296e", "-o", "jsonpath='{.items[*].spec.data}'",
		"|", "jq", "'.nfStatus'"}
	assert.Equal(t, expected, nrf.FormCheckRegistrationCmd(testNamespace, nrfID))
	assert.Equal(t, expected, cr.Args())
}

func TestCheckRegistration_ReelFirst(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
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
	cr := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
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
	cr := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.NotNil(t, cr)

	step := cr.ReelTimeout()
	assert.Nil(t, step)
}

func TestCheckRegistration_ReelEof(t *testing.T) {
	nrfID := nrf.NewNRFID("nrf123", "AMF", "0a0a2ede-be2e-40f0-8145-5ea6c565296e", "REGISTERED")
	cr := nrf.NewCheckRegistration(testNamespace, testTimeout, nrfID)
	assert.NotNil(t, cr)

	// just ensures no panics.
	cr.ReelEOF()
}
