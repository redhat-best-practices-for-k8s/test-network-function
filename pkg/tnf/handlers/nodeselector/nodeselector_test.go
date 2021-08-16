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

package nodeselector_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeselector"
)

func TestNewNodeSelector(t *testing.T) {
	newNodeSelector := nodeselector.NewNodeSelector(testTimeoutDuration, testPodName, testNamespaceName)
	assert.NotNil(t, newNodeSelector)
	assert.Equal(t, tnf.ERROR, newNodeSelector.Result())
}

func Test_ReelFirst(t *testing.T) {
	newNodeSelector := nodeselector.NewNodeSelector(testTimeoutDuration, testPodName, testNamespaceName)
	assert.NotNil(t, newNodeSelector)
	firstStep := newNodeSelector.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newNodeSelector := nodeselector.NewNodeSelector(testTimeoutDuration, testPodName, testNamespaceName)
	assert.NotNil(t, newNodeSelector)
	step := newNodeSelector.ReelMatch("", "", testInput)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newNodeSelector.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newNodeSelector := nodeselector.NewNodeSelector(testTimeoutDuration, testPodName, testNamespaceName)
	assert.NotNil(t, newNodeSelector)
	newNodeSelector.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = `<none>     <none>`
	testPodName         = "pod1"
	testNamespaceName   = "namespace1"
)
