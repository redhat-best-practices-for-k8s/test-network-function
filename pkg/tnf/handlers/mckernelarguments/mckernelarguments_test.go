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

package mckernelarguments_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/mckernelarguments"
)

func TestNewMcKernelArguments(t *testing.T) {
	newMcKernelArguments := mckernelarguments.NewMcKernelArguments(testTimeoutDuration, testMcName)
	assert.NotNil(t, newMcKernelArguments)
	assert.Equal(t, tnf.ERROR, newMcKernelArguments.Result())
}

func Test_ReelFirst(t *testing.T) {
	newMcKernelArguments := mckernelarguments.NewMcKernelArguments(testTimeoutDuration, testMcName)
	assert.NotNil(t, newMcKernelArguments)
	firstStep := newMcKernelArguments.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newMcKernelArguments := mckernelarguments.NewMcKernelArguments(testTimeoutDuration, testMcName)
	assert.NotNil(t, newMcKernelArguments)
	step := newMcKernelArguments.ReelMatch("", "", testInput)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newMcKernelArguments.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newMcKernelArguments := mckernelarguments.NewMcKernelArguments(testTimeoutDuration, testMcName)
	assert.NotNil(t, newMcKernelArguments)
	newMcKernelArguments.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = `["loglevel=7"]`
	testMcName          = "rendered-master-50f5c7dee39e913002e5ae6323b8adc9"
)
