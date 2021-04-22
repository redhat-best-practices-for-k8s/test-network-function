// Copyright (C) 2020-2021 Red Hat, Inc.
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

package stringcondition

import (
	"fmt"
	"regexp"
)

const (
	// EqualsConditionKey is the sentinel key identifying a string == comparison.
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
func (e EqualsCondition) Evaluate(match string, regex *regexp.Regexp, matchIdx int) (bool, error) {
	matches := regex.FindStringSubmatch(match)
	if len(matches) < matchIdx {
		return false, fmt.Errorf("matches \"%s\" has no index: %d", matches, matchIdx)
	}
	foundMatch := matches[matchIdx]
	return e.Expected == foundMatch, nil
}
