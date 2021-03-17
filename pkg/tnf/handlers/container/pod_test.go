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

package container_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/container"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	testTimeoutDuration = time.Second * 2
	name                = "HOST_NETWORK_CHECK"
	namespace           = "test"
	command             = "oc get pod  %s  -n %s -o json  | jq -r '.spec.hostNetwork'"
)

var (
	stringExpectedStatus             = []string{string(testcases.NullFalse)}
	sliceExpectedStatus              = []string{"NET_ADMIN", "SYS_TIME"}
	resultSliceExpectedStatus        = `["NET_ADMIN", "SYS_TIME"]`
	resultSliceExpectedStatusInvalid = `["NO_NET_ADMIN", "NO_SYS_TIME"]`
	IsNull                           = "null"
	IsNotNull                        = "not_null"
	args                             = strings.Split(fmt.Sprintf(command, name, namespace), " ")
)

func TestPod_Args(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, args, c.Args())
}

func TestPod_GetIdentifier(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, identifier.PodIdentifier, c.GetIdentifier())
}

func TestPod_ReelFirst(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Equal(t, []string{testcases.GetOutRegExp(testcases.AllowAll)}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestPod_ReelEof(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	// just ensures lack of panic
	c.ReelEOF()
}

func TestPod_ReelTimeout(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelTimeout()
	assert.Nil(t, step)
}

func TestPodTest_ReelMatch_String(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNull)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_Facts(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNull)
	assert.Nil(t, step)
	assert.NotNil(t, c.Facts())
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatch_String_NotFound(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNotNull)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatch_Array_Allow_Deny_ISNULL(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", IsNull)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatchArray_Allow_Match(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatus)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestPodTest_ReelMatch_Array_Allow_NoMatch(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatusInvalid)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatch_Array_Deny_Match(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Deny, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatus)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestPodTest_ReelMatch_Array_Deny_NotMatch(t *testing.T) {
	c := container.NewPod(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Deny, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatusInvalid)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestNewPod(t *testing.T) {
	c := container.NewPod(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, c.Result())
	assert.Equal(t, testTimeoutDuration, c.Timeout())
}
