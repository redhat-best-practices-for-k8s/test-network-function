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

package interactive_test

import (
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	mock_interactive "github.com/test-network-function/test-network-function/pkg/tnf/interactive/mocks"
)

var (
	ptySchemaFile = "../../../schemas/generic-pty.schema.json"
)

var yamlFileTestCases = []struct {
	inputFile             string
	inputSchema           string
	expectedErr           bool
	expectedResultIsValid bool
}{
	{
		inputFile:             "simple.json",
		inputSchema:           ptySchemaFile,
		expectedErr:           false,
		expectedResultIsValid: true,
	},
	{
		inputFile:             "missing_timeout.json",
		inputSchema:           ptySchemaFile,
		expectedErr:           false,
		expectedResultIsValid: false,
	},
	{
		inputFile:   "does_not_exist",
		inputSchema: ptySchemaFile,
		expectedErr: true,
	},
	{
		inputFile:   "simple.json",
		inputSchema: "does_not_exist",
		expectedErr: true,
	},
}

var yamlTemplateTestCases = []struct {
	inputTemplateFile     string
	inputValuesFile       string
	inputSchema           string
	expectedErr           bool
	expectedResultIsValid bool
}{
	{
		inputTemplateFile:     "../../../examples/pty/ssh.json.tpl",
		inputValuesFile:       "../../../examples/pty/ssh.json.tpl.values.yaml",
		inputSchema:           ptySchemaFile,
		expectedErr:           false,
		expectedResultIsValid: true,
	},
	{
		inputTemplateFile: "../../../examples/pty/ssh.json.tpl",
		inputValuesFile:   "./testdata/ssh.json.tpl.values.bad.yaml.tpl",
		inputSchema:       ptySchemaFile,
		expectedErr:       true,
	},
	{
		inputTemplateFile: "../../../examples/pty/ssh.json.tpl",
		inputValuesFile:   "does_not_exist",
		inputSchema:       ptySchemaFile,
		expectedErr:       true,
	},
	{
		inputTemplateFile: "../../../examples/pty/ssh.json.tpl",
		inputValuesFile:   "testdata/ssh.json.tpl.values.bad.nonyaml",
		inputSchema:       ptySchemaFile,
		expectedErr:       true,
	},
	{
		inputTemplateFile: "../../../examples/pty/ssh.json.tpl",
		inputValuesFile:   "../../../examples/pty/ssh.json.tpl.values.yaml",
		inputSchema:       "does_not_exist",
		expectedErr:       true,
	},
}

func getTestFile(inputFile string) string {
	return path.Join("testdata", inputFile)
}

func TestSpawnGenericPTYFromYAMLFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpawner := mock_interactive.NewMockSpawner(ctrl)
	var spawner interactive.Spawner = mockSpawner
	mockSpawner.EXPECT().Spawn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	for _, testCase := range yamlFileTestCases {
		_, result, err := interactive.SpawnGenericPTYFromYAMLFile(getTestFile(testCase.inputFile), testCase.inputSchema, &spawner)
		assert.Equal(t, testCase.expectedErr, err != nil)
		if err == nil {
			assert.Equal(t, testCase.expectedResultIsValid, result.Valid())
		}
	}
}

func TestSpawnGenericPTYFromYAMLTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpawner := mock_interactive.NewMockSpawner(ctrl)
	var spawner interactive.Spawner = mockSpawner
	mockSpawner.EXPECT().Spawn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	for _, testCase := range yamlTemplateTestCases {
		_, result, err := interactive.SpawnGenericPTYFromYAMLTemplate(testCase.inputTemplateFile, testCase.inputValuesFile, testCase.inputSchema, &spawner)
		assert.Equal(t, testCase.expectedErr, err != nil)
		if err == nil {
			assert.Equal(t, testCase.expectedResultIsValid, result.Valid())
		}
	}
}
