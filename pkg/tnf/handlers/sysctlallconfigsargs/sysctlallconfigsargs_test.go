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

package sysctlallconfigsargs_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/sysctlallconfigsargs"
)

func TestNewSysctlAllConfigsArgs(t *testing.T) {
	newSysctlAllConfigsArgs := sysctlallconfigsargs.NewSysctlAllConfigsArgs(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newSysctlAllConfigsArgs)
	assert.Equal(t, tnf.ERROR, newSysctlAllConfigsArgs.Result())
}

func Test_ReelFirst(t *testing.T) {
	newSysctlAllConfigsArgs := sysctlallconfigsargs.NewSysctlAllConfigsArgs(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newSysctlAllConfigsArgs)
	firstStep := newSysctlAllConfigsArgs.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newSysctlAllConfigsArgs := sysctlallconfigsargs.NewSysctlAllConfigsArgs(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newSysctlAllConfigsArgs)
	step := newSysctlAllConfigsArgs.ReelMatch("", "", testInput)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newSysctlAllConfigsArgs.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newSysctlAllConfigsArgs := sysctlallconfigsargs.NewSysctlAllConfigsArgs(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newSysctlAllConfigsArgs)
	newSysctlAllConfigsArgs.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = `* Applying /usr/lib/sysctl.d/10-coreos-ratelimit-kmsg.conf ...
	kernel.printk_devkmsg = ratelimit
	* Applying /usr/lib/sysctl.d/10-default-yama-scope.conf ...
	kernel.yama.ptrace_scope = 0
	* Applying /usr/lib/sysctl.d/50-coredump.conf ...
	kernel.core_pattern = |/usr/lib/systemd/systemd-coredump %P %u %g %s %t %c %h %e
	* Applying /usr/lib/sysctl.d/50-default.conf ...
	kernel.sysrq = 16
	kernel.core_uses_pid = 1
	kernel.kptr_restrict = 1
	net.ipv4.conf.all.rp_filter = 1
	net.ipv4.conf.all.accept_source_route = 0
	net.ipv4.conf.all.promote_secondaries = 1
	net.core.default_qdisc = fq_codel
	fs.protected_hardlinks = 1
	fs.protected_symlinks = 1
	* Applying /usr/lib/sysctl.d/50-libkcapi-optmem_max.conf ...
	net.core.optmem_max = 81920
	* Applying /usr/lib/sysctl.d/50-pid-max.conf ...
	kernel.pid_max = 4194304
	* Applying /etc/sysctl.d/99-sysctl.conf ...
	* Applying /etc/sysctl.d/forward.conf ...
	fs.inotify.max_user_instances = 10000
	net.ipv6.conf.all.forwarding = 1
	* Applying /etc/sysctl.d/inotify.conf ...
	fs.inotify.max_user_watches = 65536
	fs.inotify.max_user_instances = 8192
	* Applying /etc/sysctl.conf ...	
	`
	testNodeName = "crc-l6qvn-master-0"
)
