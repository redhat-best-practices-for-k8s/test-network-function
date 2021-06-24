package diagnostic

import (
	"encoding/json"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/test-network-function/test-network-function/test-network-function/identifiers"
	"github.com/test-network-function/test-network-function/test-network-function/results"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodedebug"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/nodenames"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// testSuiteSpec contains the name of the Ginkgo test specification.
	testSuiteSpec = "diagnostic"
	// defaultTimeoutSeconds contains the default timeout in secons.
	defaultTimeoutSeconds = 20
)

var (
	// defaultTestTimeout is the timeout for the test.
	defaultTestTimeout = time.Duration(defaultTimeoutSeconds) * time.Second

	// nodeSummary stores the raw JSON output of `oc get nodes -o json`
	nodeSummary = make(map[string]interface{})

	cniPlugins = make([]CniPlugin, 0)

	// nodesTestPath is the file location of the nodes.json test case relative to the project root.
	nodesTestPath = path.Join("pkg", "tnf", "handlers", "node", "nodes.json")

	// pathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// relativeNodesTestPath is the relative path to the nodes.json test case.
	relativeNodesTestPath = path.Join(pathRelativeToRoot, nodesTestPath)

	// relativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)

	// schemaPath is the path to the generic-test.schema.json JSON schema relative to the project root.
	schemaPath = path.Join("schemas", "generic-test.schema.json")
)

// createShell sets up a local shell expect.Expecter, checking errors along the way.
func createShell() *interactive.Context {
	context, err := interactive.SpawnShell(interactive.CreateGoExpectSpawner(), defaultTestTimeout, interactive.Verbose(true))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(context).ToNot(gomega.BeNil())
	return context
}

var _ = ginkgo.Describe(testSuiteSpec, func() {
	ginkgo.When("a cluster is set up and installed with OpenShift", func() {
		ginkgo.It("should report all available nodeSummary", func() {
			defer results.RecordResult(identifiers.TestExtractNodeInformationIdentifier)
			context := createShell()

			test, handlers, jsonParseResult, err := generic.NewGenericFromJSONFile(relativeNodesTestPath, relativeSchemaPath)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(jsonParseResult).ToNot(gomega.BeNil())
			gomega.Expect(jsonParseResult.Valid()).To(gomega.BeTrue())
			gomega.Expect(handlers).ToNot(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())

			tester, err := tnf.NewTest(context.GetExpecter(), *test, handlers, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(tester).ToNot(gomega.BeNil())

			result, err := tester.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(result).To(gomega.Equal(tnf.SUCCESS))

			genericTest := (*test).(*generic.Generic)
			gomega.Expect(genericTest).ToNot(gomega.BeNil())
			matches := genericTest.Matches
			gomega.Expect(len(matches)).To(gomega.Equal(1))
			match := genericTest.GetMatches()[0]
			err = json.Unmarshal([]byte(match.Match), &nodeSummary)
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("should report all CNI plugins", func() {
			defer results.RecordResult(identifiers.TestListCniPluginsIdentifier)
			testCniPlugins()
		})
	})
})

// CniPlugin holds info about a CNI plugin
type CniPlugin struct {
	Name    string
	version string
}

// GetNodeSummary returns the result of running `oc get nodes -o json`.
func GetNodeSummary() map[string]interface{} {
	return nodeSummary
}

// GetCniPlugins return the found plugins
func GetCniPlugins() []CniPlugin {
	return cniPlugins
}

func getMasterNodeName() string {
	const masterNodeLabel = "node-role.kubernetes.io/master"
	context := createShell()
	tester := nodenames.NewNodeNames(defaultTestTimeout, map[string]*string{masterNodeLabel: nil})
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	nodeNames := tester.GetNodeNames()
	gomega.Expect(nodeNames).NotTo(gomega.BeEmpty())
	return nodeNames[0]
}

func listNodeCniPlugins(nodeName string) []CniPlugin {
	const command = "jq -r .name,.cniVersion '/etc/cni/net.d/*'"
	result := []CniPlugin{}
	context := createShell()
	tester := nodedebug.NewNodeDebug(defaultTestTimeout, nodeName, command, true, true)
	test, err := tnf.NewTest(context.GetExpecter(), tester, []reel.Handler{tester}, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	testResult, err := test.Run()
	gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(len(tester.Processed)%2 == 0).To(gomega.BeTrue())
	for i := 0; i < len(tester.Processed); i += 2 {
		result = append(result, CniPlugin{
			tester.Processed[i],
			tester.Processed[i+1],
		})
	}
	return result
}

func testCniPlugins() {
	if isMinikube() {
		ginkgo.Skip("can't use 'oc debug' in minikube")
	}
	// get name of a master node
	nodeName := getMasterNodeName()
	gomega.Expect(nodeName).ToNot(gomega.BeEmpty())
	// get CNI plugins from node
	cniPlugins = listNodeCniPlugins(nodeName)
	gomega.Expect(cniPlugins).ToNot(gomega.BeNil())
}

func isMinikube() bool {
	b, _ := strconv.ParseBool(os.Getenv("TNF_MINIKUBE_ONLY"))
	return b
}
