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

package owners_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	ow "github.com/test-network-function/test-network-function/pkg/tnf/handlers/owners"
)

func Test_NewNodeNames(t *testing.T) {
	newOw := ow.NewOwners(testTimeoutDuration, testPodNamespace, testPodName)
	assert.NotNil(t, newOw)
	assert.Equal(t, testTimeoutDuration, newOw.Timeout())
	assert.Equal(t, newOw.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newOw := ow.NewOwners(testTimeoutDuration, testPodNamespace, testPodName)
	assert.NotNil(t, newOw)
	firstStep := newOw.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	for _, positiveInput := range testInputSuccessSlice {
		matches := re.FindStringSubmatch(positiveInput)
		assert.Len(t, matches, 1)
	}
}

func Test_ReelFirstNegative(t *testing.T) {
	newOw := ow.NewOwners(testTimeoutDuration, testPodNamespace, testPodName)
	assert.NotNil(t, newOw)
	firstStep := newOw.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newOw := ow.NewOwners(testTimeoutDuration, testPodNamespace, testPodName)
	assert.NotNil(t, newOw)
	for _, input := range testInputSuccessSlice {
		step := newOw.ReelMatch("", "", input)
		assert.Nil(t, step)
		assert.Equal(t, tnf.SUCCESS, newOw.Result())
	}
}

func Test_ReelMatchFail(t *testing.T) {
	newOw := ow.NewOwners(testTimeoutDuration, testPodNamespace, testPodName)
	assert.NotNil(t, newOw)
	for _, input := range testInputFailureSlice {
		step := newOw.ReelMatch("", "", input)
		assert.Nil(t, step)
		assert.Equal(t, tnf.FAILURE, newOw.Result())
	}
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newOw := ow.NewOwners(testTimeoutDuration, testPodNamespace, testPodName)
	assert.NotNil(t, newOw)
	newOw.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInputError      = ""
	testPodNamespace    = "testPodNamespace"
	testPodName         = "testPodName"
)

var (
	testInputFailureSlice = []string{
		"NAME\n",
		"NAME\nOwner\nOwner\n",
	}
	testInputSuccessSlice = []string{
		"OWNERKIND\nReplicaSet\n",
		"OWNERKIND\nDaemonSet\n",
	}
)
