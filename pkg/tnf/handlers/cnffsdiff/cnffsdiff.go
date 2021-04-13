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

package cnffsdiff

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// CnfFsDiff provides a fsdiff test implemented using the "diff-fs.sh" script in the master container.
type CnfFsDiff struct {
	result  int
	timeout time.Duration
	args    []string
	diff    string
}

const (
	// SuccessfulOutputRegex matches a successfully run "fsdiff" command.  That does not mean that no errors or drops
	// occurred during the test.
	SuccessfulOutputRegex = `(?m)empty\n`
	// AcceptAllRegex matches all strings
	AcceptAllRegex = `(?m)(.|\n)+`
)

// Args returns the command line args for the test.
func (p *CnfFsDiff) Args() []string {
	return p.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (p *CnfFsDiff) GetIdentifier() identifier.Identifier {
	return identifier.CnfFsDiffIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (p *CnfFsDiff) Timeout() time.Duration {
	return p.timeout
}

// Result returns the test result.
func (p *CnfFsDiff) Result() int {
	return p.result
}

// ReelFirst returns a step which expects the fs diff within the test timeout.
func (p *CnfFsDiff) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  p.GetReelFirstRegularExpressions(),
		Timeout: p.timeout,
	}
}

// ReelMatch checks if the test passed the first regex which means there were no installation on the container
// or the second regex which accepts everything and means that something in the container was installed.
func (p *CnfFsDiff) ReelMatch(pattern, before, match string) *reel.Step {
	if pattern == SuccessfulOutputRegex {
		p.result = tnf.SUCCESS
	} else {
		p.result = tnf.FAILURE
	}
	return nil
}

// ReelTimeout returns a step which kills the fs diff test by sending it ^C.
func (p *CnfFsDiff) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  fs diff requires no intervention on eof.
func (p *CnfFsDiff) ReelEOF() {
}

// Command returns command line args for checking the fs difference between a container and it's image
func Command(containerID string) []string {
	return []string{"/diff-fs.sh", containerID}
}

// NewFsDiff creates a new `FsDiff` test which checks the fs difference between a container and it's image
func NewFsDiff(timeout time.Duration, containerID string) *CnfFsDiff {
	return &CnfFsDiff{
		result:  tnf.ERROR,
		timeout: timeout,
		args:    Command(containerID),
	}
}

// GetReelFirstRegularExpressions returns the regular expressions used for matching in ReelFirst.
func (p *CnfFsDiff) GetReelFirstRegularExpressions() []string {
	return []string{SuccessfulOutputRegex, AcceptAllRegex}
}
