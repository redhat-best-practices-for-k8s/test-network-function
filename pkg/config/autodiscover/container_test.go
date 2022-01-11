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

func TestBuildContainers(t *testing.T) {
	subjectPod := loadPodResource(testSubjectFilePath)
	subjectContainers := buildContainers(&subjectPod)
	assert.Equal(t, 1, len(subjectContainers))

	assert.Equal(t, "tnf", subjectContainers[0].Namespace)
	assert.Equal(t, "I'mAPodName", subjectContainers[0].PodName)
	assert.Equal(t, "I'mAContainer", subjectContainers[0].ContainerName)
}
