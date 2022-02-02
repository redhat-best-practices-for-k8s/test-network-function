package diagnostic

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/test-network-function/common"

	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterversion"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodedebug"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// defaultTimeoutSeconds contains the default timeout in seconds.
	defaultTimeoutSeconds = 20
)

var (
	// defaultTestTimeout is the timeout for the test.
	defaultTestTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

	// nodeSummary stores the raw JSON output of `oc get nodes -o json`
	nodeSummary = make(map[string]interface{})

	cniPlugins = make([]CniPlugin, 0)

	versionsOcp clusterversion.ClusterVersion

	nodesHwInfo = NodesHwInfo{}

	initialRuntimeTestEnvironment config.TestEnvironment

	// csiDriver stores the csi driver JSON output of `oc get csidriver -o json`
	csiDriver = make(map[string]interface{})

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

// CniPlugin holds info about a CNI plugin
// The JSON fields come from the jq output
type CniPlugin struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Version string      `json:"version"`
	Plugins interface{} `json:"plugins"`
}

// NodeHwInfo node HW info
type NodeHwInfo struct {
	NodeName string
	Lscpu    map[string]string   // lscpu output parsed as entry to value map
	IPconfig map[string][]string // 'ip a' output parsed as interface name to output lines map
	Lsblk    interface{}         // 'lsblk -J' output un-marshaled into an unknown type
	Lspci    []string            // lspci output parsed to individual lines
}

// NodesHwInfo one master one worker
type NodesHwInfo struct {
	Master NodeHwInfo
	Worker NodeHwInfo
}

func GetDiagnosticData() []error {
	errs := []error{}
	if len(env.PodsUnderTest) == 0 {
		errs = append(errs, errors.New("nod pods under test found"))
	}
	if len(env.ContainersUnderTest) == 0 {
		errs = append(errs, errors.New("no containers under test found"))
	}

	if err := getOcpVersions(); err != nil {
		errs = append(errs, fmt.Errorf("failed to get ocp version. Error: %v", err))
	}

	if err := getNodes(); err != nil {
		errs = append(errs, fmt.Errorf("failed to get nodes info. Error: %v", err))
	}

	if err := getCniPlugins(); err != nil {
		errs = append(errs, fmt.Errorf("failed to get CNI plugins info. Error: %v", err))
	}

	if err := getClusterCSIInfo(); err != nil {
		errs = append(errs, fmt.Errorf("failed to get cluster CSI info. Error: %v", err))
	}

	if hwErrs := getNodesHwInfo(); len(hwErrs) > 0 {
		errs = append(errs, hwErrs...)
	}

	saveInitialRuntimeEnv()

	return errs
}

func getNodes() error {
	log.Infof("Getting Nodes information.")

	context := env.GetLocalShellContext()
	defer env.CloseLocalShellContext()

	tester, handlers, jsonParseResult, err := generic.NewGenericFromJSONFile(relativeNodesTestPath, relativeSchemaPath)
	if validParseResult := jsonParseResult.Valid(); err != nil || !validParseResult {
		return fmt.Errorf("failed to create handler to get nodes information (validParseResult: %v, error: %v)", validParseResult, err)
	}

	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	if err != nil {
		return fmt.Errorf("failed to create tester to get nodes information (error: %v)", err)
	}

	test.RunWithCallbacks(func() {
		genericTest := (*tester).(*generic.Generic)
		matches := genericTest.Matches
		if n := len(matches); n != 1 {
			err = fmt.Errorf("failed to parse console output for %s (len=%d)", strings.Join(genericTest.Args(), " "), n)
			return
		}

		match := matches[0].Match
		err = json.Unmarshal([]byte(match), &nodeSummary)
	}, func() {
		err = errors.New("failed to execute tester to get nodes information")
	}, func(handlerError error) {
		err = fmt.Errorf("failed to execute tester to get nodes information (error: %v)", handlerError)
	})

	return err
}

// GetNodeSummary returns the result of running `oc get nodes -o json`.
func GetNodeSummary() map[string]interface{} {
	return nodeSummary
}

// GetCniPlugins return the found plugins
func GetCniPlugins() []CniPlugin {
	return cniPlugins
}

// GetVersionsOcp return OCP versions
func GetVersionsOcp() clusterversion.ClusterVersion {
	return versionsOcp
}

