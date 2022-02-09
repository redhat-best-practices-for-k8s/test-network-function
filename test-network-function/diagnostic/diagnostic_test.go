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

// Package diagnostic provides a test suite which gathers OpenShift cluster information.
package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

//nolint:funlen
func TestGetMasterAndWorkerNodeName(t *testing.T) {
	testCases := []struct {
		testEnv          *config.TestEnvironment
		isMasterTest     bool
		expectedNodeName string
	}{
		{ // Test Case 1 - IsMaster & HasDebugPod
			testEnv: &config.TestEnvironment{
				NodesUnderTest: map[string]*config.NodeConfig{
					"node01": {
						Name: "node01",
						Node: configsections.Node{
							Labels: []string{configsections.MasterLabel},
						},
						DebugContainer: &configsections.Container{
							ContainerIdentifier: configsections.ContainerIdentifier{
								PodName: "debugTest",
							},
						},
					},
				},
			},
			expectedNodeName: "node01",
			isMasterTest:     true,
		},
		{ // Test Case 2 - IsWorker & HasDebugPod
			testEnv: &config.TestEnvironment{
				NodesUnderTest: map[string]*config.NodeConfig{
					"node01": {
						Name: "node01",
						Node: configsections.Node{
							Labels: []string{configsections.WorkerLabel},
						},
						DebugContainer: &configsections.Container{
							ContainerIdentifier: configsections.ContainerIdentifier{
								PodName: "debugTest",
							},
						},
					},
				},
			},
			expectedNodeName: "",
			isMasterTest:     true,
		},
		{ // Test Case 3 - IsMaster & No DebugPod
			testEnv: &config.TestEnvironment{
				NodesUnderTest: map[string]*config.NodeConfig{
					"node01": {
						Name: "node01",
						Node: configsections.Node{
							Labels: []string{configsections.MasterLabel},
						},
						DebugContainer: nil,
					},
				},
			},
			expectedNodeName: "",
			isMasterTest:     true,
		},
		{ // Test Case 4 - IsWorker & HasDebugPod
			testEnv: &config.TestEnvironment{
				NodesUnderTest: map[string]*config.NodeConfig{
					"node01": {
						Name: "node01",
						Node: configsections.Node{
							Labels: []string{configsections.WorkerLabel},
						},
						DebugContainer: &configsections.Container{
							ContainerIdentifier: configsections.ContainerIdentifier{
								PodName: "debugTest",
							},
						},
					},
				},
			},
			expectedNodeName: "node01",
			isMasterTest:     false,
		},
		{ // Test Case 5 - IsMaster & HasDebugPod
			testEnv: &config.TestEnvironment{
				NodesUnderTest: map[string]*config.NodeConfig{
					"node01": {
						Name: "node01",
						Node: configsections.Node{
							Labels: []string{configsections.MasterLabel},
						},
						DebugContainer: &configsections.Container{
							ContainerIdentifier: configsections.ContainerIdentifier{
								PodName: "debugTest",
							},
						},
					},
				},
			},
			expectedNodeName: "",
			isMasterTest:     false,
		},
		{ // Test Case 6 - IsWorker & No DebugPod
			testEnv: &config.TestEnvironment{
				NodesUnderTest: map[string]*config.NodeConfig{
					"node01": {
						Name: "node01",
						Node: configsections.Node{
							Labels: []string{configsections.WorkerLabel},
						},
						DebugContainer: nil,
					},
				},
			},
			expectedNodeName: "",
			isMasterTest:     false,
		},
	}

	for _, tc := range testCases {
		if tc.isMasterTest {
			assert.Equal(t, tc.expectedNodeName, getMasterNodeName(tc.testEnv))
		} else {
			assert.Equal(t, tc.expectedNodeName, getWorkerNodeName(tc.testEnv))
		}
	}
}

func TestNewNodeSummary(t *testing.T) {
	assert.NotNil(t, GetNodeSummary())
}

func TestGetVersionsOcp(t *testing.T) {
	assert.NotNil(t, GetVersionsOcp())
}

func TestNewCsiDriverInfo(t *testing.T) {
	assert.NotNil(t, GetCsiDriverInfo())
}
