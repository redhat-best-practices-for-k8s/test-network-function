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

package nodemcname_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodemcname"
)

func TestNodeMcName(t *testing.T) {
	newNodeMcName := nodemcname.NewNodeMcName(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newNodeMcName)
	assert.Equal(t, tnf.ERROR, newNodeMcName.Result())
}

func Test_ReelFirst(t *testing.T) {
	newNodeMcName := nodemcname.NewNodeMcName(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newNodeMcName)
	firstStep := newNodeMcName.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newNodeMcName := nodemcname.NewNodeMcName(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newNodeMcName)
	step := newNodeMcName.ReelMatch("", "", testInput)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newNodeMcName.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newNodeMcName := nodemcname.NewNodeMcName(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newNodeMcName)
	newNodeMcName.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = "rendered-master-50f5c7dee39e913002e5ae6323b8adc9"
	testNodeName        = "crc-l6qvn-master-0"
)
