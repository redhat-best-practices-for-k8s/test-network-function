package diagnostic

import (
	"encoding/json"
	"path"
	"strings"
	"time"

	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/test-network-function/common"
	"github.com/test-network-function/test-network-function/test-network-function/identifiers"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterversion"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodedebug"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"github.com/test-network-function/test-network-function/test-network-function/results"
)

const (
	// defaultTimeoutSeconds contains the default timeout in secons.
	defaultTimeoutSeconds = 20
)

var (
	// defaultTestTimeout is the timeout for the test.
	defaultTestTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

	// nodesTestPath is the file location of the nodes.json test case relative to the project root.
	nodesTestPath = path.Join("pkg", "tnf", "handlers", "node", "nodes.json")

	// csiDriverTestPath is the file location of the csidriver.json test case relative to the project root.
	csiDriverTestPath = path.Join("pkg", "tnf", "handlers", "csidriver", "csidriver.json")

	// relativeCsiDriverTestPath is the relative path to the csidriver.json test case.
	relativeCsiDriverTestPath = path.Join(pathRelativeToRoot, csiDriverTestPath)

	// pathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// relativeNodesTestPath is the relative path to the nodes.json test case.
	relativeNodesTestPath = path.Join(pathRelativeToRoot, nodesTestPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")

	// retrieve the singleton instance of test environment
	env *config.TestEnvironment = config.GetTestEnvironment()
)

var _ = ginkgo.Describe(common.DiagnosticTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.DiagnosticTestKey) {
		ginkgo.BeforeEach(func() {
			env.LoadAndRefresh()
			gomega.Expect(len(env.PodsUnderTest)).ToNot(gomega.Equal(0))
			gomega.Expect(len(env.ContainersUnderTest)).ToNot(gomega.Equal(0))
		})
		ginkgo.ReportAfterEach(results.RecordResult)
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestclusterVersionIdentifier)
		ginkgo.It(testID, func() {
			testOcpVersion()
		})

		testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestExtractNodeInformationIdentifier)
		ginkgo.It(testID, func() {
			context := common.GetContext()

			tester, handlers, jsonParseResult, err := generic.NewGenericFromJSONFile(relativeNodesTestPath, relativeSchemaPath)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(jsonParseResult).ToNot(gomega.BeNil())
			gomega.Expect(jsonParseResult.Valid()).To(gomega.BeTrue())
			gomega.Expect(handlers).ToNot(gomega.BeNil())
			gomega.Expect(tester).ToNot(gomega.BeNil())

			test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())

			test.RunAndValidate()

			genericTest := (*tester).(*generic.Generic)
			gomega.Expect(genericTest).ToNot(gomega.BeNil())
			matches := genericTest.Matches
			gomega.Expect(len(matches)).To(gomega.Equal(1))
			match := genericTest.GetMatches()[0]
			nodeSummary := NewNodeSummary()
			err = json.Unmarshal([]byte(match.Match), &nodeSummary)
			gomega.Expect(err).To(gomega.BeNil())
		})
		testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestListCniPluginsIdentifier)
		ginkgo.It(testID, func() {
			testCniPlugins()
		})
		testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNodesHwInfoIdentifier)
		ginkgo.It(testID, func() {
			testNodesHwInfo()
		})
		testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestClusterCsiInfoIdentifier)
		ginkgo.It(testID, func() {
			listClusterCSIInfo()
		})
	}
})

// CniPlugin holds info about a CNI plugin
type CniPlugin struct {
	Name    string
	version string
}

// NodeHwInfo node HW info
type NodeHwInfo struct {
	NodeName string
	Lscpu    map[string]string   // lscpu output parsed as entry to value map
	Ifconfig map[string][]string // ifconfig output parsed as interface name to output lines map
	Lsblk    interface{}         // 'lsblk -J' output un-marshaled into an unknown type
	Lspci    []string            // lspci output parsed to individual lines
}

// NodesHwInfo one master one worker
type NodesHwInfo struct {
	Master NodeHwInfo
	Worker NodeHwInfo
}

// GetNodeSummary returns the result of running `oc get nodes -o json`.
func NewNodeSummary() map[string]interface{} {
	return make(map[string]interface{})
}

// GetVersionsOcp return OCP versions
func GetVersionsOcp() clusterversion.ClusterVersion {
	return clusterversion.ClusterVersion{}
}

// NewCsiDriverInfo returns the CSI driver info of running `oc get csidriver -o json`.
func NewCsiDriverInfo() map[string]interface{} {
	return make(map[string]interface{})
}

func getMasterNodeName(env *config.TestEnvironment) string {
	for _, node := range env.NodesUnderTest {
		if node.IsMaster() && node.HasDebugPod() {
			return node.Name
		}
	}
	return ""
}

func getWorkerNodeName(env *config.TestEnvironment) string {
	for _, node := range env.NodesUnderTest {
		if node.IsWorker() && node.HasDebugPod() {
			return node.Name
		}
	}
	return ""
}

