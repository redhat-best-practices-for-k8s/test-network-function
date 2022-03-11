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

//nolint: funlen
func TestGetCsvInstallPlanNames(t *testing.T) {
	originalExecCommandOutput := execCommandOutput
	defer func() {
		execCommandOutput = originalExecCommandOutput
	}()

	testCases := []struct {
		csvName                     string
		csvNamespace                string
		expectedError               string
		expectedPlanNames           []string
		mockedExecCommandOutputFunc func(cmd string) string
	}{
		{
			csvName:           "csvexample1",
			csvNamespace:      "csvns1",
			expectedError:     "",
			expectedPlanNames: []string{"installPlan1"},
			mockedExecCommandOutputFunc: func(cmd string) string {
				return "installPlan1"
			},
		},
		{
			csvName:           "csvexample1",
			csvNamespace:      "csvns1",
			expectedPlanNames: []string{"installPlan1", "installPlan2"},
			mockedExecCommandOutputFunc: func(cmd string) string {
				return "installPlan1\ninstallPlan2"
			},
		},
		{
			csvName:           "csvexample1",
			csvNamespace:      "csvns1",
			expectedError:     "installplan not found",
			expectedPlanNames: []string{},
			mockedExecCommandOutputFunc: func(cmd string) string {
				return ""
			},
		},
	}

	for _, tc := range testCases {
		execCommandOutput = tc.mockedExecCommandOutputFunc
		planNames, err := getCsvInstallPlanNames(tc.csvName, tc.csvNamespace)
		if tc.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), tc.expectedError)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, planNames, tc.expectedPlanNames)
	}
}

//nolint:funlen
func TestGetInstallPlanData(t *testing.T) {
	originalExecCommandOutput := execCommandOutput
	defer func() {
		execCommandOutput = originalExecCommandOutput
	}()

	testCases := []struct {
		installPlanName                string
		namespace                      string
		expectedError                  string
		expectedBundleImage            string
		expectedCatalogSourceName      string
		expectedCatalogSourceNamespace string
		mockedExecCommandOutputFunc    func(cmd string) string
	}{
		{
			installPlanName:                "install-1",
			namespace:                      "ns1",
			expectedError:                  "",
			expectedBundleImage:            "http://bundle-csvexample1-in-csvns1:sha",
			expectedCatalogSourceName:      "catalogName1",
			expectedCatalogSourceNamespace: "catalogNamespace1",
			mockedExecCommandOutputFunc: func(cmd string) string {
				return "http://bundle-csvexample1-in-csvns1:sha,catalogName1,catalogNamespace1"
			},
		},
		{
			installPlanName:                "install-2",
			namespace:                      "ns1",
			expectedError:                  "invalid installplan info: invalid-output",
			expectedBundleImage:            "",
			expectedCatalogSourceName:      "",
			expectedCatalogSourceNamespace: "",
			mockedExecCommandOutputFunc: func(cmd string) string {
				return "invalid-output"
			},
		},
	}

	for _, tc := range testCases {
		execCommandOutput = tc.mockedExecCommandOutputFunc
		bundleImage, catalogName, catalogNamespace, err := getInstallPlanData(tc.installPlanName, tc.namespace)
		if tc.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), tc.expectedError)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, bundleImage, tc.expectedBundleImage)
		assert.Equal(t, catalogName, tc.expectedCatalogSourceName)
		assert.Equal(t, catalogNamespace, tc.expectedCatalogSourceNamespace)
	}
}

//nolint:funlen
func TestGetCatalogSourceImageIndex(t *testing.T) {
	originalExecCommandOutput := execCommandOutput
	defer func() {
		execCommandOutput = originalExecCommandOutput
	}()

	testCases := []struct {
		catalogName                 string
		catalogNamespace            string
		expectedError               string
		expectedImageIndex          string
		mockedExecCommandOutputFunc func(cmd string) string
	}{
		{
			catalogName:        "catalogName1",
			catalogNamespace:   "ns1",
			expectedError:      "",
			expectedImageIndex: "http://index1-csvexample2-in-csvns2:sha",
			mockedExecCommandOutputFunc: func(cmd string) string {
				return "http://index1-csvexample2-in-csvns2:sha"
			},
		},
		{
			catalogName:        "catalogName2",
			catalogNamespace:   "ns1",
			expectedError:      "",
			expectedImageIndex: "",
			mockedExecCommandOutputFunc: func(cmd string) string {
				return "null"
			},
		},
		{
			catalogName:        "catalogName3",
			catalogNamespace:   "ns3",
			expectedError:      "failed to get index image for catalogsource catalogName3 (ns ns3)",
			expectedImageIndex: "",
			mockedExecCommandOutputFunc: func(cmd string) string {
				return ""
			},
		},
	}

	for _, tc := range testCases {
		execCommandOutput = tc.mockedExecCommandOutputFunc
		imageIndex, err := getCatalogSourceImageIndex(tc.catalogName, tc.catalogNamespace)
		if tc.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), tc.expectedError)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, imageIndex, tc.expectedImageIndex)
	}
}

