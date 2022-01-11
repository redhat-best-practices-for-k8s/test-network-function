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

package tnf_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	expect "github.com/google/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf"
	mock_interactive "github.com/test-network-function/test-network-function/pkg/tnf/interactive/mocks"
	mock_tnf "github.com/test-network-function/test-network-function/pkg/tnf/mocks"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
	mock_reel "github.com/test-network-function/test-network-function/pkg/tnf/reel/mocks"
)

const (
	testTimeoutDuration = time.Second * 2
)

var (
	defaultTestCommand = []string{"ls"}
	errSend            = errors.New("generic send error")
)

type newTestTestCase struct {
	testCommandArgs []string
	sendReturnErr   error
	newTestErr      error
	newTestIsNil    bool
}

// Tests related to TestCase instantiation.
var newTestTestCases = map[string]newTestTestCase{
	// Replicates the idea that the test was successfully instantiated
	"successful_new_test": {
		testCommandArgs: defaultTestCommand,
		sendReturnErr:   nil,
		newTestErr:      nil,
		newTestIsNil:    false,
	},
	// Replicates the idea that the test was not successfully instantiated as the expect.Send(...) returned an error.
	"send_error": {
		testCommandArgs: defaultTestCommand,
		sendReturnErr:   errSend,
		newTestErr:      errSend,
		newTestIsNil:    true,
	},
}

func TestNewTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range newTestTestCases {
		testCommand := strings.Join(testCase.testCommandArgs, " ") + "\n"

		mockExpecter := mock_interactive.NewMockExpecter(ctrl)
		mockExpecter.EXPECT().Send(testCommand).Return(testCase.sendReturnErr)

		mockTester := mock_tnf.NewMockTester(ctrl)
		mockTester.EXPECT().Args().Return(testCase.testCommandArgs)

		mockHandler := mock_reel.NewMockHandler(ctrl)

		var expecter expect.Expecter = mockExpecter
		var errorChannel <-chan error
		testTest, err := tnf.NewTest(&expecter, mockTester, []reel.Handler{mockHandler}, errorChannel, reel.DisableTerminalPromptEmulation())
		assert.Equal(t, testCase.newTestErr, err)
		assert.Equal(t, testCase.newTestIsNil, testTest == nil)
	}
}

type testRunTestCase struct {
	testCommandArgs           []string
	reelFirstResult           *reel.Step
	testerResultResult        int
	expectBatchIsCalled       bool
	expectBatchBatchResResult []expect.BatchRes
	expectBatchBatchResErr    error
	reelMatchIsCalled         bool
	reelMatchPattern          string
	reelMatchBefore           string
	reelMatchMatch            string
	reelMatchResult           *reel.Step
	testRunErr                error
}

func fakeSentinelOutputWithReturnCode(output string, code int) string {
	return fmt.Sprintf("%s%s %s%d\n", output, reel.EndOfTestSentinel, reel.ExitKeyword, code)
}

func fakeSentinelOutput(output string) string {
	return fakeSentinelOutputWithReturnCode(output, 0)
}

func fakeOutput() string {
	return "someOutput"
}

func fakeWrongOutput() string {
	return "something else"
}

func fakeErrorCode() int {
	return 1
}

// Tests the actual state machine.
var testRunTestCases = map[string]testRunTestCase{
	"successful_run": {
		testCommandArgs:     defaultTestCommand,
		reelFirstResult:     nil,
		testerResultResult:  tnf.SUCCESS,
		expectBatchIsCalled: false,
		reelMatchIsCalled:   false,
	},
	"fail_run": {
		testCommandArgs:     defaultTestCommand,
		reelFirstResult:     nil,
		testerResultResult:  tnf.FAILURE,
		expectBatchIsCalled: false,
		reelMatchIsCalled:   false,
	},
	"error_run": {
		testCommandArgs:     defaultTestCommand,
		reelFirstResult:     nil,
		testerResultResult:  tnf.ERROR,
		expectBatchIsCalled: false,
		reelMatchIsCalled:   false,
	},
	// tests the state machine transition from ReelFirst() to ReelMatch()
	"reel_first_transition_into_reel_match": {
		testCommandArgs: defaultTestCommand,
		reelFirstResult: &reel.Step{
			Execute: "ls",
			Expect:  []string{fakeOutput()},
			Timeout: testTimeoutDuration,
		},
		testerResultResult:  tnf.ERROR,
		expectBatchIsCalled: true,
		expectBatchBatchResResult: []expect.BatchRes{
			{
				Idx:     1,
				CaseIdx: 0,
				Output:  fakeSentinelOutput(fakeOutput()),
				Match:   []string{fakeSentinelOutput(fakeOutput())}},
		},
		expectBatchBatchResErr: nil,
		reelMatchIsCalled:      true,
		reelMatchPattern:       "",
		reelMatchBefore:        "",
		reelMatchMatch:         fakeOutput(),
		reelMatchResult:        nil,
	},
	"reel_first_only": {
		testCommandArgs: defaultTestCommand,
		reelFirstResult: &reel.Step{
			Execute: "ls",
			Expect:  []string{fakeOutput()},
			Timeout: testTimeoutDuration,
		},
		testerResultResult:  tnf.ERROR,
		expectBatchIsCalled: true,
		expectBatchBatchResResult: []expect.BatchRes{
			{
				Idx:     1,
				CaseIdx: 1,
				Output:  fakeSentinelOutput(fakeWrongOutput()),
				Match:   []string{fakeSentinelOutput(fakeWrongOutput())}},
		},
		expectBatchBatchResErr: nil,
		reelMatchIsCalled:      false,
	},
	"reel_first_only_with_error_code": {
		testCommandArgs: defaultTestCommand,
		reelFirstResult: &reel.Step{
			Execute: "ls",
			Expect:  []string{fakeOutput()},
			Timeout: testTimeoutDuration,
		},
		testerResultResult:  tnf.ERROR,
		expectBatchIsCalled: true,
		expectBatchBatchResResult: []expect.BatchRes{
			{
				Idx:     1,
				CaseIdx: 1,
				Output:  fakeSentinelOutputWithReturnCode(fakeWrongOutput(), fakeErrorCode()),
				Match:   []string{fakeSentinelOutputWithReturnCode(fakeWrongOutput(), fakeErrorCode())}},
		},
		expectBatchBatchResErr: nil,
		reelMatchIsCalled:      false,
		testRunErr:             fmt.Errorf("error executing command exit code:%d", fakeErrorCode()),
	},
}

