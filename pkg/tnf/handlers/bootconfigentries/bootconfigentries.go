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

package bootconfigentries

import (
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `(?s).+`
)

// BootConfigEntries holds information regarding all files in the /boot/loader/entries/ directory.
type BootConfigEntries struct {
	bootConfigEntries []string // Output variable that stores all file names in the /boot/loader/entries/ directory
	result            int
	timeout           time.Duration
	args              []string
}

// NewBootConfigEntries creates a BootConfigEntries tnf.Test.
func NewBootConfigEntries(timeout time.Duration) *BootConfigEntries {
	return &BootConfigEntries{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"ls /host/boot/loader/entries/",
		},
	}
}

// GetBootConfigEntries returns the boot config entries extracted from /boot/loader/entries/ directory while running the BootConfigEntries tnf.Test.
func (bce *BootConfigEntries) GetBootConfigEntries() []string {
	return bce.bootConfigEntries
}

// Args returns the command line args for the test.
func (bce *BootConfigEntries) Args() []string {
	return bce.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (bce *BootConfigEntries) GetIdentifier() identifier.Identifier {
	return identifier.CurrentKernelCmdlineArgsURLIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (bce *BootConfigEntries) Timeout() time.Duration {
	return bce.timeout
}

// Result returns the test result.
func (bce *BootConfigEntries) Result() int {
	return bce.result
}

// ReelFirst returns a step which expects the grub kernel arguments within the test timeout.
func (bce *BootConfigEntries) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: bce.timeout,
	}
}

// ReelMatch ensures that there are no GrubKernelCmdlineArgs matched in the command output.
func (bce *BootConfigEntries) ReelMatch(_, _, match string, status int) *reel.Step {
	splitMatch := strings.Split(match, "\n")
	bce.bootConfigEntries = splitMatch[0 : len(splitMatch)-1]
	bce.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (bce *BootConfigEntries) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (bce *BootConfigEntries) ReelEOF() {
}
