package assertion_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/assertion"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition/stringcondition"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	equalsAnotherThingStringCondition = stringcondition.NewEqualsCondition("AnotherThing")
	equalsSomethingStringCondition    = stringcondition.NewEqualsCondition("Something")
)

var equalsAnotherThingCondition condition.Condition = equalsAnotherThingStringCondition
var equalsSomeThingCondition condition.Condition = equalsSomethingStringCondition

func TestNewAndBooleanLogic(t *testing.T) {
	logic := assertion.NewAndBooleanLogic()
	assert.Equal(t, assertion.AndBooleanLogicKey, logic.Type)
}

func TestAndBooleanLogic_Evaluate(t *testing.T) {
	for _, testCase := range andBooleanLogicTestCases {
		logic := assertion.NewAndBooleanLogic()
		actualResult, actualError := logic.Evaluate(testCase.assertions, testCase.match, testCase.regex)
		assert.Equal(t, assertion.AndBooleanLogicKey, logic.Type)
		assert.Equal(t, testCase.expectedAndBooleanLogicResult, actualResult)
		assert.Equal(t, testCase.expectedAndBooleanLogicError, actualError != nil)
	}
}
