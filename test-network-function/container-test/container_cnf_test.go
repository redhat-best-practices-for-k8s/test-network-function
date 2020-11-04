package container

import (
	expect "github.com/google/goexpect"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container/testcases"

	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

const (
	// The default test timeout.
	defaultTimeoutSeconds = 10
)

var (
	defaultTimeout = time.Duration(defaultTimeoutSeconds) * time.Second
	context        *interactive.Context
	err            error
	podsInTest     *TestPodConfig
)

//TestTemplateConfigFileMap is the names of test files configure for test under container-test/config folder
var TestTemplateConfigFileMap = map[string]string{
	"PRIVILEGED_POD": "privileged.yaml",
}

type TestPodConfig struct {
	Pod []struct {
		Name      string   `yaml:"name" json:"name"`
		Namespace string   `yaml:"namespace" json:"namespace"`
		Tests     []string `yaml:"tests" json:"tests"`
	} `yaml:"pod" json:"pod"`
}

type ConfiguredTestCase struct {
	TestCase []string `yaml:"testconfigured"`
}

var _ = ginkgo.Describe("container_test", func() {
	podsInTest, _ = getPodConf()
	gomega.Expect(podsInTest).ToNot(gomega.BeNil())
	ginkgo.When("A local shell is spawned", func() {
		goExpectSpawner := interactive.NewGoExpectSpawner()
		var spawner interactive.Spawner = goExpectSpawner
		context, err = interactive.SpawnShell(&spawner, defaultTimeout, expect.Verbose(true))
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(context).ToNot(gomega.BeNil())
		gomega.Expect(context.GetExpecter()).ToNot(gomega.BeNil())
	})
	// add pdo is running check later
	ginkgo.When("pod is in running state", func() {
		for _, pod := range podsInTest.Pod {
			for _, testType := range pod.Tests {
				testFile := loadConfiguredTestFile(testType)
				gomega.Expect(testFile).ToNot(gomega.BeNil())
				testcase := renderTestCaseSpec(testType, testFile)
				gomega.Expect(testcase).ToNot(gomega.BeNil())
				for _, testCase := range testcase.TestCase {
					if !testCase.SkipTest {
						runTestsOnPod(pod.Name, pod.Namespace, testCase)
					}

				}
			}

		}
	})

})

func runTestsOnPod(name, namespace string, testCmd testcases.BaseTestCase) {
	ginkgo.When("pod test ", func() {
		ginkgo.It("checks for "+testCmd.Name, func() {
			podInTest := container.NewPod(testCmd.Command, name, namespace, testCmd.ExptectedStatus, testCmd.ResultType, testCmd.Action, defaultTimeout)
			gomega.Expect(podInTest).ToNot(gomega.BeNil())
			test, err := tnf.NewTest(context.GetExpecter(), podInTest, []reel.Handler{podInTest}, context.GetErrorChannel())
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(test).ToNot(gomega.BeNil())
			testResult, err := test.Run()
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(testResult).To(gomega.Equal(tnf.SUCCESS))
		})
	})
}

func loadConfiguredTestFile(testName string) *ConfiguredTestCase {
	var configuredTestCase *ConfiguredTestCase
	yamlFile, err := ioutil.ReadFile("./config/" + testcases.TestTemplateFileMap[testName])
	if err != nil {
		return nil
	}
	err = yaml.Unmarshal(yamlFile, &configuredTestCase)
	if err != nil {
		return nil
	}

	return configuredTestCase
}

func renderTestCaseSpec(testName string, config *ConfiguredTestCase) *testcases.BaseTestCaseConfigSpec {
	t, err := testcases.LoadTestCaseSpecs(testName)
	if err != nil {
		return nil
	}
	for _, elem := range config.TestCase {
		for i, e := range t.TestCase {
			if e.Name == elem {
				t.TestCase[i].SkipTest = false
			}
		}

	}
	return t
}

//NewConfig  returns a new decoded Config struct
func getPodConf() (*TestPodConfig, error) {
	var file *os.File
	var err error
	// Create config structure
	config := &TestPodConfig{}
	// Open config file
	if file, err = os.Open("config.yml"); err != nil {
		return nil, err
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
