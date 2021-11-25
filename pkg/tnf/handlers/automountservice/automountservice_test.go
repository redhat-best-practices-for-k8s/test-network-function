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

package automountservice_test

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	as "github.com/test-network-function/test-network-function/pkg/tnf/handlers/automountservice"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

const (
	// adding special variable
	testTimeoutDuration = time.Second * 2
	testServiceAccount  = "default"
	testNamespace       = "tnf"
	testPodname         = "test"
	testDataDirectory   = "testData"
	testDataFileSuffix  = ".yaml"
)

type testCase struct {
	token  int
	status int
	count  int
}

var testCases = map[string]testCase{
	"podFalse":  {as.TokenIsFalse, tnf.SUCCESS, 2},
	"podTrue":   {as.TokenIsTrue, tnf.SUCCESS, 2},
	"podNotSet": {as.TokenNotSet, tnf.SUCCESS, 0},
	"saFalse":   {as.TokenIsFalse, tnf.SUCCESS, 2},
	"saTrue":    {as.TokenIsTrue, tnf.SUCCESS, 2},
	"saNotSet":  {as.TokenNotSet, tnf.SUCCESS, 0},
}

func getMockOutputFilename(testName string) string {
	return path.Join(testDataDirectory, fmt.Sprintf("%s%s", testName, testDataFileSuffix))
}

func getMockOutput(t *testing.T, testName string) string {
	fileName := getMockOutputFilename(testName)
	b, err := os.ReadFile(fileName)
	assert.Nil(t, err)
	return string(b)
}

// Test_NewAutomountService is the unit test for NewAutomountService().
func Test_NewAutomountService(t *testing.T) {
	automount := as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithServiceAccount(testServiceAccount))
	assert.NotNil(t, automount)
	assert.Equal(t, automount.Result(), tnf.ERROR)
	// test creating with pod
	automount = as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithPodname(testPodname))
	assert.NotNil(t, automount)
	assert.Equal(t, automount.Result(), tnf.ERROR)
}

// Test_Automountservice_GetIdentifier is the unit test for Automountservice_GetIdentifier().
func TestAutomountservice_GetIdentifier(t *testing.T) {
	test := as.NewAutomountservice()
	assert.Equal(t, identifier.AutomountServiceIdentifier, test.GetIdentifier())
}

// Test_Automountservice_ReelEOF is the unit test for Automountservice_ReelEOF().
func TestAutomountservice_ReelEOF(t *testing.T) {
	test := as.NewAutomountservice()
	assert.NotNil(t, test)
	test.ReelEOF()
}

func Test_Automountservice_Args(t *testing.T) {
	test := as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithServiceAccount(testServiceAccount))
	args := []string{"oc", "-n", testNamespace, "get", "serviceaccounts", testServiceAccount, "-o", "json"}
	assert.ElementsMatch(t, args, test.Args())

	test = as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithPodname(testPodname))
	args = []string{"oc", "-n", testNamespace, "get", "pods", testPodname, "-o", "json", "|", "jq", "-r", ".spec"}
	assert.ElementsMatch(t, args, test.Args())
}

// Test_Automountservice_ReelTimeout is the unit test for Automountservice}_ReelTimeout().
func TestAutomountservice_ReelTimeout(t *testing.T) {
	test := as.NewAutomountservice(as.WithTimeout(testTimeoutDuration))
	assert.NotNil(t, test)
	assert.Equal(t, testTimeoutDuration, test.Timeout())
	test.ReelTimeout()
}

// Test_Automountservice_ReelMatch is the unit test for Automountservice_ReelMatch().
func TestAutomountservice_ReelMatch(t *testing.T) {
	for filename, testcase := range testCases {
		matchMock := getMockOutput(t, filename)
		test := as.NewAutomountservice()
		assert.NotNil(t, test)
		firstStep := test.ReelFirst()
		// validate regular expression when serviceaccount is set to false
		re := regexp.MustCompile(firstStep.Expect[0])
		matches := re.FindStringSubmatch(matchMock)
		fmt.Println("-----")
		for i := 0; i < len(matches); i++ {
			fmt.Println("i=", i, " matches= ", matches[i])
		}
		fmt.Println("-----")
		assert.Len(t, matches, testcase.count)
		if len(matches) == 0 {
			matches = []string{""}
		}
		step := test.ReelMatch("", "", matches[0])
		assert.Nil(t, step)
		assert.Equal(t, tnf.SUCCESS, test.Result())
		assert.Equal(t, test.Token(), testcase.token)
	}
}
