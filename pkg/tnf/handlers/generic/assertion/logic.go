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
