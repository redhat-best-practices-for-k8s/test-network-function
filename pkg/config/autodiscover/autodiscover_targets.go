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
	"errors"
	"fmt"
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
	operatorLabelName                = "operator"
	skipConnectivityTestsLabel       = "skip_connectivity_tests"
	skipMultusConnectivityTestsLabel = "skip_multus_connectivity_tests"
	ocGetClusterCrdNamesCommand      = "kubectl get crd -o json | jq '[.items[].metadata.name]'"
	DefaultTimeout                   = 10 * time.Second
)

var (
	operatorTestsAnnotationName    = buildAnnotationName("operator_tests")
	subscriptionNameAnnotationName = buildAnnotationName("subscription_name")
	podTestsAnnotationName         = buildAnnotationName("host_resource_tests")
)

// FindTestTarget finds test targets from the current state of the cluster,
// using labels and annotations, and add them to the `configsections.TestTarget` passed in.
//nolint:funlen
func FindTestTarget(labels []configsections.Label, target *configsections.TestTarget, namespaces []string, skipHelmChartList []configsections.SkipHelmChartList) {
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
	if err != nil {
		log.Warnf("an error (%s) occurred when getting the containers to exclude from Default connectivity tests. Attempting to continue", err)
	}
	for _, id := range identifiers {
		if ns[id.Namespace] {
			target.ExcludeContainersFromConnectivityTests = append(target.ExcludeContainersFromConnectivityTests, id)
		}
	}
	identifiers, err = getContainerIdentifiersByLabel(configsections.Label{Prefix: tnfLabelPrefix, Name: skipMultusConnectivityTestsLabel, Value: anyLabelValue})
	if err != nil {
		log.Warnf("an error (%s) occurred when getting the containers to exclude from Multus connectivity tests. Attempting to continue", err)
	}
	for _, id := range identifiers {
		if ns[id.Namespace] {
			target.ExcludeContainersFromMultusConnectivityTests = append(target.ExcludeContainersFromMultusConnectivityTests, id)
		}
	}

	csvs, err := GetCSVsByLabel(operatorLabelName, anyLabelValue)
	if err != nil {
		log.Warnf("an error (%s) occurred when looking for operators by label", err)
	}
	for _, csv := range csvs.Items {
		if ns[csv.Metadata.Namespace] {
			csv := csv
			target.Operators = append(target.Operators, buildOperatorFromCSVResource(&csv, false))
		}
	}
	dps := FindTestPodSetsByLabel(labels, string(configsections.Deployment))
	target.DeploymentsUnderTest = appendPodsets(dps, ns)
	stateFulSet := FindTestPodSetsByLabel(labels, string(configsections.StateFulSet))
	target.StateFulSetUnderTest = appendPodsets(stateFulSet, ns)
	target.Nodes = GetNodesList()
	target.HelmChart = GethelmCharts(skipHelmChartList, ns)
}
func GethelmCharts(skipHelmChartList []configsections.SkipHelmChartList, ns map[string]bool) (chartslist []configsections.HelmChart) {
	charts, err := GetClusterHelmCharts()
	if err != nil {
		log.Errorf("Failed to get helm charts... is helm installed correctly? err: %s", err)
		return nil
	}
	for _, ch := range charts.Items {
		if ns[ch.Namespace] {
			if !isSkipHelmChart(ch.Name, skipHelmChartList) {
				name, version := getHelmNameVersion(ch.Chart)
				chart := configsections.HelmChart{
					Version: version,
					Name:    name,
				}
				chartslist = append(chartslist, chart)
			}
		}
	}
	return chartslist
}

// func to check if the helm is exist on the no need to check list that are under the tnf_config.yml
func isSkipHelmChart(helmName string, skipHelmChartList []configsections.SkipHelmChartList) bool {
	if len(skipHelmChartList) == 0 {
		return false
	}
	for _, helm := range skipHelmChartList {
		if helmName == helm.Name {
			log.Infof("Helm chart with name %s was skipped", helmName)
			return true
		}
	}
	return false
}

