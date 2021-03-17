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
)

func TestNewOrBooleanLogic(t *testing.T) {
	logic := assertion.NewOrBooleanLogic()
	assert.Equal(t, assertion.OrBooleanLogicKey, logic.Type)
}

func TestOrBooleanLogic_Evaluate(t *testing.T) {
	for _, testCase := range andBooleanLogicTestCases {
		logic := assertion.NewOrBooleanLogic()
		actualResult, actualError := logic.Evaluate(testCase.assertions, testCase.match, &testCase.regex)
		assert.Equal(t, assertion.OrBooleanLogicKey, logic.Type)
		assert.Equal(t, testCase.expectedOrBooleanLogicResult, actualResult)
		assert.Equal(t, testCase.expectedOrBooleanLogicError, actualError != nil)
	}
}
