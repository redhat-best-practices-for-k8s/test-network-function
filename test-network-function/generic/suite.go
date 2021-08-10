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

package generic

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/cnffsdiff"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/mckernelarguments"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodemcname"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/podnodename"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/readbootconfig"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	utils "github.com/test-network-function/test-network-function/pkg/utils"
)

const (
	testsKey = "generic"
)

//
// All actual test code belongs below here.  Utilities belong above.
//

// Runs the "generic" CNF test cases.
var _ = ginkgo.Describe(testsKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, testsKey) {

		config := common.GetTestConfiguration()
		log.Infof("Test Configuration: %s", config)

		for _, cid := range config.ExcludeContainersFromConnectivityTests {
			common.ContainersToExcludeFromConnectivityTests[cid] = ""
		}
		containersUnderTest := common.CreateContainersUnderTest(config)
		partnerContainers := common.CreatePartnerContainers(config)
		testOrchestrator := partnerContainers[config.TestOrchestrator]
		log.Info(testOrchestrator)
		log.Info(containersUnderTest)

		for _, containerUnderTest := range containersUnderTest {
			testIsRedHatRelease(containerUnderTest.Oc)
		}
		testIsRedHatRelease(testOrchestrator.Oc)

		if !common.IsMinikube() {
			for _, containersUnderTest := range containersUnderTest {
				// To be removed once Isaac's fix is merged
				testBootParams(common.GetContext(), containersUnderTest.Oc.GetPodName(), containersUnderTest.Oc.GetPodNamespace(), containersUnderTest.Oc)
			}
		}
	}
})

// testIsRedHatRelease tests whether the container attached to oc is Red Hat based.
func testIsRedHatRelease(oc *interactive.Oc) {
	pod := oc.GetPodName()
	container := oc.GetPodContainerName()
	ginkgo.When(fmt.Sprintf("%s(%s) is checked for Red Hat version", pod, container), func() {
		ginkgo.It("Should report a proper Red Hat version", func() {
			versionTester := redhat.NewRelease(common.DefaultTimeout)
			test, err := tnf.NewTest(oc.GetExpecter(), versionTester, []reel.Handler{versionTester}, oc.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
}

func getMcKernelArguments(context *interactive.Context, mcName string) map[string]string {
	mcKernelArgumentsTester := mckernelarguments.NewMcKernelArguments(common.DefaultTimeout, mcName)
	test, err := tnf.NewTest(context.GetExpecter(), mcKernelArgumentsTester, []reel.Handler{mcKernelArgumentsTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	mcKernelArguments := mcKernelArgumentsTester.GetKernelArguments()
	var mcKernelArgumentsJSON []string
	err = json.Unmarshal([]byte(mcKernelArguments), &mcKernelArgumentsJSON)
	gomega.Expect(err).To(gomega.BeNil())
	mcKernelArgumentsMap := utils.ArgListToMap(mcKernelArgumentsJSON)
	return mcKernelArgumentsMap
}

func getMcName(context *interactive.Context, nodeName string) string {
	mcNameTester := nodemcname.NewNodeMcName(common.DefaultTimeout, nodeName)
	test, err := tnf.NewTest(context.GetExpecter(), mcNameTester, []reel.Handler{mcNameTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	return mcNameTester.GetMcName()
}

func getPodNodeName(context *interactive.Context, podName, podNamespace string) string {
	podNameTester := podnodename.NewPodNodeName(common.DefaultTimeout, podName, podNamespace)
	test, err := tnf.NewTest(context.GetExpecter(), podNameTester, []reel.Handler{podNameTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	return podNameTester.GetNodeName()
}

func getCurrentKernelCmdlineArgs(targetPodOc *interactive.Oc) map[string]string {
	currentKernelCmdlineArgsTester := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(common.DefaultTimeout)
	test, err := tnf.NewTest(targetPodOc.GetExpecter(), currentKernelCmdlineArgsTester, []reel.Handler{currentKernelCmdlineArgsTester}, targetPodOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	currnetKernelCmdlineArgs := currentKernelCmdlineArgsTester.GetKernelArguments()
	currentSplitKernelCmdlineArgs := strings.Split(currnetKernelCmdlineArgs, " ")
	return utils.ArgListToMap(currentSplitKernelCmdlineArgs)
}

func getGrubKernelArgs(context *interactive.Context, nodeName string) map[string]string {
	readBootConfigTester := readbootconfig.NewReadBootConfig(common.DefaultTimeout, nodeName)
	test, err := tnf.NewTest(context.GetExpecter(), readBootConfigTester, []reel.Handler{readBootConfigTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	bootConfig := readBootConfigTester.GetBootConfig()

	splitBootConfig := strings.Split(bootConfig, "\n")
	filteredBootConfig := utils.FilterArray(splitBootConfig, func(line string) bool {
		return strings.HasPrefix(line, "options")
	})
	gomega.Expect(len(filteredBootConfig)).To(gomega.Equal(1))
	grubKernelConfig := filteredBootConfig[0]
	grubSplitKernelConfig := strings.Split(grubKernelConfig, " ")
	grubSplitKernelConfig = grubSplitKernelConfig[1:]
	return utils.ArgListToMap(grubSplitKernelConfig)
}

func testBootParams(context *interactive.Context, podName, podNamespace string, targetPodOc *interactive.Oc) {
	ginkgo.It(fmt.Sprintf("Testing boot params for the pod's node %s/%s", podNamespace, podName), func() {
		defer results.RecordResult(identifiers.TestUnalteredStartupBootParamsIdentifier)
		nodeName := getPodNodeName(context, podName, podNamespace)
		mcName := getMcName(context, nodeName)
		mcKernelArgumentsMap := getMcKernelArguments(context, mcName)
		currentKernelArgsMap := getCurrentKernelCmdlineArgs(targetPodOc)
		grubKernelConfigMap := getGrubKernelArgs(context, nodeName)

		for key, mcVal := range mcKernelArgumentsMap {
			if currentVal, ok := currentKernelArgsMap[key]; ok {
				gomega.Expect(currentVal).To(gomega.Equal(mcVal))
			}
			if grubVal, ok := grubKernelConfigMap[key]; ok {
				gomega.Expect(grubVal).To(gomega.Equal(mcVal))
			}
		}
	})
}
