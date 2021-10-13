// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"strconv"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/autodiscover"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"gopkg.in/yaml.v2"
)

const (
	configurationFilePathEnvironmentVariableKey = "TNF_CONFIGURATION_PATH"
	defaultConfigurationFilePath                = "tnf_config.yml"
	defaultTimeoutSeconds                       = 10
)

var (
	// testEnvironment is the singleton instance of `TestEnvironment`, accessed through `GetTestEnvironment`
	testEnvironment TestEnvironment
)

// getConfigurationFilePathFromEnvironment returns the test configuration file.
func getConfigurationFilePathFromEnvironment() string {
	environmentSourcedConfigurationFilePath := os.Getenv(configurationFilePathEnvironmentVariableKey)
	if environmentSourcedConfigurationFilePath != "" {
		return environmentSourcedConfigurationFilePath
	}
	return defaultConfigurationFilePath
}

// Container is a construct which follows the Container design pattern.  Essentially, a Container holds the
// pertinent information to perform a test against or using an Operating System Container.  This includes facets such
// as the reference to the interactive.Oc instance, the reference to the test configuration, and the default network
// IP address.
type Container struct {
	ContainerConfiguration  configsections.ContainerConfig
	Oc                      *interactive.Oc
	DefaultNetworkIPAddress string
	ContainerIdentifier     configsections.ContainerIdentifier
}

type NodeConfig struct {
	// same Name as the one inside Node structure
	Name string
	Node configsections.Node
	// Oc holds shell for debug pod running on the node
	Oc *interactive.Oc
	// deployment indicates if the node has a deployment
	deployment bool
}

func (n NodeConfig) IsMaster() bool {
	return n.Node.IsMaster()
}

func (n NodeConfig) IsWorker() bool {
	return n.Node.IsWorker()
}

func (n NodeConfig) HasDeployment() bool {
	return n.deployment
}

// DefaultTimeout for creating new interactive sessions (oc, ssh, tty)
var DefaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// Helper used to instantiate an OpenShift Client Session.
func getOcSession(pod, container, namespace string, timeout time.Duration, options ...interactive.Option) *interactive.Oc {
	// Spawn an interactive OC shell using a goroutine (needed to avoid cross expect.Expecter interaction).  Extract the
	// Oc reference from the goroutine through a channel.  Performs basic sanity checking that the Oc session is set up
	// correctly.
	var containerOc *interactive.Oc
	ocChan := make(chan *interactive.Oc)

	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner

	go func() {
		oc, outCh, err := interactive.SpawnOc(&spawner, pod, container, namespace, timeout, options...)
		gomega.Expect(outCh).ToNot(gomega.BeNil())
		gomega.Expect(err).To(gomega.BeNil())
		// Set up a go routine which reads from the error channel
		go func() {
			for {
				select {
				case err := <-outCh:
					log.Fatalf("OC session to container %s/%s is broken due to: %v, aborting the test run", oc.GetPodName(), oc.GetPodContainerName(), err)
				case <-oc.GetDoneChannel():
					log.Infof("stop watching the session with container %s/%s", oc.GetPodName(), oc.GetPodContainerName())
					return
				}
			}
		}()
		ocChan <- oc
	}()

	containerOc = <-ocChan

	gomega.Expect(containerOc).ToNot(gomega.BeNil())

	return containerOc
}

// Extract a container IP address for a particular device.  This is needed since container default network IP address
// is served by dhcp, and thus is ephemeral.
func getContainerDefaultNetworkIPAddress(oc *interactive.Oc, dev string) (string, error) {
	log.Infof("Getting IP Information for: %s(%s) in ns=%s", oc.GetPodName(), oc.GetPodContainerName(), oc.GetPodNamespace())
	ipTester := ipaddr.NewIPAddr(DefaultTimeout, dev)
	test, err := tnf.NewTest(oc.GetExpecter(), ipTester, []reel.Handler{ipTester}, oc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	result, err := test.Run()
	if result == tnf.SUCCESS && err == nil {
		return ipTester.GetIPv4Address(), nil
	}
	return "", fmt.Errorf("failed to get IP information for %s(%s) in ns=%s, result=%v, err=%v",
		oc.GetPodName(), oc.GetPodContainerName(), oc.GetPodNamespace(), result, err)
}

// TestEnvironment includes the representation of the current state of the test targets and parters as well as the test configuration
type TestEnvironment struct {
	ContainersUnderTest  map[configsections.ContainerIdentifier]*Container
	PartnerContainers    map[configsections.ContainerIdentifier]*Container
	DebugContainers      map[configsections.ContainerIdentifier]*Container
	PodsUnderTest        []configsections.Pod
	DeploymentsUnderTest []configsections.Deployment
	OperatorsUnderTest   []configsections.Operator
	NameSpaceUnderTest   string
	CrdNames             []string
	NodesUnderTest       map[string]*NodeConfig

	// ContainersToExcludeFromConnectivityTests is a set used for storing the containers that should be excluded from
	// connectivity testing.
	ContainersToExcludeFromConnectivityTests map[configsections.ContainerIdentifier]interface{}
	TestOrchestrator                         *Container
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
		env.Config.Partner = configsections.TestPartner{}
		env.Config.TestTarget = configsections.TestTarget{}
		env.TestOrchestrator = nil
		env.NodesUnderTest = nil
		env.Config.Nodes = nil
		env.DebugContainers = nil
		env.doAutodiscover()
	}
}

