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

package hugepages_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	hp "github.com/test-network-function/test-network-function/pkg/tnf/handlers/hugepages"
)

func Test_NewHugepages(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	assert.Equal(t, testTimeoutDuration, newHp.Timeout())
	assert.Equal(t, newHp.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	firstStep := newHp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstPositiveEmpty(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	firstStep := newHp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputEmpty)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputEmpty, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	firstStep := newHp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccessEmpty(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	step := newHp.ReelMatch("", "", testInputEmpty)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newHp.Result())
	assert.Equal(t, hp.RhelDefaultHugepages, newHp.GetHugepages())
	assert.Equal(t, hp.RhelDefaultHugepagesz, newHp.GetHugepagesz())
}

func Test_ReelMatchSuccess(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	step := newHp.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newHp.Result())
	assert.Equal(t, testExpectedHugepages, newHp.GetHugepages())
	assert.Equal(t, testExpectedHugepagesz, newHp.GetHugepagesz())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration)
	assert.NotNil(t, newHp)
	newHp.ReelEOF()
}

const (
	testTimeoutDuration    = time.Second * 2
	testInputError         = ""
	testInputEmpty         = "KARGS\n"
	testInputSuccess       = "KARGS\n[hugepages=32 default_hugepagesz=32M]\n[hugepages=64 hugepagesz=1G]\n"
	testExpectedHugepages  = 64
	testExpectedHugepagesz = 1024 * 1024
)
