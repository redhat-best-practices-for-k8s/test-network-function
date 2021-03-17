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

package stringcondition_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/condition/stringcondition"
)

func TestNewEqualsCondition(t *testing.T) {
	c := stringcondition.NewEqualsCondition("input")
	assert.Equal(t, stringcondition.EqualsConditionKey, c.Type)
	assert.Equal(t, "input", c.Expected)
}

type equalsConditionTestCase struct {
	expected       string
	match          string
	regex          regexp.Regexp
	matchIdx       int
	expectedType   string
	expectedResult bool
	expectedError  bool
}

var equalsCondtionTestCases = map[string]equalsConditionTestCase{
	"Positive Case: strings_are_equal": {
		expected:       "apple",
		match:          "apple tree",
		regex:          *regexp.MustCompile(`(\w+) tree`),
		matchIdx:       1,
		expectedType:   stringcondition.EqualsConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Case: strings_are_not_equal": {
		expected:       "apple",
		match:          "notapple tree",
		regex:          *regexp.MustCompile(`(\w+) tree`),
		matchIdx:       1,
		expectedType:   stringcondition.EqualsConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Negative Case: index_out_of_bounds": {
		expected:       "apple",
		match:          "apple tree",
		regex:          *regexp.MustCompile(`(\w+) tree`),
		matchIdx:       100,
		expectedType:   stringcondition.EqualsConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
}

func TestEqualsCondition_Evaluate(t *testing.T) {
	for _, testCase := range equalsCondtionTestCases {
		c := stringcondition.NewEqualsCondition(testCase.expected)
		actualResult, actualError := c.Evaluate(testCase.match, &testCase.regex, testCase.matchIdx)
		assert.Equal(t, testCase.expectedType, c.Type)
		assert.Equal(t, testCase.expectedResult, actualResult)
		assert.Equal(t, testCase.expectedError, actualError != nil)
	}
}
