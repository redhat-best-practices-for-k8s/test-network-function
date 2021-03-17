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
	"io"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	expect "github.com/ryandgoulding/goexpect"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function/pkg/tnf/interactive"
	mock_interactive "github.com/test-network-function/test-network-function/pkg/tnf/interactive/mocks"
)

// Note: Test coverage for this file is as high as possible short of attempting to perform multi-threaded unit tests.
// Some lines cannot be covered as they are specifically geared towards production use v.s. unit test use (mock injection).

const (
	testTimeoutDuration = time.Second * 2
)

func init() {
	interactive.UnitTestMode = true
}

var (
	defaultGoExpectArgs            = []expect.Option{expect.Verbose(true)}
	defaultStdout, defaultStdin, _ = os.Pipe()
	errStart                       = errors.New("start failed")
	errStdInPipe                   = errors.New("failed to access stdin")
)

type goExpectSpawnerTestCase struct {
	goExpectSpawnerSpawnCommand string
	goExpectSpawnerSpawnArgs    []string
	goExpectSpawnerSpawnTimeout time.Duration
	goExpectSpawnerSpawnOpts    []expect.Option

	stdinPipeShouldBeCalled bool
	stdinPipeReturnValue    io.WriteCloser
	stdinPipeReturnErr      error

	stdoutPipeShouldBeCalled bool
	stdoutPipeReturnValue    io.Reader
	stdoutPipeReturnErr      error

	startShouldBeCalled bool
	startReturnErr      error

	goExpectSpawnerSpawnReturnContextIsNil bool
	goExpectSpawnerSpawnReturnErr          error
}

var goExpectSpawnerTestCases = map[string]goExpectSpawnerTestCase{
	// 1. Test to ensure that if StdinPipe() fails, that the error cascades back out of the Spawn invocation.
	"stdin_pipe_creation_failure": {
		// The command is unimportant
		goExpectSpawnerSpawnCommand: "ls",
		goExpectSpawnerSpawnArgs:    []string{"-al"},
		goExpectSpawnerSpawnTimeout: testTimeoutDuration,
		goExpectSpawnerSpawnOpts:    defaultGoExpectArgs,

		// Sets up a scenario in which the call to StdinPipe() returns an error.  This error should cascade out of Spawn().
		stdinPipeShouldBeCalled: true,
		stdinPipeReturnValue:    nil,
		stdinPipeReturnErr:      errStdInPipe,

		stdoutPipeShouldBeCalled: false,
		stdoutPipeReturnValue:    nil,
		stdoutPipeReturnErr:      nil,

		startShouldBeCalled: false,
		startReturnErr:      nil,

		goExpectSpawnerSpawnReturnContextIsNil: true,
		goExpectSpawnerSpawnReturnErr:          errStdInPipe,
	},
	// 2. Progressing past the creation of stdin, now cause stdout to fail.
	"stdout_pipe_creation_failure": {
		// The command is unimportant
		goExpectSpawnerSpawnCommand: "ls",
		goExpectSpawnerSpawnArgs:    []string{"-al"},
		goExpectSpawnerSpawnTimeout: testTimeoutDuration,
		goExpectSpawnerSpawnOpts:    defaultGoExpectArgs,

		stdinPipeShouldBeCalled: true,
		stdinPipeReturnValue:    defaultStdin,
		stdinPipeReturnErr:      nil,

		// cause StdoutPipe() call to fail and ensure the error cascades.
		stdoutPipeShouldBeCalled: true,
		stdoutPipeReturnValue:    nil,
		stdoutPipeReturnErr:      errStdInPipe,

		startShouldBeCalled: false,
		startReturnErr:      nil,

		goExpectSpawnerSpawnReturnContextIsNil: true,
		goExpectSpawnerSpawnReturnErr:          errStdInPipe,
	},
	// 3. Progressing past the creation of stdin/stdout, now cause Start to fail.
	"start_failure": {
		// The command is unimportant
		goExpectSpawnerSpawnCommand: "ls",
		goExpectSpawnerSpawnArgs:    []string{"-al"},
		goExpectSpawnerSpawnTimeout: testTimeoutDuration,
		goExpectSpawnerSpawnOpts:    defaultGoExpectArgs,

		stdinPipeShouldBeCalled: true,
		stdinPipeReturnValue:    defaultStdin,
		stdinPipeReturnErr:      nil,

		stdoutPipeShouldBeCalled: true,
		stdoutPipeReturnValue:    defaultStdout,
		stdoutPipeReturnErr:      nil,

		// cause Start() call to fail and make sure the error cascades out of Spawn().
		startShouldBeCalled: true,
		startReturnErr:      errStart,

		goExpectSpawnerSpawnReturnContextIsNil: true,
		goExpectSpawnerSpawnReturnErr:          errStart,
	},
	// 4. Successful spawn.
	"successful_spawn": {
		// The command is unimportant
		goExpectSpawnerSpawnCommand: "ls",
		goExpectSpawnerSpawnArgs:    []string{"-al"},
		goExpectSpawnerSpawnTimeout: testTimeoutDuration,
		goExpectSpawnerSpawnOpts:    defaultGoExpectArgs,

		stdinPipeShouldBeCalled: true,
		stdinPipeReturnValue:    defaultStdin,
		stdinPipeReturnErr:      nil,

		stdoutPipeShouldBeCalled: true,
		stdoutPipeReturnValue:    defaultStdout,
		stdoutPipeReturnErr:      nil,

		startShouldBeCalled: true,
		startReturnErr:      nil,

		goExpectSpawnerSpawnReturnContextIsNil: false,
		goExpectSpawnerSpawnReturnErr:          nil,
	},
}

