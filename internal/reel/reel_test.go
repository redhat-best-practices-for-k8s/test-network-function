// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package reel_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	mock_reel "github.com/redhat-nfvpe/test-network-function/internal/reel/mocks"
	mock_interactive "github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive/mocks"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

// Tests all aspects of reel.  The few lines that are not tested include a function definition that is passed as a
// parameter to goexpect and called externally.  Since we are mocking goexpect, it is not easy to test this function
// (generateCase()).

var (
	defaultCommand   = []string{"ls"}
	reelError        = errors.New("some reel error")
	sendCommandError = errors.New("send command error")
)

type newReelTestCase struct {
	command            []string
	newReelResultErr   error
	newReelResultIsNil bool
	sendCmdErr         error
}

var newReelTestCases = map[string]newReelTestCase{
	"success": {
		command:            defaultCommand,
		newReelResultIsNil: false,
		newReelResultErr:   nil,
		sendCmdErr:         nil,
	},
	"success_with_newline": {
		command:            []string{"ls\n"},
		newReelResultIsNil: false,
		newReelResultErr:   nil,
		sendCmdErr:         nil,
	},
	"fail_to_send": {
		command:            defaultCommand,
		newReelResultErr:   sendCommandError,
		newReelResultIsNil: true,
		sendCmdErr:         sendCommandError,
	},
	"empty_command": {
		command:            []string{},
		newReelResultErr:   nil,
		newReelResultIsNil: false,
		sendCmdErr:         sendCommandError,
	},
}

// Also handles testing of Run()
func TestNewReel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range newReelTestCases {
		mockExpecter := mock_interactive.NewMockExpecter(ctrl)

		testCommand := strings.Join(testCase.command, " ")
		if !strings.HasSuffix(testCommand, "\n") {
			testCommand = testCommand + "\n"
		}
		mockExpecter.EXPECT().Send(testCommand).AnyTimes().Return(testCase.sendCmdErr)

		var expecter expect.Expecter = mockExpecter
		var errorChannel <-chan error
		r, err := reel.NewReel(&expecter, testCase.command, errorChannel)
		assert.Equal(t, testCase.newReelResultErr, err)
		assert.Equal(t, testCase.newReelResultIsNil, r == nil)
		// In the event that the Reel instance is good, perform Run()
		if r != nil {
			mockHandler := mock_reel.NewMockHandler(ctrl)
			mockHandler.EXPECT().ReelFirst().AnyTimes().Return(nil)
			err := r.Run(mockHandler)
			assert.Nil(t, err)
		}
	}
}

// Test the REEL state machine.

type reelStepTestCase struct {
	stepInput                          *reel.Step
	command                            []string
	stepReturnErr                      error
	reelErr                            error
	expectBatchExpectedInvocationCount int
	expectBatchResResult               []expect.BatchRes
	expectBatchErrResult               error
	isTimeout                          bool
}

var reelStepTestCases = map[string]reelStepTestCase{
	// the case where step is nil initially.
	"step_is_nil": {
		stepInput:                          nil,
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		expectBatchExpectedInvocationCount: 0,
		isTimeout:                          false,
	},
	"reel_err_not_nil": {
		stepInput:                          &reel.Step{},
		command:                            defaultCommand,
		stepReturnErr:                      reelError,
		reelErr:                            reelError,
		expectBatchExpectedInvocationCount: 0,
		isTimeout:                          false,
	},
	"expectationless_step": {
		stepInput:                          &reel.Step{},
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult:               []expect.BatchRes{},
		expectBatchErrResult:               nil,
		isTimeout:                          false,
	},
	"timeout_error": {
		stepInput:                          &reel.Step{Expect: []string{"expect something"}},
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult:               []expect.BatchRes{},
		expectBatchErrResult:               expect.TimeoutError(time.Second * 1),
		isTimeout:                          true,
	},
	"non_timeout_error": {
		stepInput:                          &reel.Step{Expect: []string{"expect something"}},
		command:                            defaultCommand,
		stepReturnErr:                      reelError,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult:               []expect.BatchRes{},
		expectBatchErrResult:               reelError,
		isTimeout:                          false,
	},
	"successful_reel": {
		stepInput:                          &reel.Step{Expect: []string{`.+`}},
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult: []expect.BatchRes{
			{
				Idx:    0,
				Output: "someMatch",
				Match:  []string{"someMatch"},
			},
		},
		expectBatchErrResult: nil,
		isTimeout:            false,
	},
	"another_successful_reel": {
		stepInput:                          &reel.Step{Expect: []string{`.+start`}},
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult: []expect.BatchRes{
			{
				Idx:    0,
				Output: "12345\n12start",
				Match:  []string{"12start"},
			},
		},
		expectBatchErrResult: nil,
		isTimeout:            false,
	},
	"second_match": {
		stepInput:                          &reel.Step{Expect: []string{`.+start`, `.+`}},
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult: []expect.BatchRes{
			{
				Idx:    0,
				Output: "12345",
				Match:  []string{"12345"},
			},
		},
		expectBatchErrResult: nil,
		isTimeout:            false,
	},
	"execute_exists": {
		stepInput:                          &reel.Step{Execute: "ls"},
		command:                            defaultCommand,
		stepReturnErr:                      nil,
		reelErr:                            nil,
		expectBatchExpectedInvocationCount: 1,
		expectBatchResResult:               nil,
		expectBatchErrResult:               nil,
		isTimeout:                          false,
	},
}

func TestReel_Step(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range reelStepTestCases {
		testCommand := strings.Join(testCase.command, " ")
		if !strings.HasSuffix(testCommand, "\n") {
			testCommand = testCommand + "\n"
		}

		mockExpecter := mock_interactive.NewMockExpecter(ctrl)
		mockExpecter.EXPECT().Send(testCommand).Return(nil)
		if testCase.expectBatchExpectedInvocationCount > 0 {
			mockExpecter.EXPECT().ExpectBatch(gomock.Any(), gomock.Any()).Times(testCase.expectBatchExpectedInvocationCount).Return(testCase.expectBatchResResult, testCase.expectBatchErrResult)
		}

		var expecter expect.Expecter = mockExpecter
		var errorChannel <-chan error
		r, err := reel.NewReel(&expecter, testCase.command, errorChannel)
		r.Err = testCase.reelErr
		assert.Nil(t, err)

		handler := mock_reel.NewMockHandler(ctrl)
		// timeout test only
		if testCase.isTimeout {
			handler.EXPECT().ReelTimeout().Times(1)
		}

		// successful ExpectBatch
		if len(testCase.expectBatchResResult) > 0 {
			handler.EXPECT().ReelMatch(gomock.Any(), gomock.Any(), gomock.Any())
		}

		err = r.Step(testCase.stepInput, handler)
		assert.Equal(t, testCase.stepReturnErr, err)
	}
}
