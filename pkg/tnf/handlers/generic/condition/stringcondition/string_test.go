package stringcondition_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition/stringcondition"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestNewEqualsCondition(t *testing.T) {
	c := stringcondition.NewEqualsCondition("input")
	assert.Equal(t, stringcondition.EqualsConditionKey, c.Type)
	assert.Equal(t, "input", c.Expected)
}

type equalsConditionTestCase struct {
	expected       string
	match          string
	regex          regexp.Regexp
	matchIdx       int
	expectedType   string
	expectedResult bool
	expectedError  bool
}

var equalsCondtionTestCases = map[string]equalsConditionTestCase{
	"Positive Case: strings_are_equal": {
		expected:       "apple",
		match:          "apple tree",
		regex:          *regexp.MustCompile(`(\w+) tree`),
		matchIdx:       1,
		expectedType:   stringcondition.EqualsConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Case: strings_are_not_equal": {
		expected:       "apple",
		match:          "notapple tree",
		regex:          *regexp.MustCompile(`(\w+) tree`),
		matchIdx:       1,
		expectedType:   stringcondition.EqualsConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Negative Case: index_out_of_bounds": {
		expected:       "apple",
		match:          "apple tree",
		regex:          *regexp.MustCompile(`(\w+) tree`),
		matchIdx:       100,
		expectedType:   stringcondition.EqualsConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
}

func TestEqualsCondition_Evaluate(t *testing.T) {
	for _, testCase := range equalsCondtionTestCases {
		c := stringcondition.NewEqualsCondition(testCase.expected)
		actualResult, actualError := c.Evaluate(testCase.match, testCase.regex, testCase.matchIdx)
		assert.Equal(t, testCase.expectedType, stringcondition.EqualsConditionKey)
		assert.Equal(t, testCase.expectedResult, actualResult)
		assert.Equal(t, testCase.expectedError, actualError != nil)
	}
}
