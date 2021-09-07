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
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

const (
	operatorLabelName          = "operator"
	skipConnectivityTestsLabel = "skip_connectivity_tests"
)

var (
	operatorTestsAnnotationName    = buildAnnotationName("operator_tests")
	subscriptionNameAnnotationName = buildAnnotationName("subscription_name")
	podTestsAnnotationName         = buildAnnotationName("host_resource_tests")
)

// FindTestTarget finds test targets from the current state of the cluster,
// using labels and annotations, and add them to the `configsections.TestTarget` passed in.
func FindTestTarget(labels []configsections.Label, target *configsections.TestTarget) {
	// find pods by label
	for _, l := range labels {
		pods, err := GetPodsByLabel(l)
		if err == nil {
			for i := range pods.Items {
				target.PodsUnderTest = append(target.PodsUnderTest, buildPodUnderTest(&pods.Items[i]))
				target.ContainerConfigList = append(target.ContainerConfigList, buildContainersFromPodResource(&pods.Items[i])...)
			}
		} else {
			log.Warnf("failed to query by label: %v %v", l, err)
		}
	}
	// Containers to exclude from connectivity tests are optional
	identifiers, err := getContainerIdentifiersByLabel(configsections.Label{Prefix: tnfLabelPrefix, Name: skipConnectivityTestsLabel, Value: anyLabelValue})
	target.ExcludeContainersFromConnectivityTests = identifiers

	if err != nil {
		log.Warnf("an error (%s) occurred when getting the containers to exclude from connectivity tests. Attempting to continue", err)
	}

	csvs, err := GetCSVsByLabel(operatorLabelName, anyLabelValue)
	if err == nil {
		for i := range csvs.Items {
			target.Operators = append(target.Operators, buildOperatorFromCSVResource(&csvs.Items[i]))
		}
	} else {
		log.Warnf("an error (%s) occurred when looking for operaters by label", err)
	}

	target.DeploymentsUnderTest = append(target.DeploymentsUnderTest, FindTestDeployments(labels, target)...)
}

// FindTestDeployments uses the containers' namespace to get its parent deployment. Filters out non CNF test deployments,
// currently partner and fs_diff ones.
func FindTestDeployments(targetLabels []configsections.Label, target *configsections.TestTarget) (deployments []configsections.Deployment) {
	namespaces := make(map[string]bool)

	for _, cut := range target.ContainerConfigList {
		namespace := cut.ContainerIdentifier.Namespace
		if _, alreadyProcessed := namespaces[namespace]; alreadyProcessed {
			continue
		}

		namespaces[namespace] = true

		for _, label := range targetLabels {
			deploymentResourceList, err := GetTargetDeploymentsByNamespace(namespace, label)

			if err != nil {
				log.Error("Unable to get deployment list from namespace ", namespace, ". Error: ", err)
			} else {
				for _, deploymentResource := range deploymentResourceList.Items {
					deployment := configsections.Deployment{
						Name:      deploymentResource.GetName(),
						Namespace: deploymentResource.GetNamespace(),
						Replicas:  deploymentResource.GetReplicas(),
					}

					deployments = append(deployments, deployment)
				}
			}
		}
	}

	return deployments
}

// buildPodUnderTest builds a single `configsections.Pod` from a PodResource
func buildPodUnderTest(pr *PodResource) (podUnderTest configsections.Pod) {
	var err error
	podUnderTest.Namespace = pr.Metadata.Namespace
	podUnderTest.Name = pr.Metadata.Name
	podUnderTest.ServiceAccount = pr.Spec.ServiceAccount
	podUnderTest.ContainerCount = len(pr.Spec.Containers)
	var tests []string
	err = pr.GetAnnotationValue(podTestsAnnotationName, &tests)
	if err != nil {
		log.Warnf("unable to extract tests from annotation on '%s/%s' (error: %s). Attempting to fallback to all tests", podUnderTest.Namespace, podUnderTest.Name, err)
		podUnderTest.Tests = testcases.GetConfiguredPodTests()
	} else {
		podUnderTest.Tests = tests
	}
	return
}

// buildOperatorFromCSVResource builds a single `configsections.Operator` from a CSVResource
func buildOperatorFromCSVResource(csv *CSVResource) (op configsections.Operator) {
	var err error
	op.Name = csv.Metadata.Name
	op.Namespace = csv.Metadata.Namespace

	var tests []string
	err = csv.GetAnnotationValue(operatorTestsAnnotationName, &tests)
	if err != nil {
		log.Warnf("unable to extract tests from annotation on '%s/%s' (error: %s). Attempting to fallback to all tests", op.Namespace, op.Name, err)
		op.Tests = getConfiguredOperatorTests()
	} else {
		op.Tests = tests
	}

	var subscriptionName []string
	err = csv.GetAnnotationValue(subscriptionNameAnnotationName, &subscriptionName)
	if err != nil {
		log.Warnf("unable to get a subscription name annotation from CSV %s (error: %s).", csv.Metadata.Name, err)
	} else {
		op.SubscriptionName = subscriptionName[0]
	}
	return op
}

// getConfiguredOperatorTests loads the `configuredTestFile` used by the `operator` specs and extracts
// the names of test groups from it.
func getConfiguredOperatorTests() (opTests []string) {
	configuredTests, err := testcases.LoadConfiguredTestFile(testcases.ConfiguredTestFile)
	if err != nil {
		log.Errorf("failed to load %s, continuing with no tests", testcases.ConfiguredTestFile)
		return []string{}
	}
	for _, configuredTest := range configuredTests.OperatorTest {
		opTests = append(opTests, configuredTest.Name)
	}
	log.WithField("opTests", opTests).Infof("got all tests from %s.", testcases.ConfiguredTestFile)
	return opTests
}
