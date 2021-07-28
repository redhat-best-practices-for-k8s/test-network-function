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

package common

import (
	"os"
	"path"
	"strconv"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/ipaddr"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

var (
	// PathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	PathRelativeToRoot = path.Join("..")

	// RelativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	RelativeSchemaPath = path.Join(PathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

// DefaultTimeout for creating new interactive sessions (oc, ssh, tty)
var DefaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

// ContainersToExcludeFromConnectivityTests is a set used for storing the containers that should be excluded from
// connectivity testing.
var ContainersToExcludeFromConnectivityTests = make(map[configsections.ContainerIdentifier]interface{})

// Helper used to instantiate an OpenShift Client Session.
func getOcSession(pod, container, namespace string, timeout time.Duration, options ...interactive.Option) *interactive.Oc {
	// Spawn an interactive OC shell using a goroutine (needed to avoid cross expect.Expecter interaction).  Extract the
	// Oc reference from the goroutine through a channel.  Performs basic sanity checking that the Oc session is set up
	// correctly.
	var containerOc *interactive.Oc
	ocChan := make(chan *interactive.Oc)
	var chOut <-chan error

	goExpectSpawner := interactive.NewGoExpectSpawner()
	var spawner interactive.Spawner = goExpectSpawner

	go func() {
		oc, outCh, err := interactive.SpawnOc(&spawner, pod, container, namespace, timeout, options...)
		gomega.Expect(outCh).ToNot(gomega.BeNil())
		gomega.Expect(err).To(gomega.BeNil())
		ocChan <- oc
	}()

	// Set up a go routine which reads from the error channel
	go func() {
		err := <-chOut
		gomega.Expect(err).To(gomega.BeNil())
	}()

	containerOc = <-ocChan

	gomega.Expect(containerOc).ToNot(gomega.BeNil())

	return containerOc
}

// Container is an internal construct which follows the Container design pattern.  Essentially, a Container holds the
// pertinent information to perform a test against or using an Operating System Container.  This includes facets such
// as the reference to the interactive.Oc instance, the reference to the test configuration, and the default network
// IP address.
type Container struct {
	ContainerConfiguration  configsections.Container
	Oc                      *interactive.Oc
	DefaultNetworkIPAddress string
	ContainerIdentifier     configsections.ContainerIdentifier
}

// createContainers contains the general steps involved in creating "oc" sessions and other configuration. A map of the
// aggregate information is returned.
func createContainers(containerDefinitions []configsections.Container) map[configsections.ContainerIdentifier]*Container {
	createdContainers := make(map[configsections.ContainerIdentifier]*Container)
	for _, c := range containerDefinitions {
		oc := getOcSession(c.PodName, c.ContainerName, c.Namespace, DefaultTimeout, interactive.Verbose(true))
		var defaultIPAddress = "UNKNOWN"
		if _, ok := ContainersToExcludeFromConnectivityTests[c.ContainerIdentifier]; !ok {
			defaultIPAddress = getContainerDefaultNetworkIPAddress(oc, c.DefaultNetworkDevice)
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

// Extract a container IP address for a particular device.  This is needed since container default network IP address
// is served by dhcp, and thus is ephemeral.
func getContainerDefaultNetworkIPAddress(oc *interactive.Oc, dev string) string {
	log.Infof("Getting IP Information for: %s(%s) in ns=%s", oc.GetPodName(), oc.GetPodContainerName(), oc.GetPodNamespace())
	ipTester := ipaddr.NewIPAddr(DefaultTimeout, dev)
	test, err := tnf.NewTest(oc.GetExpecter(), ipTester, []reel.Handler{ipTester}, oc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	RunAndValidateTest(test)
	return ipTester.GetIPv4Address()
}

// CreateContainersUnderTest sets up the test containers.
func CreateContainersUnderTest(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*Container {
	return createContainers(conf.ContainersUnderTest)
}

// CreatePartnerContainers sets up the partner containers.
func CreatePartnerContainers(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*Container {
	return createContainers(conf.PartnerContainers)
}

// GetContext spawns a new shell session and returns its context
func GetContext() *interactive.Context {
	context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), DefaultTimeout, interactive.Verbose(true))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(context).ToNot(gomega.BeNil())
	gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
	return context
}

// RunAndValidateTest runs the test and checks the result
func RunAndValidateTest(test *tnf.Test) {
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
}

// GetTestConfiguration returns the cnf-certification-generic-tests test configuration.
func GetTestConfiguration() *configsections.TestConfiguration {
	conf := config.GetConfigInstance()
	return &conf.Generic
}

// IsMinikube returns true when the env var is set, OCP only test would be skipped based on this flag
func IsMinikube() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))
	return b
}

// NonIntrusive is for skipping tests that would impact the CNF or test environment in an intrusive way
func NonIntrusive() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_NON_INTRUSIVE_ONLY"))
	return b
}

// ConfigurationData is used to host test configuration
type ConfigurationData struct {
	ContainersUnderTest map[configsections.ContainerIdentifier]*Container
	PartnerContainers   map[configsections.ContainerIdentifier]*Container
	TestOrchestrator    *Container
	FsDiffContainer     *Container
	needsRefresh        bool
}

// createContainersUnderTest sets up the test containers.
func createContainersUnderTest(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*Container {
	return createContainers(conf.ContainersUnderTest)
}

// createPartnerContainers sets up the partner containers.
func createPartnerContainers(conf *configsections.TestConfiguration) map[configsections.ContainerIdentifier]*Container {
	return createContainers(conf.PartnerContainers)
}

// Loadconfiguration the configuration into ConfigurationData
func Loadconfiguration(configData *ConfigurationData) {
	conf := GetTestConfiguration()
	log.Infof("Test Configuration: %s", conf)

	for _, cid := range conf.ExcludeContainersFromConnectivityTests {
		ContainersToExcludeFromConnectivityTests[cid] = ""
	}
	configData.ContainersUnderTest = createContainersUnderTest(conf)
	configData.PartnerContainers = createPartnerContainers(conf)
	configData.TestOrchestrator = configData.PartnerContainers[conf.TestOrchestrator]
	configData.FsDiffContainer = configData.PartnerContainers[conf.FsDiffMasterContainer]
	log.Info(configData.TestOrchestrator)
	log.Info(configData.ContainersUnderTest)
}

// ReloadConfiguration force the autodiscovery to run again
func ReloadConfiguration(configData *ConfigurationData) {
	if configData.needsRefresh {
		config.SetNeedsRefresh()
		Loadconfiguration(configData)
	}
	configData.needsRefresh = false
}

// SetNeedsRefresh indicate the config should be reloaded after this test
func (configData *ConfigurationData) SetNeedsRefresh() {
	configData.needsRefresh = true
}
