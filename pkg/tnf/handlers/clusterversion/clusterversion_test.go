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

package clusterversion_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	ver "github.com/test-network-function/test-network-function/pkg/tnf/handlers/clusterversion"
)

func Test_NewNodeNames(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	assert.Equal(t, testTimeoutDuration, newVer.Timeout())
	assert.Equal(t, newVer.Result(), tnf.ERROR)
}

func Test_ReelFirstPositiveOcp(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccessOcp)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccessOcp, matches[0])
}

func Test_ReelFirstPositiveMinikube(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccessMinikube)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccessMinikube, matches[0])
}

func Test_ReelFirstPositiveEmpty(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputFailure)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputFailure, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	firstStep := newVer.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccessOcp(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	step := newVer.ReelMatch("", "", testInputSuccessOcp)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newVer.Result())
	assert.Equal(t, newVer.GetVersions().Oc, "4.7.16")
	assert.Equal(t, newVer.GetVersions().Ocp, "4.8.3")
	assert.Equal(t, newVer.GetVersions().K8s, "v1.21.1+051ac4f")
}

func Test_ReelMatchSuccessMinikube(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	step := newVer.ReelMatch("", "", testInputSuccessMinikube)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newVer.Result())
	assert.Equal(t, newVer.GetVersions().Oc, "4.7.16")
	assert.Equal(t, newVer.GetVersions().Ocp, "n/a")
	assert.Equal(t, newVer.GetVersions().K8s, "v1.21.1+051ac4f")
}

func Test_ReelMatchFail(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	step := newVer.ReelMatch("", "", testInputFailure)
	assert.Nil(t, step)
	assert.Equal(t, tnf.FAILURE, newVer.Result())
	assert.Equal(t, newVer.GetVersions().Ocp, "")
	assert.Equal(t, newVer.GetVersions().Oc, "")
	assert.Equal(t, newVer.GetVersions().K8s, "")
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newVer := ver.NewClusterVersion(testTimeoutDuration)
	assert.NotNil(t, newVer)
	newVer.ReelEOF()
}

const (
	testTimeoutDuration      = time.Second * 2
	testInputError           = ""
	testInputFailure         = "NAME\n"
	testInputSuccessOcp      = "Client Version: 4.7.16\nServer Version: 4.8.3\nKubernetes Version: v1.21.1+051ac4f\n"
	testInputSuccessMinikube = "Client Version: 4.7.16\nKubernetes Version: v1.21.1+051ac4f\n"
)
