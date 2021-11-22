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

package automountservice_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	as "github.com/test-network-function/test-network-function/pkg/tnf/handlers/automountservice"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
)

const (
	// adding special variable
	testTimeoutDuration     = time.Second * 2
	testServiceAccount      = "default"
	testNamespace           = "tnf"
	testPodname             = "test"
	testServiceAccount1Yaml = `apiVersion: v1
automountServiceAccountToken: false
imagePullSecrets:
- name: default-dockercfg-wczhp
kind: ServiceAccount
metadata:
  creationTimestamp: "2021-11-03T16:56:49Z"
  name: default
  namespace: salah
  resourceVersion: "186554188"
  uid: 3dd99c2c-c6fe-4858-bf0b-bfab1af80b95
secrets:
- name: default-token-qwfxb
- name: default-dockercfg-wczhp`
	testServiceAccount2Yaml = `apiVersion: v1
automountServiceAccountToken: true`
	testServiceAccount3Yaml = `apiVersion: v1
imagePullSecrets:
- name: default-dockercfg-xqpcb
kind: ServiceAccount`
)

// Test_NewAutomountService is the unit test for NewAutomountService().
func Test_NewAutomountService(t *testing.T) {
	automount := as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithServiceAccount(testServiceAccount))
	assert.NotNil(t, automount)
	assert.Equal(t, automount.Result(), tnf.ERROR)
	// test creating with pod
	automount = as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithPodname(testPodname))
	assert.NotNil(t, automount)
	assert.Equal(t, automount.Result(), tnf.ERROR)
}

// Test_Automountservice_GetIdentifier is the unit test for Automountservice_GetIdentifier().
func TestAutomountservice_GetIdentifier(t *testing.T) {
	test := as.NewAutomountservice()
	assert.Equal(t, identifier.AutomountServiceIdentifier, test.GetIdentifier())
}

// Test_Automountservice_ReelEOF is the unit test for Automountservice_ReelEOF().
func TestAutomountservice_ReelEOF(t *testing.T) {
	test := as.NewAutomountservice()
	assert.NotNil(t, test)
	test.ReelEOF()
}

func Test_Automountservice_Args(t *testing.T) {
	test := as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithServiceAccount(testServiceAccount))
	args := []string{"oc", "-n", testNamespace, "get", "serviceaccounts", testServiceAccount, "-o", "yaml"}
	assert.ElementsMatch(t, args, test.Args())

	test = as.NewAutomountservice(as.WithNamespace(testNamespace), as.WithPodname(testPodname))
	args = []string{"oc", "-n", testNamespace, "get", "pods", testPodname, "-o", "yaml"}
	assert.ElementsMatch(t, args, test.Args())
}

// Test_Automountservice_ReelTimeout is the unit test for Automountservice}_ReelTimeout().
func TestAutomountservice_ReelTimeout(t *testing.T) {
	test := as.NewAutomountservice(as.WithTimeout(testTimeoutDuration))
	assert.NotNil(t, test)
	assert.Equal(t, testTimeoutDuration, test.Timeout())
	test.ReelTimeout()
}

// Test_Automountservice_ReelMatch is the unit test for Automountservice_ReelMatch().
func TestAutomountservice_ReelMatch(t *testing.T) {
	test := as.NewAutomountservice()
	assert.NotNil(t, test)
	firstStep := test.ReelFirst()
	// validate regular expression when serviceaccount is set to false
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testServiceAccount1Yaml)
	assert.Len(t, matches, 2)
	// validate reel match
	step := test.ReelMatch("", "", matches[0])
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, test.Result())
	assert.Equal(t, as.TokenIsFalse, test.Token())

	// validate regular expression when serviceaccount is set to true
	matches = re.FindStringSubmatch(testServiceAccount2Yaml)
	assert.Len(t, matches, 2)
	// validate reel match
	step = test.ReelMatch("", "", matches[0])
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, test.Result())
	assert.Equal(t, as.TokenIsTrue, test.Token())
	// validate regular expression when serviceaccount is not set
	test = as.NewAutomountservice()
	matches = re.FindStringSubmatch(testServiceAccount3Yaml)
	assert.Len(t, matches, 0)
	// validate reel match
	step = test.ReelMatch("", "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, test.Result())
	assert.Equal(t, as.TokenNotSet, test.Token())
}
