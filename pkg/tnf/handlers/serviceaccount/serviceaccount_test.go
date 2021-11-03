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

package serviceaccount_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	sa "github.com/test-network-function/test-network-function/pkg/tnf/handlers/serviceaccount"
)

func Test_NewServiceAccount(t *testing.T) {
	newSa := sa.NewServiceAccount(testTimeoutDuration, testPodName, testPodNamespace)
	assert.NotNil(t, newSa)
	assert.Equal(t, testTimeoutDuration, newSa.Timeout())
	assert.Equal(t, newSa.Result(), tnf.ERROR)
}

func Test_ReelFirstPositive(t *testing.T) {
	newSa := sa.NewServiceAccount(testTimeoutDuration, testPodName, testPodNamespace)
	assert.NotNil(t, newSa)
	firstStep := newSa.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testPodYaml)
	assert.Len(t, matches, 2)
	assert.Equal(t, "default", matches[1])
}

func Test_ReelFirstNegative(t *testing.T) {
	const errorInput = "not really a yaml file\njust someinput for test\nok?"
	newSa := sa.NewServiceAccount(testTimeoutDuration, testPodName, testPodNamespace)
	assert.NotNil(t, newSa)
	firstStep := newSa.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(errorInput)
	assert.Len(t, matches, 0)
}

func Test_ReelMatch(t *testing.T) {
	// Prepare input for ReelMatch
	newSa := sa.NewServiceAccount(testTimeoutDuration, testPodName, testPodNamespace)
	assert.NotNil(t, newSa)
	firstStep := newSa.ReelFirst()
	re := regexp.MustCompile(firstStep.Expect[0])
	matches := re.FindStringSubmatch(testPodYaml)
	assert.Len(t, matches, 2)
	assert.Equal(t, "default", matches[1])

	// Call ReelMatch
	step := newSa.ReelMatch("", "", matches[0], 0)
	assert.Nil(t, step)
	assert.Equal(t, tnf.SUCCESS, newSa.Result())
	assert.Equal(t, "default", newSa.GetServiceAccountName())
}

// Just ensure there are no panics.
func Test_ReelEof(t *testing.T) {
	newSa := sa.NewServiceAccount(testTimeoutDuration, testPodName, testPodNamespace)
	assert.NotNil(t, newSa)
	newSa.ReelEOF()
}

