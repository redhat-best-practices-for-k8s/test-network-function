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
	"encoding/json"
	"io/ioutil"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/assertion"
)

type assertionsTestCase struct {
	match                        string
	regex                        regexp.Regexp
	expectedUnmarshalError       bool
	expectedUnmarshalErrorString string
	expectedEvaluationResult     bool
	expectedEvaluationError      bool
}

var assertionsTestCases = map[string]assertionsTestCase{

	// Positive Test:  A bunch of assertions "and"-ed together.
	"and_composed_assertions_positive_test": {
		match:                    "iperf 1",
		regex:                    *regexp.MustCompile(`(\w+)\s(\d+)`),
		expectedUnmarshalError:   false,
		expectedEvaluationResult: true,
		expectedEvaluationError:  false,
	},

	// Positive Test:  A bunch of assertions, some of which are false, "or"-ed together.
	"or_composed_assertions_positive_test": {
		match:                    "iperf 1",
		regex:                    *regexp.MustCompile(`(\w+)\s(\d+)`),
		expectedUnmarshalError:   false,
		expectedEvaluationResult: true,
		expectedEvaluationError:  false,
	},

	// Negative Test:  When bad JSON is given.
	"not_json": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "invalid character 'n' looking for beginning of object key string",
	},

	// Negative Test:  When the "logic" key is not defined at all.
	"missing_logic": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "mandatory \"logic\" key is missing",
	},

	// Negative Test:  When the "logic" payload is incorrect.
	"logic_incorrect_type": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "json: cannot unmarshal string into Go value of type map[string]*json.RawMessage",
	},

	// Negative Test:  When the "logic"'s "type" key is not defined.
	"logic_type_missing": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "mandatory \"type\" key is missing",
	},

	// Negative Test:  When the "logic"'s "type" payload is incorrect.
	"logic_type_key_incorrect_type": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "json: cannot unmarshal bool into Go value of type string",
	},

	// Negative Test:  When non-array "assertions" is provided.
	"assertions_incorrect_type": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "json: cannot unmarshal string into Go value of type []assertion.Assertion",
	},

	// Negative Test:  When groupIdx is missing.
	"groupIdx_missing": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "required field \"groupIdx\" is missing from the JSON payload",
	},

	// Negative Test:  When a non-int groupIdx is provided.
	"groupIdx_incorrect_type": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "json: cannot unmarshal string into Go value of type int",
	},

	// Negative Test:  When conditionType is missing.
	"condition_type_missing": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "condition missing \"type\"",
	},

	// Negative Test:  When conditionType is the incorrect type.
	"condition_type_incorrect_type": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "json: cannot unmarshal bool into Go value of type string",
	},

	// Negative Test:  When conditionType refers to an undefined type.
	"condition_type_does_not_exist": {
		expectedUnmarshalError:       true,
		expectedUnmarshalErrorString: "unrecognized condition type: \"this condition type does not exist, and is for test purposes only\"",
	},
}

// getTestFile returns the file location of the test identified by testName.
func getTestFile(testName string) string {
	return path.Join("testdata", testName+".json")
}

// TestAssertions_UnmarshalJSON also tests Assertion.UnmarshalJSON.
func TestAssertions_UnmarshalJSON(t *testing.T) {
	for testName, testCase := range assertionsTestCases {
		contents, err := ioutil.ReadFile(path.Join(getTestFile(testName)))
		assert.Nil(t, err)
		assert.NotNil(t, contents)

		assertions := &assertion.Assertions{}
		err = json.Unmarshal(contents, assertions)
		if testCase.expectedUnmarshalError {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedUnmarshalErrorString, err.Error())
		} else {
			result, err := (*assertions.Logic).Evaluate(assertions.Assertions, testCase.match, &testCase.regex)
			assert.Equal(t, testCase.expectedEvaluationError, err != nil)
			assert.Equal(t, testCase.expectedEvaluationResult, result)
		}
	}
}