// GetNodesHwInfo returns an object with HW info of one master and one worker
func GetNodesHwInfo() NodesHwInfo {
	return nodesHwInfo
}

// GetCsiDriverInfo returns the CSI driver info of running `oc get csidriver -o json`.
func GetCsiDriverInfo() map[string]interface{} {
	return csiDriver
}

// GetInitialRuntimeEnv returns initial test environment
func GetInitialRuntimeEnv() config.TestEnvironment {
	return initialRuntimeTestEnvironment
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

func listNodeCniPlugins(nodeName string) ([]CniPlugin, error) {
	// This command will return a JSON array, with the name, cniVersion and plugins fields from the cat output
	const command = "cat /host/etc/cni/net.d/[0-999]* | jq -s '[ .[] | {name:.name, type:.type, version:.cniVersion, plugins: .plugins}]'"

	nodes := config.GetTestEnvironment().NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		return nil, fmt.Errorf("failed to create cni plugins handler (error: %v)", err)
	}

	result := []CniPlugin{}
	test.RunWithCallbacks(func() {
		err = json.Unmarshal([]byte(tester.Raw), &result)
	}, func() {
		err = errors.New("failed to execute tester to get CNI plugins")
	}, func(handlerError error) {
		err = fmt.Errorf("failed to execute tester to get CNI plugins (error: %v)", handlerError)
	})

	return result, err
}

func getOcpVersions() error {
	log.Infof("Getting Openshift versions.")

	context := env.GetLocalShellContext()
	defer env.CloseLocalShellContext()

	tester := clusterversion.NewClusterVersion(defaultTestTimeout)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		return fmt.Errorf("failed to create ocp versions handler test. Error: %v", err)
	}

	test.RunWithCallbacks(func() {
		versionsOcp = tester.GetVersions()
	}, func() {
		err = errors.New("ocp versions handler failed")
	}, func(handlerError error) {
		err = fmt.Errorf("<ocp versions handler failed with error: %v", handlerError)
	})

	return err
}

func getCniPlugins() error {
	log.Infof("Getting CNI plugins.")
	env = config.GetTestEnvironment()

	// CNI plugins info will be retrieved from a master node.
	nodeName := getMasterNodeName(env)
	if nodeName == "" {
		return errors.New("master node's name is empty")
	}

	var err error
	cniPlugins, err = listNodeCniPlugins(nodeName)
	if len(cniPlugins) == 0 {
		return errors.New("CNI plugins list is empty")
	}

	return err
}

func getNodeHwInfo(hwInfo *NodeHwInfo, nodeName, nodeType string) []error {
	hwInfo.NodeName = nodeName

	errs := []error{}
	var err error
	hwInfo.Lscpu, err = getNodeLscpu(nodeName)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get %s node lscpu info - %v", nodeType, err))
	}

	hwInfo.IPconfig, err = getNodeIPconfig(nodeName)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get %s node IP config - %v", nodeType, err))
	}

	hwInfo.Lsblk, err = getNodeLsblk(nodeName)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get %s node lsblk info - %v", nodeType, err))
	}

	hwInfo.Lspci, err = getNodeLspci(nodeName)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get %s node lspci info - %v", nodeType, err))
	}

	return errs
}

func getNodesHwInfo() []error {
	log.Infof("Getting Master & Worker nodes' hardware information.")

	env = config.GetTestEnvironment()
	errs := []error{}

	masterNodeName := getMasterNodeName(env)
	if masterNodeName == "" {
		errs = append(errs, fmt.Errorf("failed to get master node hw info: name is empty"))
	}

	workerNodeName := getWorkerNodeName(env)
	if workerNodeName == "" {
		errs = append(errs, fmt.Errorf("failed to get worker node hw info: name is empty"))
	}

	errs = append(errs, getNodeHwInfo(&nodesHwInfo.Master, masterNodeName, "master")...)
	errs = append(errs, getNodeHwInfo(&nodesHwInfo.Worker, workerNodeName, "worker")...)

	return errs
}

// saveInitialRuntimeEnv Saves the initial runtime environment to a global variable for use in suite_test.go
func saveInitialRuntimeEnv() {
	log.Infof("Saving initial runtime environement in diagnostics")
	initialRuntimeTestEnvironment = *config.GetTestEnvironment()
}

