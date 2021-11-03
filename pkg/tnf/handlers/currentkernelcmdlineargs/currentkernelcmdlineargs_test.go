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

package currentkernelcmdlineargs_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/currentkernelcmdlineargs"
)

func TestNewCurrentKernelCmdlineArgs(t *testing.T) {
	newCurrentKernelCmdlineArgs := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(testTimeoutDuration)
	assert.NotNil(t, newCurrentKernelCmdlineArgs)
	assert.Equal(t, tnf.ERROR, newCurrentKernelCmdlineArgs.Result())
}

func Test_ReelFirst(t *testing.T) {
	newCurrentKernelCmdlineArgs := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(testTimeoutDuration)
	assert.NotNil(t, newCurrentKernelCmdlineArgs)
	firstStep := newCurrentKernelCmdlineArgs.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newCurrentKernelCmdlineArgs := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(testTimeoutDuration)
	assert.NotNil(t, newCurrentKernelCmdlineArgs)
	step := newCurrentKernelCmdlineArgs.ReelMatch("", "", testInput, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newCurrentKernelCmdlineArgs.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newCurrentKernelCmdlineArgs := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(testTimeoutDuration)
	assert.NotNil(t, newCurrentKernelCmdlineArgs)
	newCurrentKernelCmdlineArgs.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = `BOOT_IMAGE=(hd0,gpt3)/ostree/rhcos-8db996458745a61fa6759e8612cc44d429af0417584807411e38b991b9bcedb9/vmlinuz-4.18.0-240.10.1.el8_3.x86_64 random.trust_cpu=on console=tty0 console=ttyS0,115200n8 ignition.platform.id=qemu ostree=/ostree/boot.0/rhcos/8db996458745a61fa6759e8612cc44d429af0417584807411e38b991b9bcedb9/0 root=UUID=a9fd5b50-93`
)
