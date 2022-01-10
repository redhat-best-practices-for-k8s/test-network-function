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
	"errors"
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	csvFile = "csv.json"
)

var (
	csvFilePath = path.Join(filePath, csvFile)
)

func loadCSVResource(filePath string) (csv CSVResource) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error (%s) loading CSVResource %s for testing", err, filePath)
	}
	err = jsonUnmarshal(contents, &csv)
	if err != nil {
		log.Fatalf("error (%s) loading CSVResource %s for testing", err, filePath)
	}
	return
}

func TestCSVGetAnnotationValue(t *testing.T) {
	csv := loadCSVResource(csvFilePath)
	var val []string

	err := csv.GetAnnotationValue("notPresent", &val)
	assert.Equal(t, 0, len(val))
	assert.NotNil(t, err)

	err = csv.GetAnnotationValue("test-network-function.com/operator_tests", &val)
	assert.Equal(t, []string{"OPERATOR_STATUS", "ANOTHER_TEST"}, val)
	assert.Nil(t, err)
}

func TestAnnotationUnmarshalError(t *testing.T) {
	testCases := []struct {
		expectedString string
	}{
		{
			expectedString: "error (this is an error) attempting to unmarshal value of annotation 'testKey' on CSV 'testnamespace/testname'",
		},
	}

	for _, tc := range testCases {
		csvr := CSVResource{
			Metadata: struct {
				Name        string            "json:\"name\""
				Namespace   string            "json:\"namespace\""
				Labels      map[string]string "json:\"labels\""
				Annotations map[string]string "json:\"annotations\""
			}{
				Name:      "testname",
				Namespace: "testnamespace",
			},
		}
		assert.Equal(t, tc.expectedString, csvr.annotationUnmarshalError("testKey", errors.New("this is an error")).Error())
	}
}

func TestGetCSVsByLabel(t *testing.T) {
	testCases := []struct {
		filename          string
		label             string
		value             string
		expectedCSV       string
		expectedItemCount int
	}{
		{
			filename:          "csv_output.json",
			expectedCSV:       "etcdoperator.v0.9.4",
			label:             "testLabel",
			value:             "testValue",
			expectedItemCount: 1,
		},
		{
			filename:          "csv_output_nolabel.json",
			expectedCSV:       "",
			label:             "testLabel",
			value:             "testValue",
			expectedItemCount: 0,
		},
	}

	for _, tc := range testCases {
		origFunc := executeOcGetAllCommand
		executeOcGetAllCommand = func(resourceType, labelQuery string) string {
			output, err := os.ReadFile(path.Join(filePath, tc.filename))
			assert.Nil(t, err)
			return string(output)
		}

		outputList, err := GetCSVsByLabel(tc.label, tc.value)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedItemCount, len(outputList.Items))
		if len(outputList.Items) > 0 {
			assert.Equal(t, tc.expectedCSV, outputList.Items[0].Metadata.Name)
		}

		executeOcGetAllCommand = origFunc
	}
}

func TestGetCSVsByNamespace(t *testing.T) {
	testCases := []struct {
		filename          string
		label             string
		value             string
		expectedCSV       string
		expectedItemCount int
	}{
		{
			filename:          "csv_output.json",
			expectedCSV:       "etcdoperator.v0.9.4",
			label:             "testLabel",
			value:             "testValue",
			expectedItemCount: 1,
		},
		{
			filename:          "csv_output_nolabel.json",
			expectedCSV:       "",
			label:             "testLabel",
			value:             "testValue",
			expectedItemCount: 0,
		},
	}

	for _, tc := range testCases {
		origFunc := executeOcGetAllCommand
		executeOcGetCommand = func(resourceType, labelQuery, namespace string) string {
			output, err := os.ReadFile(path.Join(filePath, tc.filename))
			assert.Nil(t, err)
			return string(output)
		}

		outputList, err := GetCSVsByLabelByNamespace(tc.label, tc.value, "testnamespace")
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedItemCount, len(outputList.Items))
		if len(outputList.Items) > 0 {
			assert.Equal(t, tc.expectedCSV, outputList.Items[0].Metadata.Name)
		}

		executeOcGetAllCommand = origFunc
	}
}
