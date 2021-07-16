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

package shutdown_test

import (
	"fmt"
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
	testTimeoutDuration = time.Second * 5
)

var (
	genericTestSchemaFile = path.Join("schemas", "generic-test.schema.json")
	shutdownFilename      = "shutdown.json"
	/* #nosec G101 */
	expectedPassPattern  = "(?m)prestop-defined"
	expectedFailPattern  = "(?m)prestop-not-defined"
	pathRelativeToRoot   = path.Join("..", "..", "..", "..")
	pathToTestSchemaFile = path.Join(pathRelativeToRoot, genericTestSchemaFile)
	testPodNameSpace     = "testnamespace"
	testPodName          = "testPodname"
)

func createTest() (*tnf.Tester, []reel.Handler, *gojsonschema.Result, error) {
	values := make(map[string]interface{})
	values["POD_NAMESPACE"] = testPodNameSpace
	values["POD_NAME"] = testPodName
	values["GO_TEMPLATE_PATH"] = "."
	return generic.NewGenericFromMap(shutdownFilename, pathToTestSchemaFile, values)
}

func TestShutdown_Args(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Nil(t, (*test).Args())
}

func TestShutdown_GetIdentifier(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, identifier.ShutdownURLIdentifier, (*test).GetIdentifier())
}

func TestShutdown_ReelFirst(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	step := handler.ReelFirst()
	expectedCommand := fmt.Sprintf("oc get pod -n %s %s -o go-template-file=./shutdown.gotemplate",
		testPodNameSpace, testPodName)
	assert.Equal(t, expectedCommand, step.Execute)
	assert.Contains(t, step.Expect, expectedPassPattern, expectedFailPattern)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestShutdown_ReelEof(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	// just ensure there isn't a panic
	handler.ReelEOF()
}

func TestShutdown_ReelTimeout(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	assert.Nil(t, handler.ReelTimeout())
}

func TestShutdown_ReelMatch(t *testing.T) {
	tester, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	step := handler.ReelMatch(expectedFailPattern, "", "prestop-not-defined")

	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, (*tester).Result())

	step = handler.ReelMatch(expectedPassPattern, "", "prestop-defined")

	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, (*tester).Result())
}
