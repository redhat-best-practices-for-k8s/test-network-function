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
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	assert.Equal(t, testTimeoutDuration, newHp.Timeout())
	assert.Equal(t, newHp.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	firstStep := newHp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstPositiveEmpty(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	firstStep := newHp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputEmpty)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputEmpty, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	firstStep := newHp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccessEmpty(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	step := newHp.ReelMatch("", "", testInputEmpty, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newHp.Result())
	assert.Equal(t, hp.RhelDefaultHugepages, newHp.GetHugepages())
	assert.Equal(t, hp.RhelDefaultHugepagesz, newHp.GetHugepagesz())
}

func Test_ReelMatchSuccess(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	step := newHp.ReelMatch("", "", testInputSuccess, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newHp.Result())
	assert.Equal(t, testExpectedHugepages, newHp.GetHugepages())
	assert.Equal(t, testExpectedHugepagesz, newHp.GetHugepagesz())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newHp := hp.NewHugepages(testTimeoutDuration, testMachineConfig)
	assert.NotNil(t, newHp)
	newHp.ReelEOF()
}

const (
	testTimeoutDuration    = time.Second * 2
	testMachineConfig      = "worker-rendered-1"
	testInputError         = ""
	testInputEmpty         = "KARGS\n"
	testInputSuccess       = "KARGS\n[skew_tick=1 nohz=on rcu_nocbs=2-19,22-39,42-59,62-79 tuned.non_isolcpus=30000300,00300003 intel_pstate=disable nosoftlockup tsc=nowatchdog intel_iommu=on iommu=pt isolcpus=managed_irq,2-19,22-39,42-59,62-79 systemd.cpu_affinity=0,1,40,41,20,21,60,61 hugepages=32 default_hugepagesz=32M hugepages=64 hugepagesz=1G nmi_watchdog=0 audit=0 mce=off processor.max_cstate=1 idle=poll intel_idle.max_cstate=0]\n"
	testExpectedHugepages  = 64
	testExpectedHugepagesz = 1024 * 1024
)
