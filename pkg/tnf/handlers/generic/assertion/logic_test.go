package assertion_test

import (
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/assertion"
	"regexp"
)

type booleanLogicTestCase struct {
	assertions                    []assertion.Assertion
	match                         string
	regex                         regexp.Regexp
	expectedType                  string
	expectedAndBooleanLogicResult bool
	expectedAndBooleanLogicError  bool
	expectedOrBooleanLogicResult  bool
	expectedOrBooleanLogicError   bool
}

var andBooleanLogicTestCases = map[string]booleanLogicTestCase{
	"Positive Test:  single true assertion": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  0,
				Condition: &equalsSomeThingCondition,
			},
		},
		match:                         "Something",
		regex:                         *regexp.MustCompile(`Something`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: true,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  true,
		expectedOrBooleanLogicError:   false,
	},
	"Positive Test:  two true assertions": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1,
				Condition: &equalsSomeThingCondition,
			},
			{
				GroupIdx:  2,
				Condition: &equalsAnotherThingCondition,
			},
		},
		match:                         "Something AnotherThing",
		regex:                         *regexp.MustCompile(`(\w+) (\w+)`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: true,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  true,
		expectedOrBooleanLogicError:   false,
	},
	"Positive Test:  mixed assertion results": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1,
				Condition: &equalsSomeThingCondition,
			},
			{
				GroupIdx:  2,
				Condition: &equalsAnotherThingCondition,
			},
		},
		match:                         "Something NotAnotherThing",
		regex:                         *regexp.MustCompile(`(\w+) (\w+)`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: false,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  true,
		expectedOrBooleanLogicError:   false,
	},
	"Positive Test:  both assertions are false": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1,
				Condition: &equalsSomeThingCondition,
			},
			{
				GroupIdx:  2,
				Condition: &equalsAnotherThingCondition,
			},
		},
		match:                         "NotSomething NotAnotherThing",
		regex:                         *regexp.MustCompile(`(\w+) (\w+)`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: false,
		expectedAndBooleanLogicError:  false,
		expectedOrBooleanLogicResult:  false,
		expectedOrBooleanLogicError:   false,
	},
	"Negative Test:  index out of bounds": {
		assertions: []assertion.Assertion{
			{
				GroupIdx:  1000,
				Condition: &equalsSomeThingCondition,
			},
		},
		match:                         "Something",
		regex:                         *regexp.MustCompile(`Something`),
		expectedType:                  assertion.AndBooleanLogicKey,
		expectedAndBooleanLogicResult: false,
		expectedAndBooleanLogicError:  true,
		expectedOrBooleanLogicResult:  false,
		expectedOrBooleanLogicError:   true,
	},
}
