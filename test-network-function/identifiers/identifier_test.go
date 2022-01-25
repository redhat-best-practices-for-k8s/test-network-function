// Copyright (C) 2022 Red Hat, Inc.
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

package identifiers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

func TestGetSuiteAndTestFromIdentifier(t *testing.T) {
	testCases := []struct {
		testIdentifier claim.Identifier
		expectedResult []string
	}{
		{
			testIdentifier: claim.Identifier{
				Url: "http://test-network-function.com/testcases/SuiteName/TestName/MyTest",
			},
			expectedResult: []string{"SuiteName", "TestName", "MyTest"},
		},
		{ // invalid formatting
			testIdentifier: claim.Identifier{
				Url:     "testURL",
				Version: "testVersion",
			},
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, GetSuiteAndTestFromIdentifier(tc.testIdentifier))
	}
}

func TestXformToGinkgoItIdentifier(t *testing.T) {
	testCases := []struct {
		testIdentifier claim.Identifier
		expectedResult string
	}{
		{
			testIdentifier: claim.Identifier{
				Url: "http://test-network-function.com/testcases/SuiteName/TestName/MyTest",
			},
			expectedResult: "SuiteName-TestName-MyTest",
		},
		{ // invalid formatting
			testIdentifier: claim.Identifier{
				Url:     "testURL",
				Version: "testVersion",
			},
			expectedResult: "testURL",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, XformToGinkgoItIdentifier(tc.testIdentifier))
	}
}

func TestXformToGinkgoItIdentifierExtended(t *testing.T) {
	testCases := []struct {
		testIdentifier claim.Identifier
		expectedResult string
	}{
		{
			testIdentifier: claim.Identifier{
				Url: "http://test-network-function.com/testcases/SuiteName/TestName/MyTest",
			},
			expectedResult: "SuiteName-TestName-MyTest-extra",
		},
		{ // invalid formatting
			testIdentifier: claim.Identifier{
				Url:     "testURL",
				Version: "testVersion",
			},
			expectedResult: "testURL-extra",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, XformToGinkgoItIdentifierExtended(tc.testIdentifier, "extra"))
	}
}