// Also covers ReelFirst() and ReelMatch().  Tests those state transitions.
func TestTest_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range testRunTestCases {
		mockExpecter := mock_interactive.NewMockExpecter(ctrl)
		testCommand := fmt.Sprintf("%s ; echo %s %s$?\n", strings.Join(testCase.testCommandArgs, " "), reel.EndOfTestSentinel, reel.ExitKeyword)
		mockExpecter.EXPECT().Send(testCommand).AnyTimes()

		// Only for test cases where ReelMatch(...) is encountered.
		if testCase.expectBatchIsCalled {
			mockExpecter.EXPECT().ExpectBatch(gomock.Any(), gomock.Any()).Return(testCase.expectBatchBatchResResult, testCase.expectBatchBatchResErr)
		}
		mockTester := mock_tnf.NewMockTester(ctrl)
		mockTester.EXPECT().Args().Return(testCase.testCommandArgs)
		mockTester.EXPECT().Result().Return(testCase.testerResultResult)

		mockHandler := mock_reel.NewMockHandler(ctrl)
		mockHandler.EXPECT().ReelFirst().Return(testCase.reelFirstResult)
		// Only for cases where ReelMatch(...) is encountered
		if testCase.reelMatchIsCalled {
			mockHandler.EXPECT().ReelMatch(testCase.reelMatchPattern, testCase.reelMatchBefore, testCase.reelMatchMatch).Return(testCase.reelMatchResult)
		}

		var expecter expect.Expecter = mockExpecter
		var errorChannel <-chan error
		test, err := tnf.NewTest(&expecter, mockTester, []reel.Handler{mockHandler}, errorChannel)
		assert.Nil(t, err)
		assert.NotNil(t, test)
		result, err := test.Run()

		assert.Equal(t, err, testCase.testRunErr)
		assert.Equal(t, result, testCase.testerResultResult)
	}
}

func TestTest_ReelTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExpecter := mock_interactive.NewMockExpecter(ctrl)
	mockExpecter.EXPECT().Send(gomock.Any()).AnyTimes()

	mockTester := mock_tnf.NewMockTester(ctrl)
	mockTester.EXPECT().Args().Return(defaultTestCommand)

	mockHandler := mock_reel.NewMockHandler(ctrl)
	mockHandler.EXPECT().ReelTimeout().Return(nil)
	var expecter expect.Expecter = mockExpecter
	var errorChannel <-chan error

	test, err := tnf.NewTest(&expecter, mockTester, []reel.Handler{mockHandler}, errorChannel)

	assert.Nil(t, err)
	step := test.ReelTimeout()
	assert.Nil(t, step)
}

func TestTest_ReelEof(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExpecter := mock_interactive.NewMockExpecter(ctrl)
	mockExpecter.EXPECT().Send(gomock.Any()).AnyTimes()

	mockTester := mock_tnf.NewMockTester(ctrl)
	mockTester.EXPECT().Args().Return(defaultTestCommand)

	mockHandler := mock_reel.NewMockHandler(ctrl)
	mockHandler.EXPECT().ReelEOF().Times(1)
	var expecter expect.Expecter = mockExpecter
	var errorChannel <-chan error

	test, err := tnf.NewTest(&expecter, mockTester, []reel.Handler{mockHandler}, errorChannel)

	assert.Nil(t, err)
	// just ensure there are no panics
	test.ReelEOF()
}