//nolint:funlen
func TestGetCsvInstallPlans(t *testing.T) {
	// Save the original functions and defer its restoration.
	originalGetCsvInstallPlanNames := getCsvInstallPlanNames
	defer func() {
		getCsvInstallPlanNames = originalGetCsvInstallPlanNames
	}()

	originalGetInstallPlanData := getInstallPlanData
	defer func() {
		getInstallPlanData = originalGetInstallPlanData
	}()

	originalGetCatalogSourceImageIndex := getCatalogSourceImageIndex
	defer func() {
		getCatalogSourceImageIndex = originalGetCatalogSourceImageIndex
	}()

	// installPlanIndex is a helper index for the mocking functions to allow
	// the testing of several installPlans
	installPlanIndex := 0
	testCases := []struct {
		csvName                          string
		csvNamespace                     string
		mockedGetCsvInstallPlanNames     func(csvName string, csvNamespace string) ([]string, error)
		mockedGetInstallPlanData         func(installPlanName string, namespace string) (bundleImagePath string, catalogSource string, catalogSourceNamespace string, err error)
		mockedGetCatalogSourceImageIndex func(catalogSourceName string, catalogSourceNamespace string) (string, error)
		expectedError                    string
		expectedInstallPlans             []configsections.InstallPlan
	}{
		// Positive TCs:
		{
			csvName:      "csvexample1",
			csvNamespace: "csvns1",
			mockedGetCsvInstallPlanNames: func(csvName string, csvNamespace string) ([]string, error) {
				return []string{"install-1"}, nil
			},
			mockedGetInstallPlanData: func(installPlanName string, namespace string) (string, string, string, error) {
				return "http://bundle1-csvexample1-in-csvns1:sha", "catalogName1", "catalogNamespace1", nil
			},
			mockedGetCatalogSourceImageIndex: func(catalogSourceName string, catalogSourceNamespace string) (string, error) {
				return "http://index-csvexample1-in-csvns1:sha", nil
			},
			expectedError:        "",
			expectedInstallPlans: []configsections.InstallPlan{{Name: "install-1", BundleImage: "http://bundle1-csvexample1-in-csvns1:sha", IndexImage: "http://index-csvexample1-in-csvns1:sha"}},
		},
		{
			csvName:      "csvexample2",
			csvNamespace: "csvns2",
			mockedGetCsvInstallPlanNames: func(csvName string, csvNamespace string) ([]string, error) {
				return []string{"install-1", "install-2"}, nil
			},
			mockedGetInstallPlanData: func(installPlanName string, namespace string) (string, string, string, error) {
				if installPlanIndex == 0 {
					return "http://bundle1-csvexample2-in-csvns2:sha", "catalogName1", "catalogNamespace1", nil
				}
				return "http://bundle2-csvexample2-in-csvns2:sha", "catalogName1", "catalogNamespace1", nil
			},
			mockedGetCatalogSourceImageIndex: func(catalogSourceName string, catalogSourceNamespace string) (string, error) {
				if installPlanIndex == 0 {
					installPlanIndex++
					return "http://index1-csvexample2-in-csvns2:sha", nil
				}
				return "http://index2-csvexample2-in-csvns2:sha", nil
			},
			expectedError: "",
			expectedInstallPlans: []configsections.InstallPlan{
				{Name: "install-1", BundleImage: "http://bundle1-csvexample2-in-csvns2:sha", IndexImage: "http://index1-csvexample2-in-csvns2:sha"},
				{Name: "install-2", BundleImage: "http://bundle2-csvexample2-in-csvns2:sha", IndexImage: "http://index2-csvexample2-in-csvns2:sha"},
			},
		},
		// Error checking TCs:
		{
			// No installPlan found for given CSV.
			csvName:      "csvexample5",
			csvNamespace: "csvns5",
			mockedGetCsvInstallPlanNames: func(csvName string, csvNamespace string) ([]string, error) {
				return []string{}, errors.New("installplan not found")
			},
			expectedError:        "installplan not found",
			expectedInstallPlans: []configsections.InstallPlan{},
		},
		{
			// Invalid output when getting installPlan data.
			csvName:      "csvexample4",
			csvNamespace: "csvns4",
			mockedGetCsvInstallPlanNames: func(csvName string, csvNamespace string) ([]string, error) {
				return []string{"install-1"}, nil
			},
			mockedGetInstallPlanData: func(installPlanName string, namespace string) (string, string, string, error) {
				return "", "", "", errors.New("invalid installplan info: invalid-output")
			},
			expectedError:        "invalid installplan info: invalid-output",
			expectedInstallPlans: []configsections.InstallPlan{},
		},
		{
			// Empty output when retrieving image index from catalog source.
			csvName:      "csvexample3",
			csvNamespace: "csvns3",
			mockedGetCsvInstallPlanNames: func(csvName string, csvNamespace string) ([]string, error) {
				return []string{"install-1"}, nil
			},
			mockedGetInstallPlanData: func(installPlanName string, namespace string) (string, string, string, error) {
				return "http://bundle1-csvexample3-in-csvns3:sha", "catalogName1", "catalogNamespace1", nil
			},
			mockedGetCatalogSourceImageIndex: func(catalogSourceName string, catalogSourceNamespace string) (string, error) {
				return "", fmt.Errorf("failed to get index image for catalogsource %s (ns %s)", catalogSourceName, catalogSourceNamespace)
			},
			expectedError:        "failed to get index image for catalogsource catalogName1 (ns catalogNamespace1)",
			expectedInstallPlans: []configsections.InstallPlan{},
		},
	}

	for _, tc := range testCases {
		getCsvInstallPlanNames = tc.mockedGetCsvInstallPlanNames
		getInstallPlanData = tc.mockedGetInstallPlanData
		getCatalogSourceImageIndex = tc.mockedGetCatalogSourceImageIndex

		installPlans, err := getCsvInstallPlans(tc.csvName, tc.csvNamespace)
		if tc.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, err.Error(), tc.expectedError)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, installPlans, tc.expectedInstallPlans)
	}
}
