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

package autodiscover_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
)

func TestGetGenericContainers(t *testing.T) {
	orchestratorPod := loadPodResource(testOrchestratorFilePath)
	orchestratorContainers := autodiscover.GetGenericContainers(&orchestratorPod)
	assert.Equal(t, 1, len(orchestratorContainers))

	subjectPod := loadPodResource(testSubjectFilePath)
	subjectContainers := autodiscover.GetGenericContainers(&subjectPod)
	assert.Equal(t, 1, len(subjectContainers))

	assert.Equal(t, "tnf", orchestratorContainers[0].Namespace)
	assert.Equal(t, "I'mAPodName", orchestratorContainers[0].PodName)
	assert.Equal(t, "I'mAContainer", orchestratorContainers[0].ContainerName)

	// Check correct order of precedence for network devices
	assert.Equal(t, "eth0", orchestratorContainers[0].DefaultNetworkDevice)
	assert.NotEqual(t, "LowerPriorityInterface", orchestratorContainers[0].DefaultNetworkDevice)
	assert.Equal(t, "eth1", subjectContainers[0].DefaultNetworkDevice)

	// Check correct IPs are chosen
	assert.Equal(t, 1, len(orchestratorContainers[0].MultusIPAddresses))
	assert.Equal(t, "1.1.1.1", orchestratorContainers[0].MultusIPAddresses[0])
	assert.NotEqual(t, "2.2.2.2", orchestratorContainers[0].MultusIPAddresses[0])
	// test-network-function.com/multusips should be used for the test subject container.
	assert.Equal(t, 2, len(subjectContainers[0].MultusIPAddresses))
	assert.Equal(t, "3.3.3.3", subjectContainers[0].MultusIPAddresses[0])
	assert.Equal(t, "4.4.4.4", subjectContainers[0].MultusIPAddresses[1])
}
