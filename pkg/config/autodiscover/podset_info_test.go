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
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	testDeploymentFile = "testdeployment.json"
	testJQFile         = "testdeploy.json"
)

var (
	testDeploymentFilePath = path.Join(filePath, testDeploymentFile)
	testJQFilePath         = path.Join(filePath, testJQFile)
)

func loadDeployment(filePath string) (deployment PodSetResource) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error (%s) loading PodSetResource %s for testing", err, filePath)
	}
	err = jsonUnmarshal(contents, &deployment)
	if err != nil {
		log.Fatalf("error (%s) unmarshalling PodSetResource %s for testing", err, filePath)
	}
	return
}

func TestPodGetAnnotationValue1(t *testing.T) {
	deployment := loadDeployment(testDeploymentFilePath)

	assert.Equal(t, "test", deployment.GetName())
	assert.Equal(t, "tnf", deployment.GetNamespace())
	assert.Equal(t, 2, deployment.GetReplicas())

	labels := deployment.GetLabels()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "test", labels["app"])
}

//nolint:funlen
func TestGetTargetDeploymentByNamespace(t *testing.T) {
	testCases := []struct {
		badExec          bool
		execErr          error
		badJSONUnmarshal bool
		jsonErr          error
	}{
		{ // no failures
			badExec:          false,
			badJSONUnmarshal: false,
		},
		{ // failure to exec
			badExec:          true,
			execErr:          errors.New("this is an error"),
			badJSONUnmarshal: false,
			jsonErr:          nil,
		},
		{ // failure to jsonUnmarshal
			badExec:          false,
			execErr:          nil,
			badJSONUnmarshal: true,
			jsonErr:          errors.New("this is an error"),
		},
	}

	origExecFunc := execCommandOutput

	for _, tc := range testCases {
		// Setup the mock functions
		if tc.badExec {
			execCommandOutput = func(command string) string {
				return ""
			}
		} else {
			execCommandOutput = func(command string) string {
				contents, err := os.ReadFile(testJQFilePath)
				assert.Nil(t, err)
				return string(contents)
			}
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
		list, err := GetTargetPodSetsByNamespace("test", configsections.Label{
			Prefix: "prefix1",
			Name:   "name1",
			Value:  "value1",
		}, string(configsections.Deployment))
		if !tc.badExec && !tc.badJSONUnmarshal {
			assert.NotNil(t, list)
			assert.Equal(t, "my-test1", list.Items[0].Metadata.Name)
		}

		// Assert the errors and cleanup
		if tc.badExec {
			assert.NotNil(t, err)
			execCommandOutput = origExecFunc
		}

		if tc.badJSONUnmarshal {
			assert.NotNil(t, err)
			jsonUnmarshal = json.Unmarshal
		}
	}
}
