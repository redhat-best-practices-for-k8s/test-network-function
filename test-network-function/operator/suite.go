package operator

import (
	"fmt"
	expect "github.com/google/goexpect"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	operatorTestConfig "github.com/redhat-nfvpe/test-network-function/pkg/tnf/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/operator"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/testcases"
	"strings"
	"time"
)

const (
	// The default test timeout.
	defaultTimeoutSeconds = 10
	configuredTestFile    = "testconfigure.yml"
	testSpecName          = "operator"
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	operatorInTest *operatorTestConfig.TnfContainerOperatorTestConfig
)

var _ = ginkgo.Describe(testSpecName, func() {

	ginkgo.When("a local shell is spawned", func() {
		goExpectSpawner := interactive.NewGoExpectSpawner()
		var spawner interactive.Spawner = goExpectSpawner
		context, err = interactive.SpawnShell(&spawner, defaultTimeout, expect.Verbose(true))
		ginkgo.It("should be created without error", func() {
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(context).ToNot(gomega.BeNil())
			gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
		})
	})
	defer ginkgo.GinkgoRecover()
	operatorInTest, err = operatorTestConfig.GetConfig()
	gomega.Expect(err).To(gomega.BeNil())
	gomega.Expect(operatorInTest).ToNot(gomega.BeNil())

	for _, operator := range operatorInTest.Operator {
		// TODO: Gather facts for operator
		for _, testType := range operator.Tests {
			testFile, err := testcases.LoadConfiguredTestFile(configuredTestFile)
			gomega.Expect(testFile).ToNot(gomega.BeNil())
			gomega.Expect(err).To(gomega.BeNil())
			testConfigure := testcases.ContainsConfiguredTest(testFile.OperatorTest, testType)
			renderedTestCase, err := testConfigure.RenderTestCaseSpec(testcases.Operator, testType)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(renderedTestCase).ToNot(gomega.BeNil())
			for _, testCase := range renderedTestCase.TestCase {
				if !testCase.SkipTest {
					if testCase.ExpectedType == "function" {
						for _, val := range testCase.ExpectedStatus {
							testCase.ExpectedStatusFn(testcases.StatusFunctionType(val))
						}
					}
					args := []interface{}{operator.Name, operator.Namespace}
					runTestsOnOperator(args, operator.Name, operator.Namespace, testCase)
				}
			}
		}
	}
})

func runTestsOnOperator(args []interface{}, name, namespace string, testCmd testcases.BaseTestCase) {
	ginkgo.When(fmt.Sprintf("operator under test is: %s/%s ", namespace, name), func() {
		ginkgo.It(fmt.Sprintf("tests for: %s", testCmd.Name), func() {
			cmdArgs := strings.Split(fmt.Sprintf(testCmd.Command, args...), " ")
			opInTest := operator.NewOperator(cmdArgs, name, namespace, testCmd.ExpectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
			gomega.Expect(opInTest).ToNot(gomega.BeNil())
			test, err := tnf.NewTest(context.GetExpecter(), opInTest, []reel.Handler{opInTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		})
	})
}
