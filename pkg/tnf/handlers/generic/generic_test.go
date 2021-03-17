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

package generic_test

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

var (
	schemaPath = path.Join("..", "..", "..", "..", "schemas", generic.TestSchemaFileName)
)

// newGenericFromJSONFileTestCase defines input and expected values in order to exercise generic.Generic.
type newGenericFromJSONFileTestCase struct {

	// expectedCreationErr is whether the test will fail with a creation error.  This might be due to schema
	// non-conformance, invalid JSON, bad file location, etc.
	expectedCreationErr bool

	// expectedCreationErrText allows the user to define a text error message that is expected from creation errors.
	// A string is needed as it is the only easy way to compare various error implementations for equality.
	expectedCreationErrText string

	// expectedTester is whether the generated tester should be non-nil.
	expectedTester bool

	// expectedTimeout is the expected test timeout.
	expectedTimeout time.Duration

	// expectedHandlers is whether the generated handlers should be non-nil.
	expectedHandlers bool

	// expectedHandlersLen is the expected length of a non-nil handlers array.
	expectedHandlersLen int

	// expectedResultIsValid is whether the JSON Schema Validation result is valid or not.
	expectedResultIsValid bool

	// expectedReelTimeoutStep is the reel.Step expected for a reel.Handler ReelTimeout event.
	expectedReelTimeoutStep *reel.Step

	// expectedArgs is the expected tnf.Test Args array.
	expectedArgs []string

	// expectedReelFirstStep is the reel.Step expected for reel.Handler ReelFirst.
	expectedReelFirstStep *reel.Step
	// expectedInitialResult is the result of the tnf.Test prior to running.  Most tests should start out as tnf.ERROR.
	expectedInitialResult int

	// matchTestCases is a way of validating various reel.Handler ReelMatch test cases.
	matchTestCases []matchTestCase
}

// matchTestCase stores the inputs/outputs expected from feeding different results to a reel.Handler ReelMatch.
type matchTestCase struct {

	// inputPattern is the reel.Handler ReelMatch pattern.
	inputPattern string

	// inputBefore is the reel.Handler ReelMatch before string.
	inputBefore string

	// inputMatch is the reel.Handler ReelMatch match string.
	inputMatch string

	// expectedReelMatchNextStep is the reel.Step expected from running this matchTestCase through ReelMatch.
	expectedReelMatchNextStep *reel.Step

	// expectedFinalResult is the expected tnf.Test Result after running ReelMatch.
	expectedFinalResult int
}

