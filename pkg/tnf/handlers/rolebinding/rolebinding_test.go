// Copyright (C) 2021 Red Hat, Inc.
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

package rolebinding_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	rb "github.com/test-network-function/test-network-function/pkg/tnf/handlers/rolebinding"
)

func Test_NewRoleBinding(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	assert.Equal(t, testTimeoutDuration, newRb.Timeout())
	assert.Equal(t, newRb.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	firstStep := newRb.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstPositiveEmpty(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	firstStep := newRb.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccessEmpty)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccessEmpty, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	firstStep := newRb.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	step := newRb.ReelMatch("", "", testInputSuccess, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newRb.Result())
	assert.Len(t, newRb.GetRoleBindings(), 0)
}

func Test_ReelMatchSuccessEmpty(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	step := newRb.ReelMatch("", "", testInputSuccessEmpty, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newRb.Result())
	assert.Len(t, newRb.GetRoleBindings(), 0)
}

func Test_ReelMatchFail(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	step := newRb.ReelMatch("", "", testInputFail, 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, newRb.Result())
	assert.Len(t, newRb.GetRoleBindings(), 2)
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newRb := rb.NewRoleBinding(testTimeoutDuration, testServiceAccount, testPodNamespace)
	assert.NotNil(t, newRb)
	newRb.ReelEOF()
}

const (
	testTimeoutDuration   = time.Second * 2
	testPodNamespace      = "testPodNamespace"
	testServiceAccount    = "testServiceAccount"
	testInputError        = ""
	testInputSuccessEmpty = "NAMESPACE\tNAME\tSERVICE_ACCOUNTS"
	testInputSuccess      = `NAMESPACE	 NAME	 SERVICE_ACCOUNTS
	testPodNamespace                                      test-deployer                                                    map[kind:ServiceAccount name:testServiceAccount namespace:testPodNamespace]	
	`
	testInputFail = `NAMESPACE	 NAME	 SERVICE_ACCOUNTS
	default                                            test-builder                                                     map[kind:ServiceAccount name:testServiceAccount namespace:testPodNamespace]
	default                                            test-default                                                     map[kind:ServiceAccount name:testServiceAccount namespace:testPodNamespace]
	testPodNamespace                                      test-deployer                                                    map[kind:ServiceAccount name:testServiceAccount namespace:testPodNamespace]
	`
)
