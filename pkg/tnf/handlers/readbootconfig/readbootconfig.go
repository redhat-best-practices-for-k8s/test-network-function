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

package readbootconfig

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `(?s).+`
)

// ReadBootConfig holds information regarding boot config in /boot directory.
type ReadBootConfig struct {
	bootConfig string // Output variable that stores the boot config as specified in the /boot directory
	result     int
	timeout    time.Duration
	args       []string
}

// NewReadBootConfig creates a ReadBootConfig tnf.Test.
func NewReadBootConfig(timeout time.Duration, nodeName /*, entryName*/ string) *ReadBootConfig {
	return &ReadBootConfig{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			// "echo", "\"cat /host/boot/loader/entries/" + entryName + "\"", "|", "oc", "debug", "--preserve-pod=true", "node/" + nodeName,
			"echo", "\"cat /host/boot/loader/entries/\\`ls /host/boot/loader/entries/ | sort | tail -n 1\\`\"", "|", "oc", "debug", "-q", "node/" + nodeName,
		},
	}
}

// GetBootConfig returns the boot config extracted from /boot directory while running the ReadBootConfig tnf.Test.
func (bce *ReadBootConfig) GetBootConfig() string {
	return bce.bootConfig
}

// Args returns the command line args for the test.
func (bce *ReadBootConfig) Args() []string {
	return bce.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (bce *ReadBootConfig) GetIdentifier() identifier.Identifier {
	return identifier.CurrentKernelCmdlineArgsURLIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (bce *ReadBootConfig) Timeout() time.Duration {
	return bce.timeout
}

// Result returns the test result.
func (bce *ReadBootConfig) Result() int {
	return bce.result
}

// ReelFirst returns a step which expects the grub kernel arguments within the test timeout.
func (bce *ReadBootConfig) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: bce.timeout,
	}
}

// ReelMatch ensures that there are no GrubKernelCmdlineArgs matched in the command output.
func (bce *ReadBootConfig) ReelMatch(_, _, match string) *reel.Step {
	bce.bootConfig = match
	bce.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (bce *ReadBootConfig) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (bce *ReadBootConfig) ReelEOF() {
}
