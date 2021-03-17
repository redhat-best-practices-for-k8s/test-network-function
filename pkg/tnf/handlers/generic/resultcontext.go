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

package generic

import (
	"encoding/json"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/assertion"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

// ResultContext evaluates the Result for a given Match.  If ComposedAssertions is not supplied, then Result is assigned
// to the reel.Handler result.  If ComposedAssertions is supplied, then the ComposedAssertions are evaluated against the
// match.  The result of ComposedAssertions evaluation is assigned to the reel.Handler's result.
type ResultContext struct {

	// Pattern is the pattern causing a match in reel.Handler ReelMatch.
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`

	// ComposedAssertions is a means of making many assertion.Assertion claims about the match.
	ComposedAssertions []assertion.Assertions `json:"composedAssertions,omitempty" yaml:"composedAssertions,omitempty"`

	// DefaultResult is the result of the test.  This is only used if ComposedAssertions is not provided.
	DefaultResult int `json:"defaultResult,omitempty" yaml:"defaultResult,omitempty"`

	// NextStep is an optional next step to take after an initial ReelMatch.
	NextStep *reel.Step `json:"nextStep,omitempty" yaml:"nextStep,omitempty"`

	// NextResultContexts is an optional array which provides the ability to make assertion.Assertions based on the next pattern match.
	NextResultContexts []*ResultContext `json:"nextResultContexts,omitempty" yaml:"nextResultContexts,omitempty"`
}

// MarshalJSON is a shim provided over the default implementation that omits empty NextResultContexts slices.  This
// custom MarshallJSON implementation is needed due to a recursive definition (type ResultContext has a property of type
// ResultContext).
func (r *ResultContext) MarshalJSON() ([]byte, error) {
	if len(r.NextResultContexts) == 0 {
		return json.Marshal(&struct {
			Pattern            string                 `json:"pattern,omitempty"`
			ComposedAssertions []assertion.Assertions `json:"composedAssertions,omitempty"`
			DefaultResult      int                    `json:"defaultResult"`
			NextStep           *reel.Step             `json:"nextStep,omitempty"`
		}{
			Pattern:            r.Pattern,
			ComposedAssertions: r.ComposedAssertions,
			DefaultResult:      r.DefaultResult,
			NextStep:           r.NextStep,
		})
	}

	// Normally, you would just augment the struct here by adding the missing NextResultContexts field.  However, since
	// NextResultContexts is recursive (i.e., it is a ResultContext), doing so causes a loop.  Thus, this requires a
	// more robust definition.
	return json.Marshal(&struct {
		Pattern            string                 `json:"pattern,omitempty"`
		ComposedAssertions []assertion.Assertions `json:"composedAssertions,omitempty"`
		DefaultResult      int                    `json:"defaultResult"`
		NextStep           *reel.Step             `json:"nextStep,omitempty"`
		NextResultContexts []*ResultContext       `json:"nextResultContexts,omitempty"`
	}{
		Pattern:            r.Pattern,
		ComposedAssertions: r.ComposedAssertions,
		DefaultResult:      r.DefaultResult,
		NextStep:           r.NextStep,
		NextResultContexts: r.NextResultContexts,
	})
}
