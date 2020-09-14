package reel

import (
	expect "github.com/google/goexpect"
	"google.golang.org/grpc/codes"
	"regexp"
	"strings"
	"time"
)

const (
	// CtrlC is the constant representing SIGINT.
	CtrlC   = "\003" // ^C
	newLine = "\n"
	sep     = " "
)

// Step is an instruction for a single REEL pass.
// To process a step, first send the `Execute` string to the target subprocess (if supplied).  Block until the
// subprocess output to stdout matches one of the regular expressions in `Expect` (if any supplied). A positive integer
// `Timeout` (seconds) prevents blocking forever.
type Step struct {
	// A command to execute using the underlying subprocess.
	Execute string `json:"execute,omitempty"`

	// An array of expected text regular expressions.  The first expectation results in a match.
	Expect []string `json:"expect,omitempty"`

	// The timeout for the Step.  A positive Timeout prevents blocking forever.
	Timeout time.Duration `json:"timeout,omitempty"`
}

// A utility method to return the important aspects of the Step container as a tuple.
func (s *Step) unpack() (string, []string, time.Duration) {
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

	// ReelMatch informs of a match event, returning the next step to perform.
	ReelMatch(pattern string, before string, match string) *Step

	// ReelTimeout informs of a timeout event, returning the next step to perform.
	ReelTimeout() *Step

	// ReelEOF informs of the eof event.
	ReelEOF()
}

// StepFunc provides a wrapper around a generic Handler.
type StepFunc func(Handler) *Step

// A Reel instance allows interaction with a target subprocess.
type Reel struct {
	// A pointer to the underlying subprocess
	expecter *expect.Expecter
	Err      error
}

// Each Step can have zero or more expectations (Step.Expect).  This method follows the Adapter design pattern;  a raw
// array of strings is turned into a corresponding array of exepct.Batcher.  This method side-effects the input
// expectations array, following the Builder design pattern.  Finally, the first match is stored in the firstMatch
// output parameter.
func batchExpectations(expectations []string, batcher []expect.Batcher, firstMatch *string) []expect.Batcher {
	if len(expectations) > 0 {
		expectCases := generateCases(expectations, firstMatch)
		batcher = append(batcher, &expect.BCas{C: expectCases})
	}
	return batcher
}

// Each Step can have zero or more expectations (Step.Expect), and if any match is found, then a match event occurs
// (representing a logical "OR" over the array). This helper follows the Adapter design pattern;  a raw array of string
// regular expressions is converted to an equivalent expect.Caser array.  The firstMatch parameter is used as an output
// parameter to store the first match found in the expectations array.  Thus, the order of expectations is important.
func generateCases(expectations []string, firstMatch *string) []expect.Caser {
	var cases []expect.Caser
	for _, expectation := range expectations {
		thisCase := generateCase(expectation, firstMatch)
		cases = append(cases, thisCase)
	}
	return cases
}

// Each Step can have zero or more expectations (Step.Expect).  This method follows the Adapter design pattern;  a
// single raw string Expectation is converted into a corresponding expect.Case.
func generateCase(expectation string, firstMatch *string) *expect.Case {
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
func generateBatcher(execute string) []expect.Batcher {
	var batcher []expect.Batcher
	if execute != "" {
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
func (reel *Reel) Step(step *Step, handler Handler) error {
	for step != nil {
		if reel.Err != nil {
			return reel.Err
		}
		exec, exp, timeout := step.unpack()
		var batcher []expect.Batcher
		batcher = generateBatcher(exec)
		var firstMatch string
		batcher = batchExpectations(exp, batcher, &firstMatch)
		results, err := (*reel.expecter).ExpectBatch(batcher, timeout)

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
				output := result.Output
				match := result.Match[0]
				matchIndex := strings.Index(output, match)
				var before string
				// special case:  the match regex may be nothing at all.
				if matchIndex != 0 {
					before = output[0 : matchIndex-1]
				} else {
					before = ""
				}
				step = handler.ReelMatch(firstMatch, before, match)
			}
		}
	}
	return nil
}

// Run the target subprocess to completion.  The first step to take is supplied by handler.  Consequent steps are
// determined by handler in response to events.  Return on first error, or when there is no next step to execute.
func (reel *Reel) Run(handler Handler) error {
	return reel.Step(handler.ReelFirst(), handler)
}

// Appends a new line to a command, if necessary.
func createExecutableCommand(command string) string {
	if !strings.HasSuffix(command, newLine) {
		return command + newLine
	}
	return command
}

// NewReel create a new `Reel` instance for interacting with a target subprocess.  The command line for the target is
// specified by the args parameter.
func NewReel(expecter *expect.Expecter, args []string, errorChannel <-chan error) (*Reel, error) {
	if len(args) > 0 {
		command := createExecutableCommand(strings.Join(args, sep))
		err := (*expecter).Send(command)
		if err != nil {
			return nil, err
		}
	}
	r := &Reel{expecter: expecter}
	go func() {
		r.Err = <-errorChannel
	}()
	return r, nil
}
