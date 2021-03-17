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

package tnf

import (
	"time"

	expect "github.com/ryandgoulding/goexpect"
	"github.com/test-network-function/test-network-function/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function/pkg/tnf/reel"
)

const (
	// ERROR represents an errored test.
	ERROR = iota
	// SUCCESS represents a successful test.
	SUCCESS
	// FAILURE represents a failed test.
	FAILURE
)

// ExitCodeMap maps a test result value to a more appropriate Unix return code.
var ExitCodeMap = map[int]int{
	SUCCESS: 0,
	FAILURE: 1,
	ERROR:   2,
}

// Tester provides the interface for a Test.
type Tester interface {
	// Args returns the CLI command as a string array.
	Args() []string

	// GetIdentifier returns the tnf.Test specific identifier.
	GetIdentifier() identifier.Identifier

	// Result returns the result of the test (ERROR, SUCCESS, or FAILURE).
	Result() int

	// Timeout returns the test timeout as a Duration.
	Timeout() time.Duration
}

// Test runs a chain of Handlers.
type Test struct {
	runner *reel.Reel
	tester Tester
	chain  []reel.Handler
}

// Run performs a test, returning the result and any encountered errors.
func (t *Test) Run() (int, error) {
	err := t.runner.Run(t)
	return t.tester.Result(), err
}

func (t *Test) dispatch(fp reel.StepFunc) *reel.Step {
	for _, handler := range t.chain {
		step := fp(handler)
		if step != nil {
			return step
		}
	}
	return nil
}

// ReelFirst calls the current Handler's ReelFirst function.
func (t *Test) ReelFirst() *reel.Step {
	fp := func(handler reel.Handler) *reel.Step {
		return handler.ReelFirst()
	}
	return t.dispatch(fp)
}

// ReelMatch calls the current Handler's ReelMatch function.
func (t *Test) ReelMatch(pattern, before, match string) *reel.Step {
	fp := func(handler reel.Handler) *reel.Step {
		return handler.ReelMatch(pattern, before, match)
	}
	return t.dispatch(fp)
}

// ReelTimeout calls the current Handler's ReelTimeout function.
func (t *Test) ReelTimeout() *reel.Step {
	fp := func(handler reel.Handler) *reel.Step {
		return handler.ReelTimeout()
	}
	return t.dispatch(fp)
}

// ReelEOF calls the current Handler's ReelEOF function.
func (t *Test) ReelEOF() {
	for _, handler := range t.chain {
		handler.ReelEOF()
	}
}

// NewTest creates a new Test given a chain of Handlers.
func NewTest(expecter *expect.Expecter, tester Tester, chain []reel.Handler, errorChannel <-chan error) (*Test, error) {
	var args []string = tester.Args()
	runner, err := reel.NewReel(expecter, args, errorChannel)
	if err != nil {
		return nil, err
	}
	return &Test{runner: runner, tester: tester, chain: chain}, nil
}
