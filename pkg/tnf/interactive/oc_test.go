// Copyright (C) 2020 Red Hat, Inc.
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

package interactive_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	expect "github.com/ryandgoulding/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	mock_interactive "github.com/test-network-function/test-network-function/pkg/tnf/interactive/mocks"
)

const (
	ocTestTimeoutDuration = time.Second * 2
)

var errSpawnOC = errors.New("some error related to spawning OC")

type ocTestCase struct {
	podName            string
	podContainerName   string
	podNamespace       string
	options            []expect.Option
	errReturnValue     error
	contextReturnValue *interactive.Context
	expectedSpawnErr   error
}

var ocTestCases = map[string]ocTestCase{
	"no_error": {
		podName:            "test",
		podContainerName:   "test",
		podNamespace:       "default",
		options:            []expect.Option{expect.Verbose(true)},
		errReturnValue:     nil,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   nil,
	},
	"error": {
		podName:            "test",
		podContainerName:   "testPod",
		podNamespace:       "default",
		options:            []expect.Option{expect.Verbose(true)},
		errReturnValue:     errSpawnOC,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   errSpawnOC,
	},
}

func TestSpawnOc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range ocTestCases {
		ctrl = gomock.NewController(t)
		mockSpawner := mock_interactive.NewMockSpawner(ctrl)
		mockSpawner.EXPECT().Spawn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testCase.contextReturnValue, testCase.errReturnValue)

		var spawner interactive.Spawner = mockSpawner
		oc, _, err := interactive.SpawnOc(&spawner, testCase.podName, testCase.podContainerName, testCase.podNamespace, ocTestTimeoutDuration, testCase.options...)
		assert.Equal(t, testCase.expectedSpawnErr, err)
		if testCase.expectedSpawnErr == nil {
			assert.Equal(t, testCase.podName, oc.GetPodName())
			assert.Equal(t, testCase.podContainerName, oc.GetPodContainerName())
			assert.Equal(t, testCase.podNamespace, oc.GetPodNamespace())
			assert.Equal(t, ocTestTimeoutDuration, oc.GetTimeout())
			assert.Nil(t, oc.GetErrorChannel())
			assert.Nil(t, oc.GetExpecter())
			// Since expect.Option is rightfully of type func, we cannot compare equality of func(s).  Thus, just check to make
			// sure there is an option.
			assert.Equal(t, 1, len(oc.GetOptions()))
		}
	}
}
