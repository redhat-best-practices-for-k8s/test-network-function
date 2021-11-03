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

package podantiaffinity_test

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
	genericTestSchemaFile = path.Join("schemas", "generic-test.schema.json")
	checkSubFilename      = "podantiaffinity.json"
	expectedFailPattern1  = "(?m)Antiaffinity missing"
	expectedPassPattern   = "(?m)OK"
	expectedFailPattern2  = "(?m)Replica count is 1"
	pathRelativeToRoot    = path.Join("..", "..", "..", "..")
	pathToTestSchemaFile  = path.Join(pathRelativeToRoot, genericTestSchemaFile)
	testDeploymentName    = "testDeployName123"
	testNamespace         = "testDeployNamespace123"
)

func createTest() (*tnf.Tester, []reel.Handler, *gojsonschema.Result, error) {
	values := make(map[string]interface{})
	values["DEPLOYMENT_NAME"] = testDeploymentName
	values["DEPLOYMENT_NAMESPACE"] = testNamespace
	return generic.NewGenericFromMap(checkSubFilename, pathToTestSchemaFile, values)
}

func TestPods_Args(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Nil(t, (*test).Args())
}

func TestPods_GetIdentifier(t *testing.T) {
	test, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, identifier.PodAntiAffinityIdentifier, (*test).GetIdentifier())
}

func TestPods_ReelFirst(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	step := handler.ReelFirst()
	expectedReelFirstExecute := fmt.Sprintf(
		"oc get deployment %s -n %s -o json | jq -r ' if .spec.replicas > 1 then "+
			"if .spec.template.spec.affinity.podAntiAffinity != null then \"OK\" else "+
			"\"Antiaffinity missing\" end else \"Replica count is 1\" end '\n",
		testDeploymentName, testNamespace)
	assert.Equal(t, expectedReelFirstExecute, step.Execute)
	assert.Contains(t, step.Expect, expectedPassPattern, expectedFailPattern1, expectedFailPattern2)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestPods_ReelEOF(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	// just ensure there isn't a panic
	handler.ReelEOF()
}

func TestPods_ReelTimeout(t *testing.T) {
	_, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]
	assert.Nil(t, handler.ReelTimeout())
}

func TestPods_ReelMatch(t *testing.T) {
	tester, handlers, jsonParseResult, err := createTest()

	assert.Nil(t, err)
	assert.True(t, jsonParseResult.Valid())
	assert.NotNil(t, handlers)

	assert.Equal(t, 1, len(handlers))
	handler := handlers[0]

	// Positive Test
	step := handler.ReelMatch(expectedPassPattern, "", "OK", 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, (*tester).Result())

	// Negative Test
	step = handler.ReelMatch(expectedFailPattern1, "", "Antiaffinity missing", 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, (*tester).Result())
	// Negative Test
	step = handler.ReelMatch(expectedFailPattern2, "", "Replica count is 1", 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, (*tester).Result())
}
