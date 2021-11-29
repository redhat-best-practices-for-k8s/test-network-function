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

package containerid_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/containerid"
)

func TestArgs(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	assert.Equal(t, "cat /proc/self/cgroup", strings.Join(c.Args(), " "))
}

func TestGetIdentifier(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	identifier := c.GetIdentifier()
	assert.Equal(t, "http://test-network-function.com/tests/generic/containerId", identifier.URL)
	assert.Equal(t, "v1.0.0", identifier.SemanticVersion)
}

func TestTimeout(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	assert.Equal(t, time.Second*5, c.Timeout())
}

func TestResult(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestReelFirst(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	rf := c.ReelFirst()
	assert.Equal(t, containerid.SuccessfulOutputRegex, rf.Expect[0])
	assert.Equal(t, time.Second*5, rf.Timeout)
}

func TestReelMatch(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	step := c.ReelMatch("", "", "crio-test.scope")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
	step = c.ReelMatch("", "", "crio-test-scope")
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, c.Result())
}

func TestReelTimeout(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	assert.Nil(t, c.ReelTimeout())
}

func TestGetID(t *testing.T) {
	c := containerid.NewContainerID(5 * time.Second)
	assert.Equal(t, c.GetID(), "")
}
