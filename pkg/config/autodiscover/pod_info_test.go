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
	"os"
	"path"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
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

func TestPodResource_getDefaultPodIPAddresses(t *testing.T) {
	type test struct {
		name     string
		testFile string
		want     []string
	}

	// Passing tests
	var testsExpectFail = []test{
		{name: "ipv4ipv6",
			testFile: "testorchestrator.json",
			want:     []string{"2.2.2.3", "fd00:10:244:1::3"},
		},
	}

	// Failing tests
	var testsExpectPass = []test{
		{name: "ipv4ipv6",
			testFile: "ipv4ipv6pod.json",
			want:     []string{"2.2.2.2", "fd00:10:244:1::3"},
		},
		{name: "ipv4",
			testFile: "ipv4pod.json",
			want:     []string{"2.2.2.2"},
		},
		{name: "ipv6",
			testFile: "ipv6pod.json",
			want:     []string{"fd00:10:244:1::3"},
		},
	}

	// Expect fail
	for _, tt := range testsExpectFail {
		t.Run(tt.name, func(t *testing.T) {
			pr := loadPodResource(path.Join(filePath, tt.testFile))
			if got := pr.getDefaultPodIPAddresses(); reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodResource.getDefaultPodIPAddresses() = %v, want %v", got, tt.want)
			}
		})
	}
	// Expect pass
	for _, tt := range testsExpectPass {
		t.Run(tt.name, func(t *testing.T) {
			pr := loadPodResource(path.Join(filePath, tt.testFile))
			if got := pr.getDefaultPodIPAddresses(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodResource.getDefaultPodIPAddresses() = %v, want %v", got, tt.want)
			}
		})
	}
}
