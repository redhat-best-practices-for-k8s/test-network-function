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
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/assertion"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/condition"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/condition/intcondition"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

var testFileName = path.Join("testdata", "nested-result-context-marshal-expected.json")

// getContrivedCondition assembles a contrived example returning the expected interface type.  This is because Go is
// incapable of resolving an implementation without an explicit cast.
func getContrivedCondition() *condition.Condition {
	// The following is needed to achieve casting a ComparisonCondition as a condition.Condition
	intEqualsOneIntCondition := intcondition.ComparisonCondition{
		Type:       "intComparison",
		Input:      1,
		Comparison: "==",
	}
	var intEqualsOneCondition condition.Condition = intEqualsOneIntCondition
	return &intEqualsOneCondition
}

// getContrivedBooleanLogic assembles a contrived example returning the expected interface type.  This is because Go is
// incapable of resolving an implementation without an explicit cast.
func getContrivedBooleanLogic() *assertion.BooleanLogic {
	andBooleanLogic := &assertion.AndBooleanLogic{Type: assertion.AndBooleanLogicKey}
	var booleanLogic assertion.BooleanLogic = andBooleanLogic
	return &booleanLogic
}

// TestResultContext_MarshalJSON tests the custom ResultContext.MarshalJSON implementation.  The easiest way to do this
// involved creating a rendered version of the expected JSON for a short example, invoking the ResultContext.MarshalJSON
// function, and comparing the result.
func TestResultContext_MarshalJSON(t *testing.T) {
	intEqualsOneCondition := getContrivedCondition()
	booleanLogic := getContrivedBooleanLogic()

	resultContext := &generic.ResultContext{
		Pattern: `(\d)+`,
		ComposedAssertions: []assertion.Assertions{
			{
				Assertions: []assertion.Assertion{
					{
						GroupIdx:  1,
						Condition: intEqualsOneCondition,
					},
				},
				Logic: booleanLogic,
			},
		},
		DefaultResult: 2,
		NextStep: &reel.Step{
			Execute: "echo 2",
			Expect:  []string{`(\d+)`},
		},
		// Triggers the recursive definition, the whole reason we needed a custom ResultContext.MarshalJSON.
		NextResultContexts: []*generic.ResultContext{
			{
				Pattern:       `(\d)+`,
				DefaultResult: 0,
			},
		},
	}

	actualContents, err := json.MarshalIndent(resultContext, "", "  ")
	assert.Nil(t, err)
	// Compare against an expected rendering which has been pre-verified.
	expectedContents, err := ioutil.ReadFile(testFileName)
	assert.Nil(t, err)
	assert.Equal(t, string(expectedContents), string(actualContents))
}
