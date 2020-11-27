// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package nrf_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/test-network-function/cnf-specific/casa/cnf/nrf"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"time"
)

const (
	defaultNamespace  = "default"
	testDataDirectory = "testdata"
	testDataSuffix    = ".txt"
)

var defaultTestTimeout = time.Second * 10

type genericRegistrationTestCase struct {
	timeout          time.Duration
	namespace        string
	expectedArgs     []string
	reelMatchPattern string
	expectedResult   int
	expectedNRFUuids []string
}

var genericRegistrationTestCases = map[string]*genericRegistrationTestCase{
	"real_test_data": {
		timeout:          defaultTestTimeout,
		namespace:        defaultNamespace,
		expectedArgs:     []string{"oc", "-n", "default", "get", "nfregistrations.mgmt.casa.io", "$(oc", "-n", "default", "get", "nfregistrations.mgmt.casa.io", "|", "awk", "{\"print", "$2\"}", "|", "xargs", "-n", "1)"},
		reelMatchPattern: nrf.OutputRegexString,
		expectedResult:   tnf.SUCCESS,
		expectedNRFUuids: []string{
			"9b971c6d-12ed-48ee-9ed7-bb99ee4d0b99",
			"a9bb40c7-bddf-443b-bf64-89ae54812ef9",
		},
	},
	"default_test_case": {
		timeout:          defaultTestTimeout,
		namespace:        defaultNamespace,
		expectedArgs:     []string{"oc", "-n", "default", "get", "nfregistrations.mgmt.casa.io", "$(oc", "-n", "default", "get", "nfregistrations.mgmt.casa.io", "|", "awk", "{\"print", "$2\"}", "|", "xargs", "-n", "1)"},
		reelMatchPattern: nrf.OutputRegexString,
		expectedResult:   tnf.SUCCESS,
		expectedNRFUuids: []string{
			"77647705-bf24-4b0d-a754-4cb7df86d169",
			"315889b0-4640-4de3-b41f-4b77643d1a9d",
			"549328de-4ead-4dac-90ca-4d8d09af3938",
			"f3d51461-83f4-4bfc-b20e-43297dd4404d",
			"0a0a2ede-be2e-40f0-8145-5ea6c565296e",
		},
	},
	"no_nrfs_registered": {
		timeout:          time.Hour * 24,
		namespace:        "someOtherNamespace",
		expectedArgs:     []string{"oc", "-n", "someOtherNamespace", "get", "nfregistrations.mgmt.casa.io", "$(oc", "-n", "someOtherNamespace", "get", "nfregistrations.mgmt.casa.io", "|", "awk", "{\"print", "$2\"}", "|", "xargs", "-n", "1)"},
		reelMatchPattern: nrf.CommandCompleteRegexString,
		expectedResult:   tnf.FAILURE,
		expectedNRFUuids: []string{},
	},
	"incorrect_match": {
		timeout:          defaultTestTimeout,
		namespace:        defaultNamespace,
		expectedArgs:     []string{"oc", "-n", "default", "get", "nfregistrations.mgmt.casa.io", "$(oc", "-n", "default", "get", "nfregistrations.mgmt.casa.io", "|", "awk", "{\"print", "$2\"}", "|", "xargs", "-n", "1)"},
		reelMatchPattern: "some_random_match_pattern",
		expectedResult:   tnf.ERROR,
		expectedNRFUuids: []string{},
	},
}

func getTestOutputFileName(name string) string {
	return fmt.Sprintf("%s%s", name, testDataSuffix)
}

func getTestOutputFile(name string) string {
	return path.Join(testDataDirectory, getTestOutputFileName(name))
}

func getTestOutputContents(name string) (string, error) {
	b, err := ioutil.ReadFile(getTestOutputFile(name))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TestNewRegistration also tests Registration.Timeout and Registration.Args
func TestNewRegistration(t *testing.T) {
	for _, testCase := range genericRegistrationTestCases {
		r := nrf.NewRegistration(testCase.timeout, testCase.namespace)
		assert.NotNil(t, r)
		assert.Equal(t, testCase.timeout, r.Timeout())
		assert.Equal(t, strings.Split(fmt.Sprintf(nrf.CheckRegistrationCommand, testCase.namespace, testCase.namespace), " "), r.Args())
	}
}

func TestRegistration_ReelFirst(t *testing.T) {
	for _, testCase := range genericRegistrationTestCases {
		r := nrf.NewRegistration(testCase.timeout, testCase.namespace)
		assert.NotNil(t, r)
		step := r.ReelFirst()
		assert.NotNil(t, step)
		assert.Equal(t, 2, len(step.Expect))
		assert.Equal(t, step.Expect[0], nrf.OutputRegexString)
		assert.Equal(t, step.Expect[1], nrf.CommandCompleteRegexString)
		assert.Equal(t, testCase.timeout, step.Timeout)
		assert.Equal(t, "", step.Execute)
	}
}

func TestRegistration_ReelMatch(t *testing.T) {
	for testCaseName, testCase := range genericRegistrationTestCases {
		r := nrf.NewRegistration(testCase.timeout, testCase.namespace)
		testCaseOutput, err := getTestOutputContents(testCaseName)
		assert.Nil(t, err)
		assert.NotNil(t, testCaseOutput)
		step := r.ReelMatch(testCase.reelMatchPattern, "", testCaseOutput)
		assert.Nil(t, step)
		assert.Equal(t, testCase.expectedResult, r.Result())
		for _, expectedUuid := range testCase.expectedNRFUuids {
			assert.NotNil(t, r.GetRegisteredNRFs()[expectedUuid])
		}
	}
}

func TestRegistration_ReelTimeout(t *testing.T) {
	r := nrf.NewRegistration(defaultTestTimeout, defaultNamespace)
	s := r.ReelTimeout()
	assert.Nil(t, s)
}

func TestRegistration_ReelEof(t *testing.T) {
	r := nrf.NewRegistration(defaultTestTimeout, defaultNamespace)
	// Just ensure it doesn't panic.
	r.ReelEOF()
}
