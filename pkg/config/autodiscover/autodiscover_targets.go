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
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config/configsections"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodenames"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

const (
	operatorLabelName           = "operator"
	skipConnectivityTestsLabel  = "skip_connectivity_tests"
	ocGetClusterCrdNamesCommand = "kubectl get crd -o json | jq '[.items[].metadata.name]'"
	DefaultTimeout              = 10 * time.Second
)

var (
	operatorTestsAnnotationName    = buildAnnotationName("operator_tests")
	subscriptionNameAnnotationName = buildAnnotationName("subscription_name")
	podTestsAnnotationName         = buildAnnotationName("host_resource_tests")
)

// FindTestTarget finds test targets from the current state of the cluster,
// using labels and annotations, and add them to the `configsections.TestTarget` passed in.
//nolint:funlen
func FindTestTarget(labels []configsections.Label, target *configsections.TestTarget, namespaces []string) {
	ns := make(map[string]bool)
	for _, n := range namespaces {
		ns[n] = true
	}
	for _, l := range labels {
		pods, err := GetPodsByLabel(l)
		if err == nil {
			for _, pod := range pods.Items {
				if ns[pod.Metadata.Namespace] {
					target.PodsUnderTest = append(target.PodsUnderTest, buildPodUnderTest(pod))
					target.ContainerList = append(target.ContainerList, buildContainers(pod)...)
				} else {
					target.NonValidPods = append(target.NonValidPods, buildPodUnderTest(pod))
				}
			}
		} else {
			log.Warnf("failed to query by label: %v %v", l, err)
		}
	}
	// Containers to exclude from connectivity tests are optional
	identifiers, err := getContainerIdentifiersByLabel(configsections.Label{Prefix: tnfLabelPrefix, Name: skipConnectivityTestsLabel, Value: anyLabelValue})
	for _, id := range identifiers {
		if ns[id.Namespace] {
			target.ExcludeContainersFromConnectivityTests = append(target.ExcludeContainersFromConnectivityTests, id)
		}
	}
	if err != nil {
		log.Warnf("an error (%s) occurred when getting the containers to exclude from connectivity tests. Attempting to continue", err)
	}
	csvs, err := GetCSVsByLabel(operatorLabelName, anyLabelValue)
	if err != nil {
		log.Warnf("an error (%s) occurred when looking for operators by label", err)
	}
	for _, csv := range csvs.Items {
		if ns[csv.Metadata.Namespace] {
			csv := csv
			target.Operators = append(target.Operators, buildOperatorFromCSVResource(&csv))
		}
	}
	dps := FindTestPodSetsByLabel(labels, string(configsections.Deployment))
	target.DeploymentsUnderTest = appendPodsets(dps, ns)
	stateFulSet := FindTestPodSetsByLabel(labels, string(configsections.StateFulSet))
	target.StateFulSetUnderTest = appendPodsets(stateFulSet, ns)
	target.Nodes = GetNodesList()
	target.Csi = getCsi()
}
func getCsi() (csiset []configsections.Csi) {
	csilist, err := GetTargetCsi()
	if err != nil {
		log.Error("Unable to get csi list  Error: ", err)
		return nil
	}
	for _, csi := range csilist {
		if csi != "" {
			pack, org := GetPackageandOrg(csi)
			csiconf := configsections.Csi{
				Name:         csi,
				Organization: org,
				Packag:       pack,
			}
			csiset = append(csiset, csiconf)
		}
	}
	return csiset
}

// func for appending the pod sets
func appendPodsets(podsets []configsections.PodSet, ns map[string]bool) (podSet []configsections.PodSet) {
	for _, ps := range podsets {
		if ns[ps.Namespace] {
			podSet = append(podSet, ps)
		}
	}
	return podSet
}

// GetNodesList Function that return a list of node and what is the type of them.
func GetNodesList() (nodes map[string]configsections.Node) {
	nodes = make(map[string]configsections.Node)
	var nodeNames []string
	context := interactive.GetContext(expectersVerboseModeEnabled)
	tester := nodenames.NewNodeNames(DefaultTimeout, map[string]*string{configsections.MasterLabel: nil})
	test, _ := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	_, err := test.Run()
	if err != nil {
		log.Error("Unable to get node list ", ". Error: ", err)
		return
	}
	nodeNames = tester.GetNodeNames()
	for i := range nodeNames {
		nodes[nodeNames[i]] = configsections.Node{
			Name:   nodeNames[i],
			Labels: []string{configsections.MasterLabel},
		}
	}

	tester = nodenames.NewNodeNames(DefaultTimeout, map[string]*string{configsections.WorkerLabel: nil})
	test, _ = tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	_, err = test.Run()
	if err != nil {
		log.Error("Unable to get node list ", ". Error: ", err)
	} else {
		nodeNames = tester.GetNodeNames()
		for i := range nodeNames {
			if _, ok := nodes[nodeNames[i]]; ok {
				var node = nodes[nodeNames[i]]
				node.Labels = append(node.Labels, configsections.WorkerLabel)
				nodes[nodeNames[i]] = node
			} else {
				nodes[nodeNames[i]] = configsections.Node{
					Name:   nodeNames[i],
					Labels: []string{configsections.WorkerLabel},
				}
			}
		}
	}

	return nodes
}

