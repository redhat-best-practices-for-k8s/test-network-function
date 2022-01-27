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

package autodiscover

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

func TestFindDebugPods(t *testing.T) {
	testCases := []struct {
		jsonFileName           string
		expectedPodName        string
		expectedContainerName  string
		expectedDebugPodAmount int
		// expectedContainersDebugList []configsections.Container
	}{
		{
			jsonFileName:           "testdata/pods_with_debug_label.json",
			expectedPodName:        "test-7dc8cf6b5f-2t4bn",
			expectedContainerName:  "test",
			expectedDebugPodAmount: 1,
		},
		{
			jsonFileName:           "testdata/empty.json",
			expectedDebugPodAmount: 0,
		},
	}

	// Spoof the executeOcGetCommand
	origFunc := executeOcGetCommand
	defer func() {
		executeOcGetCommand = origFunc
	}()

	for _, tc := range testCases {
		tp := &configsections.TestPartner{}

		executeOcGetCommand = func(resourceType, labelQuery, namespace string) string {
			output, _ := os.ReadFile(tc.jsonFileName)
			return string(output)
		}

		if tc.expectedDebugPodAmount > 0 {
			FindDebugPods(tp)
			assert.Len(t, tp.ContainersDebugList, 1) // Only assuming one debug pod in the test YAML
			assert.Equal(t, tc.expectedPodName, tp.ContainersDebugList[0].PodName)
			assert.Equal(t, tc.expectedContainerName, tp.ContainersDebugList[0].ContainerName)
		} else {
			assert.Panics(t, func() { FindDebugPods(tp) })
		}
	}
}
