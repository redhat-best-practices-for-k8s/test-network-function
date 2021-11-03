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

package nodedebug_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	nd "github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodedebug"
)

func Test_NewNodeDebug(t *testing.T) {
	newNd := nd.NewNodeDebug(testTimeoutDuration, testNodeName, testCommand, true, true)
	assert.NotNil(t, newNd)
	assert.Equal(t, testTimeoutDuration, newNd.Timeout())
	assert.Equal(t, newNd.Result(), tnf.ERROR)
	assert.Equal(t, newNd.Trim, true)
	assert.Equal(t, newNd.Split, true)
	assert.Nil(t, newNd.Processed)
}

func Test_ReelFirstPositive(t *testing.T) {
	newNd := nd.NewNodeDebug(testTimeoutDuration, testNodeName, testCommand, true, true)
	assert.NotNil(t, newNd)
	firstStep := newNd.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newNd := nd.NewNodeDebug(testTimeoutDuration, testNodeName, testCommand, true, true)
	assert.NotNil(t, newNd)
	firstStep := newNd.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatch(t *testing.T) {
	newNd := nd.NewNodeDebug(testTimeoutDuration, testNodeName, testCommand, true, true)
	assert.NotNil(t, newNd)
	step := newNd.ReelMatch("", "", testInputSuccess, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newNd.Result())
	assert.Equal(t, testInputSuccess, newNd.Raw)
	assert.Equal(t, newNd.Processed, testMatchSuccess)
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newNd := nd.NewNodeDebug(testTimeoutDuration, testNodeName, testCommand, true, true)
	assert.NotNil(t, newNd)
	newNd.ReelEOF()
}

const (
	testNodeName        = "testNode"
	testCommand         = "command"
	testTimeoutDuration = time.Second * 2
	testInputError      = ""
)

var (
	testMatchSuccess = []string{
		"some data in line 1",
		"some data in line 2",
		"some data in line 3",
	}
	testInputSuccess = "\n" + strings.Join(testMatchSuccess, "\n") + "\n"
)