func TestGoExpectSpawner_Spawn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range goExpectSpawnerTestCases {
		mockSpawnFunc := mock_interactive.NewMockSpawnFunc(ctrl)
		// coax the types
		var sFunc interactive.SpawnFunc = mockSpawnFunc
		interactive.SetSpawnFunc(&sFunc)

		if testCase.stdinPipeShouldBeCalled {
			mockSpawnFunc.EXPECT().StdinPipe().Return(testCase.stdinPipeReturnValue, testCase.stdinPipeReturnErr)
		}

		if testCase.stdoutPipeShouldBeCalled {
			mockSpawnFunc.EXPECT().StdoutPipe().Return(testCase.stdoutPipeReturnValue, testCase.stdoutPipeReturnErr)
		}

		if testCase.startShouldBeCalled {
			mockSpawnFunc.EXPECT().Start().Return(testCase.startReturnErr)
		}

		// Wait() is executed within the expect.Expect.waitForSession(...) function, and is done so through a separate
		// goroutine.  We can't make any expectations of this thread, as doing so is prone to race conditions.  Take
		// the simple way out, and just allow Wait() to be invoked any number of times.
		mockSpawnFunc.EXPECT().Wait().AnyTimes()

		// Command is always called...
		mockSpawnFunc.EXPECT().Command(testCase.goExpectSpawnerSpawnCommand, testCase.goExpectSpawnerSpawnArgs).Return(&sFunc)

		goExpectSpawner := interactive.NewGoExpectSpawner()
		context, err := goExpectSpawner.Spawn(testCase.goExpectSpawnerSpawnCommand, testCase.goExpectSpawnerSpawnArgs, testCase.goExpectSpawnerSpawnTimeout, testCase.goExpectSpawnerSpawnOpts...)
		assert.Equal(t, testCase.goExpectSpawnerSpawnReturnErr, err)
		assert.Equal(t, testCase.goExpectSpawnerSpawnReturnContextIsNil, context == nil)
	}
}

// Also tests GetExpecter() and GetErrorChannel().
func TestNewContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockExpecter := mock_interactive.NewMockExpecter(ctrl)
	var errorChannel <-chan error
	var expecter expect.Expecter = mockExpecter
	context := interactive.NewContext(&expecter, errorChannel)
	assert.Equal(t, &expecter, context.GetExpecter())
	assert.Equal(t, errorChannel, context.GetErrorChannel())
}

func TestExecSpawnFunc(t *testing.T) {
	execSpawnFunc := interactive.ExecSpawnFunc{}
	cmd := execSpawnFunc.Command("pwd")
	assert.NotNil(t, cmd)

	stdin, err := (*cmd).StdinPipe()
	assert.Nil(t, err)
	assert.NotNil(t, stdin)

	stdout, err := (*cmd).StdoutPipe()
	assert.Nil(t, err)
	assert.NotNil(t, stdout)

	err = (*cmd).Start()
	assert.Nil(t, err)

	err = (*cmd).Wait()
	assert.Nil(t, err)
}
