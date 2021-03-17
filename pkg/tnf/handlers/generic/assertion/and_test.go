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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/assertion"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/condition"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/condition/stringcondition"
)

var (
	equalsAnotherThingStringCondition = stringcondition.NewEqualsCondition("AnotherThing")
	equalsSomethingStringCondition    = stringcondition.NewEqualsCondition("Something")
)

var equalsAnotherThingCondition condition.Condition = equalsAnotherThingStringCondition
var equalsSomeThingCondition condition.Condition = equalsSomethingStringCondition

func TestNewAndBooleanLogic(t *testing.T) {
	logic := assertion.NewAndBooleanLogic()
	assert.Equal(t, assertion.AndBooleanLogicKey, logic.Type)
}

func TestAndBooleanLogic_Evaluate(t *testing.T) {
	for _, testCase := range andBooleanLogicTestCases {
		logic := assertion.NewAndBooleanLogic()
		actualResult, actualError := logic.Evaluate(testCase.assertions, testCase.match, &testCase.regex)
		assert.Equal(t, assertion.AndBooleanLogicKey, logic.Type)
		assert.Equal(t, testCase.expectedAndBooleanLogicResult, actualResult)
		assert.Equal(t, testCase.expectedAndBooleanLogicError, actualError != nil)
	}
}
