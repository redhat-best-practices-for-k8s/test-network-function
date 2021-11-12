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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	expect "github.com/google/goexpect"
	"google.golang.org/grpc/codes"
)

const (
	// EndOfTestSentinel is the emulated terminal prompt that will follow command output.
	EndOfTestSentinel = `END_OF_TEST_SENTINEL`
	// ExitKeyword keyword delimiting the command exit status
	ExitKeyword = "exit="
)

var (

	// matchSentinel This regular expression is matching stricly the sentinel and exit code.
	// This match regular expression matches commands that return no output
	matchSentinel = fmt.Sprintf("((.|\n)*%s %s[0-9]+\n)", EndOfTestSentinel, ExitKeyword)

	// EndOfTestRegexPostfix This regular expression is a postfix added to the goexpect regular expressions. This regular expression matches a
	// sentinel or marker string that is marking the end of the command output. This is because after the command
	// output, the shell might also return a prompt which is not desired. Note: this is currently the same as the string above
	// but was splitted for clarity
	EndOfTestRegexPostfix = matchSentinel
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
// This disables terminal shell management and is used only for unit testing where the terminal is not available
// In this mode, go expect operates only on strings not command/shell outputs
func DisableTerminalPromptEmulation() Option {
	return func(r *Reel) Option {
		r.disableTerminalPromptEmulation = true
		return DisableTerminalPromptEmulation()
	}
}

// Each Step can have zero or more expectations (Step.Expect).  This method follows the Adapter design pattern;  a raw
// array of strings is turned into a corresponding array of expect.Batcher.  This method side-effects the input
// expectations array, following the Builder design pattern.  Finally, the first match is stored in the firstMatch
// output parameter.
// This command translates individual expectations in the test cases (e.g. success, failure, etc) into expect.Case in go expect
// The expect.Case are later matched in order inside goexpect ExpectBatch function.
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
	// expectations created from test case matches
	for _, expectation := range expectations {
		thisCase := r.generateCase(expectation, firstMatch)
		cases = append(cases, thisCase)
	}
	// extra test case to match when commands do not return anything but exit without error. This expectation makes
	// sure that any command exiting successfully will be processed without timeout.
	thisCase := r.generateCase("", firstMatch)
	cases = append(cases, thisCase)
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

// Each Step can have exactly one execution string (Step.Execute). This method follows the Adapter design pattern;  a
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
		// firstMatchRe is the first regular expression (expectation) that has matched results
		var firstMatchRe string
		batcher = r.batchExpectations(exp, batcher, &firstMatchRe)
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

				output, outputStatus := r.stripEmulatedPromptFromOutput(result.Output)
				if outputStatus != 0 {
					return fmt.Errorf("error executing command exit code:%d", outputStatus)
				}
				match, matchStatus := r.stripEmulatedPromptFromOutput(result.Match[0])
				log.Debugf("command status: output=%s, match=%s, outputStatus=%d, matchStatus=%d", output, match, outputStatus, matchStatus)

				matchIndex := strings.Index(output, match)
				var before string
				// special case:  the match regex may be nothing at all.
				if matchIndex > 0 {
					before = output[0 : matchIndex-1]
				} else {
					before = ""
				}
				strippedFirstMatchRe := r.stripEmulatedRegularExpression(firstMatchRe)
				step = handler.ReelMatch(strippedFirstMatchRe, before, match)
			}
		}
		// This is for the last step
		if r.Err != nil {
			return r.Err
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
	wrappedCommand := fmt.Sprintf("%s ; echo %s %s$?\n", cmd, EndOfTestSentinel, ExitKeyword)
	log.Tracef("Command sent: %s", wrappedCommand)
	return wrappedCommand
}

// stripEmulatedPromptFromOutput will elide the emulated terminal prompt from the test output.
func (r *Reel) stripEmulatedPromptFromOutput(output string) (data string, status int) {
	parsed := strings.Split(output, EndOfTestSentinel)
	var err error
	if !r.disableTerminalPromptEmulation && len(parsed) == 2 {
		// if a sentinel was present, then we have at least 2 parsed results
		// if command retuned nothing parsed[0]==""
		data = parsed[0]
		status, err = strconv.Atoi(strings.Split(strings.Split(parsed[1], ExitKeyword)[1], "\n")[0])
		if err != nil {
			// Cannot parse status from output, something is wrong, fail command
			status = 1
			log.Errorf("Cannot determine command status. Error: %s", err)
		}
		// remove trailing \n if present
		data = strings.TrimRight(data, "\n")
	} else {
		// to support unit tests (without sentinel parsing)
		data = output
		status = 0
		log.Errorf("Cannot determine command status, no sentinel present. Error: %s", err)
	}
	return
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
		regularExpressionStringWithMetadata := fmt.Sprintf("%s%s", regularExpressionString, EndOfTestRegexPostfix)
		return regularExpressionStringWithMetadata
	}
	return regularExpressionString
}
