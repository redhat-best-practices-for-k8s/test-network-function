// Copyright (C) 2020-2022 Red Hat, Inc.
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

package networking

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	"github.com/test-network-function/test-network-function/pkg/utils"
)

func TestParseVariables(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputs
		// inputRes is string that include the result after we run the command ""oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].ports'""
		inputRes string

		// now is empty but maybe in the future has be not empty.
		declaredPorts map[key]string

		// expected outputs here
		expectedDeclaredPorts map[key]string
		expectedRes           string
	}{
		{
			inputRes:              "[\n  {\n    \"containerPort\": 8080,\n    \"name\": \"http-probe\",\n    \"protocol\": \"TCP\"\n  },{\n    \"containerPort\": 7878,\n    \"name\": \"http\",\n    \"protocol\": \"TCP\"\n  } \n]",
			declaredPorts:         map[key]string{},
			expectedDeclaredPorts: map[key]string{{port: 8080, protocol: "TCP"}: "http-probe", {port: 7878, protocol: "TCP"}: "http"},
			expectedRes:           "[\n  {\n    \"containerPort\": 8080,\n    \"name\": \"http-probe\",\n    \"protocol\": \"TCP\"\n  },{\n    \"containerPort\": 7878,\n    \"name\": \"http\",\n    \"protocol\": \"TCP\"\n  } \n]",
		},
		{
			inputRes:              "[\n  {\n    \"containerPort\": 8080,\n    \"name\": \"http-probe\",\n    \"protocol\": \"TCP\"\n  }\n]",
			declaredPorts:         map[key]string{},
			expectedDeclaredPorts: map[key]string{{port: 8080, protocol: "TCP"}: "http-probe"},
			expectedRes:           "[\n  {\n    \"containerPort\": 8080,\n    \"name\": \"http-probe\",\n    \"protocol\": \"TCP\"\n  }\n]",
		},
		{
			inputRes:              "[\n  {\n    \"containerPort\": 8080,\n    \"name\": \"http-probe\",\n    \"protocol\": \"UDP\"\n  }\n]",
			declaredPorts:         map[key]string{},
			expectedDeclaredPorts: map[key]string{{port: 8080, protocol: "UDP"}: "http-probe"},
			expectedRes:           "[\n  {\n    \"containerPort\": 8080,\n    \"name\": \"http-probe\",\n    \"protocol\": \"UDP\"\n  }\n]",
		},
		{
			inputRes:              "[\n \n]",
			declaredPorts:         map[key]string{},
			expectedDeclaredPorts: map[key]string{},
			expectedRes:           "[\n \n]",
		},
		{
			inputRes:              "[\n  {\n    \"containerPort\": 9000,\n    \"name\": \"http-probe\",\n    \"protocol\": \"UDP\"\n  }\n]",
			declaredPorts:         map[key]string{},
			expectedDeclaredPorts: map[key]string{{port: 9000, protocol: "UDP"}: "http-probe"},
			expectedRes:           "[\n  {\n    \"containerPort\": 9000,\n    \"name\": \"http-probe\",\n    \"protocol\": \"UDP\"\n  }\n]",
		},
	}

	for _, tc := range testCases {
		err := parseVariables(tc.inputRes, tc.declaredPorts)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedDeclaredPorts, tc.declaredPorts)
	}
}

func TestDeclaredPortList(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputs
		jsonFileName  string
		container     int
		podName       string
		podNamespace  string
		declaredPorts map[key]string

		// expected outputs here
		expectedDeclaredPorts map[key]string
	}{
		{
			jsonFileName:          "testdata/test_ports.json",
			container:             0,
			podName:               "test-54bc4c6d7-8rzch",
			podNamespace:          "tnf",
			declaredPorts:         map[key]string{},
			expectedDeclaredPorts: map[key]string{{port: 8080, protocol: "TCP"}: "http-probe", {port: 8443, protocol: "TCP"}: "https", {port: 50051, protocol: "TCP"}: "grpc"},
		},
	}

	origFunc := utils.ExecuteCommand
	defer func() {
		utils.ExecuteCommand = origFunc
	}()
	for _, tc := range testCases {
		utils.ExecuteCommand = func(command string, timeout time.Duration, context *interactive.Context) (string, error) {
			output, err := os.ReadFile(tc.jsonFileName)
			return string(output), err
		}
		err := declaredPortList(tc.container, tc.podName, tc.podNamespace, tc.declaredPorts)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedDeclaredPorts, tc.declaredPorts)
	}
}

