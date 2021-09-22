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

package versionocp_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	ver "github.com/test-network-function/test-network-function/pkg/tnf/handlers/versionocp"
)

func Test_NewNodeNames(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	assert.Equal(t, testTimeoutDuration, newVer.Timeout())
	assert.Equal(t, newVer.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstPositiveEmpty(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputFailure)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputFailure, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	step := newVer.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newVer.Result())
	assert.Len(t, newVer.GetVersions(), 3)
}

func Test_ReelMatchFail(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	step := newVer.ReelMatch("", "", testInputFailure)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, newVer.Result())
	assert.Len(t, newVer.GetVersions(), 1)
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newVer := ver.NewVersionOCP(testTimeoutDuration)
	assert.NotNil(t, newVer)
	newVer.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInputError      = ""
	testInputFailure    = "NAME\n"
	testInputSuccess    = "Client Version: 4.7.16\nServer Version: 4.8.3\nKubernetes Version: v1.21.1+051ac4f\n"
)
