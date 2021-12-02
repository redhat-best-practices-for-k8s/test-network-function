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

package autodiscover

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildContainersFromPodResource(t *testing.T) {
	subjectPod := loadPodResource(testSubjectFilePath)
	subjectContainers := buildContainersFromPodResource(&subjectPod)
	assert.Equal(t, 1, len(subjectContainers))

	assert.Equal(t, "tnf", subjectContainers[0].Namespace)
	assert.Equal(t, "I'mAPodName", subjectContainers[0].PodName)
	assert.Equal(t, "I'mAContainer", subjectContainers[0].ContainerName)

	// Check correct order of precedence for network devices
	assert.Equal(t, "eth0", subjectContainers[0].DefaultNetworkDevice)

	// test-network-function.com/multusips should be used for the test subject container.
	assert.Equal(t, 2, len(subjectContainers[0].MultusIPAddressesPerNet))
	assert.Equal(t, "3.3.3.3", subjectContainers[0].MultusIPAddressesPerNet["default/macvlan-conf1"][0])
	assert.Equal(t, "4.4.4.4", subjectContainers[0].MultusIPAddressesPerNet["default/macvlan-conf2"][0])
}
