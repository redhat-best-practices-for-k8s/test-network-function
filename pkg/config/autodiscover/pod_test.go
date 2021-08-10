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

func TestBuildPodUnderTest(t *testing.T) {
	orchestratorPodResource := loadPodResource(testOrchestratorFilePath)
	orchestratorPod := buildPodUnderTest(&orchestratorPodResource)

	subjectPodResource := loadPodResource(testSubjectFilePath)
	subjectPod := buildPodUnderTest(&subjectPodResource)

	assert.Equal(t, "tnf", orchestratorPod.Namespace)
	assert.Equal(t, "I'mAPodName", orchestratorPod.Name)
	assert.NotEqual(t, "I'mAContainer", orchestratorPod.Name)
	// no tests set on pod and the config file will not be loaded from the unit test context: no tests should be set.
	assert.Equal(t, []string{}, orchestratorPod.Tests)

	assert.Equal(t, "tnf", subjectPod.Namespace)
	assert.Equal(t, "test", subjectPod.Name)
	assert.Equal(t, []string{"OneTestName", "AnotherTestName"}, subjectPod.Tests)
}
