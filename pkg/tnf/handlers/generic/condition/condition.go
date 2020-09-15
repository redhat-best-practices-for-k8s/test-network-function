package condition

import "regexp"

const (
	// TypeKey is the JSON key indicating a condition payload.
	TypeKey = "type"
)

// Condition represents the polymorphic behavior of the ability to Evaluate.  Given a regular expression, match, and
// match index, it is useful to make assertions using Condition implementations.  For example, if a Ping test returns a
// matching summary, it is convenient to evaluate that summary indicates zero errors.
type Condition interface {

	// Evaluate evaluates a Condition implementation for groupIdx group of a matched expression.
	Evaluate(match string, regex regexp.Regexp, groupIdx int) (bool, error)
}
