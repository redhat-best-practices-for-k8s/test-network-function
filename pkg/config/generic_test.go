// Copyright (C) 2020 Red Hat, Inc.
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

package config_test

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config"
)

func TestContainerIdentifier_MarshalText(t *testing.T) {
	c := &config.ContainerIdentifier{
		Namespace:     "default",
		PodName:       "test",
		ContainerName: "test",
	}
	bytes, err := c.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t, "{\"namespace\":\"default\",\"podName\":\"test\",\"containerName\":\"test\"}", string(bytes))
}

func TestContainerIdentifier_UnmarshalText(t *testing.T) {
	bytes := []byte("{\"namespace\":\"default\",\"podName\":\"test\",\"containerName\":\"test\"}")
	c := &config.ContainerIdentifier{}
	err := c.UnmarshalText(bytes)
	assert.Nil(t, err)
	assert.Equal(t, "default", c.Namespace)
	assert.Equal(t, "test", c.PodName)
	assert.Equal(t, "test", c.ContainerName)
}

type testConfigurationTestCase struct {
	configurationType     string
	expectedConfiguration *config.TestConfiguration
	expectedMarshalErr    error
}

var expectedContainersUnderTest = map[config.ContainerIdentifier]config.Container{
	{
		Namespace:     "default",
		PodName:       "test",
		ContainerName: "test",
	}: {
		DefaultNetworkDevice: "eth0",
		MultusIPAddresses:    []string{"192.168.1.1"},
	},
}

var expectedPartnerContainers = map[config.ContainerIdentifier]config.Container{
	{
		Namespace:     "default",
		PodName:       "partner",
		ContainerName: "partner",
	}: {
		DefaultNetworkDevice: "eth0",
		MultusIPAddresses:    []string{"192.168.1.3"},
	},
}

var expectedTestOrchestrator = config.ContainerIdentifier{
	Namespace:     "default",
	PodName:       "partner",
	ContainerName: "partner",
}

var goodExpectedConfiguration = &config.TestConfiguration{
	ContainersUnderTest: expectedContainersUnderTest,
	PartnerContainers:   expectedPartnerContainers,
	TestOrchestrator:    expectedTestOrchestrator,
}

var testConfigurationTestCases = map[string]*testConfigurationTestCase{
	"good-yaml": {
		configurationType:     "yaml",
		expectedConfiguration: goodExpectedConfiguration,
		expectedMarshalErr:    nil,
	},
	"good-yml": {
		configurationType:     "yml",
		expectedConfiguration: goodExpectedConfiguration,
		expectedMarshalErr:    nil,
	},
	"good-json": {
		configurationType:     "json",
		expectedConfiguration: goodExpectedConfiguration,
		expectedMarshalErr:    nil,
	},
	"empty-json": {
		configurationType: "json",
		expectedConfiguration: &config.TestConfiguration{
			ContainersUnderTest: nil,
			PartnerContainers:   nil,
			TestOrchestrator:    config.ContainerIdentifier{},
		},
		expectedMarshalErr: nil,
	},
	"missing-key-json": {
		configurationType:     "json",
		expectedConfiguration: nil,
		expectedMarshalErr:    fmt.Errorf("couldn't Unmarshal key: namespace from map[containerName:\"test\" podName:\"test\"]"),
	},
}

func formTestFileName(name, configurationType string) string {
	return path.Join("testdata", name+"."+configurationType)
}

func TestGetConfiguration(t *testing.T) {
	for testName, testConfiguration := range testConfigurationTestCases {
		testFileName := formTestFileName(testName, testConfiguration.configurationType)
		tc, err := config.GetConfiguration(testFileName)
		if testConfiguration.expectedMarshalErr == nil {
			assert.Nil(t, err)
			assert.NotNil(t, tc)
			assert.Equal(t, testConfiguration.expectedConfiguration, tc)
		} else {
			assert.Equal(t, testConfiguration.expectedMarshalErr, err)
		}
	}
}