const (
	testTimeoutDuration = time.Second * 2
	testPodName         = "testPod"
	testPodNamespace    = "testPodNamespace"
	testPodYaml         = `apiVersion: v1
	kind: Pod
	metadata:
	  annotations:
		k8s.v1.cni.cncf.io/network-status: |-
		  [{
			  "name": "openshift-sdn",
			  "interface": "eth0",
			  "ips": [
				  "10.116.0.17"
			  ],
			  "default": true,
			  "dns": {}
		  }]
		k8s.v1.cni.cncf.io/networks-status: |-
		  [{
			  "name": "openshift-sdn",
			  "interface": "eth0",
			  "ips": [
				  "10.116.0.17"
			  ],
			  "default": true,
			  "dns": {}
		  }]
	  creationTimestamp: "2021-03-16T14:48:39Z"
	  labels:
		app: test
	  managedFields:
	  - apiVersion: v1
		fieldsType: FieldsV1
		fieldsV1:
		  f:metadata:
			f:labels:
			  .: {}
			  f:app: {}
		  f:spec:
			f:containers:
			  k:{"name":"test"}:
				.: {}
				f:command: {}
				f:image: {}
				f:imagePullPolicy: {}
				f:name: {}
				f:resources:
				  .: {}
				  f:limits:
					.: {}
					f:cpu: {}
					f:memory: {}
				  f:requests:
					.: {}
					f:cpu: {}
					f:memory: {}
				f:terminationMessagePath: {}
				f:terminationMessagePolicy: {}
			f:dnsPolicy: {}
			f:enableServiceLinks: {}
			f:restartPolicy: {}
			f:schedulerName: {}
			f:securityContext: {}
			f:terminationGracePeriodSeconds: {}
		manager: oc
		operation: Update
		time: "2021-03-16T14:48:39Z"
	  - apiVersion: v1
		fieldsType: FieldsV1
		fieldsV1:
		  f:metadata:
			f:annotations:
			  .: {}
			  f:k8s.v1.cni.cncf.io/network-status: {}
			  f:k8s.v1.cni.cncf.io/networks-status: {}
		manager: multus
		operation: Update
		time: "2021-04-04T10:10:10Z"
	  - apiVersion: v1
		fieldsType: FieldsV1
		fieldsV1:
		  f:status:
			f:conditions:
			  k:{"type":"ContainersReady"}:
				.: {}
				f:lastProbeTime: {}
				f:lastTransitionTime: {}
				f:status: {}
				f:type: {}
			  k:{"type":"Initialized"}:
				.: {}
				f:lastProbeTime: {}
				f:lastTransitionTime: {}
				f:status: {}
				f:type: {}
			  k:{"type":"Ready"}:
				.: {}
				f:lastProbeTime: {}
				f:lastTransitionTime: {}
				f:status: {}
				f:type: {}
			f:containerStatuses: {}
			f:hostIP: {}
			f:phase: {}
			f:podIP: {}
			f:podIPs:
			  .: {}
			  k:{"ip":"10.116.0.17"}:
				.: {}
				f:ip: {}
			f:startTime: {}
		manager: kubelet
		operation: Update
		time: "2021-04-04T10:10:23Z"
	  name: test
	  namespace: default
	  resourceVersion: "14788662"
	  selfLink: /api/v1/namespaces/default/pods/test
	  uid: f5967af2-e838-4fc9-a9f2-b7aa7fd40340
	spec:
	  containers:
	  - command:
		- tail
		- -f
		- /dev/null
		image: quay.io/testnetworkfunction/cnf-test-partner:latest
		imagePullPolicy: Always
		name: test
		resources:
		  limits:
			cpu: 250m
			memory: 512Mi
		  requests:
			cpu: 250m
			memory: 512Mi
		terminationMessagePath: /dev/termination-log
		terminationMessagePolicy: File
		volumeMounts:
		- mountPath: /var/run/secrets/kubernetes.io/serviceaccount
		  name: default-token-zhsqz
		  readOnly: true
	  dnsPolicy: ClusterFirst
	  enableServiceLinks: true
	  imagePullSecrets:
	  - name: default-dockercfg-mt7k6
	  nodeName: crc-mk4pg-master-0
	  priority: 0
	  restartPolicy: Always
	  schedulerName: default-scheduler
	  securityContext: {}
	  serviceAccount: default
	  serviceAccountName: default
	  terminationGracePeriodSeconds: 30
	  tolerations:
	  - effect: NoExecute
		key: node.kubernetes.io/not-ready
		operator: Exists
		tolerationSeconds: 300
	  - effect: NoExecute
		key: node.kubernetes.io/unreachable
		operator: Exists
		tolerationSeconds: 300
	  - effect: NoSchedule
		key: node.kubernetes.io/memory-pressure
		operator: Exists
	  volumes:
	  - name: default-token-zhsqz
		secret:
		  defaultMode: 420
		  secretName: default-token-zhsqz
	status:
	  conditions:
	  - lastProbeTime: null
		lastTransitionTime: "2021-03-16T14:48:39Z"
		status: "True"
		type: Initialized
	  - lastProbeTime: null
		lastTransitionTime: "2021-04-04T10:10:23Z"
		status: "True"
		type: Ready
	  - lastProbeTime: null
		lastTransitionTime: "2021-04-04T10:10:23Z"
		status: "True"
		type: ContainersReady
	  - lastProbeTime: null
		lastTransitionTime: "2021-03-16T14:48:39Z"
		status: "True"
		type: PodScheduled
	  containerStatuses:
	  - containerID: cri-o://1c8463f7b2cac4aeb5d60048f226cc76c7eb5d7ba1428dcbcdbfbd748eac143d
		image: quay.io/testnetworkfunction/cnf-test-partner:latest
		imageID: quay.io/testnetworkfunction/cnf-test-partner@sha256:e117af66264c5e6db9effb5ebe1b4c79893bc44f281676446553e98eb4041efe
		lastState: {}
		name: test
		ready: true
		restartCount: 0
		started: true
		state:
		  running:
			startedAt: "2021-04-04T10:10:22Z"
	  hostIP: 192.168.126.11
	  phase: Running
	  podIP: 10.116.0.17
	  podIPs:
	  - ip: 10.116.0.17
	  qosClass: Guaranteed
	  startTime: "2021-03-16T14:48:39Z"`
)
