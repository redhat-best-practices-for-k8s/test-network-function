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

package redhat_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

var testTimeoutDuration = time.Second * 2

// TestNewRelease also tests Args, Timeout and Result
func TestNewRelease(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	assert.NotNil(t, r)
	assert.Equal(t, strings.Split(redhat.ReleaseCommand, " "), r.Args())
	assert.Equal(t, testTimeoutDuration, r.Timeout())
	assert.Equal(t, tnf.ERROR, r.Result())
}

func TestRelease_GetIdentifier(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	assert.Equal(t, identifier.VersionIdentifier, r.GetIdentifier())
}

func TestRelease_ReelFirst(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	step := r.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Contains(t, step.Expect, redhat.VersionRegex)
	assert.Contains(t, step.Expect, redhat.NotRedHatBasedRegex)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestRelease_ReelMatch(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)

	// Positive test.
	step := r.ReelMatch(redhat.VersionRegex, "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, r.Result())

	r = redhat.NewRelease(testTimeoutDuration)

	// Negative test.
	step = r.ReelMatch(redhat.NotRedHatBasedRegex, "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, r.Result())

	// Error case.  Note, this shouldn't ever happen based on the FSM, but it is better to be defensive.
	step = r.ReelMatch("unknown regex", "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, r.Result())
}

func TestRelease_ReelTimeout(t *testing.T) {
	r := redhat.NewRelease(testTimeoutDuration)
	step := r.ReelTimeout()
	assert.Nil(t, step)
}

func TestRelease_ReelEOF(t *testing.T) {
	// just ensures no panics
	r := redhat.NewRelease(testTimeoutDuration)
	r.ReelEOF()
}

func TestRelease_VersionRegex(t *testing.T) {
	r := regexp.MustCompile(redhat.VersionRegex)
	// As we encounter more (and more) variations of /etc/redhat-release contents, do a quick compile time check to make
	// sure the VersionRegex should still match.
	assert.True(t, r.MatchString("Red Hat Enterprise Linux Server release 7.8 (Maipo)"))
	assert.True(t, r.MatchString("Red Hat Enterprise Linux release 8.2 (Ootpa)"))
	assert.False(t, r.MatchString("Won't match."))
}
