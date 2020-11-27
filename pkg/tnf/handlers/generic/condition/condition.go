// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
