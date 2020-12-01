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
	// OrBooleanLogicKey is the sentinel used for Type to identify OrBooleanLogicKey in a JSON/YAML payload.
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
		if assertionResult {
			return true, nil
		}
	}
	return false, nil
}
