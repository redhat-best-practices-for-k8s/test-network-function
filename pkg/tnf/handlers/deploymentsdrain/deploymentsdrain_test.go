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

package deploymentsdrain_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	dd "github.com/test-network-function/test-network-function/pkg/tnf/handlers/deploymentsdrain"
)

func Test_NewDeploymentsDrain(t *testing.T) {
	newDd := dd.NewDeploymentsDrain(testTimeoutDuration, testNode)
	assert.NotNil(t, newDd)
	assert.Equal(t, testTimeoutDuration, newDd.Timeout())
	assert.Equal(t, newDd.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newDd := dd.NewDeploymentsDrain(testTimeoutDuration, testNode)
	assert.NotNil(t, newDd)
	firstStep := newDd.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputSuccess)
	assert.Len(t, matches, 1)
}

func Test_ReelFirstNegative(t *testing.T) {
	newDd := dd.NewDeploymentsDrain(testTimeoutDuration, testNode)
	assert.NotNil(t, newDd)
	firstStep := newDd.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testInputError)
	assert.Len(t, matches, 0)
}

func Test_ReelMatchSuccess(t *testing.T) {
	newDd := dd.NewDeploymentsDrain(testTimeoutDuration, testNode)
	assert.NotNil(t, newDd)
	step := newDd.ReelMatch("", "", "")
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newDd.Result())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newDd := dd.NewDeploymentsDrain(testTimeoutDuration, testNode)
	assert.NotNil(t, newDd)
	newDd.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testNode            = "testNode"
	testInputError      = `node/worker-0-0 cordoned
	pod/cert-manager-webhook-5f57f59fbc-fp4j4 deleted
	pod/cdi-apiserver-58ccb6859d-kq4jj deleted
	pod/cdi-operator-5d865dbf9b-q5sfn deleted
	pod/hco-webhook-645d67694f-cx4lz deleted
	pod/hco-operator-845485f57d-csh9h deleted
	pod/hostpath-provisioner-operator-9df4ccdb-ddvbp deleted
	pod/kubevirt-ssp-operator-6594456c66-4m6wb deleted
	pod/virt-api-7d9f6484c8-sjb7z deleted
	pod/virt-api-7d9f6484c8-zh5wb deleted
	pod/virt-controller-cc5d899b8-v72fv deleted
	pod/virt-controller-cc5d899b8-wczxc deleted
	pod/virt-operator-5f47db64cc-97nlq deleted
	pod/virt-operator-5f47db64cc-n7ztz deleted
	pod/virt-template-validator-6866fc5854-g9xp5 deleted
	pod/virt-template-validator-6866fc5854-pbgtz deleted
	pod/vm-import-controller-75b64fcc85-64dkh deleted
	pod/vm-import-operator-db4787bb5-7clwz deleted
	pod/image-registry-bcb6f5f56-mkg49 deleted
	pod/migrator-598bf8d49b-9zqct deleted
	pod/grafana-74657cc99d-97fzd deleted
	pod/kube-state-metrics-6f9495ff98-gwxdv deleted
	pod/openshift-state-metrics-7466679bfb-s9ggg deleted
	pod/prometheus-adapter-76cc597d7d-4q686 deleted
	pod/prometheus-adapter-76cc597d7d-8kmzh deleted
	pod/telemeter-client-6dc8858fc-25w96 deleted
	pod/thanos-querier-5b5c97b84f-4ggt5 deleted
	pod/thanos-querier-5b5c97b84f-f9tt2 deleted
	pod/network-check-source-85c57b85d7-xhqpf deleted
	pod/cert-manager-5597cff495-9bjrb deleted
	pod/cert-manager-cainjector-bd5f9c764-87j2v deleted
	pod/cdi-deployment-84d9d6bcb8-7974j deleted
	pod/dep1-7959744869-jtt7j deleted
	pod/dep1-7959744869-ths9l deleted
	pod/dep2-75fdb9b54b-t75hj deleted
	pod/cdi-uploadproxy-67f696bcb9-6wq69 deleted
	pod/router-default-57cbd67577-pvdnq deleted
	pod/dep1-7959744869-2m787 deleted
	pod/dep2-75fdb9b54b-kclpx deleted
	pod/dep2-75fdb9b54b-v8jng deleted
	There are pending pods in node "worker-0-0" when an error occurred: timed out waiting for the condition
	pod/cluster-network-addons-operator-57bddf6b86-qkq4m
	pod/hco-webhook-5ccd87ff7c-kzwnp
	error: unable to drain node "worker-0-0", aborting command...
	
	There are pending nodes to be drained:
	 worker-0-0
	error: timed out waiting for the condition
	`
	testInputSuccess = `node/worker-0-0 cordoned
	pod/cert-manager-webhook-5f57f59fbc-fp4j4 deleted
	pod/cdi-apiserver-58ccb6859d-kq4jj deleted
	pod/cdi-operator-5d865dbf9b-q5sfn deleted
	pod/hco-webhook-645d67694f-cx4lz deleted
	pod/hco-operator-845485f57d-csh9h deleted
	pod/hostpath-provisioner-operator-9df4ccdb-ddvbp deleted
	pod/kubevirt-ssp-operator-6594456c66-4m6wb deleted
	pod/virt-api-7d9f6484c8-sjb7z deleted
	pod/virt-api-7d9f6484c8-zh5wb deleted
	pod/virt-controller-cc5d899b8-v72fv deleted
	pod/virt-controller-cc5d899b8-wczxc deleted
	pod/virt-operator-5f47db64cc-97nlq deleted
	pod/virt-operator-5f47db64cc-n7ztz deleted
	pod/virt-template-validator-6866fc5854-g9xp5 deleted
	pod/virt-template-validator-6866fc5854-pbgtz deleted
	pod/vm-import-controller-75b64fcc85-64dkh deleted
	pod/vm-import-operator-db4787bb5-7clwz deleted
	pod/image-registry-bcb6f5f56-mkg49 deleted
	pod/migrator-598bf8d49b-9zqct deleted
	pod/grafana-74657cc99d-97fzd deleted
	pod/kube-state-metrics-6f9495ff98-gwxdv deleted
	pod/openshift-state-metrics-7466679bfb-s9ggg deleted
	pod/prometheus-adapter-76cc597d7d-4q686 deleted
	pod/prometheus-adapter-76cc597d7d-8kmzh deleted
	pod/telemeter-client-6dc8858fc-25w96 deleted
	pod/thanos-querier-5b5c97b84f-4ggt5 deleted
	pod/thanos-querier-5b5c97b84f-f9tt2 deleted
	pod/network-check-source-85c57b85d7-xhqpf deleted
	pod/cert-manager-5597cff495-9bjrb deleted
	pod/cert-manager-cainjector-bd5f9c764-87j2v deleted
	pod/cdi-deployment-84d9d6bcb8-7974j deleted
	pod/dep1-7959744869-jtt7j deleted
	pod/dep1-7959744869-ths9l deleted
	pod/dep2-75fdb9b54b-t75hj deleted
	pod/cdi-uploadproxy-67f696bcb9-6wq69 deleted
	pod/router-default-57cbd67577-pvdnq deleted
	pod/dep1-7959744869-2m787 deleted
	pod/dep2-75fdb9b54b-kclpx deleted
	pod/dep2-75fdb9b54b-v8jng deleted
	There are pending pods in node "worker-0-0" when an error occurred: timed out waiting for the condition
	pod/cluster-network-addons-operator-57bddf6b86-qkq4m
	pod/hco-webhook-5ccd87ff7c-kzwnp
	error: unable to drain node "worker-0-0", aborting command...
	
	There are pending nodes to be drained:
	 worker-0-0
	error: timed out waiting for the condition
	SUCCESS
	`
)
