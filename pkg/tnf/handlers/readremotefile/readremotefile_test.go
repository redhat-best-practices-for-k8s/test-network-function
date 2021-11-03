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

package readremotefile_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/readremotefile"
)

func TestReadRemoteFile(t *testing.T) {
	newReadRemoteFile := readremotefile.NewReadRemoteFile(testTimeoutDuration, testNodeName, testRemotePath)
	assert.NotNil(t, newReadRemoteFile)
	assert.Equal(t, tnf.ERROR, newReadRemoteFile.Result())
}

func Test_ReelFirst(t *testing.T) {
	newReadRemoteFile := readremotefile.NewReadRemoteFile(testTimeoutDuration, testNodeName, testRemotePath)
	assert.NotNil(t, newReadRemoteFile)
	firstStep := newReadRemoteFile.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newReadRemoteFile := readremotefile.NewReadRemoteFile(testTimeoutDuration, testNodeName, testRemotePath)
	assert.NotNil(t, newReadRemoteFile)
	step := newReadRemoteFile.ReelMatch("", "", testInput, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newReadRemoteFile.Result())
}

// Just ensure there are no panics.
func Test_ReelEOF(t *testing.T) {
	newReadRemoteFile := readremotefile.NewReadRemoteFile(testTimeoutDuration, testNodeName, testRemotePath)
	assert.NotNil(t, newReadRemoteFile)
	newReadRemoteFile.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = ` /usr/lib/sysctl.d/, /run/sysctl.d/, and /etc/sysctl.d/.
	#
	# Vendors settings live in /usr/lib/sysctl.d/.
	# To override a whole file, create a new file with the same in
	# /etc/sysctl.d/ and put new settings there. To override
	# only specific settings, add a file with a lexically later
	# name in /etc/sysctl.d/ and put new settings there.
	#
	# For more information, see sysctl.conf(5) and sysctl.d(5).
	`
	testNodeName   = "crc-l6qvn-master-0"
	testRemotePath = "/etc/sysctl.conf"
)