func (env *TestEnvironment) doAutodiscover() {
	if len(env.Config.TargetNameSpaces) != 1 {
		log.Fatal("a single namespace should be specified in config file")
	}
	env.NameSpaceUnderTest = env.Config.TargetNameSpaces[0].Name
	if autodiscover.PerformAutoDiscovery() {
		autodiscover.FindTestTarget(env.Config.TargetPodLabels, &env.Config.TestTarget, env.NameSpaceUnderTest)
	}
	autodiscover.FindTestPartner(&env.Config.Partner, env.NameSpaceUnderTest)

	env.ContainersToExcludeFromConnectivityTests = make(map[configsections.ContainerIdentifier]interface{})

	for _, cid := range env.Config.ExcludeContainersFromConnectivityTests {
		env.ContainersToExcludeFromConnectivityTests[cid] = ""
	}
	for _, cid := range env.Config.Partner.ContainersDebugList {
		env.ContainersToExcludeFromConnectivityTests[cid.ContainerIdentifier] = ""
	}

	env.ContainersUnderTest = env.createContainers(env.Config.ContainerConfigList)
	env.PodsUnderTest = env.Config.PodsUnderTest
	env.PartnerContainers = env.createContainers(env.Config.Partner.ContainerConfigList)
	env.TestOrchestrator = env.PartnerContainers[env.Config.Partner.TestOrchestratorID]
	env.DeploymentsUnderTest = env.Config.DeploymentsUnderTest
	env.OperatorsUnderTest = env.Config.Operators
	env.DebugContainers = env.createContainers(env.Config.Partner.ContainersDebugList)
	env.NodesUnderTest = env.createNodes()
	log.Infof("Test Configuration: %+v", *env)

	env.needsRefresh = false
}


func (env *TestEnvironment) createNodes() map[string]*NodeConfig {
	log.Debug("autodiscovery: create nodes  start")
	defer log.Debug("autodiscovery: create nodes done")
	nodes := make(map[string]*NodeConfig)
	minikube, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))

	for _, n := range env.Config.Nodes {
		nodes[n.Name] = &NodeConfig{Node: n, Name: n.Name, Oc: nil, deployment: false}
	}

	for _, c := range env.ContainersUnderTest {
		nodeName := c.ContainerConfiguration.NodeName
		if _, ok := nodes[nodeName]; ok {
			nodes[nodeName].deployment = true
		} else {
			log.Warn("node ", nodeName, " has deployment, but not the right labels")
		}
	}

	if minikube {
		return nodes
	}
	for _, c := range env.DebugContainers {
		nodeName := c.ContainerConfiguration.NodeName
		if _, ok := nodes[nodeName]; ok {
			nodes[nodeName].Oc = c.Oc
		}
	}

	return nodes
}


// createContainers contains the general steps involved in creating "oc" sessions and other configuration. A map of the
// aggregate information is returned.
func (env *TestEnvironment) createContainers(containerDefinitions []configsections.ContainerConfig) map[configsections.ContainerIdentifier]*Container {
	createdContainers := make(map[configsections.ContainerIdentifier]*Container)
	for _, c := range containerDefinitions {
		oc := getOcSession(c.PodName, c.ContainerName, c.Namespace, DefaultTimeout, interactive.Verbose(true), interactive.SendTimeout(DefaultTimeout))
		var defaultIPAddress = "UNKNOWN"
		var err error
		if _, ok := env.ContainersToExcludeFromConnectivityTests[c.ContainerIdentifier]; !ok {
			defaultIPAddress, err = getContainerDefaultNetworkIPAddress(oc, c.DefaultNetworkDevice)
			if err != nil {
				log.Warnf("Adding container to the ExcludeFromConnectivityTests list due to: %v", err)
				env.ContainersToExcludeFromConnectivityTests[c.ContainerIdentifier] = ""
			}
		}
		createdContainers[c.ContainerIdentifier] = &Container{
			ContainerConfiguration:  c,
			Oc:                      oc,
			DefaultNetworkIPAddress: defaultIPAddress,
			ContainerIdentifier:     c.ContainerIdentifier,
		}
	}
	return createdContainers
}

// SetNeedsRefresh marks the config stale so that the next getInstance call will redo discovery
func (env *TestEnvironment) SetNeedsRefresh() {
	env.needsRefresh = true
}

// GetTestEnvironment provides the current state of test environment
func GetTestEnvironment() *TestEnvironment {
	return &testEnvironment
}
