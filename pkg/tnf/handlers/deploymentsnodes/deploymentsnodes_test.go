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

package deploymentsnodes_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	dn "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsnodes"
)

func Test_NewDeployments(t *testing.T) {
	newDn := dn.NewDeploymentsNodes(testTimeoutDuration, testNamespace, "")
	assert.NotNil(t, newDn)
	assert.Equal(t, testTimeoutDuration, newDn.Timeout())
	assert.Equal(t, newDn.Result(), tnf.ERROR)
	assert.NotNil(t, newDn.GetNodes())
}

func Test_ReelFirstPositive(t *testing.T) {
	newDn := dn.NewDeploymentsNodes(testTimeoutDuration, testNamespace, "")
	assert.NotNil(t, newDn)
	firstStep := newDn.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
	assert.Equal(t, testInputSuccess, matches[0])
}

func Test_ReelFirstNegative(t *testing.T) {
	newDn := dn.NewDeploymentsNodes(testTimeoutDuration, testNamespace, "")
	assert.NotNil(t, newDn)
	firstStep := newDn.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newDn := dn.NewDeploymentsNodes(testTimeoutDuration, testNamespace, "")
	assert.NotNil(t, newDn)
	step := newDn.ReelMatch("", "", testInputSuccess)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newDn.Result())
	assert.Len(t, newDn.GetNodes(), testInputSuccessNumNodes)

	nodes := newDn.GetNodes()

	for name, expected := range expectedNodes {
		node, ok := nodes[name]
		assert.True(t, ok)
		assert.Equal(t, expected, node)
	}
}

// Just ensure there are no panics.
func Test_ReelEOF(t *testing.T) {
	newDn := dn.NewDeploymentsNodes(testTimeoutDuration, testNamespace, "")
	assert.NotNil(t, newDn)
	newDn.ReelEOF()
}

const (
	testTimeoutDuration      = time.Second * 2
	testNamespace            = "testNamespace"
	testInputError           = ""
	testInputSuccessNumNodes = 2
	testInputSuccess         = `NAME                                                  NODE
								cdi-apiserver-7b7bdc4745-8mz2m                        crc-mk4pg-master-0
								cdi-deployment-5f68bcf958-qm4pf                       crc-mk4pg-master-0
								cdi-operator-5fcd9d7cb6-8h9bj                         crc-mk4pg-master-0
								cdi-uploadproxy-869d4956c9-2x4dx                      crc-mk4pg-master-0
								cluster-network-addons-operator-7d9bb9b998-66kh9      crc-mk4pg-master-0
								hostpath-provisioner-operator-86d9d964b4-zrcc6        crc-mk4pg-master-0
								hyperconverged-cluster-operator-75bcff9b79-jbh6x      crc-mk4pg-master-0
								kubemacpool-mac-controller-manager-5699f48684-4ntmx   crc-mk4pg-master-0
								kubevirt-ssp-operator-864775bd4b-x7dgb                crc-mk4pg-master-0
								nmstate-webhook-5bc9777476-n7sqj                      crc-mk4pg-master-0
								nmstate-webhook-5bc9777476-vb2jg                      crc-mk4pg-master-1
								node-maintenance-operator-6cbcd9ccd-hnd8n             crc-mk4pg-master-1
								v2v-vmware-57c468498-srmst                            crc-mk4pg-master-1
								virt-api-65c85544b-7pmkj                              crc-mk4pg-master-1
								virt-api-65c85544b-w5hvm                              crc-mk4pg-master-1
								virt-controller-6d85889844-99qhn                      crc-mk4pg-master-1
								virt-controller-6d85889844-knfvm                      crc-mk4pg-master-1
								virt-operator-799985df85-g5sgj                        crc-mk4pg-master-1
								virt-operator-799985df85-lhk49                        crc-mk4pg-master-1
								virt-template-validator-76db69664c-4znj9              crc-mk4pg-master-1
								virt-template-validator-76db69664c-grzxf              crc-mk4pg-master-1
	`
)

var (
	expectedNodes = dn.NodesMap{
		"crc-mk4pg-master-1": {
			"nmstate-webhook":           true,
			"node-maintenance-operator": true,
			"v2v-vmware":                true,
			"virt-api":                  true,
			"virt-controller":           true,
			"virt-operator":             true,
			"virt-template-validator":   true,
		},
		"crc-mk4pg-master-0": {
			"cdi-apiserver":                      true,
			"cdi-deployment":                     true,
			"cdi-operator":                       true,
			"cdi-uploadproxy":                    true,
			"cluster-network-addons-operator":    true,
			"hostpath-provisioner-operator":      true,
			"hyperconverged-cluster-operator":    true,
			"kubemacpool-mac-controller-manager": true,
			"kubevirt-ssp-operator":              true,
			"nmstate-webhook":                    true,
		},
	}
)
