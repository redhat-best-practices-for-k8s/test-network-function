package reel

import (
	"bufio"
)

// A handler feeding steps from `scanner`.
// On match event the next step will execute the next line from `scanner`,
// will expect to match `prompt` and complete within `timeout` seconds.
type LineFeeder struct {
	active  bool
	timeout int
	prompt  string
	scanner *bufio.Scanner
}

// Returns no step; a line feeder only feeds steps on match events.
func (f *LineFeeder) ReelFirst() *Step {
	return nil
}

// On match, return a step which will execute the next line from the feeder's
// scanner. If the scanner is exhausted, return no step.
func (f *LineFeeder) ReelMatch(pattern string, before string, match string) *Step {
	if f.scanner.Scan() {
		command := f.scanner.Text()
		if command == "" {
			// the empty string will be omitted and result in no command sent
			// send single space to execute "no command"
			command = " "
		}
		f.active = true
		return &Step{
			Execute: command,
			Expect:  []string{f.prompt},
			Timeout: f.timeout,
		}
	}
	f.active = false
	return nil
}

// On timeout, return a step which kills an active subprocess by sending it ^C.
// Otherwise, return no step.
func (f *LineFeeder) ReelTimeout() *Step {
	if f.active {
		f.active = false
		return &Step{
			Execute: CTRL_C,
			Expect:  []string{f.prompt},
			Timeout: f.timeout,
		}
	}
	return nil
}

// On eof, take no action.
func (f *LineFeeder) ReelEof() {
	// empty
}

// Create a new `LineFeeder` which feeds steps executing lines from `scanner`,
// expecting to match `prompt` and complete within `timeout`.
func NewLineFeeder(timeout int, prompt string, scanner *bufio.Scanner) *LineFeeder {
	return &LineFeeder{
		timeout: timeout,
		prompt:  prompt,
		scanner: scanner,
	}
}
