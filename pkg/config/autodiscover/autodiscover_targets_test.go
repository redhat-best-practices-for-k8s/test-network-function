// Copyright (C) 2021 Red Hat, Inc.
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

package autodiscover

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

//nolint:funlen
func TestFindTestCrdNames(t *testing.T) {
	// Simple function to check existence of a string in a slice
	contains := func(s []string, e string) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}

	testCases := []struct {
		crdFilters   []configsections.CrdFilter
		badUnmarshal bool
		expectedCRDs []string
	}{
		{
			crdFilters: []configsections.CrdFilter{
				{
					NameSuffix: "metal3.io",
				},
			},
			expectedCRDs: []string{
				"provisionings.metal3.io",
				"baremetalhosts.metal3.io",
			},
			badUnmarshal: false,
		},
		{
			crdFilters: []configsections.CrdFilter{
				{
					NameSuffix: "k8s.io",
				},
			},
			expectedCRDs: []string{
				"storagestates.migration.k8s.io",
				"storageversionmigrations.migration.k8s.io",
			},
			badUnmarshal: false,
		},
		{ // fail to unmarshal the JSON correctly
			crdFilters: []configsections.CrdFilter{
				{
					NameSuffix: "k8s.io",
				},
			},
			expectedCRDs: []string{},
			badUnmarshal: true,
		},
	}

	// Spoof the executeCommand func
	origFunc := utils.ExecuteCommandAndValidate
	utils.ExecuteCommandAndValidate = func(command string, timeout time.Duration, context *interactive.Context, failureCallbackFun func()) string {
		fileContents, err := os.ReadFile("testdata/crd_output.json")
		assert.Nil(t, err)
		return string(fileContents)
	}

	for _, tc := range testCases {
		if tc.badUnmarshal {
			jsonUnmarshal = func(data []byte, v interface{}) error {
				return errors.New("this is an error")
			}
		}

		// Compare the expected to the actual
		output := FindTestCrdNames(tc.crdFilters)
		for _, i := range tc.expectedCRDs {
			assert.True(t, contains(output, i))
		}
	}

	utils.ExecuteCommandAndValidate = origFunc
	jsonUnmarshal = json.Unmarshal
}
