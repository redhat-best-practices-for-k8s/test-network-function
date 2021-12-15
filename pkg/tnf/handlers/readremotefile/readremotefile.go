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

package readremotefile

import (
	"time"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	successfulOutputRegex = `(?s).+`
)

// ReadRemoteFile holds information regarding remote file contents at a specified node and path.
type ReadRemoteFile struct {
	remoteFileContents string // Output variable that stores the remote file contents
	result             int
	timeout            time.Duration
	args               []string
}

// NewReadRemoteFile creates a ReadRemoteFile tnf.Test.
func NewReadRemoteFile(timeout time.Duration, nodeName, filePath string) *ReadRemoteFile {
	return &ReadRemoteFile{
		timeout: timeout,
		result:  tnf.ERROR,
		args: []string{
			"echo", "\"cat /host" + filePath + "\"", "|", "oc", "debug", "-q", "node/" + nodeName,
		},
	}
}

// GetRemoteFileContents returns the file contents extracted from the specified path while running the ReadRemoteFile tnf.Test.
func (handler *ReadRemoteFile) GetRemoteFileContents() string {
	return handler.remoteFileContents
}

// Args returns the command line args for the test.
func (handler *ReadRemoteFile) Args() []string {
	return handler.args
}

// GetIdentifier returns the tnf.Test specific identifier.
func (handler *ReadRemoteFile) GetIdentifier() identifier.Identifier {
	return identifier.ReadRemoteFileURLIdentifier
}

// Timeout returns the timeout in seconds for the test.
func (handler *ReadRemoteFile) Timeout() time.Duration {
	return handler.timeout
}

// Result returns the test result.
func (handler *ReadRemoteFile) Result() int {
	return handler.result
}

// ReelFirst returns a step which expects the grub kernel arguments within the test timeout.
func (handler *ReadRemoteFile) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{successfulOutputRegex},
		Timeout: handler.timeout,
	}
}

// ReelMatch just forwards the output to handler.remoteFileContents.
func (handler *ReadRemoteFile) ReelMatch(_, _, match string) *reel.Step {
	handler.remoteFileContents = match
	handler.result = tnf.SUCCESS
	return nil
}

// ReelTimeout does nothing;  no action is necessary upon timeout.
func (handler *ReadRemoteFile) ReelTimeout() *reel.Step {
	return nil
}

// ReelEOF does nothing;  no action is necessary on EOF.
func (handler *ReadRemoteFile) ReelEOF() {
}
