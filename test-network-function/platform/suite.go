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

package platform

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/cnffsdiff"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/containerid"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/currentkernelcmdlineargs"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/hugepages"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/mckernelarguments"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodehugepages"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodemcname"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodetainted"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/podnodename"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/readbootconfig"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/sysctlallconfigsargs"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	utils "github.com/test-network-function/test-network-function/pkg/utils"
)

//
// All actual test code belongs below here.  Utilities belong above.
//

func getTaintedBitValues() []string {
	return []string{"proprietary module was loaded",
		"module was force loaded",
		"kernel running on an out of specification system",
		"module was force unloaded",
		"processor reported a Machine Check Exception (MCE)",
		"bad page referenced or some unexpected page flags",
		"taint requested by userspace application",
		"kernel died recently, i.e. there was an OOPS or BUG",
		"ACPI table overridden by user",
		"kernel issued warning",
		"staging driver was loaded",
		"workaround for bug in platform firmware applied",
		"externally-built (“out-of-tree”) module was loaded",
		"unsigned module was loaded",
		"soft lockup occurred",
		"kernel has been live patched",
		"auxiliary taint, defined for and used by distros",
		"kernel was built with the struct randomization plugin",
	}
}

var _ = ginkgo.Describe(common.PlatformAlterationTestKey, func() {
	if testcases.IsInFocus(ginkgoconfig.GinkgoConfig.FocusStrings, common.PlatformAlterationTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})
		// use this boolean to turn off tests that require OS packages
		if !common.IsMinikube() {
			testContainersFsDiff(env)
			testTainted(env)
			testHugepages(env)
			testBootParams(env)
			testSysctlConfigs(env)
		}
		testIsRedHatRelease(env)
	}
})

// testIsRedHatRelease fetch the configuration and test containers attached to oc is Red Hat based.
func testIsRedHatRelease(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestIsRedHatReleaseIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("should report a proper Red Hat version")
		defer results.RecordResult(identifiers.TestIsRedHatReleaseIdentifier)
		for _, cut := range env.ContainersUnderTest {
			testContainerIsRedHatRelease(cut)
		}
	})
}