// func to get the name and verstion need to split the number that have dots and the string valuse
// we could have a chart name like orion-ld-1.0.1 version=1.0.1 and name is orion-ld
func getHelmNameVersion(nameVersion string) (name, version string) {
	nameversion := strings.Split(nameVersion, "-")
	for k, val := range nameversion {
		if strings.Contains(val, ".") {
			version = val
			continue
		}
		if k == 0 {
			name = val
		} else {
			name = name + "-" + val
		}
	}
	return name, version
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

	podUnderTest.DefaultNetworkIPAddresses = pr.getDefaultPodIPAddresses()

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
func buildOperatorFromCSVResource(csv *CSVResource, istest bool) (op *configsections.Operator) {
	var err error
	op = &configsections.Operator{}
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
	if !istest {
		op.InstallPlans, err = getCsvInstallPlans(op.Name, op.Namespace)
		if err != nil {
			log.Errorf("Failed to get operator bundle and index image for csv %s (ns %s), error: %s", op.Name, op.Namespace, err)
		}
		op.Packag, op.Org, op.Version = csv.PackOrgVersion(op.Name)
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

var getCsvInstallPlanNames = func(csvName, csvNamespace string) ([]string, error) {
	installPlanCmd := fmt.Sprintf("oc get installplan -n %s | grep %q | awk '{ print $1 }'", csvNamespace, csvName)
	out := execCommandOutput(installPlanCmd)
	if out == "" {
		return []string{}, errors.New("installplan not found")
	}

	return strings.Split(out, "\n"), nil
}

var getInstallPlanData = func(installPlanName, namespace string) (bundleImagePath, catalogSource, catalogSourceNamespace string, err error) {
	const installPlanNumFields = 3
	const bundleImageIndex = 0
	const catalogSourceIndex = 1
	const catalogSourceNamespaceIndex = 2

	infoFromInstallPlanCmd := fmt.Sprintf("oc get installplan -n %s -o go-template="+
		"'{{range .items}}{{ if eq .metadata.name %q}}{{ range .status.bundleLookups }}"+
		"{{ .path }},{{ .catalogSourceRef.name }},{{ .catalogSourceRef.namespace }}{{end}}{{end}}{{end}}'", namespace, installPlanName)

	out := execCommandOutput(infoFromInstallPlanCmd)
	installPlanFields := strings.Split(out, ",")
	if len(installPlanFields) != installPlanNumFields {
		return "", "", "", fmt.Errorf("invalid installplan info: %s", out)
	}

	return installPlanFields[bundleImageIndex], installPlanFields[catalogSourceIndex], installPlanFields[catalogSourceNamespaceIndex], nil
}

var getCatalogSourceImageIndex = func(catalogSourceName, catalogSourceNamespace string) (string, error) {
	const nullOutput = "null"
	indexImageCmd := fmt.Sprintf("oc get catalogsource -n %s %s -o json | jq -r .spec.image", catalogSourceNamespace, catalogSourceName)
	indexImage := execCommandOutput(indexImageCmd)
	if indexImage == "" {
		return "", fmt.Errorf("failed to get index image for catalogsource %s (ns %s)", catalogSourceName, catalogSourceNamespace)
	}

	// In case there wasn't a catalogsource for this installplan, jq will return null, so leave it empty.
	if indexImage == nullOutput {
		indexImage = ""
	}

	return indexImage, nil
}

// getCsvInstallPlans provides the bundle image and index image of each installplan for a given CSV.
// These variables are saved in the `configsections.Operator` in order to be used by DCI,
// which obtains them from the claim.json and provides them to preflight suite.
func getCsvInstallPlans(csvName, csvNamespace string) (installPlans []configsections.InstallPlan, err error) {
	installPlanNames, err := getCsvInstallPlanNames(csvName, csvNamespace)
	if err != nil {
		return []configsections.InstallPlan{}, err
	}

	for _, installPlanName := range installPlanNames {
		bundleImage, catalogSourceName, catalogSourceNamespace, err := getInstallPlanData(installPlanName, csvNamespace)
		if err != nil {
			return []configsections.InstallPlan{}, err
		}

		indexImage, err := getCatalogSourceImageIndex(catalogSourceName, catalogSourceNamespace)
		if err != nil {
			return []configsections.InstallPlan{}, err
		}

		installPlans = append(installPlans, configsections.InstallPlan{Name: installPlanName, BundleImage: bundleImage, IndexImage: indexImage})
	}

	return installPlans, nil
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
