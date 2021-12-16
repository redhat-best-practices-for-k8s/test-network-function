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

package common

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsNonOcpCluster(t *testing.T) {
	testCases := []struct {
		isNonOCPCluster bool
	}{
		{
			isNonOCPCluster: true,
		},
		{
			isNonOCPCluster: false,
		},
	}

	defer os.Unsetenv("TNF_NON_OCP_CLUSTER")
	for _, tc := range testCases {
		if tc.isNonOCPCluster {
			os.Setenv("TNF_NON_OCP_CLUSTER", "true")
			assert.Equal(t, tc.isNonOCPCluster, IsNonOcpCluster())
		} else {
			os.Setenv("TNF_NON_OCP_CLUSTER", "false")
			assert.Equal(t, tc.isNonOCPCluster, IsNonOcpCluster())
		}
	}
}

func TestIntrusive(t *testing.T) {
	testCases := []struct {
		isIntrusive bool
	}{
		{
			isIntrusive: true,
		},
		{
			isIntrusive: false,
		},
	}

	defer os.Unsetenv("TNF_NON_INTRUSIVE_ONLY")
	for _, tc := range testCases {
		if tc.isIntrusive {
			os.Setenv("TNF_NON_INTRUSIVE_ONLY", "false")
			assert.Equal(t, tc.isIntrusive, Intrusive())
		} else {
			os.Setenv("TNF_NON_INTRUSIVE_ONLY", "true")
			assert.Equal(t, tc.isIntrusive, Intrusive())
		}
	}
}

func TestLogLevel(t *testing.T) {
	testCases := []struct {
		logLevel         string
		expectedLogLevel string
	}{
		{
			logLevel:         "high",
			expectedLogLevel: "high",
		},
		{
			logLevel:         "",
			expectedLogLevel: "debug",
		},
	}

	defer os.Unsetenv("LOG_LEVEL")
	for _, tc := range testCases {
		os.Setenv("LOG_LEVEL", tc.logLevel)
		assert.Equal(t, tc.expectedLogLevel, logLevel())
	}
}

func TestSetLogLevel(t *testing.T) {
	testCases := []struct {
		logLevel         string
		expectedLogLevel log.Level
	}{
		{
			logLevel:         "high",
			expectedLogLevel: log.DebugLevel,
		},
		{
			logLevel:         "",
			expectedLogLevel: log.DebugLevel,
		},
		{
			logLevel:         "trace",
			expectedLogLevel: log.TraceLevel,
		},
	}

	defer os.Unsetenv("LOG_LEVEL")
	for _, tc := range testCases {
		os.Setenv("LOG_LEVEL", tc.logLevel)
		SetLogLevel()
		assert.Equal(t, tc.expectedLogLevel, log.GetLevel())
	}
}
