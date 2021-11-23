// Copyright (C) 2020-2021 Red Hat, Inc.
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

package imagepullpolicy_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/imagepullpolicy"
)

// Test_NewImagepullpolicy is the unit test for NewImagepullpolicy().
func Test_NewImagepullpolicy(t *testing.T) {
	handler := imagepullpolicy.NewImagepullpolicy(testTimeoutDuration, testPodNamespace, testPodName, testContainerCount)
	assert.NotNil(t, handler)
	assert.Equal(t, testTimeoutDuration, handler.Timeout())
	assert.Equal(t, handler.Result(), tnf.ERROR)
	// Todo: Write test.
}

func Test_ReelFirstPositive(t *testing.T) {
	handler := imagepullpolicy.NewImagepullpolicy(testTimeoutDuration, testPodNamespace, testPodName, testContainerCount)
	assert.NotNil(t, handler)
	firstStep := handler.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])

	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
}

func Test_ReelFirstNegative(t *testing.T) {
	handler := imagepullpolicy.NewImagepullpolicy(testTimeoutDuration, testPodNamespace, testPodName, testContainerCount)
	assert.NotNil(t, handler)
	firstStep := handler.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

// Test_Imagepullpolicy_Args is the unit test for Imagepullpolicy_Args().
func TestImagepullpolicy_Args(t *testing.T) {
	// Todo: Write test.
}

// Test_Imagepullpolicy_GetIdentifier is the unit test for Imagepullpolicy_GetIdentifier().
func TestImagepullpolicy_GetIdentifier(t *testing.T) {
	// Todo: Write test.
}

// Test_Imagepullpolicy_ReelFirst is the unit test for Imagepullpolicy_ReelFirst().
func TestImagepullpolicy_ReelFirst(t *testing.T) {
	// Todo: Write test.
}

// Test_Imagepullpolicy_ReelEOF is the unit test for Imagepullpolicy_ReelEOF().
func TestImagepullpolicy_ReelEOF(t *testing.T) {
	handler := imagepullpolicy.NewImagepullpolicy(testTimeoutDuration, testPodNamespace, testPodName, testContainerCount)
	assert.NotNil(t, handler)
	handler.ReelEOF()
	// Todo: Write test.
}

// Test_Imagepullpolicy_ReelTimeout is the unit test for Imagepullpolicy}_ReelTimeout().
func TestImagepullpolicy_ReelTimeout(t *testing.T) {
	// Todo: Write test.
}

// Test_Imagepullpolicy_ReelMatch is the unit test for Imagepullpolicy_ReelMatch().
func TestImagepullpolicy_ReelMatch(t *testing.T) {
	// Todo: Write test.
}

const (
	testTimeoutDuration = time.Second * 1
	testContainerCount  = 1
	testInputError      = "Always"
	testPodNamespace    = "testPodNamespace"
	testPodName         = "testPodName"
)

var (
	testInputSuccess = "IfNotPresent"
)
