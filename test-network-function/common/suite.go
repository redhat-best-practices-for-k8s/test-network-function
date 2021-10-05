package common

import (
	"fmt"
	"path"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
)

var (
	// PathRelativeToRoot is used to calculate relative filepaths for the `test-network-function` executable entrypoint.
	pathRelativeToRoot = path.Join("..")

	// TestFile is the file location of the command.json test case relative to the project root.
	testFile = path.Join("pkg", "tnf", "handlers", "command", "command.json")
	// RelativeSchemaPath is the relative path to the generic-test.schema.json JSON schema.
	relativeSchemaPath = path.Join(pathRelativeToRoot, schemaPath)
	// pathToTestFile is the relative path to the command.json test case.
	pathToTestFile = path.Join(pathRelativeToRoot, testFile)
)

// TeardownNodeDebugSession closes the session opened with nodes
//
func TeardownNodeDebugSession() {
	if IsMinikube() {
		return
	}
	log.Info("test suite teardown: start")
	env := config.GetTestEnvironment()
	env.LoadAndRefresh()
	for _, node := range env.NodesUnderTest {
		log.Debug("should close session with ", node.Name)
	}
	commonContext := interactive.GetContext()
	for _, node := range env.NodesUnderTest {
		const command = "exit "
		context := node.Oc
		if context != nil {
			node.Oc.Close()
			time.Sleep(1 * time.Second)
			log.Info("send exit command to pod=", node.Oc.GetPodName())
			if false {
				err := (*context.GetExpecter()).Send(command)
				if err != nil {
					log.Error("Error when sending a command to exit pod")
				}
				//nolint:gomnd
				time.Sleep(100 * time.Millisecond)
			}
			killPodCommand := fmt.Sprintf("oc delete pod %s", node.Oc.GetPodName())
			err := (*commonContext.GetExpecter()).Send(killPodCommand)
			if err != nil {
				log.Error("kil pod error")
			}
		} else {
			log.Warn("Oc context of node=", node.Name, " is nil")
		}
	}
	log.Info("test suite teardown: done")
}

var _ = ginkgo.Describe(commonTestKey, func() {
	ginkgo.BeforeSuite(func() {
		log.Info("test suite setup")
	})
	ginkgo.AfterSuite(func() {
		TeardownNodeDebugSession()
		log.Info("After Suite")
	})
})

//nolint:deadcode
func killPod(podName string, timeout time.Duration) {
	command := fmt.Sprintf("oc delete pod %s", podName)
	log.Debug("kill pod [start] podname=", podName, " using command= ", command)
	defer log.Debug("kill pod [done] podname=", podName)
	values := make(map[string]interface{})
	values["COMMAND"] = command
	values["TIMEOUT"] = timeout
	context := interactive.GetContext()
	tester, handler, result, err := generic.NewGenericFromMap(pathToTestFile, relativeSchemaPath, values)

	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(result).ToNot(gomega.BeNil())
	gomega.Expect(result.Valid()).To(gomega.BeTrue())
	gomega.Expect(handler).ToNot(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	test, err := tnf.NewTest(context.GetExpecter(), *tester, handler, context.GetErrorChannel())
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(tester).ToNot(gomega.BeNil())

	test.RunAndValidate()
}
