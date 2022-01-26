// Copyright (C) 2022 Red Hat, Inc.
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

package accesscontrol

import (
	"errors"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

func TestParseCrOutput(t *testing.T) {
	testCases := []struct {
		rawOutput      string
		expectedOutput map[string]string
		expectedErr    error
	}{
		{
			rawOutput: "aws-ebs-csi-driver-operator,openshift-cloud-credential-operator",
			expectedOutput: map[string]string{
				"aws-ebs-csi-driver-operator": "openshift-cloud-credential-operator",
			},
			expectedErr: nil,
		},
		{
			rawOutput: "abcd1234,openshift-cloud-credential-operator",
			expectedOutput: map[string]string{
				"abcd1234": "openshift-cloud-credential-operator",
			},
			expectedErr: nil,
		},
		{
			rawOutput:      "openshift-cloud-credential-operator",
			expectedOutput: map[string]string{},
			expectedErr:    errors.New("failed to parse output line openshift-cloud-credential-operator"),
		},
		{
			rawOutput:      "",
			expectedOutput: map[string]string{},
			expectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		crMap, err := parseCrOutput(tc.rawOutput)
		assert.Equal(t, tc.expectedErr, err)
		if err == nil {
			assert.Equal(t, tc.expectedOutput, crMap)
		}
	}
}

func TestGetCrsNamespaces(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	origFunc := utils.ExecuteCommandAndValidate
	defer func() {
		utils.ExecuteCommandAndValidate = origFunc
	}()

	utils.ExecuteCommandAndValidate = func(command string, timeout time.Duration, context *interactive.Context, failureCallbackFun func()) string {
		return "aws-ebs-csi-driver-operator,openshift-cloud-credential-operator"
	}

	crsNamespaces, err := getCrsNamespaces("test123", "testCRD", nil)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		"aws-ebs-csi-driver-operator": "openshift-cloud-credential-operator",
	}, crsNamespaces)
}

//nolint:funlen
func TestAddFailedTcInfo(t *testing.T) {
	testCases := []struct {
		tc          string
		pod         string
		namespace   string
		contID      int
		existingMap map[string][]failedTcInfo
		expectedMap map[string][]failedTcInfo
	}{
		{
			tc:          "tc1",
			pod:         "pod1",
			namespace:   "ns1",
			contID:      1,
			existingMap: map[string][]failedTcInfo{},
			expectedMap: map[string][]failedTcInfo{
				"pod1": {
					{
						tc:           "tc1",
						containerIdx: 1,
						ns:           "ns1",
					},
				},
			},
		},
		{
			tc:        "tc1",
			pod:       "pod1",
			namespace: "ns1",
			contID:    1,
			existingMap: map[string][]failedTcInfo{
				"pod1": {
					{
						tc:           "tc1",
						containerIdx: 1,
						ns:           "ns1",
					},
				},
			},
			expectedMap: map[string][]failedTcInfo{
				"pod1": {
					{
						tc:           "tc1",
						containerIdx: 1,
						ns:           "ns1",
					},
					{
						tc:           "tc1",
						containerIdx: 1,
						ns:           "ns1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		addFailedTcInfo(tc.existingMap, tc.tc, tc.pod, tc.namespace, tc.contID)
		assert.Equal(t, tc.expectedMap, tc.existingMap)
	}
}
