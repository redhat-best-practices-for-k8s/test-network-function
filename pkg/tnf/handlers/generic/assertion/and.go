package assertion

import (
	"regexp"
)

const (
	// AndBooleanLogicKey is the sentinel used for Type to identify AndBooleanLogic in a JSON/YAML payload.
	AndBooleanLogicKey = "and"
)

// AndBooleanLogic is an implementation of the BooleanLogic interface.  AndBooleanLogic dictates that all assertions in
// a assertion.Assertions must evaluate as true.  Although AndBooleanLogic is exposed for serialization purposes,
// it is recommended to instantiate instances of AndBooleanLogic using NewAndBooleanLogic.
type AndBooleanLogic struct {
	// Type stores the sentinel which represents the type of BooleanLogic implemented.
	Type string `json:"type" yaml:"type"`
}

// NewAndBooleanLogic creates an instance of AndBooleanLogic.
func NewAndBooleanLogic() *AndBooleanLogic {
	return &AndBooleanLogic{Type: AndBooleanLogicKey}
}

// Evaluate evaluates an arbitrarily sized array of Assertion and ensures each assertion passes.
func (a AndBooleanLogic) Evaluate(assertions []Assertion, match string, regex regexp.Regexp) (bool, error) {
	// TODO This could be multi-threaded, but is unlikely worth doing from a risk-reward standpoint.
	for _, assertion := range assertions {
		assertionResult, err := (*assertion.Condition).Evaluate(match, regex, assertion.GroupIdx)
		// exit early if the condition is false or an error is encountered
		if assertionResult == false || err != nil {
			return assertionResult, err
		}
	}
	return true, nil
}
