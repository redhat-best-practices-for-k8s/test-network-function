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

package checksubscription_test

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
	testTimeoutDuration = time.Second * 10
)

var (
	genericTestSchemaFile     = path.Join("schemas", "generic-test.schema.json")
	checkSubFilename          = "check-subscription.json"
	expectedFailPattern       = "(?m)Error from server.*"
	expectedPassPattern       = testSubscriptionName
	pathRelativeToRoot        = path.Join("..", "..", "..", "..")
	pathToTestSchemaFile      = path.Join(pathRelativeToRoot, genericTestSchemaFile)
	testSubscriptionName      = "testSubName123"
	testSubscriptionNamespace = "testSubNamespace123"
)

func createTest() (*tnf.Tester, []reel.Handler, *gojsonschema.Result, error) {
	values := make(map[string]interface{})
	values["SUBSCRIPTION_NAME"] = testSubscriptionName
	values["SUBSCRIPTION_NAMESPACE"] = testSubscriptionNamespace
	return generic.NewGenericFromMap(checkSubFilename, pathToTestSchemaFile, values)
}

func TestNodes_Args(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Nil(t, (*test).Args())
}

func TestNodes_GetIdentifier(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, identifier.CheckSubscriptionURLIdentifier, (*test).GetIdentifier())
}

func TestNodes_ReelFirst(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	step := handler.ReelFirst()
	expectedReelFirstExecute := fmt.Sprintf("oc get subscription %s -n %s -o json | jq -r '.metadata.name'\n", testSubscriptionName, testSubscriptionNamespace)
	assert.Equal(t, expectedReelFirstExecute, step.Execute)
	assert.Contains(t, step.Expect, expectedPassPattern, expectedFailPattern)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestNodes_ReelEOF(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	// just ensure there isn't a panic
	handler.ReelEOF()
}

func TestNodes_ReelTimeout(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	assert.Nil(t, handler.ReelTimeout())
}

// Also tests GetNodes() and Result()
func TestNodes_ReelMatch(t *testing.T) {
	tester, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]

	// Positive Test
	step := handler.ReelMatch(expectedPassPattern, "", testSubscriptionName)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, (*tester).Result())

	// Negative Test
	step = handler.ReelMatch(expectedFailPattern, "", "Error from server")
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, (*tester).Result())
}
