package stringcondition

import (
	"fmt"
	"regexp"
)

const (
	EqualsConditionKey = "equals"
)

// EqualsCondition is an implementation of the condition.Condition interface which evaluates string equality of a match
// against Expected.  Although EqualsCondition is exported for serialization reasons, it is recommended to instantiate
// new instances of EqualsCondition using NewEqualsCondition.
type EqualsCondition struct {
	// Type stores the sentinel which represents the type of Condition implemented.
	Type string `json:"type" yaml:"type"`
	// Expected is the expected string value.
	Expected string `json:"expected,omitempty" yaml:"expected,omitempty"`
}

// NewEqualsCondition creates an EqualsCondition.
func NewEqualsCondition(expected string) *EqualsCondition {
	return &EqualsCondition{Type: EqualsConditionKey, Expected: expected}
}

// Evaluate evaluates string equality for a match against Expected.
func (e EqualsCondition) Evaluate(match string, regex regexp.Regexp, matchIdx int) (bool, error) {
	matches := regex.FindStringSubmatch(match)
	if len(matches) < matchIdx {
		return false, fmt.Errorf("matches \"%s\" has no index: %d", matches, matchIdx)
	}
	foundMatch := matches[matchIdx]
	return e.Expected == foundMatch, nil
}
