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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

func TestBuildLabelQuery(t *testing.T) {
	testCases := []struct {
		testLabel      configsections.Label
		expectedOutput string
	}{
		{
			testLabel: configsections.Label{
				Prefix: "testprefix",
				Name:   "testname",
				Value:  "testvalue",
			},
			expectedOutput: "testprefix/testname=testvalue",
		},
		{
			testLabel: configsections.Label{
				Prefix: "testprefix",
				Name:   "testname",
				Value:  "", // empty value
			},
			expectedOutput: "testprefix/testname",
		},
		{
			testLabel: configsections.Label{
				Prefix: "", // empty value
				Name:   "testname",
				Value:  "testvalue",
			},
			expectedOutput: "testname=testvalue",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, buildLabelQuery(tc.testLabel))
	}
}

//nolint:funlen
func TestGetContainersByLabel(t *testing.T) {
	testCases := []struct {
		expectedOutput []configsections.Container
		prefix         string
		name           string
		value          string
		filename       string
	}{
		{
			prefix:   "testprefix",
			name:     "testname",
			value:    "testvalue",
			filename: "testdata/testpods_withlabel.json",
			expectedOutput: []configsections.Container{
				{
					ContainerIdentifier: configsections.ContainerIdentifier{
						Namespace:        "kube-system",
						PodName:          "coredns-78fcd69978-cc94v",
						ContainerName:    "coredns",
						NodeName:         "minikube",
						ContainerUID:     "cf794b9e8c2448815b8b5a47b354c9bf9414a04f6fa567ac3b059851ed6757ab",
						ContainerRuntime: "docker",
					},
					ImageSource: &configsections.ContainerImageSource{
						Registry:   "k8s.gcr.io",
						Repository: "coredns",
						Name:       "coredns",
						Tag:        "v1.8.4",
						Digest:     "",
					},
				},
			},
		},
		{
			prefix:         "test1",
			name:           "",
			value:          "", // no value
			filename:       "testdata/testpods_empty.json",
			expectedOutput: []configsections.Container{},
		},
	}
	origCommand := executeOcGetAllCommand
	defer func() {
		executeOcGetAllCommand = origCommand
	}()
	for _, tc := range testCases {
		executeOcGetAllCommand = func(resourceType, labelQuery string) string {
			file, _ := os.ReadFile(tc.filename)
			return string(file)
		}
		containers, _ := getContainersByLabel(configsections.Label{
			Prefix: tc.prefix,
			Name:   tc.name,
			Value:  tc.value,
		})
		assert.Equal(t, tc.expectedOutput, containers)
	}
}

func TestPerformAutoDiscovery(t *testing.T) {
	defer os.Unsetenv(disableAutodiscoverEnvVar)
	testCases := []struct {
		autoDiscoverEnabled bool
	}{
		{autoDiscoverEnabled: true},
		{autoDiscoverEnabled: false},
	}

	for _, tc := range testCases {
		if !tc.autoDiscoverEnabled {
			os.Setenv(disableAutodiscoverEnvVar, "true")
			assert.False(t, PerformAutoDiscovery())
		} else {
			os.Setenv(disableAutodiscoverEnvVar, "false")
			assert.True(t, PerformAutoDiscovery())
		}
	}
}

//nolint:funlen
func TestGetContainerIdentifiersByLabel(t *testing.T) {
	testCases := []struct {
		expectedOutput []configsections.ContainerIdentifier
		prefix         string
		name           string
		value          string
		filename       string
	}{
		{
			expectedOutput: []configsections.ContainerIdentifier{
				{
					Namespace:        "kube-system",
					PodName:          "coredns-78fcd69978-cc94v",
					ContainerName:    "coredns",
					NodeName:         "minikube",
					ContainerUID:     "cf794b9e8c2448815b8b5a47b354c9bf9414a04f6fa567ac3b059851ed6757ab",
					ContainerRuntime: "docker",
				},
			},
			prefix:   "testprefix",
			name:     "testname",
			value:    "testvalue",
			filename: "testdata/testpods_withlabel.json",
		},

		{
			prefix:         "test1",
			name:           "",
			value:          "", // no value
			filename:       "testdata/testpods_empty.json",
			expectedOutput: []configsections.ContainerIdentifier{},
		},
	}

	origCommand := executeOcGetAllCommand
	defer func() {
		executeOcGetAllCommand = origCommand
	}()

	for _, tc := range testCases {
		executeOcGetAllCommand = func(resourceType, labelQuery string) string {
			file, _ := os.ReadFile(tc.filename)
			return string(file)
		}

		identifiers, err := getContainerIdentifiersByLabel(configsections.Label{
			Prefix: tc.prefix,
			Name:   tc.name,
			Value:  tc.value,
		})

		assert.Nil(t, err)
		assert.Equal(t, tc.expectedOutput, identifiers)
	}
}

func TestBuildContainerImageSource(t *testing.T) {
	testCases := []struct {
		expectedOutput configsections.ContainerImageSource
		url            string
	}{
		{
			expectedOutput: configsections.ContainerImageSource{
				Registry:   "k8s.gcr.io",
				Repository: "coredns",
				Name:       "coredns",
				Tag:        "v1.8.0",
				Digest:     "",
			},
			url: "k8s.gcr.io/coredns/coredns:v1.8.0",
		},
		{
			expectedOutput: configsections.ContainerImageSource{
				Registry:   "quay.io",
				Repository: "rh-nfv-int",
				Name:       "testpmd-operator",
				Tag:        "",
				Digest:     "sha256:3e8fc703c71a7ccaca24b7312f8fcb3495370c46e7abc12975757b76430addf5",
			},
			url: "quay.io/rh-nfv-int/testpmd-operator@sha256:3e8fc703c71a7ccaca24b7312f8fcb3495370c46e7abc12975757b76430addf5",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, *(buildContainerImageSource(tc.url)))
	}
}
