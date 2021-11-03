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

package mckernelarguments

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

// McKernelArguments holds information regarding the kernel arguments of an mc.
type McKernelArguments struct {
	kernelArguments string // Output variable that stores the kernel arguments of the mc
	result          int
	timeout         time.Duration
	args            []string
}

// NewMcKernelArguments creates a NodeMcName tnf.Test.
func NewMcKernelArguments(timeout time.Duration, mcName string) *McKernelArguments {
	return &McKernelArguments{
		timeout: timeout,
		result:  tnf.ERROR,
		args:    []string{dependencies.OcBinaryName, "get", "mc", mcName, "-o", "jsonpath=\"{.spec.kernelArguments}\""},
	}
}

// GetKernelArguments returns the kernel arguments of the mc extracted while running the McKernelArguments tnf.Test.
func (mka *McKernelArguments) GetKernelArguments() string {
	return mka.kernelArguments
}

// Args returns the command line args for the test.
func (mka *McKernelArguments) Args() []string {
	return mka.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (mka *McKernelArguments) GetIdentifier() identifier.Identifier {
	return identifier.NodeMcNameIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (mka *McKernelArguments) Timeout() time.Duration {
	return mka.timeout
}

// Result returns the test result.
func (mka *McKernelArguments) Result() int {
	return mka.result
}

// ReelFirst returns a step which expects the mc kernel arguments within the test timeout.
func (mka *McKernelArguments) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: mka.timeout,
	}
}

// ReelMatch ensures that there are no McKernelArguments matched in the command output.
func (mka *McKernelArguments) ReelMatch(_, _, match string, status int) *reel.Step {
	mka.kernelArguments = match
	mka.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (mka *McKernelArguments) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (mka *McKernelArguments) ReelEOF() {
}
