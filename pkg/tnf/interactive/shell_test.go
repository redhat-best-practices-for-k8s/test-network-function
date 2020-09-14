package interactive_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	mock_interactive "github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var shellSomeSpawnError = errors.New("some error related to spawning shell")

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
		errReturnValue:     shellSomeSpawnError,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   shellSomeSpawnError,
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
