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

package operator_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/operator"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	testTimeoutDuration = time.Second * 2
	name                = "CSV_INSTALLED"
	namespace           = "test"
	command             = "oc get csv %s -n %s -o json | jq -r '.status.phase'"
)

var (
	stringExpectedStatus             = []string{string(testcases.NullFalse)}
	sliceExpectedStatus              = []string{"Running", "Installed"}
	sliceExpectedStatusInvalid       = []string{"Not_Running", "Not_Installed"}
	resultSliceExpectedStatus        = `["Running", "Installed"]`
	resultSliceExpectedStatusInvalid = `["Not_Running", "Not_Installed"]`
	args                             = strings.Split(fmt.Sprintf(command, name, namespace), " ")
)

func TestOperator_Args(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, args, c.Args())
}

func TestOperator_GetIdentifier(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, identifier.OperatorIdentifier, c.GetIdentifier())
}

func TestOperator_ReelFirst(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelFirst()
	assert.Equal(t, "", step.Execute)
	assert.Equal(t, []string{testcases.GetOutRegExp(testcases.AllowAll)}, step.Expect)
	assert.Equal(t, testTimeoutDuration, step.Timeout)
}

func TestOperator_ReelEof(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	// just ensures lack of panic
	c.ReelEOF()
}

func TestOperator_ReelTimeout(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelTimeout()
	assert.Nil(t, step)
}

func TestOperatorTest_ReelMatch_String(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", "null")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestOperatorTest_Facts(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", "null")
	assert.Nil(t, step)
	assert.NotNil(t, c.Facts())
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestOperatorTest_ReelMatch_Array_Allow_Deny_ISNULL(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", `null`)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestOperatorTest_ReelMatch_Array_Allow_Match(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatus)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, c.Result())
}

func TestOperatorTest_ReelMatch_Array_Allow_NoMatch(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatusInvalid)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestOperatorTest_ReelMatch_Array_Deny_Match(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, sliceExpectedStatus, testcases.ArrayType, testcases.Deny, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatus)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestOperatorTest_ReelMatch_Array_Deny_NotMatch(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, sliceExpectedStatusInvalid, testcases.ArrayType, testcases.Deny, testTimeoutDuration)
	step := c.ReelMatch("", "", resultSliceExpectedStatusInvalid)
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestOperatorTest_ReelMatch_StringNoFound(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.StringType, testcases.Allow, testTimeoutDuration)
	step := c.ReelMatch("", "", "not_null")
	assert.Nil(t, step)
	assert.Equal(t, tnf.ERROR, c.Result())
}

func TestNewOperator(t *testing.T) {
	c := operator.NewOperator(args, name, namespace, stringExpectedStatus, testcases.ArrayType, testcases.Allow, testTimeoutDuration)
	assert.Equal(t, tnf.ERROR, c.Result())
	assert.Equal(t, testTimeoutDuration, c.Timeout())
}
