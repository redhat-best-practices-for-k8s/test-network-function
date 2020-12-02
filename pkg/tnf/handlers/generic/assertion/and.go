// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

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
func (a AndBooleanLogic) Evaluate(assertions []Assertion, match string, regex *regexp.Regexp) (bool, error) {
	// TODO This could be multi-threaded, but is unlikely worth doing from a risk-reward standpoint.
	for _, assertion := range assertions {
		assertionResult, err := (*assertion.Condition).Evaluate(match, regex, assertion.GroupIdx)
		// exit early if the condition is false or an error is encountered
		if !assertionResult || err != nil {
			return assertionResult, err
		}
	}
	return true, nil
}
