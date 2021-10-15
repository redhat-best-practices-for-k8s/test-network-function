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
	"log"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	filePath             = "testdata"
	testOrchestratorFile = "testorchestrator.json"
	testSubjectFile      = "testtarget.json"
)

var (
	testOrchestratorFilePath = path.Join(filePath, testOrchestratorFile)
	testSubjectFilePath      = path.Join(filePath, testSubjectFile)
)

func loadPodResource(filePath string) (pod PodResource) {
	contents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error (%s) loading PodResource %s for testing", err, filePath)
	}
	err = jsonUnmarshal(contents, &pod)
	if err != nil {
		log.Fatalf("error (%s) loading PodResource %s for testing", err, filePath)
	}
	return
}

func TestPodGetAnnotationValue(t *testing.T) {
	pod := loadPodResource(testOrchestratorFilePath)
	var val string
	err := pod.GetAnnotationValue("notPresent", &val)
	assert.Equal(t, "", val)
	assert.NotNil(t, err)

	err = pod.GetAnnotationValue("test-network-function.com/defaultnetworkinterface", &val)
	assert.Equal(t, "eth0", val)
	assert.Nil(t, err)
}
