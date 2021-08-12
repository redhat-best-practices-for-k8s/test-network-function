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

package readbootconfig_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/bootconfigentries"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/readbootconfig"
)

func TestReadBootConfig(t *testing.T) {
	newReadBootConfig := readbootconfig.NewReadBootConfig(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newReadBootConfig)
	assert.Equal(t, tnf.ERROR, newReadBootConfig.Result())
}

func Test_ReelFirst(t *testing.T) {
	newReadBootConfig := readbootconfig.NewReadBootConfig(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newReadBootConfig)
	firstStep := newReadBootConfig.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInput)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInput, matches[0])
}

func Test_ReelMatch(t *testing.T) {
	newBootConfig := bootconfigentries.NewBootConfigEntries(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newBootConfig)
	step := newBootConfig.ReelMatch("", "", testInput)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newBootConfig.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newBootConfig := bootconfigentries.NewBootConfigEntries(testTimeoutDuration, testNodeName)
	assert.NotNil(t, newBootConfig)
	newBootConfig.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testInput           = `title Red Hat Enterprise Linux CoreOS
	version 2
	options random.trust_cpu=on console=tty0
	linux /ostree/rhcos-8db99645874
	initrd /ostree/rhcos-8db99645874`
	testNodeName = "crc-l6qvn-master-0"
	// testBootEntryName = "ostree-2-rhcos.conf"
)
