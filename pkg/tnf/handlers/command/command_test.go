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

package command

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
)

const (
	testTimeoutDuration = time.Second * 1
)

var (
	testcommandTargetLabels = "oc get %s -n %s -o json -l %s"
)

// Command_ReelFirst
func TestCommand_ReelFirst(t *testing.T) {
	handler := NewCommand(testTimeoutDuration, testcommandTargetLabels)
	assert.NotNil(t, handler)
	firstStep := handler.ReelFirst()
	assert.NotNil(t, firstStep.Expect[0])
	_ = regexp.MustCompile(firstStep.Expect[0])
}

// Command_ReelEof
func TestCommand_ReelEof(t *testing.T) {
	handler := NewCommand(testTimeoutDuration, testcommandTargetLabels)
	assert.NotNil(t, handler)
	handler.ReelEOF()
}

// Command_ReelTimeout
func TestCommand_ReelTimeout(t *testing.T) {
	handler := NewCommand(testTimeoutDuration, testcommandTargetLabels)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.ReelTimeout())
}

// NewCommand
func TestNewCommand(t *testing.T) {
	handler := NewCommand(testTimeoutDuration, testcommandTargetLabels)
	assert.NotNil(t, handler)
	assert.Equal(t, testTimeoutDuration, handler.Timeout())
	assert.Equal(t, handler.Result(), tnf.ERROR)
}
