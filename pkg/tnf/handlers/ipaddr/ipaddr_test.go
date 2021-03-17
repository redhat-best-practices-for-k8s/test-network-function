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

package ipaddr_test

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

const (
	testDataDirectory   = "testdata"
	testDataFileSuffix  = ".txt"
	testTimeoutDuration = time.Second * 2
)

type TestCase struct {
	device              string
	pattern             string
	expectedResult      int
	expectedIpv4Address string
}

var testCases = map[string]TestCase{
	"device_exists": {
		device:              "eth0",
		pattern:             ipaddr.SuccessfulOutputRegex,
		expectedResult:      tnf.SUCCESS,
		expectedIpv4Address: "172.17.0.7",
	},
	"device_does_not_exist": {
		device:              "dne",
		pattern:             ipaddr.DeviceDoesNotExistRegex,
		expectedResult:      tnf.ERROR,
		expectedIpv4Address: "",
	},
}

func getMockOutputFilename(testName string) string {
	return path.Join(testDataDirectory, fmt.Sprintf("%s%s", testName, testDataFileSuffix))
}

func getMockOutput(t *testing.T, testName string) string {
	fileName := getMockOutputFilename(testName)
	b, err := ioutil.ReadFile(fileName)
	assert.Nil(t, err)
	return string(b)
}

func TestNewIpAddr(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		assert.NotNil(t, ipAddr)
		assert.Equal(t, tnf.ERROR, ipAddr.Result())
		assert.Equal(t, []string{"ip", "addr", "show", "dev", testCase.device}, ipAddr.Args())
	}
}

func TestIpAddr_Args(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, []string{"ip", "addr", "show", "dev", testCase.device}, ipAddr.Args())
	}
}

func TestIPAddr_GetIdentifier(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, identifier.IPAddrIdentifier, ipAddr.GetIdentifier())
	}
}

func TestIpAddr_Timeout(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, testTimeoutDuration, ipAddr.Timeout())
	}
}

func TestIpAddr_ReelFirst(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		step := ipAddr.ReelFirst()
		assert.Equal(t, "", step.Execute)
		assert.Contains(t, step.Expect, ipaddr.SuccessfulOutputRegex)
		assert.Equal(t, testTimeoutDuration, step.Timeout)
	}
}

func TestIpAddr_Result(t *testing.T) {
	for testName, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		assert.Equal(t, tnf.ERROR, ipAddr.Result())
		step := ipAddr.ReelMatch(testCase.pattern, "", getMockOutput(t, testName))
		assert.Nil(t, step)
		assert.Equal(t, testCase.expectedResult, ipAddr.Result())
	}
}

func TestIpAddr_GetIpv4Address(t *testing.T) {
	for testName, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		step := ipAddr.ReelMatch(testCase.pattern, "", getMockOutput(t, testName))
		assert.Nil(t, step)
		assert.Equal(t, testCase.expectedIpv4Address, ipAddr.GetIPv4Address())
	}
}

func TestIpAddr_ReelTimeout(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		assert.Nil(t, ipAddr.ReelTimeout())
	}
}

// Ensure there are no panics.
func TestIpAddr_ReelEof(t *testing.T) {
	for _, testCase := range testCases {
		ipAddr := ipaddr.NewIPAddr(testTimeoutDuration, testCase.device)
		ipAddr.ReelEOF()
	}
}
