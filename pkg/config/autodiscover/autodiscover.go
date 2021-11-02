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
	"path"
	"strconv"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
)

const (
	disableAutodiscoverEnvVar = "TNF_DISABLE_CONFIG_AUTODISCOVER"
	tnfLabelPrefix            = "test-network-function.com"
	labelTemplate             = "%s/%s"
	// anyLabelValue is the value that will allow any value for a label when building the label query.
	anyLabelValue    = ""
	ocCommand        = "oc get %s -n %s -o json -l %s"
	ocCommandTimeOut = time.Second * 10
)

var (
	// PathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")
	// TestFile is the file location of the command.json test case relative to the project root.
	TestFile = path.Join("pkg", "tnf", "handlers", "command", "command.json")
	// RelativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)
	// pathToTestFile is the relative path to the command.json test case.
	pathToTestFile = path.Join(pathRelativeToRoot, TestFile)
	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
	// commandDriver stores the csi driver JSON output.
	commandDriver = make(map[string]interface{})

	expectersVerboseModeEnabled = false
)

// PerformAutoDiscovery checks the environment variable to see if autodiscovery should be performed
func PerformAutoDiscovery() (doAuto bool) {
	doAuto, _ = strconv.ParseBool(os.Getenv(disableAutodiscoverEnvVar))
	return !doAuto
}

func buildLabelName(labelNS, labelName string) string {
	if labelNS == "" {
		return labelName
	}
	return fmt.Sprintf(labelTemplate, labelNS, labelName)
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

func executeOcGetCommand(resourceType, labelQuery, namespace string) (string, error) {
	ocCommandToExecute := fmt.Sprintf(ocCommand, resourceType, namespace, labelQuery)
	match, err := executeOcCommand(ocCommandToExecute)
	if err != nil {
		log.Error("can't run command, ", ocCommandToExecute, "Error=", err)
		return "", err
	}
	err = jsonUnmarshal([]byte(match), &commandDriver)
	gomega.Expect(err).To(gomega.BeNil())
	return match, err
}

func executeOcCommand(command string) (string, error) {
	values := make(map[string]interface{})
	values["COMMAND"] = command
	values["TIMEOUT"] = ocCommandTimeOut.Nanoseconds()
	context := interactive.GetContext(expectersVerboseModeEnabled)
	tester, handler, result, err := generic.NewGenericFromMap(pathToTestFile, relativeSchemaPath, values)

	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handler).ToNot(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	test, err := tnf.NewTest(context.GetExpecter(), *tester, handler, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())
	if err != nil {
		return "", err
	}
	test.RunAndValidate()

	genericTest := (*tester).(*generic.Generic)
	gomega.Expect(genericTest).ToNot(gomega.BeNil())

	matches := genericTest.Matches
	gomega.Expect(len(matches)).To(gomega.Equal(1))
	match := genericTest.GetMatches()[0]
	return match.Match, nil
}

// getContainersByLabel builds `config.Container`s from containers in pods matching a label.
func getContainersByLabel(label configsections.Label, namespace string) (containers []configsections.ContainerConfig, err error) {
	pods, err := GetPodsByLabel(label, namespace)
	if err != nil {
		return nil, err
	}
	for i := range pods.Items {
		containers = append(containers, buildContainersFromPodResource(pods.Items[i])...)
	}
	return containers, nil
}

// getContainerIdentifiersByLabel builds `config.ContainerIdentifier`s from containers in pods matching a label.
func getContainerIdentifiersByLabel(label configsections.Label, namespace string) (containerIDs []configsections.ContainerIdentifier, err error) {
	containers, err := getContainersByLabel(label, namespace)
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
func getContainerByLabel(label configsections.Label, namespace string) (container configsections.ContainerConfig, err error) {
	containers, err := getContainersByLabel(label, namespace)
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

// EnableExpectersVerboseMode enables the verbose mode for expecters (Sent/Match output)
func EnableExpectersVerboseMode() {
	expectersVerboseModeEnabled = true
}
