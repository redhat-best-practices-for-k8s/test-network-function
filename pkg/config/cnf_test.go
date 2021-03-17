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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	tnfConfig "github.com/test-network-function/test-network-function/pkg/config"
	"github.com/test-network-function/test-network-function/pkg/tnf/testcases"
)

var (
	file       *os.File
	jsonFile   *os.File
	err        error
	test       tnfConfig.TnfContainerOperatorTestConfig
	configPath *string
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
	// operatorStatus specifies status of the CSV in the cluster
	operatorStatus = "Succeeded"
	// organization is under which operator is published
	organization = "redhat-marketplace"
	// permissionName type of permission
	permissionName = "Cluster-wide-permission"
	// permissionRole type of role
	permissionRole = "ClusterRole"
	// imageRepository for  container images
	imageRepository = "rhel8"
	// testNameSpace k8s namespace
	testNameSpace = "default"
)

func loadCnfConfig() {
	// CNF only
	test.CNFs = []tnfConfig.Cnf{
		{
			Name:      cnfName,
			Namespace: testNameSpace,
			Status:    "",
			Tests:     []string{testcases.PrivilegedPod},
			CertifiedContainerRequestInfos: []tnfConfig.CertifiedContainerRequestInfo{
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
	operator := tnfConfig.Operator{}
	operator.Name = operatorName
	operator.Namespace = operatorNameSpace
	operator.Status = operatorStatus
	setCrdsAndInstances(&operator)
	dep := tnfConfig.Deployment{}
	dep.Name = deploymentName
	dep.Replicas = deploymentReplicas
	operator.Deployments = append(operator.Deployments, dep)
	setCnfAndPermissions(&operator)
	operator.Tests = []string{testcases.OperatorStatus}
	operator.CertifiedOperatorRequestInfos = []tnfConfig.CertifiedOperatorRequestInfo{
		{
			Name:         operatorPackageName,
			Organization: organization,
		},
	}
	test.Operator = append(test.Operator, operator)
	// CNF only
	loadCnfConfig()
}

func setCrdsAndInstances(op *tnfConfig.Operator) {
	crd := tnfConfig.Crd{}
	crd.Name = crdNameOne
	crd.Namespace = testNameSpace
	instance := tnfConfig.Instance{}
	instance.Name = instanceNameOne
	crd.Instances = append(crd.Instances, instance)
	op.CRDs = append(op.CRDs, crd)
	crd2 := tnfConfig.Crd{}
	crd2.Name = crdNameTwo
	crd2.Namespace = testNameSpace
	instance2 := tnfConfig.Instance{}
	instance2.Name = instanceNameTwo
	crd2.Instances = append(crd2.Instances, instance2)
	op.CRDs = append(op.CRDs, crd2)
}

func setCnfAndPermissions(op *tnfConfig.Operator) {
	cnf := tnfConfig.Cnf{}
	cnf.Name = cnfName
	cnf.Namespace = testNameSpace
	cnf.Tests = []string{testcases.PrivilegedPod}
	permission := tnfConfig.Permission{}
	permission.Name = permissionName
	permission.Role = permissionRole
	op.Permissions = append(op.Permissions, permission)
	op.CNFs = append(op.CNFs, cnf)
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
	test = tnfConfig.TnfContainerOperatorTestConfig{}
	switch configType {
	case fullConfig:
		loadFullConfig()
	case cnfConfig:
		loadCnfConfig()
	case operatorConfig:
		loadOperatorConfig()
	}
	err = test.SaveConfig(file.Name())
	if err != nil {
		log.Fatal(err)
	}
}

func setupJSON(configType string) {
	jsonFile, err = ioutil.TempFile(".", "test-json-config.json")
	if err != nil {
		log.Fatal(err)
	}
	test = tnfConfig.TnfContainerOperatorTestConfig{}
	switch configType {
	case fullConfig:
		loadFullConfig()
	case cnfConfig:
		loadCnfConfig()
	case operatorConfig:
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
	setup(fullConfig)
	defer (teardown)()
	cfg, err := tnfConfig.NewConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, len(cfg.Operator), 1)
	assert.Equal(t, cfg.CNFs[0].Name, cnfName)
	assert.Nil(t, err)
}

func TestCnfConfigLoad(t *testing.T) {
	setup(cnfConfig)
	defer (teardown)()
	cfg, err := tnfConfig.NewConfig(file.Name())
	assert.NotNil(t, cfg)
	assert.Equal(t, cfg.CNFs[0].Name, cnfName)
	assert.Nil(t, err)
}

func TestOperatorConfigLoad(t *testing.T) {
	setup(operatorConfig)
	defer (teardown)()
	cfg, err := tnfConfig.NewConfig(file.Name())
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
			tt := tt // pin
			err := tnfConfig.ValidateConfigPath(tt.path)
			if err == nil && tt.error != nil {
				assert.Fail(t, err.Error())
			}
		})
	}
}

// TestConfigLoadFunction ... Test if config is read correctly
func TestConfigLoadFunction(t *testing.T) {
	setup(fullConfig)
	defer (teardown)()
	path, _ := os.Getwd()
	var tests = []struct {
		args  []string
		conf  tnfConfig.TnfContainerOperatorTestConfig
		error string
	}{
		{[]string{"./operator-test"}, test, "no such file or directory"},
		{[]string{"./operator-test", "-config", "config_not_exists"}, test, "no such file or directory"},
		{[]string{"./operator-test", "-config", path}, test, "is a directory, not a normal file"},
		{[]string{"./operator-test", "-config", file.Name()}, test, ""},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			tt := tt // pin
			os.Args = tt.args
			if len(tt.args) > 2 {
				configPath = &tt.args[2]
			}
			testConfig, err := tnfConfig.GetConfig()
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
	setupJSON(fullConfig)
	jsonCfg, err := tnfConfig.NewConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(fullConfig)
	yamlCfg, err := tnfConfig.NewConfig(file.Name())
	assert.Nil(t, err)
	assert.NotNil(t, yamlCfg)
	assert.Equal(t, yamlCfg.Operator, jsonCfg.Operator)
	assert.Equal(t, yamlCfg.CNFs, jsonCfg.CNFs)
}

func TestCnfJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(cnfConfig)
	jsonCfg, err := tnfConfig.NewConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(cnfConfig)
	yamlCfg, err := tnfConfig.NewConfig(file.Name())
	assert.Nil(t, err)
	assert.NotNil(t, yamlCfg)
	assert.Equal(t, yamlCfg.CNFs, jsonCfg.CNFs)
}

func TestOperatorJsonConfig(t *testing.T) {
	defer (teardown)()
	// json
	setupJSON(operatorConfig)
	jsonCfg, err := tnfConfig.NewConfig(jsonFile.Name())
	assert.NotNil(t, jsonCfg)
	assert.Nil(t, err)
	// yaml
	setup(operatorConfig)
	yamlCfg, err := tnfConfig.NewConfig(file.Name())
	assert.Nil(t, err)
	assert.Equal(t, yamlCfg.Operator, jsonCfg.Operator)
}

func TestNewConfig(t *testing.T) {
	Cfg, err := tnfConfig.NewConfig("error.file")
	assert.NotNil(t, err)
	assert.Nil(t, Cfg)
	defer (teardown)()
	file, err = ioutil.TempFile(".", "test-empty.yml")
	assert.Nil(t, err)
	_, err = tnfConfig.NewConfig(file.Name())
	assert.NotNil(t, err)
}
