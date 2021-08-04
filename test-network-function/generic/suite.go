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
	"strconv"
	"strings"

	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/bootconfigentries"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/currentkernelcmdlineargs"
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
		configData := common.ConfigurationData{}
		configData.SetNeedsRefresh()
		ginkgo.BeforeEach(func() {
			common.ReloadConfiguration(&configData)
		})

		testIsRedHatRelease(&configData)

		if !common.IsMinikube() {
			// To be removed once Isaac's fix
			// is merged
			testBootParams(&configData)
		}
	}
})

// testIsRedHatRelease fetch the configuration and test containers attached to oc is Red Hat based.
func testIsRedHatRelease(configData *common.ConfigurationData) {
	ginkgo.It("Should report a proper Red Hat version", func() {
		for _, cut := range configData.ContainersUnderTest {
			testContainerIsRedHatRelease(cut)
		}
		testContainerIsRedHatRelease(configData.TestOrchestrator)
	})
}

// testContainerIsRedHatRelease tests whether the container attached to oc is Red Hat based.
func testContainerIsRedHatRelease(cut *common.Container) {
	podName := cut.Oc.GetPodName()
	containerName := cut.Oc.GetPodContainerName()
	context := cut.Oc
	ginkgo.By(fmt.Sprintf("%s(%s) is checked for Red Hat version", podName, containerName))
	versionTester := redhat.NewRelease(common.DefaultTimeout)
	test, err := tnf.NewTest(context.GetExpecter(), versionTester, []reel.Handler{versionTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
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

func getBootEntryIndex(bootEntry string) (int, error) {
	return strconv.Atoi(strings.Split(bootEntry, "-")[1])
}

func getMaxIndexEntry(bootConfigEntries []string) string {
	maxIndex, err := getBootEntryIndex(bootConfigEntries[0])
	gomega.Expect(err).To(gomega.BeNil())
	maxIndexEntryName := bootConfigEntries[0]
	for _, bootEntry := range bootConfigEntries {
		if entryIndex, err2 := getBootEntryIndex(bootEntry); entryIndex > maxIndex {
			maxIndex = entryIndex
			gomega.Expect(err2).To(gomega.BeNil())
			maxIndexEntryName = bootEntry
		}
	}

	return maxIndexEntryName
}

func getGrubKernelArgs(context *interactive.Context, nodeName string) map[string]string {
	bootConfigEntriesTester := bootconfigentries.NewBootConfigEntries(common.DefaultTimeout, nodeName)
	test, err := tnf.NewTest(context.GetExpecter(), bootConfigEntriesTester, []reel.Handler{bootConfigEntriesTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	common.RunAndValidateTest(test)
	bootConfigEntries := bootConfigEntriesTester.GetBootConfigEntries()

	maxIndexEntryName := getMaxIndexEntry(bootConfigEntries)

	readBootConfigTester := readbootconfig.NewReadBootConfig(common.DefaultTimeout, nodeName, maxIndexEntryName)
	test, err = tnf.NewTest(context.GetExpecter(), readBootConfigTester, []reel.Handler{readBootConfigTester}, context.GetErrorChannel())
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

// testBootParams test Boot parameters of nodes
func testBootParams(configData *common.ConfigurationData) {
	ginkgo.It("generic-boot-param", func() {
		for _, cut := range configData.ContainersUnderTest {
			context := common.GetContext()
			podName := cut.Oc.GetPodName()
			podNamespace := cut.Oc.GetPodNamespace()
			targetPodOc := cut.Oc
			defer results.RecordResult(identifiers.TestUnalteredStartupBootParamsIdentifier)
			ginkgo.By(fmt.Sprintf("Testing boot params for the pod's node %s/%s", podNamespace, podName))
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
		}
	})
}