func TestListeningPortList(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputs
		jsonFileName   string
		commandlisten  []string
		nodeOc         *interactive.Context
		listeningPorts map[key]string

		// expected outputs here
		expectedlisteningPorts map[key]string
	}{
		{
			jsonFileName:           "testdata/test_listening_port.json",
			commandlisten:          []string{"nsenter -t 4380 -n", "ss -tulwnH"},
			nodeOc:                 nil,
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{{port: 8080, protocol: "TCP"}: ""},
		},
		{
			jsonFileName:           "testdata/test_listening_port.json",
			commandlisten:          []string{},
			nodeOc:                 nil,
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{},
		},
	}
	origFunc := utils.ExecuteCommand
	defer func() {
		utils.ExecuteCommand = origFunc
	}()
	for _, tc := range testCases {
		utils.ExecuteCommand = func(command string, timeout time.Duration, context *interactive.Context) (string, error) {
			output, err := os.ReadFile(tc.jsonFileName)
			return string(output), err
		}
		err := listeningPortList(tc.commandlisten, tc.nodeOc, tc.listeningPorts)
		assert.Nil(t, err)
	}
}

func TestCheckIfListenIsDeclared(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputs
		listeningPorts map[key]string
		declaredPorts  map[key]string

		// expected outputs here
		expectedres map[key]string
	}{
		{
			listeningPorts: map[key]string{},
			declaredPorts:  map[key]string{},
			expectedres:    map[key]string{},
		},
		{
			listeningPorts: map[key]string{{port: 8080, protocol: "TCP"}: ""},
			declaredPorts:  map[key]string{{port: 8080, protocol: "TCP"}: "http-probe"},
			expectedres:    map[key]string{},
		},
	}
	for _, tc := range testCases {
		res := checkIfListenIsDeclared(tc.listeningPorts, tc.declaredPorts)
		assert.Equal(t, len(res), 0)
		assert.Equal(t, len(tc.listeningPorts), len(tc.declaredPorts))
  }
}

func TestFilterIPListPerVersion(t *testing.T) {
	type args struct {
		ipList    []string
		ipVersion ipVersion
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "IPv4",
			args: args{ipList: []string{"2.2.2.2", "3.3.3.3", "fd00:10:244:1::3", "fd00:10:244:1::4"}, ipVersion: IPv4},
			want: []string{"2.2.2.2", "3.3.3.3"},
		},
		{name: "IPv6",
			args: args{ipList: []string{"2.2.2.2", "3.3.3.3", "fd00:10:244:1::3", "fd00:10:244:1::4"}, ipVersion: IPv6},
			want: []string{"fd00:10:244:1::3", "fd00:10:244:1::4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterIPListPerVersion(tt.args.ipList, tt.args.ipVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterIPListPerVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getIPVersion(t *testing.T) {
	type args struct {
		aIP string
	}
	tests := []struct {
		name    string
		args    args
		want    ipVersion
		wantErr bool
	}{
		{name: "GoodIPv4",
			args:    args{aIP: "2.2.2.2"},
			want:    IPv4,
			wantErr: false,
		},
		{name: "GoodIPv6",
			args:    args{aIP: "fd00:10:244:1::3"},
			want:    IPv6,
			wantErr: false,
		},
		{name: "BadIPv4",
			args:    args{aIP: "2.hfh.2.2"},
			want:    "",
			wantErr: true,
		},
		{name: "BadIPv6",
			args:    args{aIP: "fd00:10:ono;ogmo:1::3"},
			want:    "",
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIPVersion(tt.args.aIP)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIPVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIPVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
