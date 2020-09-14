package reel

import (
	"fmt"
	"strings"
)

// Printer provides handler printing event interaction.
// A printer is only used for the side effect of printing: it never feeds steps.
// `trimr` specifies a string used in filtering out text appearing before a
// match string. (When that text is empty after being right-trimmed using
// `trimr`, it is not printed.)
type Printer struct {
	trimr string
}

// ReelFirst Returns no step; a printer does not feed steps.
func (p *Printer) ReelFirst() *Step {
	return nil
}

// ReelMatch is called when a pattern matches, and prints `before` and `match` strings, unless `before` is empty when
// right-trimmed using the printer's `trimr` string, in which case only print `match`.
// Returns no step; a printer does not feed steps.
func (p *Printer) ReelMatch(pattern string, before string, match string) *Step {
	if strings.TrimRight(before, p.trimr) == "" {
		fmt.Print(match)
	} else {
		fmt.Print(before, match)
	}
	return nil
}

// ReelTimeout prints timeout messaging to the log on timeout.
// Returns no step; a printer does not feed steps.
func (p *Printer) ReelTimeout() *Step {
	fmt.Println("(timeout)")
	return nil
}

// ReelEOF prints eof messaging to the log on eof.
func (p *Printer) ReelEOF() {
	fmt.Println("(eof)")
}

// NewPrinter create a new Printer using `trimr` to filter output text before a match.
func NewPrinter(trimr string) *Printer {
	return &Printer{
		trimr: trimr,
	}
}
