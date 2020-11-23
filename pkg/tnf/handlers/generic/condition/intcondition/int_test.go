package intcondition_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition/intcondition"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestNewIsIntCondtion(t *testing.T) {
	c := intcondition.NewIsIntCondition()
	assert.NotNil(t, c)
	assert.Equal(t, intcondition.IsIntConditionKey, c.Type)
}

type isIntEvaluationTestCase struct {
	match          string
	regex          regexp.Regexp
	matchIdx       int
	expectedType   string
	expectedResult bool
	expectedError  bool
}

var isIntConditionTestCases = map[string]isIntEvaluationTestCase{
	"working": {
		match:          "this is message 1",
		regex:          *regexp.MustCompile(`[^\d]+(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.IsIntConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"notAnInt": {
		match:          "not an int",
		regex:          *regexp.MustCompile(`\w+\s(\w+)`),
		matchIdx:       1,
		expectedType:   intcondition.IsIntConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
	"indexOOB": {
		match:          "out of bounds",
		regex:          *regexp.MustCompile(`.+`),
		matchIdx:       4,
		expectedType:   intcondition.IsIntConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
}

func TestIsIntCondition_Evaluate(t *testing.T) {
	for _, testCase := range isIntConditionTestCases {
		c := intcondition.NewIsIntCondition()
		actualResult, actualError := c.Evaluate(testCase.match, testCase.regex, testCase.matchIdx)
		assert.Equal(t, testCase.expectedType, c.Type)
		assert.Equal(t, testCase.expectedResult, actualResult)
		assert.Equal(t, testCase.expectedError, actualError != nil)
	}
}

func TestNewIntComparisonCondition(t *testing.T) {
	c := intcondition.NewComparisonCondition(1, "2")
	assert.NotNil(t, c)
	assert.Equal(t, intcondition.ComparisonConditionKey, c.Type)
}

type intComparisonTestCase struct {
	input          int
	comparison     string
	match          string
	regex          regexp.Regexp
	matchIdx       int
	expectedType   string
	expectedResult bool
	expectedError  bool
}

var intComparisonTestCases = map[string]intComparisonTestCase{
	"Positive Test: testEqual_True": {
		input:          1,
		comparison:     intcondition.Equal,
		match:          "does 1 equal 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testEqual_False": {
		input:          2,
		comparison:     intcondition.Equal,
		match:          "does 1 equal 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Positive Test: testLessThan_True": {
		input:          2,
		comparison:     intcondition.LessThan,
		match:          "is 1 less than 2?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testLessThan_False": {
		input:          1,
		comparison:     intcondition.LessThan,
		match:          "is 1 less than 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Positive Test: testLessThanOrEqual_True_LessThan": {
		input:          2,
		comparison:     intcondition.LessThanOrEqual,
		match:          "is 1 less than or equal to 2?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testLessThanOrEqual_True_Equal": {
		input:          1,
		comparison:     intcondition.LessThanOrEqual,
		match:          "is 1 less than or equal to 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testLessThanOrEqual_False": {
		input:          0,
		comparison:     intcondition.LessThanOrEqual,
		match:          "is 1 less than or equal to 0?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Positive Test: testGreaterThan_True": {
		input:          0,
		comparison:     intcondition.GreaterThan,
		match:          "is 1 greater than 0?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testGreaterThan_False": {
		input:          1,
		comparison:     intcondition.LessThan,
		match:          "is 1 greater than 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Positive Test: testGreaterThanOrEqual_True_GreaterThan": {
		input:          0,
		comparison:     intcondition.GreaterThanOrEqual,
		match:          "is 1 greater than or equal to 0?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testGreaterThanOrEqual_True_Equal": {
		input:          1,
		comparison:     intcondition.GreaterThanOrEqual,
		match:          "is 1 greater than or equal to 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testGreaterThanOrEqual_False": {
		input:          2,
		comparison:     intcondition.GreaterThanOrEqual,
		match:          "is 1 greater than or equal to 2?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Positive Test: testNotEqual_True": {
		input:          2,
		comparison:     intcondition.NotEqual,
		match:          "is 1 no equal to 2?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: true,
		expectedError:  false,
	},
	"Positive Test: testNotEqual_False": {
		input:          1,
		comparison:     intcondition.NotEqual,
		match:          "is 1 no equal to 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  false,
	},
	"Negative Test: testBadCondition": {
		input:          1,
		comparison:     "badcondition",
		match:          "is 1 badcondition to 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
	"Negative Test: testNonIntInput": {
		input:          1,
		comparison:     intcondition.Equal,
		match:          "is nonint equal to 1?",
		regex:          *regexp.MustCompile(`\w+\s(\w+)`),
		matchIdx:       1,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
	"Negative Test: matchIdxOutOfBounds": {
		input:          1,
		comparison:     intcondition.Equal,
		match:          "is 1 equal to 1?",
		regex:          *regexp.MustCompile(`\w+\s(\d+)`),
		matchIdx:       5000,
		expectedType:   intcondition.ComparisonConditionKey,
		expectedResult: false,
		expectedError:  true,
	},
}

func TestIntComparisonCondition_Evaluate(t *testing.T) {
	for _, testCase := range intComparisonTestCases {
		c := intcondition.NewComparisonCondition(testCase.input, testCase.comparison)
		actualResult, actualError := c.Evaluate(testCase.match, testCase.regex, testCase.matchIdx)
		assert.Equal(t, testCase.expectedType, intcondition.ComparisonConditionKey)
		assert.Equal(t, testCase.expectedResult, actualResult)
		assert.Equal(t, testCase.expectedError, actualError != nil)
	}
}
