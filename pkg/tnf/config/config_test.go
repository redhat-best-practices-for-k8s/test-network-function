package config_test

import (
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/config"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/testcases"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

var (
	file       *os.File
	jsonFile   *os.File
	err        error
	test       config.TnfContainerOperatorTestConfig
	configPath *string
)

const (
	// FullConfig represents full configuration, including Operator and CNF
	FullConfig = "full_config"
	// CnfConfig represents CNF configuration only
	CnfConfig = "cnf_only_config"
	// OperatorConfig represents operators configuration only
	OperatorConfig = "operator_only_config"
)

func loadCnfConfig() {
	// CNF only
	test.CNFs = []config.Cnf{{Name: "cnf_only",
		Namespace: "test",
		Status:    "",
		Tests:     []string{testcases.PrivilegedPod},
	}}
	test.CnfAvailableTestCases = nil
	for key := range testcases.CnfTestTemplateFileMap {
		test.CnfAvailableTestCases = append(test.CnfAvailableTestCases, key)
	}
}

func loadOperatorConfig() {
	operator := config.Operator{}
	operator.Name = "etcdoperator.v0.9.4"
	operator.Namespace = "my-etcd"
	operator.Status = "Succeeded"
	operator.AutoGenerate = "false"
	crd := config.Crd{}
	crd.Name = "test.crd.one"
	crd.Namespace = "default"
	instance := config.Instance{}
	instance.Name = "Instance_one"
	crd.Instances = append(crd.Instances, instance)
	operator.CRDs = append(operator.CRDs, crd)
	crd2 := config.Crd{}
	crd2.Name = "test.crd.two"
	crd2.Namespace = "default"
	instance2 := config.Instance{}
	instance2.Name = "Instance_two"
	crd2.Instances = append(crd2.Instances, instance2)
	operator.CRDs = append(operator.CRDs, crd2)
	dep := config.Deployment{}
	dep.Name = "deployment1"
	dep.Replicas = "1"
	operator.Deployments = append(operator.Deployments, dep)
	cnf := config.Cnf{}
	cnf.Name = "cnf_one"
	cnf.Namespace = "test"
	cnf.Tests = []string{testcases.PrivilegedPod}
	permission := config.Permission{}
	permission.Name = "Cluster_wide_permission"
	permission.Role = "ClusterRole"
	operator.Permissions = append(operator.Permissions, permission)

	operator.CNFs = append(operator.CNFs, cnf)
	operator.Tests = []string{testcases.OperatorStatus}
	test.Operator = append(test.Operator, operator)
	// CNF only
	test.CNFs = []config.Cnf{{
		Name:      "cnf_only",
		Namespace: "test",
		Status:    "",
		Tests:     []string{testcases.PrivilegedPod},
	}}
	test.CnfAvailableTestCases = nil
	for key := range testcases.CnfTestTemplateFileMap {
		test.CnfAvailableTestCases = append(test.CnfAvailableTestCases, key)
	}
}

func loadFullConfig() {
	loadOperatorConfig()
	loadCnfConfig()
}

func setup(configType string) {
	file, err = ioutil.TempFile(".", "test-config.yml")
	if err != nil {
		log.Fatal(err)
	}
	test = config.TnfContainerOperatorTestConfig{}
	switch configType {
	case "full_config":
		loadFullConfig()
	case "cnf_only_config":
		loadCnfConfig()
	case "operator_only_config":
		loadOperatorConfig()
	case "empty":
	}
	err = test.SaveConfig(file.Name())
	if err != nil {
		log.Fatal(err)
	}
}

func setupJSON(configType string) {
	jsonFile, err = ioutil.TempFile(".", "test-json-config.yml")
	if err != nil {
		log.Fatal(err)
	}
	test = config.TnfContainerOperatorTestConfig{}
	switch configType {
	case "full_config":
		loadFullConfig()
	case "cnf_only_config":
		loadCnfConfig()
	case "operator_only_config":
		loadOperatorConfig()
	}
	err = test.SaveConfigAsJSON(jsonFile.Name())
	if err != nil {
		log.Fatal(err)
	}
}

func teardown() {
	if file != nil {
		os.Remove(file.Name())
	}
	if jsonFile != nil {
		os.Remove(jsonFile.Name())
	}
}

func TestFullConfigLoad(t *testing.T) {
	setup(FullConfig)
	defer (teardown)()
	cfg, err := config.NewConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Operator), 1)
	assert.Equal(t, cfg.CNFs[0].Name, "cnf_only")
	assert.Nil(t, err)
}

