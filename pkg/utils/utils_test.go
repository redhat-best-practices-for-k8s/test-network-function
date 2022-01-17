// Copyright (C) 2021-2022 Red Hat, Inc.
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

	"github.com/stretchr/testify/assert"
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

func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		testString     string
		testSlice      []string
		expectedExists bool
	}{
		{
			testString: "apples",
			testSlice: []string{
				"bananas",
				"oranges",
				"apples",
			},
			expectedExists: true,
		},
		{
			testString: "tacos",
			testSlice: []string{
				"burritos",
				"enchiladas",
			},
			expectedExists: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedExists, StringInSlice(tc.testSlice, tc.testString))
	}
}
