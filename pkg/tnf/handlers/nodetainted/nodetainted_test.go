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

package nodetainted_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	nt "github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodetainted"
)

func Test_NewNodeTainted(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	assert.Equal(t, testTimeoutDuration, newNt.Timeout())
	assert.Equal(t, newNt.Result(), tnf.ERROR)
}

func Test_ReelFirstPositiveSuccess(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	firstStep := newNt.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testMatchSuccess, matches[0])
}

func Test_ReelFirstPositiveFailure(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	firstStep := newNt.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputFailure)
	assert.Len(t, matches, 1)
	assert.Equal(t, testMatchFailure, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	firstStep := newNt.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	step := newNt.ReelMatch("", "", testMatchSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newNt.Result())
}

func Test_ReelMatchFail(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	step := newNt.ReelMatch("", "", testMatchFailure)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, newNt.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newNt := nt.NewNodeTainted(testTimeoutDuration)
	assert.NotNil(t, newNt)
	newNt.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInputError      = ""
	testInputFailure    = "1\n"
	testInputSuccess    = "0\n"
	testMatchSuccess    = "0"
	testMatchFailure    = "1"
)
