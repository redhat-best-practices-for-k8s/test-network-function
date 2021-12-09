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
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

const (
	disableAutodiscoverEnvVar = "TNF_DISABLE_CONFIG_AUTODISCOVER"
	tnfLabelPrefix            = "test-network-function.com"
	labelTemplate             = "%s/%s"
	// anyLabelValue is the value that will allow any value for a label when building the label query.
	anyLabelValue    = ""
	ocCommand        = "oc get %s -n %s -o json -l %s"
	ocAllCommand     = "oc get %s -A -o json -l %s"
	ocCommandTimeOut = time.Second * 10
)

var (
	expectersVerboseModeEnabled = false
)

// PerformAutoDiscovery checks the environment variable to see if autodiscovery should be performed
func PerformAutoDiscovery() (doAuto bool) {
	doAuto, _ = strconv.ParseBool(os.Getenv(disableAutodiscoverEnvVar))
	return !doAuto
}

func buildLabelName(labelPrefix, labelName string) string {
	if labelPrefix == "" {
		return labelName
	}
	return fmt.Sprintf(labelTemplate, labelPrefix, labelName)
}

func buildAnnotationName(annotationName string) string {
	return buildLabelName(tnfLabelPrefix, annotationName)
}

func buildLabelQuery(label configsections.Label) string {
	fullLabelName := buildLabelName(label.Prefix, label.Name)
	if label.Value != anyLabelValue {
		return fmt.Sprintf("%s=%s", fullLabelName, label.Value)
	}
	return fullLabelName
}

var executeOcGetCommand = func(resourceType, labelQuery, namespace string) string {
	ocCommandToExecute := fmt.Sprintf(ocCommand, resourceType, namespace, labelQuery)
	match := utils.ExecuteCommandAndValidate(ocCommandToExecute, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("can't run command: ", ocCommandToExecute)
	})
	return match
}

var executeOcGetAllCommand = func(resourceType, labelQuery string) string {
	ocCommandToExecute := fmt.Sprintf(ocAllCommand, resourceType, labelQuery)
	match := utils.ExecuteCommandAndValidate(ocCommandToExecute, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("can't run command: ", ocCommandToExecute)
	})
	return match
}

// getContainersByLabel builds `config.Container`s from containers in pods matching a label.
func getContainersByLabel(label configsections.Label) (containers []configsections.ContainerConfig, err error) {
	pods, err := GetPodsByLabel(label)
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		containers = append(containers, buildContainersFromPodResource(pods.Items[i])...)
	}
	return containers, nil
}

// getContainersByLabelByNamespace builds `config.Container`s from containers in pods matching a label.
func getContainersByLabelByNamespace(label configsections.Label, namespace string) (containers []configsections.ContainerConfig, err error) {
	pods, err := GetPodsByLabelByNamespace(label, namespace)
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		containers = append(containers, buildContainersFromPodResource(pods.Items[i])...)
	}
	return containers, nil
}

// getContainerIdentifiersByLabel builds `config.ContainerIdentifier`s from containers in pods matching a label.
func getContainerIdentifiersByLabel(label configsections.Label) (containerIDs []configsections.ContainerIdentifier, err error) {
	containers, err := getContainersByLabel(label)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		containerIDs = append(containerIDs, c.ContainerIdentifier)
	}
	return containerIDs, nil
}

// getContainerByLabelByNamespace returns exactly one container with the given label. If any other number of containers is found
// then an error is returned along with an empty `config.Container`.
func getContainerByLabelByNamespace(label configsections.Label, namespace string) (container configsections.ContainerConfig, err error) {
	containers, err := getContainersByLabelByNamespace(label, namespace)
	if err != nil {
		return container, err
	}
	if len(containers) != 1 {
		return container, fmt.Errorf("expected exactly one container, got %d for label %s/%s=%s", len(containers), label.Prefix, label.Name, label.Value)
	}
	return containers[0], nil
}

// buildContainersFromPodResource builds `configsections.Container`s from a `PodResource`
func buildContainersFromPodResource(pr *PodResource) (containers []configsections.ContainerConfig) {
	for _, containerResource := range pr.Spec.Containers {
		var err error
		var container configsections.ContainerConfig
		container.Namespace = pr.Metadata.Namespace
		container.PodName = pr.Metadata.Name
		container.ContainerName = containerResource.Name
		container.NodeName = pr.Spec.NodeName
		// This is to have access to the pod namespace
		for _, cs := range pr.Status.ContainerStatuses {
			if cs.Name == container.ContainerName {
				container.ContainerUID = ""
				split := strings.Split(cs.ContainerID, "//")
				if len(split) > 0 {
					container.ContainerUID = split[len(split)-1]
				}
			}
		}
		container.DefaultNetworkDevice, err = pr.getDefaultNetworkDeviceFromAnnotations()
		if err != nil {
			log.Warnf("error encountered getting default network device: %s", err)
		}

		container.MultusIPAddressesPerNet, err = pr.getPodIPsPerNet()
		if err != nil {
			log.Warnf("error encountered getting multus IPs: %s", err)
			err = nil
		}

		containers = append(containers, container)
	}
	return containers
}

// EnableExpectersVerboseMode enables the verbose mode for expecters (Sent/Match output)
func EnableExpectersVerboseMode() {
	expectersVerboseModeEnabled = true
}
