package tnf_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	mock_interactive "github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive/mocks"
	mock_tnf "github.com/redhat-nfvpe/test-network-function/pkg/tnf/mocks"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/reel"
	mock_reel "github.com/redhat-nfvpe/test-network-function/pkg/tnf/reel/mocks"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

const (
	testTimeoutDuration = time.Second * 2
)

var (
	defaultTestCommand = []string{"ls"}
	sendError          = errors.New("generic send error")
	runnerError        = errors.New("runner run error")
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
		sendReturnErr:   sendError,
		newTestErr:      sendError,
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
		testTest, err := tnf.NewTest(&expecter, mockTester, []reel.Handler{mockHandler}, errorChannel)
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
			Expect:  []string{`someOutput`},
			Timeout: testTimeoutDuration,
		},
		testerResultResult:  tnf.ERROR,
		expectBatchIsCalled: true,
		expectBatchBatchResResult: []expect.BatchRes{
			{Idx: 0, Output: "someOutput", Match: []string{"someOutput"}},
		},
		expectBatchBatchResErr: nil,
		reelMatchIsCalled:      true,
		reelMatchPattern:       "",
		reelMatchBefore:        "",
		reelMatchMatch:         "someOutput",
		reelMatchResult:        nil,
	},
}

// Also covers ReelFirst() and ReelMatch().  Tests those state transitions.
func TestTest_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range testRunTestCases {
		mockExpecter := mock_interactive.NewMockExpecter(ctrl)
		testCommand := strings.Join(testCase.testCommandArgs, " ") + "\n"
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
		// Since we have no control over the t.runner, just make the assertion that err is nil.  In these cases, it
		// always should be nil, as it is mocked.
		assert.Nil(t, err)
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
