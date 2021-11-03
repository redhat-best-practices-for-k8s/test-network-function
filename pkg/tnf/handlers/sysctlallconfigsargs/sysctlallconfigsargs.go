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

package sysctlallconfigsargs

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `(?s).+`
)

// SysctlAllConfigsArgs holds information regarding the all args in all sysctl conf files in ordered
// in the same way as they are loaded by the os
type SysctlAllConfigsArgs struct {
	cmdOutput string // Output variable that stores the return value of "sysctl --system".
	result    int
	timeout   time.Duration
	args      []string
}

// NewSysctlAllConfigsArgs creates a SysctlAllConfigsArgs tnf.Test.
func NewSysctlAllConfigsArgs(timeout time.Duration) *SysctlAllConfigsArgs {
	return &SysctlAllConfigsArgs{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"sysctl --system",
		},
	}
}

// GetSysctlAllConfigsArgs returns the output of the "sysctl --system" command while running the SysctlConfigFilesList tnf.Test.
func (handler *SysctlAllConfigsArgs) GetSysctlAllConfigsArgs() string {
	return handler.cmdOutput
}

// Args returns the command line args for the test.
func (handler *SysctlAllConfigsArgs) Args() []string {
	return handler.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (handler *SysctlAllConfigsArgs) GetIdentifier() identifier.Identifier {
	return identifier.SysctlConfigFilesListURLIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (handler *SysctlAllConfigsArgs) Timeout() time.Duration {
	return handler.timeout
}

// Result returns the test result.
func (handler *SysctlAllConfigsArgs) Result() int {
	return handler.result
}

// ReelFirst returns a step which expects the result within the test timeout.
func (handler *SysctlAllConfigsArgs) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: handler.timeout,
	}
}

// ReelMatch passes the result to cmdOutput.
func (handler *SysctlAllConfigsArgs) ReelMatch(_, _, match string, status int) *reel.Step {
	handler.cmdOutput = match
	handler.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (handler *SysctlAllConfigsArgs) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (handler *SysctlAllConfigsArgs) ReelEOF() {
}
