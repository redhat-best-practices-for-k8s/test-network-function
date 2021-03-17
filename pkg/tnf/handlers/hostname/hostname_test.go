// Copyright (C) 2020 Red Hat, Inc.
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

package hostname_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/hostname"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

const (
	testTimeoutDuration = time.Second * 2
)

func TestHostname_Args(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	assert.Equal(t, []string{"hostname"}, h.Args())
}

func TestHostname_GetIdentifier(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	assert.Equal(t, identifier.HostnameIdentifier, h.GetIdentifier())
}

func TestHostname_ReelFirst(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	step := h.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Equal(t, []string{hostname.SuccessfulOutputRegex}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestHostname_ReelEof(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	// just ensures lack of panic
	h.ReelEOF()
}

func TestHostname_ReelTimeout(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	step := h.ReelTimeout()
	assert.Nil(t, step)
}

// Also tests GetHostname() and Result()
func TestHostname_ReelMatch(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	matchHostname := "testHostname"
	step := h.ReelMatch("", "", matchHostname)
	assert.Nil(t, step)
	assert.Equal(t, matchHostname, h.GetHostname())
	assert.Equal(t, tnf.SUCCESS, h.Result())
}

func TestNewHostname(t *testing.T) {
	h := hostname.NewHostname(testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, h.Result())
	assert.Equal(t, testTimeoutDuration, h.Timeout())
}
