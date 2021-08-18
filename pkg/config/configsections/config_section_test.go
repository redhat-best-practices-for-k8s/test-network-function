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

package configsections

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"gopkg.in/yaml.v2"
)

var (
	file     *os.File
	jsonFile *os.File
	err      error
	test     TestConfiguration
)

const (
	// cnfConfig represents CNF configuration only
	cnfConfig = "cnf_only_config"
	// cnfName name of the cnf
	cnfName = "cnf-test-one"
	// crdNameOne name of the crd
	crdNameOne = "crd-test-one"
	// crdNameTwo name of the crd
	crdNameTwo = "crd-test-two"
	// deploymentName is the name of the deployment
	deploymentName = "deployment-one"
	// deploymentReplicas no of replicas
	deploymentReplicas = "1"
	// fullConfig represents full configuration, including Operator and CNF
	fullConfig = "full_config"
	// instanceNameOne name of the instance
	instanceNameOne = "instance-one"
	// instanceNameTwo name of the instance
	instanceNameTwo = "instance-two"
	// operatorConfig represents operators configuration only
	operatorConfig = "operator_only_config"
	// operatorName name of the operator
	operatorName = "etcdoperator.v0.9.4"
	// operatorNameSpace is test namespace for an operator
	operatorNameSpace = "my-etcd"
	// testNameSpace k8s namespace
	testNameSpace = "default"
)

const (
	// filePerm is the permissions these tests will use when creating config files in test setup
	filePerm = 0644
)

func saveConfig(c *TestConfiguration, configPath string) (err error) {
	bytes, _ := yaml.Marshal(c)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	return
}

func saveConfigAsJSON(c *TestConfiguration, configPath string) (err error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	return
}

// newConfig  returns a new decoded TnfContainerOperatorTestConfig struct
func newConfig(configPath string) (*TestConfiguration, error) {
	// Create config structure
	conf := &TestConfiguration{}
	// Open config file
	if file, err = os.Open(configPath); err != nil {
		return nil, err
	}
	defer file.Close()
	// Init new YAML decode
	d := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err = d.Decode(&conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func loadPodConfig() {
	test.PodsUnderTest = []Pod{
		{
			Name:      cnfName,
			Namespace: testNameSpace,
			Tests:     []string{testcases.PrivilegedPod},
		},
	}
}

func loadOperatorConfig() {
	operator := Operator{}
	operator.Name = operatorName
	operator.Namespace = operatorNameSpace
	setCrdsAndInstances()
	dep := Deployment{}
	dep.Name = deploymentName
	dep.Replicas = deploymentReplicas
	operator.Tests = []string{testcases.OperatorStatus}
	test.Operators = append(test.Operators, operator)
	loadPodConfig()
}

func setCrdsAndInstances() {
	crd := Crd{}
	crd.Name = crdNameOne
	crd.Namespace = testNameSpace
	instance := Instance{}
	instance.Name = instanceNameOne
	crd.Instances = append(crd.Instances, instance)
	crd2 := Crd{}
	crd2.Name = crdNameTwo
	crd2.Namespace = testNameSpace
	instance2 := Instance{}
	instance2.Name = instanceNameTwo
	crd2.Instances = append(crd2.Instances, instance2)
}

func loadFullConfig() {
	loadOperatorConfig()
	loadPodConfig()
}

func setup(configType string) {
	file, err = ioutil.TempFile(".", "test-config.yml")
	if err != nil {
		log.Fatal(err)
	}
	test = TestConfiguration{}
	switch configType {
	case fullConfig:
		loadFullConfig()
	case cnfConfig:
		loadPodConfig()
	case operatorConfig:
		loadOperatorConfig()
	}
	err = saveConfig(&test, file.Name())
	if err != nil {
		log.Fatal(err)
	}
}

func setupJSON(configType string) {
	jsonFile, err = ioutil.TempFile(".", "test-json-config.json")
	if err != nil {
		log.Fatal(err)
	}
	test = TestConfiguration{}
	switch configType {
	case fullConfig:
		loadFullConfig()
	case cnfConfig:
		loadPodConfig()
	case operatorConfig:
		loadOperatorConfig()
	}
	err = saveConfigAsJSON(&test, jsonFile.Name())
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
	setup(fullConfig)
	defer (teardown)()
	cfg, err := newConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Operators), 1)
	assert.Equal(t, cfg.PodsUnderTest[0].Name, cnfName)
	assert.Nil(t, err)
}

func TestPodConfigLoad(t *testing.T) {
	setup(cnfConfig)
	defer (teardown)()
	cfg, err := newConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, cfg.PodsUnderTest[0].Name, cnfName)
	assert.Nil(t, err)
}

func TestOperatorConfigLoad(t *testing.T) {
	setup(operatorConfig)
	defer (teardown)()
	cfg, err := newConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Operators), 1)
	assert.Nil(t, err)
}

func TestFullJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(fullConfig)
	jsonCfg, err := newConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(fullConfig)
	yamlCfg, err := newConfig(file.Name())
	assert.Nil(t, err)
	assert.NotNil(t, yamlCfg)
	assert.Equal(t, yamlCfg.Operators, jsonCfg.Operators)
	assert.Equal(t, yamlCfg.PodsUnderTest, jsonCfg.PodsUnderTest)
}

func TestCnfJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(cnfConfig)
	jsonCfg, err := newConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(cnfConfig)
	yamlCfg, err := newConfig(file.Name())
	assert.Nil(t, err)
	assert.NotNil(t, yamlCfg)
	assert.Equal(t, yamlCfg.PodsUnderTest, jsonCfg.PodsUnderTest)
}

func TestOperatorJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(operatorConfig)
	jsonCfg, err := newConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(operatorConfig)
	yamlCfg, err := newConfig(file.Name())
	assert.Nil(t, err)
	assert.Equal(t, yamlCfg.Operators, jsonCfg.Operators)
}
