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

package nodenames_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	nn "github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodenames"
)

func Test_NewNodeNames(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, testLabelsMap)
	assert.NotNil(t, newNn)
	assert.Equal(t, testTimeoutDuration, newNn.Timeout())
	assert.Equal(t, newNn.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, nil)
	assert.NotNil(t, newNn)
	firstStep := newNn.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstPositiveEmpty(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, nil)
	assert.NotNil(t, newNn)
	firstStep := newNn.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputFailure)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputFailure, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, nil)
	assert.NotNil(t, newNn)
	firstStep := newNn.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, nil)
	assert.NotNil(t, newNn)
	step := newNn.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newNn.Result())
	assert.Len(t, newNn.GetNodeNames(), 2)
}

func Test_ReelMatchFail(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, nil)
	assert.NotNil(t, newNn)
	step := newNn.ReelMatch("", "", testInputFailure)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, newNn.Result())
	assert.Len(t, newNn.GetNodeNames(), 0)
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newNn := nn.NewNodeNames(testTimeoutDuration, nil)
	assert.NotNil(t, newNn)
	newNn.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInputError      = ""
	testInputFailure    = "NAME\n"
	testInputSuccess    = "NAME\nnode1-fga23-vm\nnode2-xda3s-vm\n"
)

var (
	emptyString   = ""
	valueString   = "value"
	testLabelsMap = map[string]*string{
		"nameOnly":   nil,
		"emptyValue": &emptyString,
		"name":       &valueString,
	}
)
