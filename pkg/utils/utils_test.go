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

package utils

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
)

const (
	testString1        = "{{\"Quoted line with new line\n char and also some others \b chars not commonly used like \f, \t, \\ and \r.\"}}"
	testEscapedString1 = `{{\"Quoted line with new line\\n char and also some others \u0008 chars not commonly used like \u000c, \t, \\ and \r.\"}}`
)

func TestEscapeToJSONstringFormat(t *testing.T) {
	escapedString, err := escapeToJSONstringFormat(testString1)
	assert.Nil(t, err)
	assert.Equal(t, testEscapedString1, escapedString)
}

func TestArgListToMap(t *testing.T) {
	testCases := []struct {
		argList     []string
		expectedMap map[string]string
	}{
		{
			argList: []string{"key1=value1", "key2=value2"},
			expectedMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			argList:     []string{},
			expectedMap: map[string]string{},
		},
		{
			argList: []string{"key1=value1", "key2=value2", "key3"},
			expectedMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "",
			},
		},
	}

	for _, tc := range testCases {
		assert.True(t, reflect.DeepEqual(tc.expectedMap, ArgListToMap(tc.argList)))
	}
}

func TestFilterArray(t *testing.T) {
	stringFilter := func(incomingVar string) bool {
		return strings.Contains(incomingVar, "test")
	}

	testCases := []struct {
		arrayToFilter []string
		expectedArray []string
	}{
		{
			arrayToFilter: []string{"test1", "test2"},
			expectedArray: []string{"test1", "test2"},
		},
		{
			arrayToFilter: []string{"apples", "oranges"},
			expectedArray: []string{},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedArray, FilterArray(tc.arrayToFilter, stringFilter))
	}
}

func TestAddNsenterPrefix(t *testing.T) {
	testCases := []struct {
		containerID    string
		expectedString string
	}{
		{
			containerID:    "1337",
			expectedString: `nsenter -t 1337 -n `,
		},
		{
			containerID:    "",
			expectedString: `nsenter -t  -n `,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedString, AddNsenterPrefix(tc.containerID))
	}
}

func TestModuleInTree(t *testing.T) {
	testCases := []struct {
		fakeOutput string
		isInTree   bool
	}{
		{
			fakeOutput: `filename:
			alias:
			version:
			license:
			srcversion:
			depends:
			retpoline:
			intree:
			name:
			vermagic:`,
			isInTree: true,
		},
		{
			fakeOutput: `filename:
			alias:
			version:
			license:
			srcversion:
			depends:
			retpoline:
			name:
			vermagic:`,
			isInTree: false,
		},
	}

	origFunc := RunCommandInNode
	defer func() {
		RunCommandInNode = origFunc
	}()
	for _, tc := range testCases {
		RunCommandInNode = func(nodeName string, nodeOc *interactive.Oc, command string, timeout time.Duration) string {
			return tc.fakeOutput
		}
		assert.Equal(t, tc.isInTree, ModuleInTree("testNode", "testModule", nil))
	}
}

func TestGetModulesFromNode(t *testing.T) {
	testCases := []struct {
		fakeOutput     string
		expectedOutput []string
	}{
		{
			fakeOutput: `xt_nat
			ip_vs_sh
			vboxsf
			vboxguest`,
			expectedOutput: []string{
				"xt_nat",
				"ip_vs_sh",
				"vboxsf",
				"vboxguest",
			},
		},
	}

	origFunc := RunCommandInNode
	defer func() {
		RunCommandInNode = origFunc
	}()
	for _, tc := range testCases {
		RunCommandInNode = func(nodeName string, nodeOc *interactive.Oc, command string, timeout time.Duration) string {
			return strings.TrimSpace(tc.fakeOutput)
		}
		assert.Equal(t, tc.expectedOutput, GetModulesFromNode("testNode", nil))
	}
}

//nolint:funlen
func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		testSlice       []string
		testString      string
		containsFeature bool
		expected        bool
	}{
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "apples",
			containsFeature: false,
			expected:        true,
		},
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "tacos",
			containsFeature: false,
			expected:        false,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: true, // Note: Turn 'on' the contains check
			expected:        true,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, StringInSlice(tc.testSlice, tc.testString, tc.containsFeature))
	}
}
