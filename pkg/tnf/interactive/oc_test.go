package interactive_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive"
	mock_interactive "github.com/redhat-nfvpe/test-network-function/pkg/tnf/interactive/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	ocTestTimeoutDuration = time.Second * 2
)

var ocSomeSpawnError = errors.New("some error related to spawning OC")

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
		errReturnValue:     ocSomeSpawnError,
		contextReturnValue: &interactive.Context{},
		expectedSpawnErr:   ocSomeSpawnError,
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
