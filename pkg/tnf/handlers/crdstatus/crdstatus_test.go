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

package crdstatus

import (
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
)

func Test_NewCrdStatus(t *testing.T) {
	handler := NewCrdStatus(testTimeoutDuration, testCrdTargetLabels)
	assert.NotNil(t, handler)
	assert.Equal(t, testTimeoutDuration, handler.Timeout())
	assert.Equal(t, handler.Result(), tnf.ERROR)
}

func Test_ReelFirst(t *testing.T) {
	handler := NewCrdStatus(testTimeoutDuration, testCrdTargetLabels)
	assert.NotNil(t, handler)
	firstStep := handler.ReelFirst()
	assert.NotNil(t, firstStep.Expect[0])
	_ = regexp.MustCompile(firstStep.Expect[0])
}

func Test_ReelMatchSuccess(t *testing.T) {
	handler := NewCrdStatus(testTimeoutDuration, testCrdTargetLabels)
	assert.NotNil(t, handler)
	jsonString, err := os.ReadFile(testValidCrdJSONPath)
	assert.Nil(t, err)
	step := handler.ReelMatch("", "", string(jsonString))
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, handler.Result())
	assert.Equal(t, len(handler.CrdItems), 2)
	assert.NotNil(t, handler.CrdItems[0].Status)
	assert.Nil(t, handler.CrdItems[1].Status)
}

func Test_ReelMatchFailure(t *testing.T) {
	handler := NewCrdStatus(testTimeoutDuration, testCrdTargetLabels)
	assert.NotNil(t, handler)
	step := handler.ReelMatch("", "", testInvalidCrdJSON)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, handler.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	handler := NewCrdStatus(testTimeoutDuration, testCrdTargetLabels)
	assert.NotNil(t, handler)
	handler.ReelEOF()
}

func Test_ReelTimeout(t *testing.T) {
	handler := NewCrdStatus(testTimeoutDuration, testCrdTargetLabels)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.ReelTimeout())
}

const (
	testTimeoutDuration  = time.Second * 1
	testInvalidCrdJSON   = "invalid json string"
	testValidCrdJSONPath = "testdata/testcrds.json"
)

var (
	testCrdTargetLabels = []string{"test-network-function.com/generic=target"}
)