func TestCnfConfigLoad(t *testing.T) {
	setup(CnfConfig)
	defer (teardown)()
	cfg, err := config.NewConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, cfg.CNFs[0].Name, "cnf_only")
	assert.Nil(t, err)
}
func TestOperatorConfigLoad(t *testing.T) {
	setup(OperatorConfig)
	defer (teardown)()
	cfg, err := config.NewConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Operator), 1)
	assert.Nil(t, err)
}

func TestValidateConfigPath(t *testing.T) {
	var tests = []struct {
		path  string
		error error
	}{
		{".", fmt.Errorf("'.' is a directory, not a normal file")},
		{"./config", fmt.Errorf("./config is a directory, not a normal file")},
		{"./config.yml", nil},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := config.ValidateConfigPath(tt.path)
			if err == nil && tt.error != nil {
				assert.Fail(t, err.Error())
			}

		})
	}
}

// TestConfigLoadFunction ... Test if config is read correctly
func TestConfigLoadFunction(t *testing.T) {
	setup(FullConfig)
	defer (teardown)()
	path, _ := os.Getwd()
	var tests = []struct {
		args  []string
		conf  config.TnfContainerOperatorTestConfig
		error string
	}{
		{[]string{"./operator-test"}, test, "no such file or directory"},
		{[]string{"./operator-test", "-config", "config_not_exists"}, test, "no such file or directory"},
		{[]string{"./operator-test", "-config", path}, test, "is a directory, not a normal file"},
		{[]string{"./operator-test", "-config", file.Name()}, test, ""},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			os.Args = tt.args
			if len(tt.args) > 2 {
				configPath = &tt.args[2]
			}
			testConfig, err := config.GetConfig()
			if err == nil {
				assert.Nil(t, err)
				assert.NotNil(t, testConfig)
				assert.Equal(t, *testConfig, tt.conf)
			} else {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.error)
			}
		})
	}
}

func TestFullJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(FullConfig)
	jsonCfg, err := config.NewConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(FullConfig)
	yamlCfg, err := config.NewConfig(file.Name())
	assert.Nil(t, err)
	assert.NotNil(t, yamlCfg)
	assert.Equal(t, yamlCfg.Operator, jsonCfg.Operator)
	assert.Equal(t, yamlCfg.CNFs, jsonCfg.CNFs)

}

func TestCnfJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(CnfConfig)
	jsonCfg, err := config.NewConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(CnfConfig)
	yamlCfg, err := config.NewConfig(file.Name())
	assert.Nil(t, err)
	assert.NotNil(t, yamlCfg)
	assert.Equal(t, yamlCfg.CNFs, jsonCfg.CNFs)
}

func TestOperatorJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(OperatorConfig)
	jsonCfg, err := config.NewConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(OperatorConfig)
	yamlCfg, err := config.NewConfig(file.Name())
	assert.Nil(t, err)
	assert.Equal(t, yamlCfg.Operator, jsonCfg.Operator)
}

func TestNewConfig(t *testing.T) {
	Cfg, err := config.NewConfig("error.file")
	assert.NotNil(t, err)
	assert.Nil(t, Cfg)
	defer (teardown)()
	file, err = ioutil.TempFile(".", "test-empty.yml")
	assert.Nil(t, err)
	_, err = config.NewConfig(file.Name())
	assert.NotNil(t, err)
}
