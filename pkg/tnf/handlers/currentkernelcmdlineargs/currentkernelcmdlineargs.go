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

package currentkernelcmdlineargs

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/dependencies"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `.+`
)

// CurrentKernelCmdlineArgs holds information regarding /proc/cmdline.
type CurrentKernelCmdlineArgs struct {
	kernelArguments string // Output variable that stores the kernel arguments of the mc
	result          int
	timeout         time.Duration
	args            []string
}

// NewCurrentKernelCmdlineArgs creates a CurrentKernelCmdlineArgs tnf.Test.
func NewCurrentKernelCmdlineArgs(timeout time.Duration) *CurrentKernelCmdlineArgs {
	return &CurrentKernelCmdlineArgs{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    []string{dependencies.CatBinaryName, "/proc/cmdline"},
	}
}

// GetKernelArguments returns the kernel arguments extracted from /proc/cmdline while running the CurrentKernelCmdlineArgs tnf.Test.
func (cmdlineArgs *CurrentKernelCmdlineArgs) GetKernelArguments() string {
	return cmdlineArgs.kernelArguments
}

// Args returns the command line args for the test.
func (cmdlineArgs *CurrentKernelCmdlineArgs) Args() []string {
	return cmdlineArgs.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (cmdlineArgs *CurrentKernelCmdlineArgs) GetIdentifier() identifier.Identifier {
	return identifier.CurrentKernelCmdlineArgsURLIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (cmdlineArgs *CurrentKernelCmdlineArgs) Timeout() time.Duration {
	return cmdlineArgs.timeout
}

// Result returns the test result.
func (cmdlineArgs *CurrentKernelCmdlineArgs) Result() int {
	return cmdlineArgs.result
}

// ReelFirst returns a step which expects the /proc/cmdline kernel arguments within the test timeout.
func (cmdlineArgs *CurrentKernelCmdlineArgs) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: cmdlineArgs.timeout,
	}
}

// ReelMatch ensures that there are no McKernelArguments matched in the command output.
func (cmdlineArgs *CurrentKernelCmdlineArgs) ReelMatch(_, _, match string, status int) *reel.Step {
	cmdlineArgs.kernelArguments = match
	cmdlineArgs.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (cmdlineArgs *CurrentKernelCmdlineArgs) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (cmdlineArgs *CurrentKernelCmdlineArgs) ReelEOF() {
}