var newGenericFromJSONFileTestCases = map[string]newGenericFromJSONFileTestCase{
	// Positive Test:  "testdata/base.json" is used to implement a basic base image test.  A number of sub-test-cases
	// are defined by matchTestCases.  Namely, we test:
	// 1) When RHEL 7.8 Maipo version is matched successfully.
	// 2) When Unknown Base Image is matched.
	// 3) When RHEL 10.10 Maipo version is matched, but doesn't live up to our imposed assertions (7.8).
	// This importantly tests the framework, and not given "base.json" test implementation.  "base.json" is arbitrary
	// and just used to validate the generic.Generic implementation.
	"base": {
		expectedCreationErr: false,
		expectedTester:      true,
		expectedTimeout:     time.Duration(2000000000),
		expectedHandlers:    true,
		expectedHandlersLen: 1,
		// This implementation returns the first command in ReelFirst().
		expectedArgs:            nil,
		expectedInitialResult:   tnf.ERROR,
		expectedResultIsValid:   true,
		expectedReelTimeoutStep: nil,
		expectedReelFirstStep: &reel.Step{
			Execute: "if [ -e /etc/redhat-release ]; then cat /etc/redhat-release; else echo \"Unknown Base Image\"; fi\n",
			Expect: []string{
				"(?m)Unknown Base Image",
				"(?m)Red Hat Enterprise Linux Server release (\\d+\\.\\d+) \\((\\w+)\\)",
				"(?m)contrived match",
			},
			Timeout: time.Duration(2000000000),
		},
		matchTestCases: []matchTestCase{
			// Positive Test:  The match is valid, and the expected version is correct (7.8)
			{
				inputPattern:              "(?m)Red Hat Enterprise Linux Server release (\\d+\\.\\d+) \\((\\w+)\\)",
				inputBefore:               "",
				inputMatch:                "Red Hat Enterprise Linux Server release 7.8 (Maipo)",
				expectedReelMatchNextStep: nil,
				expectedFinalResult:       tnf.SUCCESS,
			},
			// Positive Test:  The match is valid, and the container is not RHEL based.
			{
				inputPattern:              "(?m)Unknown Base Image",
				inputBefore:               "",
				inputMatch:                "Unknown Base Image",
				expectedReelMatchNextStep: nil,
				expectedFinalResult:       tnf.FAILURE,
			},
			// Positive Test:  The match is valid, but the derived version (10.10) is not expected (7.8)
			{
				inputPattern:              "(?m)Red Hat Enterprise Linux Server release (\\d+\\.\\d+) \\((\\w+)\\)",
				inputBefore:               "",
				inputMatch:                "Red Hat Enterprise Linux Server release 10.10 (Maipo)",
				expectedReelMatchNextStep: nil,
				expectedFinalResult:       tnf.FAILURE,
			},
			// Negative Test:  The match is invalid.
			{
				inputPattern:              "unknown pattern",
				inputBefore:               "",
				inputMatch:                "unknown pattern",
				expectedReelMatchNextStep: nil,
				expectedFinalResult:       tnf.ERROR,
			},
			// Positive Test:  A chained example.
			{
				inputPattern: "(?m)contrived match",
				inputBefore:  "",
				inputMatch:   "contrived match",
				expectedReelMatchNextStep: &reel.Step{
					Execute: "ls -al\n",
					Expect:  []string{"(?m).+"},
					Timeout: 2000000000,
				},
				expectedFinalResult: tnf.ERROR,
			},
		},
	},
	// Negative Test:  The supplied file doesn't exist, so make sure that an appropriate error is emitted.
	"file_does_not_exist": {
		expectedCreationErr:     true,
		expectedCreationErrText: "open testdata/file_does_not_exist.json: no such file or directory",
	},
	// Negative Test:  The supplied file doesn't validate against the generic-test.schema.json file.
	"test_schema_error": {
		expectedCreationErr:     true,
		expectedCreationErrText: "json: cannot unmarshal bool into Go struct field Generic.description of type string",
	},
	// Negative Test:  Garbage is supplied in the given file;  ensure that an appropriate error message is emitted.
	"not_json": {
		expectedCreationErr:     true,
		expectedCreationErrText: "invalid character 'h' in literal true (expecting 'r')",
	},
	// Positive Test:  The test input is all fine, but the actual tnf.Test makes an assertion that fails.  In this case,
	// we match "7.8" as groupIdx 1, and then assert that "7.8" is an integer.  Since it is a string, the test should
	// report tnf.Error (the types are incompatible).
	"assertion_error": {
		expectedCreationErr: false,
		expectedTester:      true,
		expectedTimeout:     time.Duration(2000000000),
		expectedHandlers:    true,
		expectedHandlersLen: 1,
		// This implementation returns the first command in ReelFirst().
		expectedArgs:            nil,
		expectedInitialResult:   tnf.ERROR,
		expectedResultIsValid:   true,
		expectedReelTimeoutStep: nil,
		expectedReelFirstStep: &reel.Step{
			Execute: "if [ -e /etc/redhat-release ]; then cat /etc/redhat-release; else echo \"Unknown Base Image\"; fi\n",
			Expect:  []string{"(?m)Red Hat Enterprise Linux Server release (\\d+\\.\\d+) \\((\\w+)\\)"},
			Timeout: time.Duration(2000000000),
		},
		matchTestCases: []matchTestCase{
			// Positive Test:  The match is valid, and the expected version is correct (7.8)
			{
				inputPattern:              "(?m)Red Hat Enterprise Linux Server release (\\d+\\.\\d+) \\((\\w+)\\)",
				inputBefore:               "",
				inputMatch:                "Red Hat Enterprise Linux Server release 7.8 (Maipo)",
				expectedReelMatchNextStep: nil,
				expectedFinalResult:       tnf.ERROR,
			},
		},
	},
}

func getTestFileLocation(testName string) string {
	return path.Join("testdata", testName+".json")
}

// TestGeneric tests all aspects of generic.Generic.
func TestGeneric(t *testing.T) {
	for testName, testCase := range newGenericFromJSONFileTestCases {
		testFile := getTestFileLocation(testName)
		tester, handlers, result, err := generic.NewGenericFromJSONFile(testFile, schemaPath)
		// this assertion also prevents `tester` from being `nil` inside the following `if`
		assert.Equal(t, testCase.expectedCreationErr, err != nil)
		if !testCase.expectedCreationErr {
			assert.Equal(t, testCase.expectedTester, tester != nil)
			assert.Equal(t, testCase.expectedTimeout, (*tester).Timeout()) //nolint:staticcheck

			assert.Equal(t, testCase.expectedHandlers, handlers != nil)
			if testCase.expectedHandlers {
				assert.Equal(t, testCase.expectedHandlersLen, len(handlers))
				firstHandler := handlers[0]

				assert.Equal(t, testCase.expectedArgs, (*tester).Args())
				assert.Equal(t, testCase.expectedInitialResult, (*tester).Result())
				assert.Equal(t, testCase.expectedReelTimeoutStep, firstHandler.ReelTimeout())

				// Test ReelFirst()
				actualReelFirst := firstHandler.ReelFirst()
				assert.Equal(t, testCase.expectedReelFirstStep, actualReelFirst)

				// Just ensure that ReelEOF doesn't cause a panic
				firstHandler.ReelEOF()

				// Test ReelMatch() cases.
				for _, reelMatchTestCase := range testCase.matchTestCases {
					actualReelMatchStep := firstHandler.ReelMatch(reelMatchTestCase.inputPattern,
						reelMatchTestCase.inputBefore, reelMatchTestCase.inputMatch)
					assert.Equal(t, reelMatchTestCase.expectedReelMatchNextStep, actualReelMatchStep)
					assert.Equal(t, reelMatchTestCase.expectedFinalResult, (*tester).Result())
				}
			}
			assert.NotNil(t, result)
			if result != nil {
				assert.Equal(t, testCase.expectedResultIsValid, result.Valid())
			}
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedCreationErrText, err.Error())
		}
	}
}
