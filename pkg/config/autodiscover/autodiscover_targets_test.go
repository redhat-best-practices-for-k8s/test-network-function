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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
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

func TestGetConfiguredOperatorTests(t *testing.T) {
	// Note: Without testconfigure.yml being set, this only
	// covers a subset of the code in the function.
	opTests := getConfiguredOperatorTests()
	assert.Nil(t, opTests)
}

func TestAppendPodsets(t *testing.T) {
	testCases := []struct {
		podsets         []configsections.PodSet
		namespaces      map[string]bool
		expectedPodSets []configsections.PodSet
	}{
		{
			podsets: []configsections.PodSet{
				{
					Name:      "testpod1",
					Namespace: "namespace1",
				},
			},
			namespaces: map[string]bool{
				"namespace1": true,
			},
			expectedPodSets: []configsections.PodSet{
				{
					Name:      "testpod1",
					Namespace: "namespace1",
				},
			},
		},
		{ // 'namespace1' does not exist, no podset available.
			podsets: []configsections.PodSet{
				{
					Name:      "testpod1",
					Namespace: "namespace1",
				},
			},
			namespaces: map[string]bool{
				"namespace2": true,
			},
			expectedPodSets: nil,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedPodSets, appendPodsets(tc.podsets, tc.namespaces))
	}
}

//nolint:funlen
func TestFindTestPodSetsByLabel(t *testing.T) {
	testCases := []struct {
		targetLabels           []configsections.Label
		resourceTypeDeployment string
		filename               string
		expectedPodSets        []configsections.PodSet
	}{
		{ // Test Case 1 - nothing found
			targetLabels: []configsections.Label{
				{
					Name:  "label1",
					Value: "value1",
				},
			},
			resourceTypeDeployment: string(configsections.Deployment),
			filename:               "testdata/empty.json",
			expectedPodSets:        nil,
		},
		{ // Test Case 2 - Found one deployment matching labels
			targetLabels: []configsections.Label{
				{
					Name:  "app",
					Value: "mydeploy",
				},
			},
			resourceTypeDeployment: string(configsections.Deployment),
			filename:               "testdata/test_deploy_matching_label.json",
			expectedPodSets: []configsections.PodSet{
				{
					Name:      "mydeploy",
					Namespace: "default",
					Type:      configsections.Deployment,
				},
			},
		},
	}

	for _, tc := range testCases {
		// spoof the output from execCommandOutput
		origFunc := execCommandOutput
		execCommandOutput = func(command string) string {
			output, err := os.ReadFile(tc.filename)
			assert.Nil(t, err)
			return string(output)
		}

		podsets := FindTestPodSetsByLabel(tc.targetLabels, tc.resourceTypeDeployment)

		if len(tc.expectedPodSets) > 0 {
			// Note: We are assuming that [0] is populated with the data we need.
			assert.Equal(t, tc.expectedPodSets[0].Name, podsets[0].Name)
			assert.Equal(t, tc.expectedPodSets[0].Namespace, podsets[0].Namespace)
			assert.Equal(t, tc.expectedPodSets[0].Type, podsets[0].Type)
		} else {
			assert.Nil(t, podsets)
		}

		execCommandOutput = origFunc
	}
}

func TestSetBundleAndIndexImage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	defer ginkgo.GinkgoRecover()
	testCases := []struct {
		csvName      string
		csvNamespace string
		indexImage   string
		bundleImage  string
	}{
		{
			csvName:      "csvexample1",
			csvNamespace: "csvns1",
			indexImage:   "http://index-csvexample1-in-csvns1:sha",
			bundleImage:  "http://bundle-csvexample1-in-csvns1:sha",
		},
		{
			csvName:      "csvexample2",
			csvNamespace: "csvns2",
			indexImage:   "http://index-csvexample2-in-csvns2:sha",
			bundleImage:  "http://bundle-csvexample2-in-csvns2:sha",
		},
	}

	for _, tc := range testCases {
		// The real method takes csvName and csvNamespace and execute several commands
		// to obtain indexImage and bundleImage. Here we will do the same but with
		// dummy commands (just echo)
		obtainedIndexImage := execCommandOutput(fmt.Sprintf("echo http://index-%s-in-%s:sha", tc.csvName, tc.csvNamespace))
		obtainedBundleImage := execCommandOutput(fmt.Sprintf("echo http://bundle-%s-in-%s:sha", tc.csvName, tc.csvNamespace))

		assert.Equal(t, tc.indexImage, obtainedIndexImage)
		assert.Equal(t, tc.bundleImage, obtainedBundleImage)
	}
}
