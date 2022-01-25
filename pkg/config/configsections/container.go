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

package configsections

import (
	"fmt"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
)

const (
	defaultTimeoutSeconds = 10
)

var (
	expectersVerboseModeEnabled = false
	// DefaultTimeout for creating new interactive sessions (oc, ssh, tty)
	DefaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
)

// Container is a construct which follows the Container design pattern.  Essentially, a Container holds the
// pertinent information to perform a test against or using an Operating System Container.  This includes facets such
// as the reference to the interactive.Oc instance, the reference to the test configuration, and the default network
// IP address.
type Container struct {
	ContainerIdentifier
	Oc          *interactive.Oc
	ImageSource *ContainerImageSource
}

type ContainerImageSource struct {
	Registry string
	ContainerImageIdentifier
}

// Tag and Digest should not be populated at the same time. Digest takes precedence if both are populated
type ContainerImageIdentifier struct {
	// Name is the name of the image that you want to check if exists in the RedHat catalog
	Name string `yaml:"name" json:"name"`

	// Repository is the name of the repository `rhel8` of the container
	// This is valid for container only and required field
	Repository string `yaml:"repository" json:"repository"`

	// Tag is the optional image tag. "latest" is implied if not specified
	Tag string `yaml:"tag" json:"tag"`

	// Digest is the image digest following the "@" in a URL, e.g. image@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2
	Digest string `yaml:"digest" json:"digest"`
}

// Helper used to instantiate an OpenShift Client Session.
func GetOcSession(pod, container, namespace string, timeout time.Duration, options ...interactive.Option) *interactive.Oc {
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
			log.Debugf("start watching the session with container %s/%s", oc.GetPodName(), oc.GetPodContainerName())
			select {
			case err := <-outCh:
				log.Fatalf("OC session to container %s/%s is broken due to: %v, aborting the test run", oc.GetPodName(), oc.GetPodContainerName(), err)
			case <-oc.GetDoneChannel():
				log.Debugf("stop watching the session with container %s/%s", oc.GetPodName(), oc.GetPodContainerName())
			}
		}()
		ocChan <- oc
	}()

	containerOc = <-ocChan

	gomega.Expect(containerOc).ToNot(gomega.BeNil())

	return containerOc
}

func (c *Container) GetOc() *interactive.Oc {
	if c.Oc == nil {
		c.Oc = GetOcSession(c.PodName, c.ContainerName, c.Namespace, DefaultTimeout, interactive.Verbose(expectersVerboseModeEnabled), interactive.SendTimeout(DefaultTimeout))
	}
	return c.Oc
}

func (c *Container) CloseOc() {
	if c.Oc != nil {
		c.Oc.Close()
		c.Oc = nil
	}
}

// ContainerIdentifier is a complex key representing a unique container.
type ContainerIdentifier struct {
	Namespace        string `yaml:"namespace" json:"namespace"`
	PodName          string `yaml:"podName" json:"podName"`
	ContainerName    string `yaml:"containerName" json:"containerName"`
	NodeName         string `yaml:"nodeName" json:"nodeName"`
	ContainerUID     string `yaml:"containerUID" json:"containerUID"`
	ContainerRuntime string `yaml:"containerRuntime" json:"containerRuntime"`
}

func (cid *ContainerIdentifier) String() string {
	return fmt.Sprintf("node:%s ns:%s podName:%s containerName:%s containerUID:%s containerRuntime:%s",
		cid.NodeName,
		cid.Namespace,
		cid.PodName,
		cid.ContainerName,
		cid.ContainerUID,
		cid.ContainerRuntime,
	)
}