func getNodeLscpu(nodeName string) (map[string]string, error) {
	const command = "lscpu"
	const numSplitSubstrings = 2

	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		return nil, fmt.Errorf("failed to create ocp versions handler test. Error: %v", err)
	}

	result := map[string]string{}
	test.RunWithCallbacks(func() {
		for _, line := range tester.Processed {
			fields := strings.SplitN(line, ":", numSplitSubstrings)
			result[fields[0]] = strings.TrimSpace(fields[1])
		}
	}, func() {
		err = errors.New("node lscpu test failed")
	}, func(handlerError error) {
		err = fmt.Errorf("node lscpu test failed with error: %v", handlerError)
	})

	return result, err
}

func getNodeIPconfig(nodeName string) (map[string][]string, error) {
	const command = "ip a"
	const numSplitSubstrings = 3

	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		return nil, fmt.Errorf("failed to create ocp versions handler test. Error: %v", err)
	}

	result := map[string][]string{}
	test.RunWithCallbacks(func() {
		deviceName := ""
		for _, line := range tester.Processed {
			if line == "" {
				continue
			}
			if line[0] != ' ' {
				fields := strings.SplitN(line, ":", numSplitSubstrings)
				deviceName = fields[1]
				line = fields[2]
			}
			result[deviceName] = append(result[deviceName], strings.TrimSpace(line))
		}
	}, func() {
		err = errors.New("node lscpu test failed")
	}, func(handlerError error) {
		err = fmt.Errorf("node lscpu test failed with error: %v", handlerError)
	})

	return result, err
}

func getNodeLsblk(nodeName string) (interface{}, error) {
	const command = "lsblk -J"

	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, false, false)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		return nil, fmt.Errorf("failed to create ocp versions handler test. Error: %v", err)
	}

	result := map[string]interface{}{}
	test.RunWithCallbacks(func() {
		err = json.Unmarshal([]byte(tester.Raw), &result)
	}, func() {
		err = errors.New("node lscpu test failed")
	}, func(handlerError error) {
		err = fmt.Errorf("node lscpu test failed with error: %v", handlerError)
	})

	return result, err
}

func getNodeLspci(nodeName string) ([]string, error) {
	const command = "lspci"

	env = config.GetTestEnvironment()
	nodes := env.NodesUnderTest
	context := nodes[nodeName].DebugContainer.GetOc()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	if err != nil {
		return nil, fmt.Errorf("failed to create ocp versions handler test. Error: %v", err)
	}

	result := []string{}
	test.RunWithCallbacks(func() {
		result = tester.Processed
	}, func() {
		err = errors.New("node lscpu test failed")
	}, func(handlerError error) {
		err = fmt.Errorf("node lscpu test failed with error: %v", handlerError)
	})
	return result, err
}

// check CSI driver info in cluster
func getClusterCSIInfo() error {
	log.Infof("Getting cluster CSI information.")

	context := env.GetLocalShellContext()
	tester, handlers, jsonParseResult, err := generic.NewGenericFromJSONFile(relativeCsiDriverTestPath, common.RelativeSchemaPath)
	if validParseResult := jsonParseResult.Valid(); err != nil || !validParseResult {
		return fmt.Errorf("failed to create handler to get cluster CSI info (validParseResult: %v, error: %v)", validParseResult, err)
	}

	test, err := tnf.NewTest(context.GetExpecter(), *tester, handlers, context.GetErrorChannel())
	if err != nil {
		return fmt.Errorf("failed to create tester to get cluster CSI info (error: %v)", err)
	}

	test.RunWithCallbacks(func() {
		genericTest := (*tester).(*generic.Generic)
		matches := genericTest.Matches
		if n := len(matches); n != 1 {
			err = fmt.Errorf("failed to parse console output for %s (len=%d)", strings.Join(genericTest.Args(), " "), n)
			return
		}

		match := matches[0].Match
		err = json.Unmarshal([]byte(match), &csiDriver)
	}, func() {
		err = errors.New("failed to execute tester to get cluster CSI information")
	}, func(handlerError error) {
		err = fmt.Errorf("failed to execute tester to get cluster CSI information (error: %v)", handlerError)
	})

	return err
}
