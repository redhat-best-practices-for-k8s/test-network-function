package reel

import (
	"bufio"
)

type LineFeeder struct {
	active  bool
	timeout int
	prompt  string
	scanner *bufio.Scanner
}

func (f *LineFeeder) ReelFirst() *Step {
	return nil
}
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
func (f *LineFeeder) ReelEof() *Step {
	return nil
}

func NewLineFeeder(timeout int, prompt string, scanner *bufio.Scanner) *LineFeeder {
	return &LineFeeder{
		timeout: timeout,
		prompt:  prompt,
		scanner: scanner,
	}
}
