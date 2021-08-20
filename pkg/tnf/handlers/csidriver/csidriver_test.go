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

package csidriver_test

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
	testTimeoutDuration = time.Second * 10
)

var (
	genericTestSchemaFile = path.Join("schemas", "generic-test.schema.json")
	csiDriverFilename     = "csidriver.json"
	/* #nosec G101 */
	expectedPassPattern  = "(?m)(.|\n)+"
	pathRelativeToRoot   = path.Join("..", "..", "..", "..")
	pathToTestSchemaFile = path.Join(pathRelativeToRoot, genericTestSchemaFile)
)

func createTest() (*tnf.Tester, []reel.Handler, *gojsonschema.Result, error) {
	return generic.NewGenericFromJSONFile(csiDriverFilename, pathToTestSchemaFile)
}

func TestCSIs_Args(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Nil(t, (*test).Args())
}

func TestCSIs_GetIdentifier(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, identifier.CSIDriverIdentifier, (*test).GetIdentifier())
}

func TestCSIs_ReelFirst(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	step := handler.ReelFirst()
	assert.Equal(t, "oc get csidriver -o json\n", step.Execute)
	assert.Equal(t, []string{expectedPassPattern}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestCSIs_ReelEOF(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	// just ensure there isn't a panic
	handler.ReelEOF()
}

func TestCSIs_ReelTimeout(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()
	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	assert.Nil(t, handler.ReelTimeout())
}

func TestCSIs_ReelMatch(t *testing.T) {
	tester, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]

	// Positive Test
	step := handler.ReelMatch(expectedPassPattern, "", "anythingMatches")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, (*tester).Result())
}
