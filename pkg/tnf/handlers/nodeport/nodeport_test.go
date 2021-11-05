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

package nodeport_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	np "github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodeport"
)

func Test_NewNodePort(t *testing.T) {
	newNp(t)
	assert.Equal(t, testTimeoutDuration, testNp.Timeout())
	assert.Equal(t, testNp.Result(), tnf.ERROR)
}

func Test_Success(t *testing.T) {
	match(t, testSuccessInput, false)
}

func Test_Failure(t *testing.T) {
	match(t, testFailureInput, false)
}

func Test_Error(t *testing.T) {
	match(t, testErrorInput, true)
}

func Test_ReelEof(t *testing.T) {
	newNp(t)
	testNp.ReelEOF()
}

const (
	testTimeoutDuration   = time.Second * 2
	testPodNamespace      = "testPodNamespace"
	testErrorInput        = ""
	testSuccessInput      = "TYPE\n"
	testFailureInput      = "TYPE\nNodePort\n"
	testNumMatchesNoError = 1
	testNumMatchesError   = 0
)

var (
	testNp *np.NodePort // a NodePort object used by all tests. New object is created by calling newNp()
)

func newNp(t *testing.T) {
	testNp = np.NewNodePort(testTimeoutDuration, testPodNamespace)
	assert.NotNil(t, testNp)
}

func match(t *testing.T, input string, expectError bool) {
	newNp(t)
	firstStep := testNp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(input)
	if expectError {
		assert.Len(t, matches, testNumMatchesError)
		return
	}
	assert.Len(t, matches, testNumMatchesNoError)
	step := testNp.ReelMatch("", "", matches[0])
	assert.Nil(t, step)
}