// FindTestPodSetsByLabel uses the containers' namespace to get its parent deployment/statefulset. Filters out non CNF test podsets,deployment/statefulset,
// currently partner and fs_diff ones.
func FindTestPodSetsByLabel(targetLabels []configsections.Label, resourceTypeDeployment string) (podsets []configsections.PodSet) {
	configType := configsections.Deployment
	if resourceTypeDeployment == string(configsections.StateFulSet) {
		configType = configsections.StateFulSet
	}
	for _, label := range targetLabels {
		podsetResourceList, err := GetTargetPodSetsByLabel(label, resourceTypeDeployment)
		if err != nil {
			log.Error("Unable to get deployment list  Error: ", err)
		} else {
			for _, podsetResource := range podsetResourceList.Items {
				podset := configsections.PodSet{
					Name:      podsetResource.GetName(),
					Namespace: podsetResource.GetNamespace(),
					Replicas:  podsetResource.GetReplicas(),
					Hpa:       podsetResource.GetHpa(),
					Type:      configType,
				}

				podsets = append(podsets, podset)
			}
		}
	}
	return podsets
}

// buildPodUnderTest builds a single `configsections.Pod` from a PodResource
func buildPodUnderTest(pr *PodResource) (podUnderTest *configsections.Pod) {
	var err error
	podUnderTest = &configsections.Pod{}
	podUnderTest.Namespace = pr.Metadata.Namespace
	podUnderTest.Name = pr.Metadata.Name
	podUnderTest.ServiceAccount = pr.Spec.ServiceAccount
	podUnderTest.ContainerCount = len(pr.Spec.Containers)
	podUnderTest.DefaultNetworkDevice, err = pr.getDefaultNetworkDeviceFromAnnotations()
	if err != nil {
		log.Warnf("error encountered getting default network device: %s", err)
	}
	podUnderTest.MultusIPAddressesPerNet, err = pr.getPodIPsPerNet()
	if err != nil {
		log.Warnf("error encountered getting multus IPs: %s", err)
	}
	var tests []string
	err = pr.GetAnnotationValue(podTestsAnnotationName, &tests)
	if err != nil {
		log.Warnf("unable to extract tests from annotation on '%s/%s' (error: %s). Attempting to fallback to all tests", podUnderTest.Namespace, podUnderTest.Name, err)
		podUnderTest.Tests = testcases.GetConfiguredPodTests()
	} else {
		podUnderTest.Tests = tests
	}

	if pr.Metadata.OwnerReferences != nil {
		podUnderTest.IsManaged = true
	}

	// Get a list of all the containers present in the pod
	allContainersInPod := buildContainers(pr)
	if len(allContainersInPod) > 0 {
		// Pick the first container in the list to use as the network context
		podUnderTest.ContainerList = allContainersInPod
	} else {
		log.Errorf("There are no containers in pod %s in namespace %s", podUnderTest.Name, podUnderTest.Namespace)
	}
	return podUnderTest
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
// the names of test groups from it.  Returns slice of strings.
func getConfiguredOperatorTests() []string {
	var opTests []string
	configuredTests, err := testcases.LoadConfiguredTestFile(testcases.ConfiguredTestFile)
	if err != nil {
		log.Errorf("failed to load %s, continuing with no tests", testcases.ConfiguredTestFile)
		return opTests
	}
	for _, configuredTest := range configuredTests.OperatorTest {
		opTests = append(opTests, configuredTest.Name)
	}
	log.WithField("opTests", opTests).Infof("got all tests from %s.", testcases.ConfiguredTestFile)
	return opTests
}

// getClusterCrdNames returns a list of crd names found in the cluster.
func getClusterCrdNames() ([]string, error) {
	out := utils.ExecuteCommandAndValidate(ocGetClusterCrdNamesCommand, ocCommandTimeOut, interactive.GetContext(expectersVerboseModeEnabled), func() {
		log.Error("can't run command: ", ocGetClusterCrdNamesCommand)
	})

	var crdNamesList []string
	err := jsonUnmarshal([]byte(out), &crdNamesList)
	if err != nil {
		return nil, err
	}

	return crdNamesList, nil
}

// FindTestCrdNames gets a list of CRD names based on configured groups.
func FindTestCrdNames(crdFilters []configsections.CrdFilter) []string {
	clusterCrdNames, err := getClusterCrdNames()
	if err != nil {
		log.Errorf("Unable to get cluster CRD.")
		return []string{}
	}

	var targetCrdNames []string
	for _, crdName := range clusterCrdNames {
		for _, crdFilter := range crdFilters {
			if strings.HasSuffix(crdName, crdFilter.NameSuffix) {
				targetCrdNames = append(targetCrdNames, crdName)
				break
			}
		}
	}
	return targetCrdNames
}
