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

	"github.com/golang/mock/gomock"
	expect "github.com/ryandgoulding/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	mock_interactive "github.com/test-network-function/test-network-function/pkg/tnf/interactive/mocks"
)

var errSpawnShell = errors.New("some error related to spawning shell")

type shellTestCase struct {
	options            []expect.Option
	errReturnValue     error
	contextReturnValue *interactive.Context
	expectedSpawnErr   error
}

var shellTestCases = map[string]shellTestCase{
	"no_error": {
		options:            []expect.Option{expect.Verbose(true)},
		errReturnValue:     nil,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   nil,
	},
	"error": {
		options:            []expect.Option{expect.Verbose(true)},
		errReturnValue:     errSpawnShell,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   errSpawnShell,
	},
}

func TestSpawnShell(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range shellTestCases {
		ctrl = gomock.NewController(t)
		mockSpawner := mock_interactive.NewMockSpawner(ctrl)
		mockSpawner.EXPECT().Spawn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testCase.contextReturnValue, testCase.errReturnValue)

		var spawner interactive.Spawner = mockSpawner
		_, err := interactive.SpawnShell(&spawner, ocTestTimeoutDuration, testCase.options...)
		assert.Equal(t, testCase.expectedSpawnErr, err)
	}
}
