package assertion_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/assertion"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewOrBooleanLogic(t *testing.T) {
	logic := assertion.NewOrBooleanLogic()
	assert.Equal(t, assertion.OrBooleanLogicKey, logic.Type)
}

func TestOrBooleanLogic_Evaluate(t *testing.T) {
	for _, testCase := range andBooleanLogicTestCases {
		logic := assertion.NewOrBooleanLogic()
		actualResult, actualError := logic.Evaluate(testCase.assertions, testCase.match, testCase.regex)
		assert.Equal(t, assertion.OrBooleanLogicKey, logic.Type)
		assert.Equal(t, testCase.expectedOrBooleanLogicResult, actualResult)
		assert.Equal(t, testCase.expectedOrBooleanLogicError, actualError != nil)
	}
}
