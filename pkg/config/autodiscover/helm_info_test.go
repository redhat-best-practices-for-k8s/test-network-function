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
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

const (
	testHelmChart = "testhelmchart.json"
)

var (
	testHelmChartPath = path.Join(filePath, testHelmChart)
)

//nolint:funlen
func TestGetClusterHelmCharts(t *testing.T) {
	testCases := []struct {
		badJSONUnmarshal bool
		jsonErr          error
	}{
		{ // no failures
			badJSONUnmarshal: false,
		},
		{ // failure to jsonUnmarshal
			badJSONUnmarshal: true,
			jsonErr:          errors.New("this is an error"),
		},
	}

	for _, tc := range testCases {
		// Setup the mock functions
		execCommandOutput = func(command string) string {
			contents, err := os.ReadFile(testHelmChartPath)
			assert.Nil(t, err)
			return string(contents)
		}
		if tc.badJSONUnmarshal {
			jsonUnmarshal = func(data []byte, v interface{}) error {
				return tc.jsonErr
			}
		} else {
			// use the "real" function
			jsonUnmarshal = json.Unmarshal
		}

		// Run the function and compare the list output
		list, err := GetClusterHelmCharts()
		if !tc.badJSONUnmarshal {
			assert.NotNil(t, list)
			assert.Equal(t, "my-test1", list.Items[0].Name)
		}

		if tc.badJSONUnmarshal {
			assert.NotNil(t, err)
			jsonUnmarshal = json.Unmarshal
		}
	}
}
