// Copyright (C) 2020-2022 Red Hat, Inc.
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

package config

import (
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

func TestGetConfigurationFilePathFromEnvironment(t *testing.T) {
	defer os.Unsetenv(configurationFilePathEnvironmentVariableKey)
	testCases := []struct {
		envTestPath  string
		expectedPath string
	}{
		{ // Custom config
			envTestPath:  "testconfig.yml",
			expectedPath: "testconfig.yml",
		},
		{ // Default config
			envTestPath:  "",
			expectedPath: defaultConfigurationFilePath,
		},
	}

	for _, tc := range testCases {
		os.Setenv(configurationFilePathEnvironmentVariableKey, tc.envTestPath)
		assert.Equal(t, tc.expectedPath, getConfigurationFilePathFromEnvironment())
	}
}

func TestIsMaster(t *testing.T) {
	testCases := []struct {
		label          []string
		expectedMaster bool
	}{
		{
			label: []string{
				configsections.MasterLabel,
			},
			expectedMaster: true,
		},
		{
			label: []string{
				configsections.WorkerLabel,
			},
			expectedMaster: false,
		},
		{ // Check if a master is also labeled as a worker, still considered a master.
			label: []string{
				configsections.MasterLabel,
				configsections.WorkerLabel,
			},
			expectedMaster: true,
		},
	}

	for _, tc := range testCases {
		n := NodeConfig{
			Node: configsections.Node{
				Labels: tc.label,
			},
		}
		assert.Equal(t, tc.expectedMaster, n.IsMaster())
	}
}

func TestIsWorker(t *testing.T) {
	testCases := []struct {
		label          []string
		expectedWorker bool
	}{
		{
			label: []string{
				configsections.MasterLabel,
			},
			expectedWorker: false,
		},
		{
			label: []string{
				configsections.WorkerLabel,
			},
			expectedWorker: true,
		},
		{ // Check if a master labeled a worker is still considered a worker.
			label: []string{
				configsections.WorkerLabel,
				configsections.MasterLabel,
			},
			expectedWorker: true,
		},
	}

	for _, tc := range testCases {
		n := NodeConfig{
			Node: configsections.Node{
				Labels: tc.label,
			},
		}
		assert.Equal(t, tc.expectedWorker, n.IsWorker())
	}
}

func TestHasPodset(t *testing.T) {
	testCases := []struct {
		podset bool
	}{
		{
			podset: true,
		},
		{
			podset: false,
		},
	}

	for _, tc := range testCases {
		n := NodeConfig{
			podset: tc.podset,
		}
		assert.Equal(t, tc.podset, n.HasPodset())
	}
}

func TestHasDebugPod(t *testing.T) {
	testCases := []struct {
		hasDebug bool
	}{
		{
			hasDebug: true,
		},
		{
			hasDebug: false,
		},
	}

	for _, tc := range testCases {
		var n NodeConfig
		if tc.hasDebug {
			n.DebugContainer = &configsections.Container{}
			assert.True(t, n.HasDebugPod())
		} else {
			n.DebugContainer = nil
			assert.False(t, n.HasDebugPod())
		}
	}
}

func TestCreateNodes(t *testing.T) {
	nodes := map[string]configsections.Node{
		"master1": {
			Name: "master1",
			Labels: []string{
				configsections.MasterLabel,
			},
		},
	}
	expectedNodeConfig := map[string]*NodeConfig{
		"master1": {
			podset: true,
			Node: configsections.Node{
				Name: "master1",
				Labels: []string{
					configsections.MasterLabel,
				},
			},
			Name:           "master1",
			DebugContainer: nil,
		},
	}

	env := &TestEnvironment{}
	env.ContainersUnderTest = map[configsections.ContainerIdentifier]*configsections.Container{
		{
			NodeName: "mynode1",
		}: {
			ContainerIdentifier: configsections.ContainerIdentifier{
				NodeName: "mynode1",
			},
		},
	}
	createdNodes := env.createNodes(nodes)
	assert.Equal(t, expectedNodeConfig["master1"].Name, createdNodes["master1"].Name)
	assert.Equal(t, expectedNodeConfig["master1"].Node, createdNodes["master1"].Node)
	assert.Equal(t, expectedNodeConfig["master1"].Node.Labels, createdNodes["master1"].Node.Labels)
	assert.Equal(t, expectedNodeConfig["master1"].DebugContainer, createdNodes["master1"].DebugContainer)
	assert.False(t, createdNodes["master1"].HasPodset())
}

func TestSetNeedsRefresh(t *testing.T) {
	testEnv := &TestEnvironment{}
	testEnv.SetNeedsRefresh()
	assert.True(t, testEnv.needsRefresh)
}

func TestAttachDebugPodsToNodes(t *testing.T) {
	testEnv := &TestEnvironment{
		DebugContainers: map[configsections.ContainerIdentifier]*configsections.Container{
			{
				PodName:  "debug1",
				NodeName: "node1",
			}: {
				ContainerIdentifier: configsections.ContainerIdentifier{
					PodName:  "debug1",
					NodeName: "node1",
				},
			},
		},
		NodesUnderTest: map[string]*NodeConfig{
			"node1": {
				Name: "node1",
			},
		},
	}

	testEnv.AttachDebugPodsToNodes()
	assert.Equal(t, "debug1", testEnv.NodesUnderTest["node1"].DebugContainer.PodName)
}

func TestReset(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	defer ginkgo.GinkgoRecover()
	testEnv := &TestEnvironment{}
	testEnv.NodesUnderTest = map[string]*NodeConfig{
		"node1": {
			Name: "node1",
			Node: configsections.Node{
				Labels: []string{
					configsections.WorkerLabel,
				},
			},
			debug: true,
		},
	}
	origFunc := utils.ExecuteCommandAndValidate
	utils.ExecuteCommandAndValidate = func(command string, timeout time.Duration, context *interactive.Context, failureCallbackFun func()) string {
		return ""
	}
	defer func() {
		utils.ExecuteCommandAndValidate = origFunc
	}()
	testEnv.reset()
	assert.Equal(t, testEnv.Config.Partner, configsections.TestPartner{})
	assert.Equal(t, testEnv.Config.TestTarget, configsections.TestTarget{})
	assert.Nil(t, testEnv.NodesUnderTest["node1"].Node.Labels)
	assert.Nil(t, testEnv.NameSpacesUnderTest)
	assert.Nil(t, testEnv.NodesUnderTest)
	assert.Nil(t, testEnv.Config.Nodes)
	assert.Nil(t, testEnv.DebugContainers)
}

func TestLabelNodes(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	defer ginkgo.GinkgoRecover()
	testEnv := &TestEnvironment{}
	testEnv.NodesUnderTest = map[string]*NodeConfig{
		"node1": {
			Name: "node1",
			Node: configsections.Node{
				Labels: []string{
					configsections.WorkerLabel,
				},
			},
			// debug: true,
		},
		"node2": {
			Name: "node2",
			Node: configsections.Node{
				Labels: []string{
					configsections.MasterLabel,
				},
			},
			// debug: true,
		},
	}

	origFunc := utils.ExecuteCommandAndValidate
	utils.ExecuteCommandAndValidate = func(command string, timeout time.Duration, context *interactive.Context, failureCallbackFun func()) string {
		return ""
	}
	defer func() {
		utils.ExecuteCommandAndValidate = origFunc
	}()

	testEnv.labelNodes()
	assert.True(t, testEnv.NodesUnderTest["node1"].debug)
	assert.True(t, testEnv.NodesUnderTest["node2"].debug)
}
