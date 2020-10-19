package testcases

import (
	"encoding/json"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/container/testcases/data"
)

//TestTemplateDataMap  is map of available json data test case templates
var TestTemplateDataMap = map[string]string{
	"PRIVILEGED_POD": data.PrivilegedPodJSON,
}

//TestTemplateFileMap is map of configured test case filenames
var TestTemplateFileMap = map[string]string{
	"PRIVILEGED_POD": "privileged.yml",
}

//OutRegExp types of available regular expression to parse matched result
var OutRegExp = map[string]string{
	"ALLOW_ALL":     `.+`,
	"NULL_FALSE":    `^\b(null|false)\b$`,
	"TRUE":          `^\b(true)\b$`,
	"NULL":          `^\b(null)\b$`,
	"NOT_SET":       `^\b(null|false)\b$`,
	"ROOT_USER":     `0`,
	"NON_ROOT_USER": `^(\d([^0]+)|null)$`,
	"ERROR":         "Unknown out expression set",
}

func getOutRegExp(key string) string {
	if val, ok := OutRegExp[key]; ok {
		return val
	}
	return OutRegExp["ERROR"]
}

//BaseTestCaseConfigSpec slcie of test configurations template
type BaseTestCaseConfigSpec struct {
	TestCase []BaseTestCase `yaml:"testcase" json:"testcase"`
}

//BaseTestCase spec of available test template
type BaseTestCase struct {
	Name            string   `yaml:"name" json:"name"`
	SkipTest        bool     `yaml:"skiptest" json:"skiptest"`
	Command         string   `yaml:"command" json:"command"`
	ExptectedStatus []string `yaml:"expectedstatus" json:"expectedstatus"`
	ResultType      string   `default:"string" yaml:"resulttype" json:"resulttype"`
	Action          string   `yaml:"action" json:"action"`
}

//LoadTestCaseSpecs loads base test template data into a struct
func LoadTestCaseSpecs(name string) (*BaseTestCaseConfigSpec, error) {
	var testCaseConfigSpec BaseTestCaseConfigSpec
	err := json.Unmarshal([]byte(TestTemplateDataMap[name]), &testCaseConfigSpec)
	if err != nil {
		return nil, err
	}
	return &testCaseConfigSpec, nil
}

//GetOutRegExp check and get available regular expression to parse
func GetOutRegExp(key string) string {
	if val, ok := OutRegExp[key]; ok {
		return val
	}
	return OutRegExp["ERROR"]
}
