// Copyright (C) 2020-2022 Red Hat, Inc.
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

package command

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"

	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/xeipuuv/gojsonschema"
)

const (
	testTimeoutDuration = time.Second * 1
)

var (
	pathRelativeToRoot    = path.Join("..", "..", "..", "..")
	genericTestSchemaFile = path.Join("schemas", "generic-test.schema.json")
	checkSubFilename      = "command.json"
	expectedPassPattern   = "(?m).*"
	pathToTestSchemaFile  = path.Join(pathRelativeToRoot, genericTestSchemaFile)
	testCommand           = "oc get pods -n tnf"
)

func createTest() (*tnf.Tester, []reel.Handler, *gojsonschema.Result, error) {
	values := make(map[string]interface{})
	values["COMMAND"] = testCommand
	values["TIMEOUT"] = testTimeoutDuration.Nanoseconds()
	return generic.NewGenericFromMap(checkSubFilename, pathToTestSchemaFile, values)
}
func TestCommand_Args(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)
	assert.Nil(t, (*test).Args())
}

func TestCommand_GetIdentifier(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)
	assert.Equal(t, identifier.CommandIdentifier, (*test).GetIdentifier())
}

func TestCommand_ReelFirst(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)
	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	step := handler.ReelFirst()
	assert.Equal(t, testCommand, step.Execute)
	assert.Contains(t, step.Expect, expectedPassPattern)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestCommand_ReelEOF(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)
	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	// just ensure there isn't a panic
	handler.ReelEOF()
}

func TestCommand_ReelTimeout(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)
	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	assert.Nil(t, handler.ReelTimeout())
}

func TestCommand_ReelMatch(t *testing.T) {
	tester, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)
	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	// Positive Test
	step := handler.ReelMatch(expectedPassPattern, "", "OK")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, (*tester).Result())
}