func listNodeCniPlugins(nodeName string) []CniPlugin {
	const command = "cat /host/etc/cni/net.d/* | chroot /host jq -r .name,.cniVersion"
	result := []CniPlugin{}
	nodes := config.GetTestEnvironment().NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	gomega.Expect(len(tester.Processed)%2 == 0).To(gomega.BeTrue())
	for i := 0; i < len(tester.Processed); i += 2 {
		result = append(result, CniPlugin{
			tester.Processed[i],
			tester.Processed[i+1],
		})
	}
	return result
}

func testOcpVersion() {
	context := common.GetContext()
	tester := clusterversion.NewClusterVersion(defaultTestTimeout)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
}

func testCniPlugins() {
	if common.IsNonOcpCluster() {
		ginkgo.Skip("can't use 'oc debug' in minikube")
	}
	// get name of a master node
	env = config.GetTestEnvironment()
	nodeName := getMasterNodeName(env)
	gomega.Expect(nodeName).ToNot(gomega.BeEmpty())
	// get CNI plugins from node
	cniPlugins := listNodeCniPlugins(nodeName)
	gomega.Expect(cniPlugins).ToNot(gomega.BeNil())
}

func testNodesHwInfo() {
	if common.IsNonOcpCluster() {
		ginkgo.Skip("can't use 'oc debug' in minikube")
	}
	env = config.GetTestEnvironment()
	masterNodeName := getMasterNodeName(env)
	gomega.Expect(masterNodeName).ToNot(gomega.BeEmpty())
	workerNodeName := getWorkerNodeName(env)
	gomega.Expect(workerNodeName).ToNot(gomega.BeEmpty())

	nodesHwInfo := NodesHwInfo{}
	nodesHwInfo.Master.NodeName = masterNodeName
	nodesHwInfo.Master.Lscpu = getNodeLscpu(masterNodeName)
	nodesHwInfo.Master.Ifconfig = getNodeIfconfig(masterNodeName)
	nodesHwInfo.Master.Lsblk = getNodeLsblk(masterNodeName)
	nodesHwInfo.Master.Lspci = getNodeLspci(masterNodeName)
	nodesHwInfo.Worker.NodeName = workerNodeName
	nodesHwInfo.Worker.Lscpu = getNodeLscpu(workerNodeName)
	nodesHwInfo.Worker.Ifconfig = getNodeIfconfig(workerNodeName)
	nodesHwInfo.Worker.Lsblk = getNodeLsblk(workerNodeName)
	nodesHwInfo.Worker.Lspci = getNodeLspci(workerNodeName)
}

func getNodeLscpu(nodeName string) map[string]string {
	const command = "lscpu"
	const numSplitSubstrings = 2
	result := map[string]string{}
	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	for _, line := range tester.Processed {
		fields := strings.SplitN(line, ":", numSplitSubstrings)
		result[fields[0]] = strings.TrimSpace(fields[1])
	}
	return result
}

func getNodeIfconfig(nodeName string) map[string][]string {
	const command = "chroot /host ifconfig"
	const numSplitSubstrings = 2
	result := map[string][]string{}
	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	deviceName := ""
	for _, line := range tester.Processed {
		if line == "" {
			continue
		}
		if line[0] != ' ' {
			fields := strings.SplitN(line, ":", numSplitSubstrings)
			deviceName = fields[0]
			line = fields[1]
		}
		result[deviceName] = append(result[deviceName], strings.TrimSpace(line))
	}
	return result
}

func getNodeLsblk(nodeName string) interface{} {
	const command = "lsblk -J"
	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, false, false)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	result := map[string]interface{}{}
	err = json.Unmarshal([]byte(tester.Raw), &result)
	gomega.Expect(err).To(gomega.BeNil())
	return result
}

func getNodeLspci(nodeName string) []string {
	const command = "lspci"
	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	test.RunAndValidate()
	return tester.Processed
}

// check CSI driver info in cluster
func listClusterCSIInfo() {
	if common.IsNonOcpCluster() {
		ginkgo.Skip("CSI is not checked in minikube")
	}
	context := common.GetContext()
	tester, handlers, result, err := generic.NewGenericFromJSONFile(relativeCsiDriverTestPath, common.RelativeSchemaPath)
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handlers).ToNot(gomega.BeNil())
	gomega.Expect(len(handlers)).To(gomega.Equal(1))
	gomega.Expect(tester).ToNot(gomega.BeNil())
	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(test).ToNot(gomega.BeNil())
	test.RunAndValidate()
	genericTest := (*tester).(*generic.Generic)
	gomega.Expect(genericTest).ToNot(gomega.BeNil())
	matches := genericTest.Matches
	gomega.Expect(len(matches)).To(gomega.Equal(1))
	match := genericTest.GetMatches()[0]
	csiDriver := NewCsiDriverInfo()
	err = json.Unmarshal([]byte(match.Match), &csiDriver)
	gomega.Expect(err).To(gomega.BeNil())
}
