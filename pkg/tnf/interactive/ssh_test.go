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

var sshSomeSpawnError = errors.New("some error related to spawning SSH")

type sshTestCase struct {
	user               string
	host               string
	options            []expect.Option
	errReturnValue     error
	contextReturnValue *interactive.Context
	expectedSpawnErr   error
}

var sshTestCases = map[string]sshTestCase{
	"no_error": {
		options:            []expect.Option{expect.Verbose(true)},
		errReturnValue:     nil,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   nil,
	},
	"error": {
		options:            []expect.Option{expect.Verbose(true)},
		errReturnValue:     sshSomeSpawnError,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   sshSomeSpawnError,
	},
}

func TestSpawnSsh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, testCase := range sshTestCases {
		ctrl = gomock.NewController(t)
		mockSpawner := mock_interactive.NewMockSpawner(ctrl)
		mockSpawner.EXPECT().Spawn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testCase.contextReturnValue, testCase.errReturnValue)

		var spawner interactive.Spawner = mockSpawner
		_, err := interactive.SpawnSsh(&spawner, testCase.user, testCase.host, ocTestTimeoutDuration, testCase.options...)
		assert.Equal(t, testCase.expectedSpawnErr, err)
	}
}
