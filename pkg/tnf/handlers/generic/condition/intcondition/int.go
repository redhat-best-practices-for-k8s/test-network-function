package intcondition

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	// Equal is the sentinel key identifying an int == comparison.
	Equal = "=="
	// GreaterThan is the sentinel key identifying the int > comparison.
	GreaterThan = ">"
	// GreaterThanOrEqual is the sentinel key identifying the int >= comparison.
	GreaterThanOrEqual = ">="
	// ComparisonConditionKey is the sentinel key indicating the condition type is an int comparison.
	ComparisonConditionKey = "intComparison"
	// IsIntConditionKey is the sentinel key identifying that the condition is checking that the argument is an integer.
	IsIntConditionKey = "isInt"
	// LessThan is the sentinel key identifying the int < comparison.
	LessThan = "<"
	// LessThanOrEqual is the sentinel key identifying the int <= comparison.
	LessThanOrEqual = "<="
	// NotEqual is the sentinel key identifying the int != comparison.
	NotEqual = "!="
)

// IsIntCondition is an implementation of the condition.Condition interface which evaluates whether a match string is an
// integer.  Although IsIntCondition is exported for serialization purposes, it is recommended to instantiate new
// instances of IsIntCondition using NewIsIntCondition.
type IsIntCondition struct {
	// Type stores the sentinel which represents the type of Condition implemented.
	Type string `json:"type" yaml:"type"`
}

// NewIsIntCondition creates an IsIntCondition.
func NewIsIntCondition() *IsIntCondition {
	return &IsIntCondition{Type: IsIntConditionKey}
}

// Evaluate evaluates whether a match string is an integer.
func (i IsIntCondition) Evaluate(match string, regex regexp.Regexp, matchIdx int) (bool, error) {
	matches := regex.FindStringSubmatch(match)
	if len(matches) < matchIdx {
		return false, fmt.Errorf("matches \"%s\" has no index: %d", matches, matchIdx)
	}
	foundMatch := matches[matchIdx]
	_, err := strconv.Atoi(foundMatch)
	return err == nil, err
}

// ComparisonCondition is an implementation of the condition.Condition interface which converts a match string to an
// integer, then checks integer equality against Input.
type ComparisonCondition struct {
	Type       string `json:"type" yaml:"type"`
	Input      int    `json:"input" yaml:"input"`
	Comparison string `json:"comparison" yaml:"comparison"`
}

// NewComparisonCondition creates an ComparisonCondition.
func NewComparisonCondition(input int, comparison string) *ComparisonCondition {
	return &ComparisonCondition{Type: ComparisonConditionKey, Input: input, Comparison: comparison}
}

// Evaluate evaluates whether a match can be converted to an integer, then performs an integer equality test against
// Input.
func (i ComparisonCondition) Evaluate(match string, regex regexp.Regexp, matchIdx int) (bool, error) {
	matches := regex.FindStringSubmatch(match)
	if len(matches) < matchIdx {
		return false, fmt.Errorf("matches \"%s\" has no index: %d", matches, matchIdx)
	}
	foundMatch := matches[matchIdx]
	val, err := strconv.Atoi(foundMatch)
	if err != nil {
		return false, fmt.Errorf("match \"%s\" cannot be converted to an integer", foundMatch)
	}
	return i.evaluateComparison(val)
}

// evaluateComparison does the comparison evaluation based on the supported comparative operators.
func (i ComparisonCondition) evaluateComparison(actual int) (bool, error) {
	switch i.Comparison {
	case Equal:
		return actual == i.Input, nil
	case LessThan:
		return actual < i.Input, nil
	case LessThanOrEqual:
		return actual <= i.Input, nil
	case GreaterThan:
		return actual > i.Input, nil
	case GreaterThanOrEqual:
		return actual >= i.Input, nil
	case NotEqual:
		return actual != i.Input, nil
	default:
		return false, errors.New(fmt.Sprintf("unknown comparative operator: %s", i.Comparison))
	}
}
