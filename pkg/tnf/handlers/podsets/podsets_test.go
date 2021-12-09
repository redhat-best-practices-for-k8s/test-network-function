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

package podsets_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	ps "github.com/test-network-function/test-network-function/pkg/tnf/handlers/podsets"
)

func Test_NewPodSets(t *testing.T) {
	newDp := ps.NewPodSets(testTimeoutDuration, testNamespace, resourceType)
	assert.NotNil(t, newDp)
	assert.Equal(t, testTimeoutDuration, newDp.Timeout())
	assert.Equal(t, newDp.Result(), tnf.ERROR)
	assert.NotNil(t, newDp.GetPodSets())
}

func Test_ReelFirstPositive(t *testing.T) {
	newDp := ps.NewPodSets(testTimeoutDuration, testNamespace, resourceType)
	assert.NotNil(t, newDp)
	firstStep := newDp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newDp := ps.NewPodSets(testTimeoutDuration, testNamespace, resourceType)
	assert.NotNil(t, newDp)
	firstStep := newDp.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newDp := ps.NewPodSets(testTimeoutDuration, testNamespace, resourceType)
	assert.NotNil(t, newDp)
	step := newDp.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newDp.Result())
	assert.Len(t, newDp.GetPodSets(), testInputSuccessNumLines)

	expectedDeployments := ps.PodSetMap{
		"testNamespace:cdi-apiserver":                   {1, 1, 1, 1, 0, 1},
		"testNamespace:hyperconverged-cluster-operator": {1, 0, 1, 0, 1, 1},
		"testNamespace:virt-api":                        {2, 2, 2, 2, 0, 1},
		"testNamespace:vm-import-operator":              {0, 0, 0, 0, 0, 1},
	}
	deployments := newDp.GetPodSets()

	for name, expected := range expectedDeployments {
		deployment, ok := deployments[name]
		assert.True(t, ok)
		assert.Equal(t, expected, deployment)
	}
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newDp := ps.NewPodSets(testTimeoutDuration, testNamespace, resourceType)
	assert.NotNil(t, newDp)
	newDp.ReelEOF()
}

const (
	resourceType             = "deployment"
	testTimeoutDuration      = time.Second * 2
	testNamespace            = "testNamespace"
	testInputError           = ""
	testInputSuccessNumLines = 17
	testInputSuccess         = `NAME                                 REPLICAS   READY    UPDATED   AVAILABLE   UNAVAILABLE  CURRENT 
	cdi-apiserver                        1          1        1         1           <none>           1
	cdi-deployment                       1          1        1         1           <none>           1
	cdi-operator                         1          1        1         1           <none>           1
	cdi-uploadproxy                      1          1        1         1           <none>           1
	cluster-network-addons-operator      1          1        1         1           <none>           1
	hostpath-provisioner-operator        1          1        1         1           <none>           1
	hyperconverged-cluster-operator      1          <none>   1         <none>      1                1
	kubemacpool-mac-controller-manager   1          1        1         1           <none>           1
	kubevirt-ssp-operator                1          <none>   1         <none>      1                1
	nmstate-webhook                      2          2        2         2           <none>           1
	node-maintenance-operator            1          <none>   1         <none>      1                1
	v2v-vmware                           1          1        1         1           <none>           1
	virt-api                             2          2        2         2           <none>           1
	virt-controller                      2          2        2         2           <none>           1
	virt-operator                        2          2        2         2           <none>           1
	virt-template-validator              2          2        2         2           <none>           1
	vm-import-operator                   0          <none>   <none>    <none>      <none>           1`
)
