// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

const (
	testTimeoutDuration = 10 * time.Second
	testNamespace       = "default"
	testDebugDaemonset  = "debug"
	testDataDirectory   = "testdata"
)

type DaemonSetTest struct {
	daemonset Status
	result    int
}

var testCases = map[string]DaemonSetTest{
	"valid_daemonset": {daemonset: Status{
		Name:         "debug",
		Desired:      1,
		Current:      1,
		Ready:        1,
		Available:    1,
		Misscheduled: 0,
	}, result: tnf.SUCCESS},
	"non_valid_daemonset": {daemonset: Status{
		Name:         "test",
		Desired:      2,
		Current:      1,
		Ready:        1,
		Available:    1,
		Misscheduled: 0,
	}, result: tnf.SUCCESS},
	"non_valid_output": {daemonset: Status{
		Name:         "",
		Desired:      0,
		Current:      0,
		Ready:        0,
		Available:    0,
		Misscheduled: 0,
	}, result: tnf.ERROR},
}

func getMockOutputFilename(testName string) string {
	return path.Join(testDataDirectory, testName)
}

func getMockOutput(t *testing.T, testName string) string {
	fileName := getMockOutputFilename(testName)
	b, err := os.ReadFile(fileName)
	assert.Nil(t, err)
	return string(b)
}

// Test_NewDaemonSet is the unit test for NewDaemonSet().
func Test_NewDaemonSet(t *testing.T) {
	newDs := NewDaemonSet(testTimeoutDuration, testDebugDaemonset, testNamespace)
	assert.NotNil(t, newDs)
	assert.Equal(t, testTimeoutDuration, newDs.Timeout())
	assert.Equal(t, newDs.Result(), tnf.ERROR)
	assert.NotNil(t, newDs.GetStatus())
}

// Test_DaemonSet_GetIdentifier is the unit test for DaemonSet_GetIdentifier().
func TestDaemonSet_GetIdentifier(t *testing.T) {
	newDs := NewDaemonSet(testTimeoutDuration, testDebugDaemonset, testNamespace)
	assert.Equal(t, identifier.DaemonSetIdentifier, newDs.GetIdentifier())
}

// Test_DaemonSet_ReelEOF is the unit test for DaemonSet_ReelEOF().
func TestDaemonSet_ReelEOF(t *testing.T) {
	newDs := NewDaemonSet(testTimeoutDuration, testDebugDaemonset, testNamespace)
	assert.NotNil(t, newDs)
	newDs.ReelEOF()
}

// Test_DaemonSet_ReelTimeout is the unit test for DaemonSet}_ReelTimeout().
func TestDaemonSet_ReelTimeout(t *testing.T) {
	ds := NewDaemonSet(testTimeoutDuration, "debug", "default")
	step := ds.ReelTimeout()
	assert.Nil(t, step)
}

// Test_DaemonSet_ReelMatch is the unit test for DaemonSet_ReelMatch().
func TestDaemonSet_ReelMatch(t *testing.T) {
	for testName, testCase := range testCases {
		fmt.Println("process case ", testName)
		ds := NewDaemonSet(testTimeoutDuration, testCase.daemonset.Name, "default")
		matchMock := getMockOutput(t, testName)
		step := ds.ReelMatch("", "", matchMock)
		assert.Nil(t, step)
		assert.Equal(t, testCase.daemonset, ds.GetStatus())
		assert.Equal(t, testCase.result, ds.result)
	}
}
