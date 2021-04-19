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

package nodehugepages_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	nh "github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodehugepages"
)

func Test_NewNodeHugepages(t *testing.T) {
	newNh := nh.NewNodeHugepages(testTimeoutDuration, testNode, testExpectedHugepagesz, testExpectedHugepages)
	assert.NotNil(t, newNh)
	assert.Equal(t, testTimeoutDuration, newNh.Timeout())
	assert.Equal(t, newNh.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newNh := nh.NewNodeHugepages(testTimeoutDuration, testNode, testExpectedHugepagesz, testExpectedHugepages)
	assert.NotNil(t, newNh)
	firstStep := newNh.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newNh := nh.NewNodeHugepages(testTimeoutDuration, testNode, testExpectedHugepagesz, testExpectedHugepages)
	assert.NotNil(t, newNh)
	firstStep := newNh.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newNh := nh.NewNodeHugepages(testTimeoutDuration, testNode, testExpectedHugepagesz, testExpectedHugepages)
	assert.NotNil(t, newNh)
	step := newNh.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newNh.Result())
}

func Test_ReelMatchFailure(t *testing.T) {
	newNh := nh.NewNodeHugepages(testTimeoutDuration, testNode, testExpectedHugepagesz, testExpectedHugepages)
	assert.NotNil(t, newNh)
	step := newNh.ReelMatch("", "", testInputFailure)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, newNh.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newNh := nh.NewNodeHugepages(testTimeoutDuration, testNode, testExpectedHugepagesz, testExpectedHugepages)
	assert.NotNil(t, newNh)
	newNh.ReelEOF()
}

const (
	testTimeoutDuration    = time.Second * 2
	testNode               = "testNode"
	testInputError         = ""
	testInputSuccess       = "HugePages_Total:       64\nHugepagesize:       1048576 kB\n"
	testInputFailure       = "HugePages_Total:       32\nHugepagesize:       1000000 kB\n"
	testExpectedHugepages  = 64
	testExpectedHugepagesz = 1024 * 1024
)
