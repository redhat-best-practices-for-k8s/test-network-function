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
	"fmt"
	"os"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/utils"
	"gopkg.in/yaml.v2"
)

const (
	configurationFilePathEnvironmentVariableKey = "TNF_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "tnf_config.yml"
	defaultTimeoutSeconds                       = 10
)

var (
	// testEnvironment is the singleton instance of `TestEnvironment`, accessed through `GetTestEnvironment`
	testEnvironment             TestEnvironment
	expectersVerboseModeEnabled = false
)

// getConfigurationFilePathFromEnvironment returns the test configuration file.
func getConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

type NodeConfig struct {
	// same Name as the one inside Node structure
	Name string
	Node configsections.Node
	// pointer to the container of the debug pod running on the node
	DebugContainer *configsections.Container
	// podset indicates if the node has a podset,deployment/statefulset
	podset bool
	// debug indicates if the node should have a debug pod
	debug bool
}

func (n NodeConfig) IsMaster() bool {
	return n.Node.IsMaster()
}

func (n NodeConfig) IsWorker() bool {
	return n.Node.IsWorker()
}

func (n NodeConfig) HasPodset() bool {
	return n.podset
}
func (n NodeConfig) HasDebugPod() bool {
	return n.DebugContainer != nil
}

// DefaultTimeout for creating new interactive sessions (oc, ssh, tty)
var DefaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// Extract a container IP address for a particular device.  This is needed since container default network IP address
// is served by dhcp, and thus is ephemeral.
func getContainerDefaultNetworkIPAddress(initiatingPodNodeOc *interactive.Oc, nodeName, containerID, runtime, dev string) (string, error) {
	log.Infof("Getting IP Information for: %s(%s) in ns=%s", initiatingPodNodeOc.GetPodName(), initiatingPodNodeOc.GetPodContainerName(), initiatingPodNodeOc.GetPodNamespace())
	containerPID := utils.GetContainerPID(nodeName, initiatingPodNodeOc, containerID, runtime)
	ipTester := ipaddr.NewIPAddrNsenter(DefaultTimeout, containerPID, dev)
	test, err := tnf.NewTest(initiatingPodNodeOc.GetExpecter(), ipTester, []reel.Handler{ipTester}, initiatingPodNodeOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	result, err := test.Run()
	if result == tnf.SUCCESS && err == nil {
		return ipTester.GetIPv4Address(), nil
	}
	return "", err
}

// TestEnvironment includes the representation of the current state of the test targets and partners as well as the test configuration
type TestEnvironment struct {
	ContainersUnderTest  map[configsections.ContainerIdentifier]*configsections.Container
	PartnerContainers    map[configsections.ContainerIdentifier]*configsections.Container
	DebugContainers      map[configsections.ContainerIdentifier]*configsections.Container
	PodsUnderTest        []*configsections.Pod
	DeploymentsUnderTest []configsections.PodSet
	StateFulSetUnderTest []configsections.PodSet
	OperatorsUnderTest   []configsections.Operator
	NameSpacesUnderTest  []string
	CrdNames             []string
	NodesUnderTest       map[string]*NodeConfig
	Csi                  []configsections.Csi

	// ContainersToExcludeFromConnectivityTests is a set used for storing the containers that should be excluded from
	// connectivity testing.
	ContainersToExcludeFromConnectivityTests map[configsections.ContainerIdentifier]interface{}
	Config                                   configsections.TestConfiguration
	// loaded tracks if the config has been loaded to prevent it being reloaded.
	loaded bool
	// set when an intrusive test has done something that would cause Pod/Container to be recreated
	needsRefresh bool
}

// loadConfigFromFile loads a config file once.
func (env *TestEnvironment) loadConfigFromFile(filePath string) error {
	if env.loaded {
		return fmt.Errorf("cannot load config from file when a config is already loaded")
	}
	log.Info("Loading config from file: ", filePath)

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(contents, &env.Config)
	if err != nil {
		return err
	}
	env.loaded = true
	return nil
}

// LoadAndRefresh loads the config file if not loaded already and performs autodiscovery if needed
func (env *TestEnvironment) LoadAndRefresh() {
	if !env.loaded {
		filePath := getConfigurationFilePathFromEnvironment()
		log.Debugf("GetConfigInstance before config loaded, loading from file: %s", filePath)
		err := env.loadConfigFromFile(filePath)
		if err != nil {
			log.Fatalf("unable to load configuration file: %s", err)
		}
		env.doAutodiscover()
	} else if env.needsRefresh {
		env.reset()
		env.doAutodiscover()
	}
}

// Resets the environment during the drain test since all the connections are affected
func (env *TestEnvironment) reset() {
	log.Debug("clean up environment Test structure")
	env.ResetOc()
	env.Config.Partner = configsections.TestPartner{}
	env.Config.TestTarget = configsections.TestTarget{}
	// Delete Oc debug sessions before re-creating them
	for name, node := range env.NodesUnderTest {
		if node.debug {
			autodiscover.DeleteDebugLabel(name)
		}
	}
	env.NameSpacesUnderTest = nil
	env.NodesUnderTest = nil
	env.Config.Nodes = nil
	env.DebugContainers = nil
}

// Resets the environment during the intrusive tests since all the connections are affected
func (env *TestEnvironment) ResetOc() {
	log.Debug("Reset Oc sessions")
	// Delete Oc debug sessions before re-creating them
	for _, node := range env.NodesUnderTest {
		if node.HasDebugPod() {
			log.Infof("Closing session to node %s", node.Name)
			node.DebugContainer.CloseOc()
		}
	}
	// Delete all remaining sessions before re-creating them
	for _, cut := range env.ContainersUnderTest {
		cut.CloseOc()
	}

	// Delete all remaining partner sessions before re-creating them
	for _, cut := range env.PartnerContainers {
		cut.CloseOc()
	}
}

func (env *TestEnvironment) doAutodiscover() {
	log.Debug("start auto discovery")
	for _, ns := range env.Config.TargetNameSpaces {
		env.NameSpacesUnderTest = append(env.NameSpacesUnderTest, ns.Name)
	}

	if autodiscover.PerformAutoDiscovery() {
		autodiscover.FindTestTarget(env.Config.TargetPodLabels, &env.Config.TestTarget, env.NameSpacesUnderTest)
	}

	env.ContainersToExcludeFromConnectivityTests = make(map[configsections.ContainerIdentifier]interface{})

	for _, cid := range env.Config.ExcludeContainersFromConnectivityTests {
		env.ContainersToExcludeFromConnectivityTests[cid] = ""
	}

	env.ContainersUnderTest = env.createContainers(env.Config.ContainerList)
	env.PodsUnderTest = env.Config.PodsUnderTest

	// Discover nodes early on since they might be used to run commands by discovery
	// But after getting a node list in FindTestTarget() and a container under test list in env.ContainersUnderTest
	env.discoverNodes()

	env.recordPodsDefaultIP(env.PodsUnderTest)
	for _, cid := range env.Config.Partner.ContainersDebugList {
		env.ContainersToExcludeFromConnectivityTests[cid.ContainerIdentifier] = ""
	}
	env.DeploymentsUnderTest = env.Config.DeploymentsUnderTest
	env.StateFulSetUnderTest = env.Config.StateFulSetUnderTest
	env.Csi = env.Config.Csi
	env.OperatorsUnderTest = env.Config.Operators
	env.CrdNames = autodiscover.FindTestCrdNames(env.Config.CrdFilters)

	log.Infof("Test Configuration: %+v", *env)

	env.needsRefresh = false
}

// labelNodes add label to specific nodes so that node selector in debug daemonset
// can be scheduled
func (env *TestEnvironment) labelNodes() {
	var masterNode, workerNode string
	// make sure at least one worker and one master has debug set to true
	for name, node := range env.NodesUnderTest {
		if node.IsMaster() && masterNode == "" {
			masterNode = name
		}
		if node.IsMaster() && node.debug {
			masterNode = ""
			break
		}
	}
	for name, node := range env.NodesUnderTest {
		if node.IsWorker() && workerNode == "" {
			workerNode = name
		}
		if node.IsWorker() && node.debug {
			workerNode = ""
			break
		}
	}
	if masterNode != "" {
		env.NodesUnderTest[masterNode].debug = true
	}
	if workerNode != "" {
		env.NodesUnderTest[workerNode].debug = true
	}
	// label all nodes
	for nodeName, node := range env.NodesUnderTest {
		if node.debug {
			autodiscover.AddDebugLabel(nodeName)
		}
	}
}

// create Nodes data from podset
func (env *TestEnvironment) createNodes(nodes map[string]configsections.Node) map[string]*NodeConfig {
	log.Debug("autodiscovery: create nodes  start")
	defer log.Debug("autodiscovery: create nodes done")
	nodesConfig := make(map[string]*NodeConfig)
	for _, n := range nodes {
		nodesConfig[n.Name] = &NodeConfig{Node: n, Name: n.Name}
	}
	for _, c := range env.ContainersUnderTest {
		nodeName := c.NodeName
		if _, ok := nodesConfig[nodeName]; ok {
			nodesConfig[nodeName].podset = true
			nodesConfig[nodeName].debug = true
		} else {
			log.Warn("node ", nodeName, " has podset, but not the right labels")
		}
	}
	return nodesConfig
}

// attach debug pod session to node session
func (env *TestEnvironment) AttachDebugPodsToNodes() {
	for _, c := range env.DebugContainers {
		nodeName := c.NodeName
		if _, ok := env.NodesUnderTest[nodeName]; ok {
			env.NodesUnderTest[nodeName].DebugContainer = c
		}
	}
}

// discoverNodes find all the nodes in the cluster
// label the ones with deployment
// attach them to debug pods
func (env *TestEnvironment) discoverNodes() {
	env.NodesUnderTest = env.createNodes(env.Config.Nodes)

	expectedDebugPods := 0
	// Wait for the previous deployment's pod to fully terminate
	autodiscover.CheckDebugDaemonset(expectedDebugPods)
	env.labelNodes()

	for _, node := range env.NodesUnderTest {
		if node.debug {
			expectedDebugPods++
		}
	}
	autodiscover.CheckDebugDaemonset(expectedDebugPods)
	autodiscover.FindDebugPods(&env.Config.Partner)
	for _, debugPod := range env.Config.Partner.ContainersDebugList {
		env.ContainersToExcludeFromConnectivityTests[debugPod.ContainerIdentifier] = ""
	}
	env.DebugContainers = env.createContainers(env.Config.Partner.ContainersDebugList)

	env.AttachDebugPodsToNodes()
}

// createContainers contains the general steps involved in creating "oc" sessions and other configuration. A map of the
// aggregate information is returned. No IP is populated yet in this step
func (env *TestEnvironment) createContainers(containerDefinitions []configsections.Container) map[configsections.ContainerIdentifier]*configsections.Container {
	createdContainers := make(map[configsections.ContainerIdentifier]*configsections.Container)
	for _, c := range containerDefinitions {
		oc := configsections.GetOcSession(c.PodName, c.ContainerName, c.Namespace, DefaultTimeout, interactive.Verbose(expectersVerboseModeEnabled), interactive.SendTimeout(DefaultTimeout))
		createdContainers[c.ContainerIdentifier] = &configsections.Container{
			ContainerIdentifier: c.ContainerIdentifier,
			Oc:                  oc,
		}
	}
	return createdContainers
}

// recordContainersDefaultIP default IP populated in container map
func (env *TestEnvironment) recordPodsDefaultIP(pods []*configsections.Pod) {
	for _, p := range pods {
		// the first container is used to get the network namespace
		c := p.ContainerList[0]
		var defaultIPAddress = "UNKNOWN"
		var err error
		if _, ok := env.ContainersToExcludeFromConnectivityTests[c.ContainerIdentifier]; !ok {
			if env.NodesUnderTest[c.NodeName].HasDebugPod() {
				defaultIPAddress, err = getContainerDefaultNetworkIPAddress(env.NodesUnderTest[c.NodeName].DebugContainer.Oc,
					c.NodeName,
					c.ContainerUID,
					c.ContainerRuntime,
					p.DefaultNetworkDevice)
				if err != nil {
					log.Warnf("Failed to get default network ip, Adding container pod:%s container:%s ns:%s to the ExcludeFromConnectivityTests list due to: %v",
						c.PodName,
						c.ContainerName,
						c.Namespace,
						err)
					env.ContainersToExcludeFromConnectivityTests[c.ContainerIdentifier] = ""
				}
			}
		}
		p.DefaultNetworkIPAddress = defaultIPAddress
	}
}

// SetNeedsRefresh marks the config stale so that the next getInstance call will redo discovery
func (env *TestEnvironment) SetNeedsRefresh() {
	env.needsRefresh = true
}

// GetTestEnvironment provides the current state of test environment
func GetTestEnvironment() *TestEnvironment {
	return &testEnvironment
}

// EnableExpectersVerboseMode enables the verbose mode for expecters (Sent/Match output)
func EnableExpectersVerboseMode() {
	expectersVerboseModeEnabled = true

	autodiscover.EnableExpectersVerboseMode()
}
