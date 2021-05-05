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

package config_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/config"
	"gopkg.in/yaml.v2"
)

const (
	tmpfileNameBase = "test-request-config.yml"
)

// marshalFunc and unmarshalfunc define the signature for marshal and unmarshal methods, respectively.
type marshalFunc func(interface{}) ([]byte, error)
type unmarshalFunc func([]byte, interface{}) error

// test data
const (
	// bananas are in the fruit bowl
	containerImageNameBanana = "banana"
	imageRepositoryFruitBowl = "fruitbowl"

	// apples are in the fridge
	containerImageNameApple = "apple"
	imageRepositoryFridge   = "fridge"
)

var (
	fruitbowlRequestInfo = config.CertifiedContainerRequestInfo{
		Name:       containerImageNameBanana,
		Repository: imageRepositoryFruitBowl,
	}
	fridgeRequestInfo = config.CertifiedContainerRequestInfo{
		Name:       containerImageNameApple,
		Repository: imageRepositoryFridge,
	}
)

var (
	// tempFiles stores file pointers for closing in the case of a test failure.
	tempFiles []*os.File
)

// setupRequestTest writes the result of `populateRequestConfig` to a temporary file for loading in a test
func setupRequestTest(marshalFun marshalFunc) (tempfileName string) {
	tempfile, err := ioutil.TempFile(".", tmpfileNameBase)
	if err != nil {
		log.Fatal(err)
	}
	requestConfig := buildRequestConfig()
	saveRequestConfig(marshalFun, requestConfig, tempfile.Name())
	tempFiles = append(tempFiles, tempfile)
	return tempfile.Name()
}

// loadRequestConfig reads `tmpPath`, unmarshals it using `unmarshalFun`, and returns the resulting `config.File`
func loadRequestConfig(tmpPath string, unmarshalFun unmarshalFunc) (conf *config.File) {
	contents, err := ioutil.ReadFile(tmpPath)
	if err != nil {
		log.Fatal(err)
	}

	conf = &config.File{}
	err = unmarshalFun(contents, conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}

// saveRequestConfig calls `marshalFun` on `c`, then writes the result to `configPath`
func saveRequestConfig(marshalFun marshalFunc, c *config.File, configPath string) {
	bytes, err := marshalFun(c)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(configPath, bytes, filePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanupTempfiles() {
	for _, f := range tempFiles {
		os.Remove(f.Name())
	}
	tempFiles = make([]*os.File, 0)
}

func buildRequestConfig() *config.File {
	conf := &config.File{}
	conf.CertifiedContainerInfo = []config.CertifiedContainerRequestInfo{
		fruitbowlRequestInfo,
		fridgeRequestInfo,
	}
	return conf
}

func RequestTest(t *testing.T, marshalFun marshalFunc, unmarshalFun unmarshalFunc) {
	defer (cleanupTempfiles)()
	cfg := loadRequestConfig(setupRequestTest(marshalFun), unmarshalFun)
	assert.Equal(t, len(cfg.CertifiedContainerInfo), 2)
	assert.Equal(t, cfg.CertifiedContainerInfo[0], fruitbowlRequestInfo)
	assert.Equal(t, cfg.CertifiedContainerInfo[1], fridgeRequestInfo)
}

func TestRequestInfos(t *testing.T) {
	RequestTest(t, yaml.Marshal, yaml.Unmarshal)
	RequestTest(t, json.Marshal, json.Unmarshal)
}
