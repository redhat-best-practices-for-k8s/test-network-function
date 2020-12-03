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
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"path"
	"regexp"
	"time"
)

const (
	genericTestSchemaFileName = "generic-test.schema.json"
)

var (
	genericTestSchemaFileLocation = path.Join("..", "..", "..", "..", "schemas", genericTestSchemaFileName)
)

// Generic is a construct for defining an arbitrary simple test with prescriptive confines.  Essentially, the definition
// of the state machine for a Generic reel.Handler is restricted in this facade, since most common use cases do not need
// to perform too much heavy lifting that would otherwise require a Custom reel.Handler implementation.  Although
// Generic is exported for serialization reasons, it is recommended to instantiate new instances of Generic using
// NewGenericFromJSONFile, is tailored to properly initialize a Generic.
type Generic struct {

	// Arguments is the Unix command array.  Arguments is optional;  a command can also be issued using ReelFirstStep.
	Arguments []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`

	// Description is a textual description of the overall functionality that is tested.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// FailureReason optionally stores extra information pertaining to why the test failed.
	FailureReason string `json:"failureReason,omitempty" yaml:"failureReason,omitempty"`

	// Matches contains an in order array of matches.
	Matches []Match `json:"matches,omitempty" yaml:"matches,omitempty"`

	// ReelFirstStep is the first Step returned by reel.ReelFirst().
	ReelFirstStep *reel.Step `json:"reelFirstStep,omitempty" yaml:"reelFirstStep,omitempty"`

	// ReelFirstStep is the first Step returned by reel.ReelFirst().
	ReelMatchStep *reel.Step `json:"reelMatchStep,omitempty" yaml:"reelMatchStep,omitempty"`

	// ResultContexts provides the ability to make assertion.Assertions based on the given pattern matched.
	ResultContexts []*ResultContext `json:"resultContexts,omitempty" yaml:"resultContexts,omitempty"`

	// reelMatchResultMap is an internal construct used to save time on lookups.  Since evaluation order of reel.Step
	// Expect regular expressions is important, the end user should define the order (ResultContexts) and realize that
	// the evaluating each regular expression is O(n).  However, when making lookups after the fact, the match pattern
	// has already been found, so ordering does not matter.  This solution duplicates data, but utilizing extra RAM on
	// the bastion server is not a concern.  Performance is favored over memory frugality.
	reelMatchResultMap map[string]int

	// ReelTimeoutStep is the reel.Step to take upon timeout.
	ReelTimeoutStep *reel.Step `json:"reelTimeoutStep,omitempty" yaml:"reelTimeoutStep,omitempty"`

	// TestResult is the result of running the tnf.Test.  0 indicates SUCCESS, 1 indicates FAILURE, 2 indicates ERROR.
	TestResult int `json:"testResult" yaml:"testResult"`

	// TestTimeout prevents the Test from running forever.
	TestTimeout time.Duration `json:"testTimeout,omitempty" yaml:"testTimeout,omitempty"`

	// currentReelMatchResultContexts is used to persist the current ResultContext over multiple invocations of ReelMatch.
	currentReelMatchResultContexts []*ResultContext
}

// init initializes a Generic, including building up the reelMatchResultMap.  reelMatchResultMap is pre-built for
// performance reasons.
func (g *Generic) init() {
	g.currentReelMatchResultContexts = g.ResultContexts
	g.reelMatchResultMap = map[string]int{}
	for _, resultContext := range g.currentReelMatchResultContexts {
		g.reelMatchResultMap[resultContext.Pattern] = resultContext.DefaultResult
	}
}

// Args returns the command line arguments as an array of type string.
func (g *Generic) Args() []string {
	return g.Arguments
}

// Timeout returns the test timeout.
func (g *Generic) Timeout() time.Duration {
	return g.TestTimeout
}

// Result returns the test result.
func (g *Generic) Result() int {
	return g.TestResult
}

// ReelFirst returns the first step to perform.
func (g *Generic) ReelFirst() *reel.Step {
	return g.ReelFirstStep
}

// findResultContext is an internal helper function used to search an array of ResultContext instances for a given
// pattern.  Since order of ResultContext is important, this operation is O(n).
func (g *Generic) findResultContext(pattern string) *ResultContext {
	for _, context := range g.currentReelMatchResultContexts {
		if context.Pattern == pattern {
			return context
		}
	}
	return nil
}

// ReelMatch informs of a match event, returning the next step to perform.
func (g *Generic) ReelMatch(pattern string, before string, match string) *reel.Step {
	m := &Match{Pattern: pattern, Before: before, Match: match}
	g.Matches = append(g.Matches, *m)

	resultContext := g.findResultContext(pattern)
	if resultContext == nil {
		g.FailureReason = "the pattern provided to ReelMatch is not defined in ReelFirst"
		g.TestResult = tnf.ERROR
		return nil
	}
	composedAssertions := resultContext.ComposedAssertions
	if len(composedAssertions) > 0 {
		for _, composedAssertion := range composedAssertions {
			regex := regexp.MustCompile(pattern)
			success, err := (*composedAssertion.Logic).Evaluate(composedAssertion.Assertions, match, *regex)
			if err != nil {
				// exit immediately on a test error.
				g.FailureReason = err.Error()
				g.TestResult = tnf.ERROR
				return nil
			} else if !success {
				// exit immediately on failure
				g.TestResult = tnf.FAILURE
				return nil
			}
			// only report success if nothing else is left
			if resultContext.NextStep == nil {
				g.TestResult = tnf.SUCCESS
				return nil
			}
		}
	}

	// Else, see if we have more work to do.  If not, return defaultResult.
	if resultContext.NextStep == nil {
		g.TestResult = resultContext.DefaultResult
		return nil
	}

	g.currentReelMatchResultContexts = resultContext.NextResultContexts
	return resultContext.NextStep
}

// ReelTimeout informs of a timeout event, returning the next step to perform.
func (g *Generic) ReelTimeout() *reel.Step {
	return g.ReelTimeoutStep
}

// ReelEOF informs of the eof event.
func (g *Generic) ReelEOF() {
	// do nothing.
}

// NewGenericFromJSONFile instantiates and initializes a Generic from a JSON-serialized file.
func NewGenericFromJSONFile(filename string) (*tnf.Tester, []reel.Handler, *gojsonschema.Result, error) {
	g, result, err := createGeneric(filename)
	if err != nil {
		return nil, nil, result, err
	}
	// poor man's polymorphism
	var tester tnf.Tester = g
	var handler reel.Handler = g
	return &tester, []reel.Handler{handler}, result, nil
}

// validateJSONTestAgainstSchema validates an input schema against the generic-test.schema.json JSON schema.
func validateJSONTestAgainstSchema(filename string) (*gojsonschema.Result, error) {
	schemaBytes, err := ioutil.ReadFile(genericTestSchemaFileLocation)
	if err != nil {
		return nil, err
	}
	schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))

	inputBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	inputLoader := gojsonschema.NewStringLoader(string(inputBytes))

	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, err
	}
	return schema.Validate(inputLoader)
}

// createGeneric is a helper function for instantiating and initializing a Generic tnf.Test.
func createGeneric(filename string) (*Generic, *gojsonschema.Result, error) {
	result, err := validateJSONTestAgainstSchema(filename)
	if err != nil {
		return nil, result, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, result, err
	}

	g := &Generic{}
	err = json.Unmarshal(bytes, g)
	if err != nil {
		return nil, result, err
	}
	g.init()
	return g, result, nil
}
