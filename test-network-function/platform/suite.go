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
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"

	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/base/redhat"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/cnffsdiff"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/containerid"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/currentkernelcmdlineargs"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/mckernelarguments"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodemcname"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodetainted"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/podnodename"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/readbootconfig"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/sysctlallconfigsargs"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	utils "github.com/test-network-function/test-network-function/pkg/utils"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	RhelDefaultHugepagesz = 2048 // kB
	RhelDefaultHugepages  = 0
	HugepagesParam        = "hugepages"
	HugepageszParam       = "hugepagesz"
	DefaultHugepagesz     = "default_hugepagesz"
)

var (
	commandHandlerFilePath = path.Join(common.PathRelativeToRoot, "pkg", "tnf", "handlers", "command", "command.json")
	mcGetterCommandTimeout = time.Second * 30
)

type hugePagesConfig struct {
	hugepagesSize  int // size in kb
	hugepagesCount int
}

// numaHugePagesPerSize maps a numa id to an array of hugePagesConfig structs.
type numaHugePagesPerSize map[int][]hugePagesConfig

// String is the stringer implementation for the numaHugePagesPerSize type so debug/info
// lines look better.
func (numaHugepages numaHugePagesPerSize) String() string {
	// Order numa ids/indexes
	numaIndexes := make([]int, 0)
	for numaIdx := range numaHugepages {
		numaIndexes = append(numaIndexes, numaIdx)
	}
	sort.Ints(numaIndexes)

	str := ""
	for _, numaIdx := range numaIndexes {
		hugepagesPerSize := numaHugepages[numaIdx]
		str += fmt.Sprintf("Numa=%d ", numaIdx)
		for _, hugepages := range hugepagesPerSize {
			str += fmt.Sprintf("[Size=%dkB Count=%d] ", hugepages.hugepagesSize, hugepages.hugepagesCount)
		}
	}
	return str
}

// machineConfig maps a json machineconfig object to get the KernelArguments and systemd units info.
type machineConfig struct {
	Spec struct {
		KernelArguments []string `json:"kernelArguments"`
		Config          struct {
			Systemd struct {
				Units []systemdHugePagesUnit `json:"units"`
			}
		} `json:"config"`
	} `json:"spec"`
}

// systemdHugePagesUnit maps a systemd unit in a machineconfig json object.
type systemdHugePagesUnit struct {
	Contents string `json:"contents"`
	Name     string `json:"name"`
}

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
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.PlatformAlterationTestKey) {
		env := config.GetTestEnvironment()
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})
		ginkgo.ReportAfterEach(results.RecordResult)
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
				containerOC := cut.Oc
				nodeName := cut.ContainerConfiguration.NodeName
				ginkgo.By(fmt.Sprintf("%s(%s) should not install new packages after starting", podName, containerName))
				nodeOc := env.NodesUnderTest[nodeName].Oc
				test := newContainerFsDiffTest(nodeName, nodeOc, containerOC)
				test.RunWithFailureCallback(func() {
					badContainers = append(badContainers, containerName)
					ginkgo.By(fmt.Sprintf("pod %s container %s did update/install/modify additional packages", podName, containerName))
				})
			}
			gomega.Expect(badContainers).To(gomega.BeNil())
		})
	})
}

