package assertion

import (
	"regexp"
)

const (
	OrBooleanLogicKey = "or"
)

// OrBooleanLogic is an implementation of the BooleanLogic interface.  OrBooleanLogic dictates that at least one
// assertion in a Assertions evaluates as true.  Although OrBooleanLogic is exposed for serialization purposes,
// it is recommended to instantiate instances of OrBooleanLogic using NewOrBooleanLogic.
type OrBooleanLogic struct {
	// Type stores the sentinel which represents the type of BooleanLogic implemented.
	Type string `json:"type" yaml:"type"`
}

// NewOrBooleanLogic creates an instance of OrBooleanLogic.
func NewOrBooleanLogic() *OrBooleanLogic {
	return &OrBooleanLogic{Type: OrBooleanLogicKey}
}

// Evaluate evaluates an arbitrarily sized array of Assertion and ensures at least one assertion passes.
func (o OrBooleanLogic) Evaluate(assertions []Assertion, match string, regex regexp.Regexp) (bool, error) {
	for _, assertion := range assertions {
		assertionResult, err := (*assertion.Condition).Evaluate(match, regex, assertion.GroupIdx)
		if err != nil {
			return false, err
		}
		if assertionResult == true {
			return true, nil
		}
	}
	return false, nil
}
