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

func TestGetContainersByLabel(t *testing.T) {
	testCases := []struct {
		expectedOutput []configsections.ContainerConfig
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
			expectedOutput: []configsections.ContainerConfig{
				{
					ContainerIdentifier: configsections.ContainerIdentifier{
						Namespace:     "kube-system",
						PodName:       "coredns-78fcd69978-cc94v",
						ContainerName: "coredns",
						NodeName:      "minikube",
					},
					MultusIPAddressesPerNet: map[string][]string{},
				},
			},
		},
		{
			prefix:         "test1",
			name:           "",
			value:          "", // no value
			filename:       "testdata/testpods_empty.json",
			expectedOutput: []configsections.ContainerConfig(nil),
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
					Namespace:     "kube-system",
					PodName:       "coredns-78fcd69978-cc94v",
					ContainerName: "coredns",
					NodeName:      "minikube",
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
			expectedOutput: []configsections.ContainerIdentifier(nil),
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