// newContainerFsDiffTest  test that the CUT didn't install new packages after starting, and report through Ginkgo.
func newContainerFsDiffTest(nodeName string, nodeOc, targetContainerOC *interactive.Oc) *tnf.Test {
	targetContainerOC.GetExpecter()
	containerIDTester := containerid.NewContainerID(common.DefaultTimeout)
	test, err := tnf.NewTest(targetContainerOC.GetExpecter(), containerIDTester, []reel.Handler{containerIDTester}, targetContainerOC.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	containerID := containerIDTester.GetID()
	fsDiffTester := cnffsdiff.NewFsDiff(common.DefaultTimeout, containerID, nodeName)
	test, err = tnf.NewTest(nodeOc.GetExpecter(), fsDiffTester, []reel.Handler{fsDiffTester}, nodeOc.GetErrorChannel())
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
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestSysctlConfigsIdentifier)
	ginkgo.It(testID, func() {
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
			if !node.HasDebugPod() {
				continue
			}
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

func runAndValidateCommand(command string, context *interactive.Context, failureCallbackFun func()) (match string) {
	log.Debugf("Launching generic command handler for cmd: %s", command)

	values := make(map[string]interface{})
	values["COMMAND"] = command
	values["TIMEOUT"] = mcGetterCommandTimeout.Nanoseconds()

	tester, handlers, result, err := generic.NewGenericFromMap(commandHandlerFilePath, common.RelativeSchemaPath, values)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(test).ToNot(gomega.BeNil())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidateWithFailureCallback(failureCallbackFun)

	genericTest := (*tester).(*generic.Generic)
	gomega.Expect(genericTest).ToNot(gomega.BeNil())

	matches := genericTest.Matches
	gomega.Expect(len(matches)).To(gomega.Equal(1))
	return genericTest.GetMatches()[0].Match
}

func hugepageSizeToInt(s string) int {
	num, _ := strconv.Atoi(s[:len(s)-1])
	unit := s[len(s)-1]
	switch unit {
	case 'M':
		num *= 1024
	case 'G':
		num *= 1024 * 1024
	}

	return num
}

func logMcKernelArgumentsHugepages(hugepagesPerSize map[int]int, defhugepagesz int) {
	logStr := fmt.Sprintf("MC KernelArguments hugepages config: default_hugepagesz=%d-kB", defhugepagesz)
	for size, count := range hugepagesPerSize {
		logStr += fmt.Sprintf(", size=%dkB - count=%d", size, count)
	}
	log.Info(logStr)
}

// getMcHugepagesFromMcKernelArguments gets the hugepages params from machineconfig's kernelArguments
func getMcHugepagesFromMcKernelArguments(mc *machineConfig) (hugepagesPerSize map[int]int, defhugepagesz int, err error) {
	const KeyValueSplitLen = 2

	defhugepagesz = RhelDefaultHugepagesz
	hugepagesPerSize = map[int]int{}

	hugepagesz := 0
	for _, arg := range mc.Spec.KernelArguments {
		keyValueSlice := strings.Split(arg, "=")
		if len(keyValueSlice) != KeyValueSplitLen {
			// Some kernel arguments don't come in name=value
			continue
		}

		key, value := keyValueSlice[0], keyValueSlice[1]
		if key == HugepagesParam {
			if _, sizeFound := hugepagesPerSize[hugepagesz]; !sizeFound {
				return map[int]int{}, RhelDefaultHugepagesz, fmt.Errorf("found hugepages count without size in kernelArguments: %v", mc.Spec.KernelArguments)
			}
			hugepages, _ := strconv.Atoi(value)
			hugepagesPerSize[hugepagesz] = hugepages
		}

		if key == HugepageszParam {
			hugepagesz = hugepageSizeToInt(value)
			// Create new map entry for this size
			hugepagesPerSize[hugepagesz] = 0
		}

		if key == DefaultHugepagesz {
			defhugepagesz = hugepageSizeToInt(value)
			// In case only default_hugepagesz and hugepages values are provided. The actual value should be
			// parsed next and this default value overwritten.
			hugepagesPerSize[hugepagesz] = RhelDefaultHugepages
		}
	}

	if len(hugepagesPerSize) == 0 {
		hugepagesPerSize[RhelDefaultHugepagesz] = RhelDefaultHugepages
		log.Warnf("No hugepages size found in node's machineconfig. Defaulting to size=%dkB (count=%d)", RhelDefaultHugepagesz, RhelDefaultHugepages)
	}

	logMcKernelArgumentsHugepages(hugepagesPerSize, defhugepagesz)
	return hugepagesPerSize, defhugepagesz, nil
}

// getNodeNumaHugePages gets the actual node's hugepages config based on /sys/devices/system/node/nodeX files.
func getNodeNumaHugePages(node *config.NodeConfig) (hugepages numaHugePagesPerSize, err error) {
	const cmd = "for file in `find /sys/devices/system/node/ -name nr_hugepages`; do echo $file count:`cat $file` ; done"
	const outputRegex = `node(\d+).*hugepages-(\d+)kB.* count:(\d+)`
	const numRegexFields = 4

	// This command must run inside the node, so we'll need the node's context to run commands inside the debug daemonset pod.
	context := interactive.NewContext(node.Oc.GetExpecter(), node.Oc.GetErrorChannel())
	var commandErr error
	hugepagesCmdOut := runAndValidateCommand(cmd, context, func() {
		commandErr = fmt.Errorf("failed to get node %s hugepages per numa", node.Name)
	})
	if commandErr != nil {
		return numaHugePagesPerSize{}, commandErr
	}

	hugepages = numaHugePagesPerSize{}
	r := regexp.MustCompile(outputRegex)
	for _, line := range strings.Split(hugepagesCmdOut, "\n") {
		values := r.FindStringSubmatch(line)
		if len(values) != numRegexFields {
			return numaHugePagesPerSize{}, fmt.Errorf("failed to parse node's numa hugepages output line:%s", line)
		}

		numaNode, _ := strconv.Atoi(values[1])
		hpSize, _ := strconv.Atoi(values[2])
		hpCount, _ := strconv.Atoi(values[3])

		hugepagesCfg := hugePagesConfig{
			hugepagesCount: hpCount,
			hugepagesSize:  hpSize,
		}

		if numaHugepagesCfg, exists := hugepages[numaNode]; exists {
			numaHugepagesCfg = append(numaHugepagesCfg, hugepagesCfg)
			hugepages[numaNode] = numaHugepagesCfg
		} else {
			hugepages[numaNode] = []hugePagesConfig{hugepagesCfg}
		}
	}

	log.Infof("Node %s hugepages: %s", node.Name, hugepages)
	return hugepages, nil
}

// getMachineConfig gets the machineconfig in json format does the unmarshalling.
func getMachineConfig(mcName string) (machineConfig, error) {
	var commandErr error

	// Local shell context is needed for the command handler.
	context := common.GetContext()
	mcJSON := runAndValidateCommand(fmt.Sprintf("oc get mc %s -o json", mcName), context, func() {
		commandErr = fmt.Errorf("failed to get json machineconfig %s", mcName)
	})
	if commandErr != nil {
		return machineConfig{}, commandErr
	}

	var mc machineConfig
	err := json.Unmarshal([]byte(mcJSON), &mc)
	if err != nil {
		return machineConfig{}, fmt.Errorf("failed to unmarshall (err: %v)", err)
	}

	return mc, nil
}

// getMcSystemdUnitsHugepagesConfig gets the hugepages information from machineconfig's systemd units.
func getMcSystemdUnitsHugepagesConfig(mc *machineConfig) (hugepages numaHugePagesPerSize, err error) {
	const UnitContentsRegexMatchLen = 4
	hugepages = numaHugePagesPerSize{}

	r := regexp.MustCompile(`(?ms)HUGEPAGES_COUNT=(\d+).*HUGEPAGES_SIZE=(\d+).*NUMA_NODE=(\d+)`)
	for _, unit := range mc.Spec.Config.Systemd.Units {
		unit.Name = strings.Trim(unit.Name, "\"")
		if !strings.Contains(unit.Name, "hugepages-allocation") {
			continue
		}
		unit.Contents = strings.Trim(unit.Contents, "\"")
		values := r.FindStringSubmatch(unit.Contents)
		if len(values) < UnitContentsRegexMatchLen {
			return numaHugePagesPerSize{}, fmt.Errorf("unable to get hugepages values from mc (contents=%s)", unit.Contents)
		}

		numaNode, _ := strconv.Atoi(values[3])
		hpSize, _ := strconv.Atoi(values[2])
		hpCount, _ := strconv.Atoi(values[1])

		hugepagesCfg := hugePagesConfig{
			hugepagesCount: hpCount,
			hugepagesSize:  hpSize,
		}

		if numaHugepagesCfg, exists := hugepages[numaNode]; exists {
			numaHugepagesCfg = append(numaHugepagesCfg, hugepagesCfg)
			hugepages[numaNode] = numaHugepagesCfg
		} else {
			hugepages[numaNode] = []hugePagesConfig{hugepagesCfg}
		}
	}

	if len(hugepages) > 0 {
		log.Infof("Machineconfig's systemd.units hugepages: %v", hugepages)
	} else {
		log.Infof("No hugepages found in machineconfig system.units")
	}

	return hugepages, nil
}

// testNodeHugepagesWithMcSystemd compares the node's hugepages values against the mc's systemd units ones.
func testNodeHugepagesWithMcSystemd(nodeName string, nodeNumaHugePages, mcSystemdHugepages numaHugePagesPerSize) (bool, error) {
	// Iterate through mc's numas and make sure they exist and have the same sizes and values in the node.
	for mcNumaIdx, mcNumaHugepageCfgs := range mcSystemdHugepages {
		nodeNumaHugepageCfgs, exists := nodeNumaHugePages[mcNumaIdx]
		if !exists {
			return false, fmt.Errorf("node %s has no hugepages config for machine config's numa %d", nodeName, mcNumaIdx)
		}

		// For this numa, iterate through each of the mc's hugepages sizes and compare with node ones.
		for _, mcHugepagesCfg := range mcNumaHugepageCfgs {
			configMatching := false
			for _, nodeHugepagesCfg := range nodeNumaHugepageCfgs {
				if nodeHugepagesCfg.hugepagesSize == mcHugepagesCfg.hugepagesSize && nodeHugepagesCfg.hugepagesCount == mcHugepagesCfg.hugepagesCount {
					log.Infof("MC numa=%d, hugepages count:%d, size:%d match node ones: %s",
						mcNumaIdx, mcHugepagesCfg.hugepagesCount, mcHugepagesCfg.hugepagesSize, nodeNumaHugePages)
					configMatching = true
					break
				}
			}
			if !configMatching {
				return false, fmt.Errorf(fmt.Sprintf("MC numa=%d, hugepages (count:%d, size:%d) not matching node ones: %s",
					mcNumaIdx, mcHugepagesCfg.hugepagesCount, mcHugepagesCfg.hugepagesSize, nodeNumaHugePages))
			}
		}
	}

	return true, nil
}

// testNodeHugepagesWithKernelArgs compares node hugepages against kernelArguments config.
// The total count of hugepages of the size defined in the kernelArguments must match the kernArgs' hugepages value.
// For other sizes, the sum should be 0.
func testNodeHugepagesWithKernelArgs(nodeName string, nodeNumaHugePages numaHugePagesPerSize, kernelArgsHugepagesPerSize map[int]int) (bool, error) {
	for size, count := range kernelArgsHugepagesPerSize {
		total := 0
		for numaIdx, numaHugepages := range nodeNumaHugePages {
			found := false
			for _, hugepages := range numaHugepages {
				if hugepages.hugepagesSize == size {
					total += hugepages.hugepagesCount
					found = true
					break
				}
			}
			if !found {
				return false, fmt.Errorf("node %s: numa %d has no hugepages of size %d", nodeName, numaIdx, size)
			}
		}

		if total == count {
			log.Infof("kernelArguments' hugepages count:%d, size:%d match total node ones for that size.", count, size)
		} else {
			return false, fmt.Errorf("node %s: total hugepages of size %d won't match (node count=%d, expected=%d)",
				nodeName, size, total, count)
		}
	}

	return true, nil
}

func getNodeMachineConfig(nodeName string, machineconfigs map[string]machineConfig) machineConfig {
	mcName := strings.Trim(getMcName(common.GetContext(), nodeName), "\"")
	log.Infof("Node %s is using machineconfig %s", nodeName, mcName)

	if mc, exists := machineconfigs[mcName]; exists {
		log.Infof("MC %s: json already parsed.", mcName)
		return mc
	}

	mc, err := getMachineConfig(mcName)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Unable to unmarshall mc %s from node %s", mcName, nodeName))
	}
	machineconfigs[mcName] = mc

	return mc
}

func testHugepages(env *config.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestHugepagesNotManuallyManipulated)
	ginkgo.It(testID, func() {
		// Map to save already retrieved and parsed machineconfigs.
		machineconfigs := map[string]machineConfig{}
		var badNodes []string

		for _, node := range env.NodesUnderTest {
			if !node.IsWorker() || !node.HasDebugPod() {
				continue
			}

			ginkgo.By(fmt.Sprintf("Should get node %s numa's hugepages values.", node.Name))
			nodeNumaHugePages, err := getNodeNumaHugePages(node)
			if err != nil {
				ginkgo.Fail(fmt.Sprintf("Unable to get node hugepages values from node %s", node.Name))
			}

			// Get and parse node's machineconfig, in case it's not already parsed.
			mc := getNodeMachineConfig(node.Name, machineconfigs)

			ginkgo.By("Should parse machineconfig's kernelArguments and systemd's hugepages units.")
			mcSystemdHugepages, err := getMcSystemdUnitsHugepagesConfig(&mc)
			if err != nil {
				ginkgo.Fail(fmt.Sprintf("Failed to get MC systemd hugepages config. Error: %v", err))
			}

			// KernelArguments params will only be used in case no systemd units were found.
			if len(mcSystemdHugepages) == 0 {
				ginkgo.By("Comparing MC KernelArguments hugepages info against node values.")
				hugepagesPerSize, _, err := getMcHugepagesFromMcKernelArguments(&mc)
				if err != nil {
					ginkgo.Fail(fmt.Sprintf("Unable to get mc kernelArguments hugepages from node %s. Error: %v", node.Name, err))
				}
				if pass, err := testNodeHugepagesWithKernelArgs(node.Name, nodeNumaHugePages, hugepagesPerSize); !pass {
					log.Error(err)
					badNodes = append(badNodes, node.Name)
				}
			} else {
				ginkgo.By("Comparing MC Systemd hugepages info against node values.")
				if pass, err := testNodeHugepagesWithMcSystemd(node.Name, nodeNumaHugePages, mcSystemdHugepages); !pass {
					log.Error(err)
					badNodes = append(badNodes, node.Name)
				}
			}
		}
		gomega.Expect(badNodes).To(gomega.BeNil())
	})
}
