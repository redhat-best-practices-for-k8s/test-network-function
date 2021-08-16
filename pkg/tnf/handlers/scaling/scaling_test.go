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

package scaling_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/scaling"
)

func Test_NewScaling(t *testing.T) {
	handler := scaling.NewScaling(testTimeoutDuration, testPodNamespace, testDeploymentName, testReplicaCount)
	assert.NotNil(t, handler)
	assert.Equal(t, testTimeoutDuration, handler.Timeout())
	assert.Equal(t, handler.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	handler := scaling.NewScaling(testTimeoutDuration, testPodNamespace, testDeploymentName, testReplicaCount)
	assert.NotNil(t, handler)
	firstStep := handler.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])

	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
}

func Test_ReelFirstNegative(t *testing.T) {
	handler := scaling.NewScaling(testTimeoutDuration, testPodNamespace, testDeploymentName, testReplicaCount)
	assert.NotNil(t, handler)
	firstStep := handler.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	handler := scaling.NewScaling(testTimeoutDuration, testPodNamespace, testDeploymentName, testReplicaCount)
	assert.NotNil(t, handler)

	step := handler.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, handler.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	handler := scaling.NewScaling(testTimeoutDuration, testPodNamespace, testDeploymentName, testReplicaCount)
	assert.NotNil(t, handler)
	handler.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 1
	testReplicaCount    = 2
	testInputError      = ""
	testPodNamespace    = "testPodNamespace"
	testDeploymentName  = "testDeploymentName"
)

var (
	testInputSuccess = fmt.Sprintf("deployment.apps/%s scaled\n", testDeploymentName)
)
