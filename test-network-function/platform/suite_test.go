// Copyright (C) 2020-2022 Red Hat, Inc.
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

package platform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// Sizes, in KBs.
	oneGB = 1024 * 1024 // 1G
	twoMB = 2 * 1024    // 2M: also RHEL's default hugepages size
)

var (
	// No hugepages params
	testKernelArgsHpNoParams = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "nmi_watchdog=0"}

	// Single param
	testKernelArgsHpSingleParam1 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "hugepages=16", "nmi_watchdog=0"}
	testKernelArgsHpSingleParam2 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "default_hugepagesz=1G", "nmi_watchdog=0"}
	testKernelArgsHpSingleParam3 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "default_hugepagesz=2M", "nmi_watchdog=0"}
	testKernelArgsHpSingleParam4 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "hugepagesz=1G", "nmi_watchdog=0"}

	// Default size + size only
	testKernelArgsHpDefParamsOnly = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "default_hugepagesz=1G", "hugepagesz=1G", "nmi_watchdog=0"}

	// size + count pairs.
	testKernelArgsHpPair1 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "hugepagesz=1G", "hugepages=16", "nmi_watchdog=0"}
	testKernelArgsHpPair2 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "hugepagesz=2M", "hugepages=256", "nmi_watchdog=0"}
	testKernelArgsHpPair3 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "hugepagesz=1G", "hugepages=16", "hugepagesz=2M", "hugepages=256", "nmi_watchdog=0"}

	// default size + (size+count) pairs
	testKernelArgsHpDefSizePlusPairs1 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "default_hugepagesz=2M", "hugepagesz=1G", "hugepages=16", "nmi_watchdog=0"}
	testKernelArgsHpDefSizePlusPairs2 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "default_hugepagesz=1G", "hugepagesz=2M", "hugepages=256", "nmi_watchdog=0"}
	testKernelArgsHpDefSizePlusPairs3 = []string{"systemd.cpu_affinity=0,1,40,41,20,21,60,61", "default_hugepagesz=1G", "hugepagesz=1G", "hugepages=16", "hugepagesz=2M", "hugepages=256", "nmi_watchdog=0"}
)

func Test_printTainted(t *testing.T) {
	assert.Equal(t, printTainted(2048), "workaround for bug in platform firmware applied, ")
	assert.Equal(t, printTainted(32769), "proprietary module was loaded, kernel has been live patched, ")
}

//nolint:funlen
func Test_hugepagesFromKernelArgsFunc(t *testing.T) {
	testCases := []struct {
		expectedHugepagesDefSize int
		expectedHugepagesPerSize map[int]int
		kernelArgs               []string
	}{
		// No params
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{twoMB: 0},
			kernelArgs:               testKernelArgsHpNoParams,
		},

		// Single params TCs.
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{twoMB: 16},
			kernelArgs:               testKernelArgsHpSingleParam1,
		},
		{
			expectedHugepagesDefSize: oneGB,
			expectedHugepagesPerSize: map[int]int{oneGB: 0},
			kernelArgs:               testKernelArgsHpSingleParam2,
		},
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{twoMB: 0},
			kernelArgs:               testKernelArgsHpSingleParam3,
		},
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{oneGB: 0},
			kernelArgs:               testKernelArgsHpSingleParam4,
		},
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{oneGB: 16},
			kernelArgs:               testKernelArgsHpPair1,
		},

		// Default sizes Tc:
		{
			expectedHugepagesDefSize: oneGB,
			expectedHugepagesPerSize: map[int]int{oneGB: 0},
			kernelArgs:               testKernelArgsHpDefParamsOnly,
		},

		// size+count pairs
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{oneGB: 16},
			kernelArgs:               testKernelArgsHpPair1,
		},
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{twoMB: 256},
			kernelArgs:               testKernelArgsHpPair2,
		},
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{oneGB: 16, twoMB: 256},
			kernelArgs:               testKernelArgsHpPair3,
		},

		// default size + (size+count) pairs
		{
			expectedHugepagesDefSize: twoMB,
			expectedHugepagesPerSize: map[int]int{twoMB: 0, oneGB: 16},
			kernelArgs:               testKernelArgsHpDefSizePlusPairs1,
		},
		{
			expectedHugepagesDefSize: oneGB,
			expectedHugepagesPerSize: map[int]int{oneGB: 0, twoMB: 256},
			kernelArgs:               testKernelArgsHpDefSizePlusPairs2,
		},
		{
			expectedHugepagesDefSize: oneGB,
			expectedHugepagesPerSize: map[int]int{oneGB: 16, twoMB: 256},
			kernelArgs:               testKernelArgsHpDefSizePlusPairs3,
		},
	}

	mc := machineConfig{}
	for _, tc := range testCases {
		// Prepare fake MC object: only KernelArguments is needed.
		mc.Spec.KernelArguments = tc.kernelArgs

		// Call the function under test.
		hugepagesPerSize, defSize := getMcHugepagesFromMcKernelArguments(&mc)

		assert.Equal(t, defSize, tc.expectedHugepagesDefSize)
		assert.Equal(t, hugepagesPerSize, tc.expectedHugepagesPerSize)
	}
}
