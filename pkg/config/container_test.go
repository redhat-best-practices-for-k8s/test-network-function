// Copyright (C) 2020 Red Hat, Inc.
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

package config_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
	"gopkg.in/yaml.v2"
)

var (
	file     *os.File
	jsonFile *os.File
	err      error
	test     config.File
)

const (
	// cnfConfig represents CNF configuration only
	cnfConfig = "cnf_only_config"
	// cnfName name of the cnf
	cnfName = "cnf-test-one"
	// containerImageName name of the container image name
	containerImageName = "rhel8/nginx-116"
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
	// operatorPackageName operator package name in the bundle
	operatorPackageName = "amq-streams"
	// organization is under which operator is published
	organization = "redhat-marketplace"
	// imageRepository for  container images
	imageRepository = "rhel8"
	// testNameSpace k8s namespace
	testNameSpace = "default"
)

const (
	filePerm = 0644
)

func saveConfig(c *config.File, configPath string) (err error) {
	bytes, _ := yaml.Marshal(c)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	return
}

func saveConfigAsJSON(c *config.File, configPath string) (err error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	return
}

// newConfig  returns a new decoded TnfContainerOperatorTestConfig struct
func newConfig(configPath string) (*config.File, error) {
	// Create config structure
	conf := &config.File{}
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

func loadCnfConfig() {
	// CNF only
	test.CNFs = []config.Cnf{
		{
			Name:      cnfName,
			Namespace: testNameSpace,
			Tests:     []string{testcases.PrivilegedPod},
			CertifiedContainerRequestInfos: []config.CertifiedContainerRequestInfo{
				{
					Name:       containerImageName,
					Repository: imageRepository,
				},
			},
		},
	}
	test.CnfAvailableTestCases = nil
	for key := range testcases.CnfTestTemplateFileMap {
		test.CnfAvailableTestCases = append(test.CnfAvailableTestCases, key)
	}
}

func loadOperatorConfig() {
	operator := config.Operator{}
	operator.Name = operatorName
	operator.Namespace = operatorNameSpace
	setCrdsAndInstances()
	dep := config.Deployment{}
	dep.Name = deploymentName
	dep.Replicas = deploymentReplicas
	operator.Tests = []string{testcases.OperatorStatus}
	operator.CertifiedOperatorRequestInfos = []config.CertifiedOperatorRequestInfo{
		{
			Name:         operatorPackageName,
			Organization: organization,
		},
	}
	test.Operators = append(test.Operators, operator)
	// CNF only
	loadCnfConfig()
}

func setCrdsAndInstances() {
	crd := config.Crd{}
	crd.Name = crdNameOne
	crd.Namespace = testNameSpace
	instance := config.Instance{}
	instance.Name = instanceNameOne
	crd.Instances = append(crd.Instances, instance)
	crd2 := config.Crd{}
	crd2.Name = crdNameTwo
	crd2.Namespace = testNameSpace
	instance2 := config.Instance{}
	instance2.Name = instanceNameTwo
	crd2.Instances = append(crd2.Instances, instance2)
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
	test = config.File{}
	switch configType {
	case fullConfig:
		loadFullConfig()
	case cnfConfig:
		loadCnfConfig()
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
	test = config.File{}
	switch configType {
	case fullConfig:
		loadFullConfig()
	case cnfConfig:
		loadCnfConfig()
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
	assert.Equal(t, cfg.CNFs[0].Name, cnfName)
	assert.Nil(t, err)
}

func TestCnfConfigLoad(t *testing.T) {
	setup(cnfConfig)
	defer (teardown)()
	cfg, err := newConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, cfg.CNFs[0].Name, cnfName)
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
	assert.Equal(t, yamlCfg.CNFs, jsonCfg.CNFs)
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
	assert.Equal(t, yamlCfg.CNFs, jsonCfg.CNFs)
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
