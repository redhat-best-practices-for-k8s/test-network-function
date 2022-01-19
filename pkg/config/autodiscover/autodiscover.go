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
	ocCommandTimeOut = time.Second * 15
)

var (
	expectersVerboseModeEnabled = false
)

// PerformAutoDiscovery checks the environment variable to see if autodiscovery should be performed
func PerformAutoDiscovery() bool {
	doAuto, _ := strconv.ParseBool(os.Getenv(disableAutodiscoverEnvVar))
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

// getContainersByLabel builds `configsections.Container`s from containers in pods matching a label.
// Returns slice of ContainerConfig, error.
func getContainersByLabel(label configsections.Label) ([]configsections.Container, error) {
	pods, err := GetPodsByLabel(label)
	if err != nil {
		return nil, err
	}
	containers := []configsections.Container{}
	for i := range pods.Items {
		containers = append(containers, buildContainers(pods.Items[i])...)
	}
	return containers, nil
}

// getContainerIdentifiersByLabel builds `config.ContainerIdentifier`s from containers in pods matching a label.
// Returns slice of ContainerIdentifier, error.
func getContainerIdentifiersByLabel(label configsections.Label) ([]configsections.ContainerIdentifier, error) {
	containers, err := getContainersByLabel(label)
	if err != nil {
		return nil, err
	}
	containerIDs := []configsections.ContainerIdentifier{}
	for _, c := range containers {
		containerIDs = append(containerIDs, c.ContainerIdentifier)
	}
	return containerIDs, nil
}

// buildContainers builds a container list
// Returns slice of Container
func buildContainers(pr *PodResource) []configsections.Container {
	containers := []configsections.Container{}
	for _, containerResource := range pr.Spec.Containers {
		var container configsections.Container
		container.Namespace = pr.Metadata.Namespace
		container.PodName = pr.Metadata.Name
		container.ContainerName = containerResource.Name
		container.NodeName = pr.Spec.NodeName
		container.ImageSource = buildContainerImageSource(containerResource.Image)
		// This is to have access to the pod namespace
		for _, cs := range pr.Status.ContainerStatuses {
			if cs.Name == container.ContainerName {
				container.ContainerUID = ""
				split := strings.Split(cs.ContainerID, "://")
				if len(split) > 0 {
					container.ContainerUID = split[len(split)-1]
					container.ContainerRuntime = split[0]
				}
			}
		}

		log.Debugf("added container: %s", container.String())
		containers = append(containers, container)
	}
	return containers
}

//nolint:gomnd
func buildContainerImageSource(url string) *configsections.ContainerImageSource {
	source := configsections.ContainerImageSource{}
	urlSegments := strings.Split(url, "/")
	n := len(urlSegments)
	if n > 2 {
		source.Registry = strings.Join(urlSegments[:n-2], "/")
	}
	if n > 1 {
		source.Repository = urlSegments[n-2]
	}
	colonIndex := strings.Index(urlSegments[n-1], ":")
	atIndex := strings.Index(urlSegments[n-1], "@")
	if atIndex == -1 {
		if colonIndex == -1 {
			source.Name = urlSegments[n-1]
		} else {
			source.Name = urlSegments[n-1][:colonIndex]
			source.Tag = urlSegments[n-1][colonIndex+1:]
		}
	} else {
		source.Name = urlSegments[n-1][:atIndex]
		source.Digest = urlSegments[n-1][atIndex+1:]
	}
	return &source
}

// EnableExpectersVerboseMode enables the verbose mode for expecters (Sent/Match output)
func EnableExpectersVerboseMode() {
	expectersVerboseModeEnabled = true
}
