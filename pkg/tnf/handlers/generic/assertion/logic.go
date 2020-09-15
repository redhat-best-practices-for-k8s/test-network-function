package assertion

import (
	"regexp"
)

const (
	// LogicKey is the JSON key used to represent a BooleanLogic payload.
	LogicKey = "logic"
	// TypeKey is the JSON key used to represent a BooleanLogic type.
	TypeKey = "type"
)

// BooleanLogic represents boolean logic.  Given a set of conditions, it is useful to make assertions over the set using
// some sort of boolean logic ("and" and "or", for example).
type BooleanLogic interface {

	// Evaluate evaluates assertions against a match using an implementation-dependent BooleanLogic.  Evaluate returns
	// whether or not the expression evaluated to true, and an optional error if the expression could not be evaluated
	// properly.  An example of an error case might be making an assertion for a match index that does not exist.
	Evaluate(assertions []Assertion, match string, regex regexp.Regexp) (bool, error)
}
