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
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
)

const (
	genericLabelName = "generic"
	underTestValue   = "target"
	// partnerLabelValue = "partner"
	orchestratorValue          = "orchestrator"
	fsDiffMasterValue          = "fs_diff_master"
	skipConnectivityTestsLabel = "skip_connectivity_tests"
)

// BuildGenericConfig builds a `configsections.TestConfiguration` from the current state of the cluster,
// using labels and annotations to populate the data.
func BuildGenericConfig() (conf configsections.TestConfiguration) {
	var partnerContainers []configsections.Container // PartnerContainers is built from all non-target containers

	// an orchestrator must be identified
	orchestrator, err := getContainerByLabel(genericLabelName, orchestratorValue)
	if err != nil {
		log.Fatalf("failed to identify a single test orchestrator container: %s", err)
	}
	partnerContainers = append(partnerContainers, orchestrator)
	conf.TestOrchestrator = orchestrator.ContainerIdentifier

	// there must be containers to test
	containersUnderTest, err := getContainersByLabel(genericLabelName, underTestValue)
	if err != nil {
		log.Fatalf("found no containers to test: %s", err)
	}
	conf.ContainersUnderTest = containersUnderTest

	// the FS Diff master container is optional
	fsDiffMasterContainer, err := getContainerByLabel(genericLabelName, fsDiffMasterValue)
	if err == nil {
		partnerContainers = append(partnerContainers, fsDiffMasterContainer)
		conf.FsDiffMasterContainer = fsDiffMasterContainer.ContainerIdentifier
	} else {
		log.Warnf("an error (%s) occurred when getting the FS Diff Master Container. Attempting to continue", err)
	}

	// Containers to exclude from connectivity tests are optional
	connectivityExcludedContainers, err := getContainerIdentifiersByLabel(skipConnectivityTestsLabel, AnyLabelValue)
	if err != nil {
		log.Warnf("an error (%s) occurred when getting the containers to exclude from connectivity tests. Attempting to continue", err)
	}
	conf.ExcludeContainersFromConnectivityTests = connectivityExcludedContainers

	conf.PartnerContainers = partnerContainers
	return conf
}

// getContainersByLabel builds `config.Container`s from containers in pods matching a label.
func getContainersByLabel(labelName, labelValue string) (containers []configsections.Container, err error) {
	pods, err := GetPodsByLabel(labelName, labelValue)
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		containers = append(containers, BuildContainersFromPodResource(&pods.Items[i])...)
	}
	return containers, nil
}

// getContainerIdentifiersByLabel builds `config.ContainerIdentifier`s from containers in pods matching a label.
func getContainerIdentifiersByLabel(labelName, labelValue string) (containerIDs []configsections.ContainerIdentifier, err error) {
	containers, err := getContainersByLabel(labelName, labelValue)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		containerIDs = append(containerIDs, c.ContainerIdentifier)
	}
	return containerIDs, nil
}

// getContainerByLabel returns exactly one container with the given label. If any other number of containers is found
// then an error is returned along with an empty `config.Container`.
func getContainerByLabel(labelName, labelValue string) (container configsections.Container, err error) {
	containers, err := getContainersByLabel(labelName, labelValue)
	if err != nil {
		return container, err
	}
	if len(containers) != 1 {
		return container, fmt.Errorf("expected exactly one container, got %d for label %s=%s", len(containers), labelName, labelValue)
	}
	return containers[0], nil
}

// BuildContainersFromPodResource builds `configsections.Container`s from a `PodResource`
func BuildContainersFromPodResource(pr *PodResource) (containers []configsections.Container) {
	for _, containerResource := range pr.Spec.Containers {
		var err error
		var container configsections.Container
		container.Namespace = pr.Metadata.Namespace
		container.PodName = pr.Metadata.Name
		container.ContainerName = containerResource.Name
		container.DefaultNetworkDevice, err = pr.getDefaultNetworkDeviceFromAnnotations()
		if err != nil {
			log.Warnf("error encountered getting default network device: %s", err)
		}
		container.MultusIPAddresses, err = pr.getPodIPs()
		if err != nil {
			log.Warnf("error encountered getting multus IPs: %s", err)
			err = nil
		}

		containers = append(containers, container)
	}
	return
}
