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

package assertion_test

import (
	"regexp"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/assertion"
)

type booleanLogicTestCase struct {
	assertions                    []assertion.Assertion
	match                         string
	regex                         regexp.Regexp
	expectedType                  string
	expectedAndBooleanLogicResult bool
	expectedAndBooleanLogicError  bool
	expectedOrBooleanLogicResult  bool
	expectedOrBooleanLogicError   bool
}

var andBooleanLogicTestCases = map[string]booleanLogicTestCase{
	"Positive Test:  single true assertion": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  0,
				Condition: &equalsSomeThingCondition,
			},
		},
		match:                         "Something",
		regex:                         *regexp.MustCompile(`Something`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: true,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  true,
		expectedOrBooleanLogicError:   false,
	},
	"Positive Test:  two true assertions": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1,
				Condition: &equalsSomeThingCondition,
			},
			{
				GroupIdx:  2,
				Condition: &equalsAnotherThingCondition,
			},
		},
		match:                         "Something AnotherThing",
		regex:                         *regexp.MustCompile(`(\w+) (\w+)`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: true,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  true,
		expectedOrBooleanLogicError:   false,
	},
	"Positive Test:  mixed assertion results": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1,
				Condition: &equalsSomeThingCondition,
			},
			{
				GroupIdx:  2,
				Condition: &equalsAnotherThingCondition,
			},
		},
		match:                         "Something NotAnotherThing",
		regex:                         *regexp.MustCompile(`(\w+) (\w+)`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: false,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  true,
		expectedOrBooleanLogicError:   false,
	},
	"Positive Test:  both assertions are false": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1,
				Condition: &equalsSomeThingCondition,
			},
			{
				GroupIdx:  2,
				Condition: &equalsAnotherThingCondition,
			},
		},
		match:                         "NotSomething NotAnotherThing",
		regex:                         *regexp.MustCompile(`(\w+) (\w+)`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: false,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  false,
		expectedOrBooleanLogicError:   false,
	},
	"Negative Test:  index out of bounds": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1000,
				Condition: &equalsSomeThingCondition,
			},
		},
		match:                         "Something",
		regex:                         *regexp.MustCompile(`Something`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: false,
		expectedAndBooleanLogicError:  true,
		expectedOrBooleanLogicResult:  false,
		expectedOrBooleanLogicError:   true,
	},
}
