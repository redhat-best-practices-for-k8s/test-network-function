package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	file, err := ioutil.TempFile("", "testconfig.yml")
	defer os.Remove(file.Name())
	if err != nil {
		fmt.Println(err)
		assert.Fail(t, "failed to parse valid test file")
	}
	cfg, err := NewConfig(file.Name())
	assert.Nil(t, cfg)
	assert.NotNil(t, err)
	cfg, err = NewConfig(file.Name() + "bad_file")
	assert.Nil(t, cfg)
	assert.NotNil(t, err)

	cfg, err = NewConfig("config.yml")
	assert.NotNil(t, cfg)
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
			err := validateConfigPath(tt.path)
			if err == nil && tt.error != nil {
				assert.Fail(t, err.Error())
			}

		})
	}
}

//TestConfigLoadFunction ... Test if config is read correctly
func TestConfigLoadFunction(t *testing.T) {
	var test Config
	test.Csv.Name = "etcdoperator.v0.9.4"
	test.Csv.Namespace = "my-etcd"
	test.Csv.Status = "Succeeded"
	path, _ := os.Getwd()

	var tests = []struct {
		args  []string
		conf  Config
		error string
	}{
		{[]string{"./operator-test"}, test, "no such file or directory"},
		{[]string{"./operator-test", "-config", "config_not_exists"}, test, "no such file or directory"},
		{[]string{"./operator-test", "-config", path}, test, "is a directory, not a normal file"},
		{[]string{"./operator-test", "-config", "config.yml"}, test, ""},
		{[]string{"./operator-test", "-config", path + "/config.yml"}, test, ""},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			os.Args = tt.args
			if len(tt.args) > 2 {
				configPath = &tt.args[2]
			}
			testConfig, err := GetConfig()
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