// testContainerIsRedHatRelease tests whether the container attached to oc is Red Hat based.
func testContainerIsRedHatRelease(cut *config.Container) {
	podName := cut.Oc.GetPodName()
	containerName := cut.Oc.GetPodContainerName()
	context := cut.Oc
	ginkgo.By(fmt.Sprintf("%s(%s) is checked for Red Hat version", podName, containerName))
	versionTester := redhat.NewRelease(common.DefaultTimeout)
	test, err := tnf.NewTest(context.GetExpecter(), versionTester, []reel.Handler{versionTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
}

// testContainersFsDiff test that all CUT didn't install new packages are starting
func testContainersFsDiff(env *config.TestEnvironment) {
	ginkgo.Context("Container does not have additional packages installed", func() {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUnalteredBaseImageIdentifier)
		ginkgo.It(testID, func() {
			var badContainers []string
			for _, cut := range env.ContainersUnderTest {
				podName := cut.Oc.GetPodName()
				containerName := cut.Oc.GetPodContainerName()
				context := cut.Oc
				nodeName := cut.ContainerConfiguration.NodeName
				ginkgo.By(fmt.Sprintf("%s(%s) should not install new packages after starting", podName, containerName))
				test := newContainerFsDiffTest(nodeName, context)
				test.RunAndValidateWithFailureCallback(func() {
					badContainers = append(badContainers, containerName)
					ginkgo.By(fmt.Sprintf("pod %s container %s did update/install/modify additional packages", podName, containerName))
				})
			}
			gomega.Expect(badContainers).To(gomega.BeNil())
		})
	})
}

// newContainerFsDiffTest  test that the CUT didn't install new packages after starting, and report through Ginkgo.
func newContainerFsDiffTest(nodeName string, targetContainerOC *interactive.Oc) *tnf.Test {
	defer results.RecordResult(identifiers.TestUnalteredBaseImageIdentifier)
	targetContainerOC.GetExpecter()
	containerIDTester := containerid.NewContainerID(common.DefaultTimeout)
	test, err := tnf.NewTest(targetContainerOC.GetExpecter(), containerIDTester, []reel.Handler{containerIDTester}, targetContainerOC.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	containerID := containerIDTester.GetID()
	context := common.GetContext()
	fsDiffTester := cnffsdiff.NewFsDiff(common.DefaultTimeout, containerID, nodeName)
	test, err = tnf.NewTest(context.GetExpecter(), fsDiffTester, []reel.Handler{fsDiffTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	return test
}
func getMcKernelArguments(context *interactive.Context, mcName string) map[string]string {
	mcKernelArgumentsTester := mckernelarguments.NewMcKernelArguments(common.DefaultTimeout, mcName)
	test, err := tnf.NewTest(context.GetExpecter(), mcKernelArgumentsTester, []reel.Handler{mcKernelArgumentsTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
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
	test.RunAndValidate()
	return mcNameTester.GetMcName()
}

func getPodNodeName(context *interactive.Context, podName, podNamespace string) string {
	podNameTester := podnodename.NewPodNodeName(common.DefaultTimeout, podName, podNamespace)
	test, err := tnf.NewTest(context.GetExpecter(), podNameTester, []reel.Handler{podNameTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	return podNameTester.GetNodeName()
}

func getCurrentKernelCmdlineArgs(targetContainerOc *interactive.Oc) map[string]string {
	currentKernelCmdlineArgsTester := currentkernelcmdlineargs.NewCurrentKernelCmdlineArgs(common.DefaultTimeout)
	test, err := tnf.NewTest(targetContainerOc.GetExpecter(), currentKernelCmdlineArgsTester, []reel.Handler{currentKernelCmdlineArgsTester}, targetContainerOc.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	currnetKernelCmdlineArgs := currentKernelCmdlineArgsTester.GetKernelArguments()
	currentSplitKernelCmdlineArgs := strings.Split(currnetKernelCmdlineArgs, " ")
	return utils.ArgListToMap(currentSplitKernelCmdlineArgs)
}

func getGrubKernelArgs(context *interactive.Oc) map[string]string {
	readBootConfigTester := readbootconfig.NewReadBootConfig(common.DefaultTimeout)
	test, err := tnf.NewTest(context.GetExpecter(), readBootConfigTester, []reel.Handler{readBootConfigTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
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

// Creates a map describing the final sysctl key-value pair out of the results of "sysctl --system"
func parseSysctlSystemOutput(sysctlSystemOutput string) map[string]string {
	retval := make(map[string]string)
	splitConfig := strings.Split(sysctlSystemOutput, "\n")
	for _, line := range splitConfig {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "*") {
			continue
		}

		keyValRegexp := regexp.MustCompile(`( \S+)(\s*)=(\s*)(\S+)`) // A line is of the form "kernel.yama.ptrace_scope = 0"
		if !keyValRegexp.MatchString(line) {
			continue
		}
		regexResults := keyValRegexp.FindStringSubmatch(line)
		key := regexResults[1]
		val := regexResults[4]
		retval[key] = val
	}
	return retval
}

func getSysctlConfigArgs(context *interactive.Oc) map[string]string {
	sysctlAllConfigsArgsTester := sysctlallconfigsargs.NewSysctlAllConfigsArgs(common.DefaultTimeout)
	test, err := tnf.NewTest(context.GetExpecter(), sysctlAllConfigsArgsTester, []reel.Handler{sysctlAllConfigsArgsTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	sysctlAllConfigsArgs := sysctlAllConfigsArgsTester.GetSysctlAllConfigsArgs()

	return parseSysctlSystemOutput(sysctlAllConfigsArgs)
}

func testBootParams(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUnalteredStartupBootParamsIdentifier)
	ginkgo.It(testID, func() {
		context := common.GetContext()
		for _, cut := range env.ContainersUnderTest {
			podName := cut.Oc.GetPodName()
			podNameSpace := cut.Oc.GetPodNamespace()
			targetContainerOc := cut.Oc
			testBootParamsHelper(context, podName, podNameSpace, targetContainerOc)
		}
	})
}
func testBootParamsHelper(context *interactive.Context, podName, podNamespace string, targetContainerOc *interactive.Oc) {
	ginkgo.By(fmt.Sprintf("Testing boot params for the pod's node %s/%s", podNamespace, podName))
	defer results.RecordResult(identifiers.TestUnalteredStartupBootParamsIdentifier)
	nodeName := getPodNodeName(context, podName, podNamespace)
	mcName := getMcName(context, nodeName)
	mcKernelArgumentsMap := getMcKernelArguments(context, mcName)
	currentKernelArgsMap := getCurrentKernelCmdlineArgs(targetContainerOc)
	env := config.GetTestEnvironment()
	nodeOC := env.NodesUnderTest[nodeName].Oc
	grubKernelConfigMap := getGrubKernelArgs(nodeOC)

	for key, mcVal := range mcKernelArgumentsMap {
		if currentVal, ok := currentKernelArgsMap[key]; ok {
			gomega.Expect(currentVal).To(gomega.Equal(mcVal))
		}
		if grubVal, ok := grubKernelConfigMap[key]; ok {
			gomega.Expect(grubVal).To(gomega.Equal(mcVal))
		}
	}
}

func testSysctlConfigs(env *config.TestEnvironment) {
	ginkgo.It("platform-sysctl-config", func() {
		for _, podUnderTest := range env.PodsUnderTest {
			podName := podUnderTest.Name
			podNameSpace := podUnderTest.Namespace
			testSysctlConfigsHelper(podName, podNameSpace)
		}
	})
}
func testSysctlConfigsHelper(podName, podNamespace string) {
	ginkgo.By(fmt.Sprintf("Testing sysctl config files for the pod's node %s/%s", podNamespace, podName))
	context := common.GetContext()
	nodeName := getPodNodeName(context, podName, podNamespace)
	env := config.GetTestEnvironment()
	nodeOc := env.NodesUnderTest[nodeName].Oc
	combinedSysctlSettings := getSysctlConfigArgs(nodeOc)
	mcName := getMcName(context, nodeName)
	mcKernelArgumentsMap := getMcKernelArguments(context, mcName)
	for key, sysctlConfigVal := range combinedSysctlSettings {
		if mcVal, ok := mcKernelArgumentsMap[key]; ok {
			gomega.Expect(mcVal).To(gomega.Equal(sysctlConfigVal))
		}
	}
}

func printTainted(bitmap uint64) string {
	values := getTaintedBitValues()
	var out string
	for i := 0; i < 32; i++ {
		bit := (bitmap >> i) & 1
		if bit == 1 {
			out += fmt.Sprintf("%s, ", values[i])
		}
	}
	return out
}

func testTainted(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestNonTaintedNodeKernelsIdentifier)
	ginkgo.It(testID, func() {
		ginkgo.By("Testing tainted nodes in cluster")

		var taintedNodes []string
		for _, node := range env.NodesUnderTest {
			context := node.Oc
			tester := nodetainted.NewNodeTainted(common.DefaultTimeout)
			test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunWithFailureCallback(func() {
				taintedNodes = append(taintedNodes, node.Name)
			})
			taintedBitmap, err := strconv.ParseUint(tester.Match, 10, 32) //nolint:gomnd // base 10 and uint32
			var message string
			if err != nil {
				message = fmt.Sprintf("Could not decode tainted kernel causes (code=%d) for node %s\n", taintedBitmap, node.Name)
			} else if taintedBitmap != 0 {
				message = fmt.Sprintf("Decoded tainted kernel causes (code=%d) for node %s : %s\n", taintedBitmap, node.Name, printTainted(taintedBitmap))
			} else {
				message = fmt.Sprintf("Decoded tainted kernel causes (code=%d) for node %s : None\n", taintedBitmap, node.Name)
			}
			_, err = ginkgo.GinkgoWriter.Write([]byte(message))
			if err != nil {
				log.Errorf("Ginkgo writer could not write because: %s", err)
			}
		}
		gomega.Expect(taintedNodes).To(gomega.BeNil())
	})
}

func getNodeMcHugepages(nodeName string) (hugePagesCount, hugePagesSize int) {
	context := common.GetContext()
	mcName := getMcName(context, nodeName)
	hugepageTester := hugepages.NewHugepages(common.DefaultTimeout, mcName)
	test, err := tnf.NewTest(context.GetExpecter(), hugepageTester, []reel.Handler{hugepageTester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	return hugepageTester.GetHugepages(), hugepageTester.GetHugepagesz()
}

func testHugepages(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestHugepagesNotManuallyManipulated)
	ginkgo.It(testID, func() {
		defer results.RecordResult(identifiers.TestHugepagesNotManuallyManipulated)

		var badNodes []string
		for _, node := range env.NodesUnderTest {
			if !node.IsWorker() {
				continue
			}
			nodeOc := node.Oc
			ginkgo.By("Should return machineconfig hugepages configuration of node " + node.Name)
			nodeHugePagesCount, nodeHugePagesSize := getNodeMcHugepages(node.Name)

			ginkgo.By(fmt.Sprintf("Node's machine config hugepages=%d/hugepagesz=%d values should match the actual ones in the node.",
				nodeHugePagesCount, nodeHugePagesSize))
			tester := nodehugepages.NewNodeHugepages(common.DefaultTimeout, nodeHugePagesSize, nodeHugePagesCount)
			test, err := tnf.NewTest(nodeOc.GetExpecter(), tester, []reel.Handler{tester}, nodeOc.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			test.RunWithFailureCallback(func() {
				badNodes = append(badNodes, node.Name)
				ginkgo.By(fmt.Sprintf("Node=%s hugepage config does not match machineconfig", node.Name))
			})
		}
		gomega.Expect(badNodes).To(gomega.BeNil())
	})
}
