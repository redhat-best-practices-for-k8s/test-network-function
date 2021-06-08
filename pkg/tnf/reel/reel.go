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

package reel

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	"google.golang.org/grpc/codes"
)

const (
	// EndOfTestSentinel is the emulated terminal prompt that will follow command output.
	EndOfTestSentinel = `END_OF_TEST_SENTINEL`
)

var (
	// endOfTestSentinelCutset is used to trim a match of the EndOfTestSentinel.
	endOfTestSentinelCutset = fmt.Sprintf("%s\n", EndOfTestSentinel)
	// EndOfTestRegexPostfix is the postfix added to regular expressions to match the emulated terminal prompt
	// (EndOfTestSentinel)
	EndOfTestRegexPostfix = fmt.Sprintf("((.|\n)*%s\n)", EndOfTestSentinel)
)

// Step is an instruction for a single REEL pass.
// To process a step, first send the `Execute` string to the target subprocess (if supplied).  Block until the
// subprocess output to stdout matches one of the regular expressions in `Expect` (if any supplied). A positive integer
// `Timeout` prevents blocking forever.
type Step struct {
	// Execute is a Unix command to execute using the underlying subprocess.
	Execute string `json:"execute,omitempty" yaml:"execute,omitempty"`

	// Expect is an array of expected text regular expressions.  The first expectation results in a match.
	Expect []string `json:"expect,omitempty" yaml:"expect,omitempty"`

	// Timeout is the timeout for the Step.  A positive Timeout prevents blocking forever.
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// A utility method to return the important aspects of the Step container as a tuple.
func (s *Step) unpack() (execute string, expect []string, timeout time.Duration) { //nolint:gocritic // Ignoring shadowed name `expect`; it makes sense
	return s.Execute, s.Expect, s.Timeout
}

// Whether or not the Step has expectations.
func (s *Step) hasExpectations() bool {
	return len(s.Expect) > 0
}

// A Handler implements desired programmatic control.
type Handler interface {
	// ReelFirst returns the first step to perform.
	ReelFirst() *Step

	// ReelMatch informs of a match event, returning the next step to perform.  ReelMatch takes three arguments:
	// `pattern` represents the regular expression pattern which was matched.
	// `before` contains all output preceding `match`.
	// `match` is the text matched by `pattern`.
	ReelMatch(pattern string, before string, match string) *Step

	// ReelTimeout informs of a timeout event, returning the next step to perform.
	ReelTimeout() *Step

	// ReelEOF informs of the eof event.
	ReelEOF()
}

// StepFunc provides a wrapper around a generic Handler.
type StepFunc func(Handler) *Step

// Option is a function pointer to enable lightweight optionals for Reel.
type Option func(reel *Reel) Option

// A Reel instance allows interaction with a target subprocess.
type Reel struct {
	// A pointer to the underlying subprocess
	expecter *expect.Expecter
	Err      error
	// disableTerminalPromptEmulation determines whether terminal prompt emulation should be disabled.
	disableTerminalPromptEmulation bool
}

// DisableTerminalPromptEmulation disables terminal prompt emulation for the reel.Reel.
func DisableTerminalPromptEmulation() Option {
	return func(r *Reel) Option {
		r.disableTerminalPromptEmulation = true
		return DisableTerminalPromptEmulation()
	}
}

// Each Step can have zero or more expectations (Step.Expect).  This method follows the Adapter design pattern;  a raw
// array of strings is turned into a corresponding array of exepct.Batcher.  This method side-effects the input
// expectations array, following the Builder design pattern.  Finally, the first match is stored in the firstMatch
// output parameter.
func (r *Reel) batchExpectations(expectations []string, batcher []expect.Batcher, firstMatch *string) []expect.Batcher {
	if len(expectations) > 0 {
		expectCases := r.generateCases(expectations, firstMatch)
		batcher = append(batcher, &expect.BCas{C: expectCases})
	}
	return batcher
}

// Each Step can have zero or more expectations (Step.Expect), and if any match is found, then a match event occurs
// (representing a logical "OR" over the array). This helper follows the Adapter design pattern;  a raw array of string
// regular expressions is converted to an equivalent expect.Caser array.  The firstMatch parameter is used as an output
// parameter to store the first match found in the expectations array.  Thus, the order of expectations is important.
func (r *Reel) generateCases(expectations []string, firstMatch *string) []expect.Caser {
	var cases []expect.Caser
	for _, expectation := range expectations {
		thisCase := r.generateCase(expectation, firstMatch)
		cases = append(cases, thisCase)
	}
	return cases
}

// Each Step can have zero or more expectations (Step.Expect).  This method follows the Adapter design pattern;  a
// single raw string Expectation is converted into a corresponding expect.Case.
func (r *Reel) generateCase(expectation string, firstMatch *string) *expect.Case {
	expectation = r.addEmulatedRegularExpression(expectation)
	return &expect.Case{R: regexp.MustCompile(expectation), T: func() (expect.Tag, *expect.Status) {
		if *firstMatch == "" {
			*firstMatch = expectation
		}
		return expect.OKTag, expect.NewStatus(codes.OK, "state reached")
	}}
}

// Each Step can have exactly one execution string (Step.Execute).  This method follows the Adapter design pattern;  a
// single raw execution string is converted into a corresponding expect.Batcher.  The function returns an array of
// expect.Batcher, as it is expected that there are likely expectations to follow.
func (r *Reel) generateBatcher(execute string) []expect.Batcher {
	var batcher []expect.Batcher
	if execute != "" {
		execute = r.wrapTestCommand(execute)
		batcher = append(batcher, &expect.BSnd{S: execute})
	}
	return batcher
}

// Determines if an error is an expect.TimeoutError.
func isTimeout(err error) bool {
	_, ok := err.(expect.TimeoutError)
	return ok
}

// Step performs `step`, then, in response to events, consequent steps fed by `handler`.
// Return on first error, or when there is no next step to perform.
func (r *Reel) Step(step *Step, handler Handler) error {
	for step != nil {
		if r.Err != nil {
			return r.Err
		}
		exec, exp, timeout := step.unpack()
		var batcher []expect.Batcher
		batcher = r.generateBatcher(exec)
		var firstMatch string
		batcher = r.batchExpectations(exp, batcher, &firstMatch)
		results, err := (*r.expecter).ExpectBatch(batcher, timeout)

		if !step.hasExpectations() {
			return nil
		}

		if err != nil {
			if isTimeout(err) {
				step = handler.ReelTimeout()
			} else {
				return err
			}
		} else {
			if len(results) > 0 {
				result := results[0]

				output := r.stripEmulatedPromptFromOutput(result.Output)
				match := r.stripEmulatedPromptFromOutput(result.Match[0])

				matchIndex := strings.Index(output, match)
				var before string
				// special case:  the match regex may be nothing at all.
				if matchIndex > 0 {
					before = output[0 : matchIndex-1]
				} else {
					before = ""
				}
				step = handler.ReelMatch(r.stripEmulatedRegularExpression(firstMatch), before, match)
			}
		}
	}
	return nil
}

// Run the target subprocess to completion.  The first step to take is supplied by handler.  Consequent steps are
// determined by handler in response to events.  Return on first error, or when there is no next step to execute.
func (r *Reel) Run(handler Handler) error {
	return r.Step(handler.ReelFirst(), handler)
}

// Appends a new line to a command, if necessary.
func (r *Reel) createExecutableCommand(command string) string {
	command = r.wrapTestCommand(command)
	return command
}

// NewReel create a new `Reel` instance for interacting with a target subprocess.  The command line for the target is
// specified by the args parameter.
func NewReel(expecter *expect.Expecter, args []string, errorChannel <-chan error, opts ...Option) (*Reel, error) {
	r := &Reel{}
	for _, o := range opts {
		o(r)
	}
	if len(args) > 0 {
		command := r.createExecutableCommand(strings.Join(args, " "))
		err := (*expecter).Send(command)
		if err != nil {
			return nil, err
		}
	}
	r.expecter = expecter

	go func() {
		r.Err = <-errorChannel
	}()
	return r, nil
}

// wrapTestCommand will wrap a test command in syntax to postfix a terminal emulation prompt.
func (r *Reel) wrapTestCommand(cmd string) string {
	if !r.disableTerminalPromptEmulation {
		return WrapTestCommand(cmd)
	}
	if !strings.HasSuffix(cmd, "\n") {
		cmd += "\n"
	}
	return cmd
}

// WrapTestCommand wraps cmd so that the output will end in an emulated terminal prompt.
func WrapTestCommand(cmd string) string {
	cmd = strings.TrimRight(cmd, "\n")
	return fmt.Sprintf("%s && echo %s\n", cmd, EndOfTestSentinel)
}

// stripEmulatedPromptFromOutput will elide the emulated terminal prompt from the test output.
func (r *Reel) stripEmulatedPromptFromOutput(output string) string {
	if !r.disableTerminalPromptEmulation {
		return strings.TrimRight(strings.TrimRight(output, endOfTestSentinelCutset), "\n")
	}
	return output
}

// stripEmulatedRegularExpression will elide the modified part of the terminal prompt regular expression.
func (r *Reel) stripEmulatedRegularExpression(match string) string {
	if !r.disableTerminalPromptEmulation && len(match) > len(EndOfTestRegexPostfix) {
		return match[0 : len(match)-len(EndOfTestRegexPostfix)]
	}
	return match
}

// addEmulatedRegularExpression will append the additional regular expression to capture the emulated terminal prompt.
func (r *Reel) addEmulatedRegularExpression(regularExpressionString string) string {
	if !r.disableTerminalPromptEmulation {
		return fmt.Sprintf("%s%s", regularExpressionString, EndOfTestRegexPostfix)
	}
	return regularExpressionString
}
