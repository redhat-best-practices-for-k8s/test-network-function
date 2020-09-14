package tnf

import (
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"time"
)

const (
	// SUCCESS represents a successful test.
	SUCCESS = iota
	// FAILURE represents a failed test.
	FAILURE
	// ERROR represents an errored test.
	ERROR
)

// Tester provides the interface for a Test.
type Tester interface {
	Args() []string
	Timeout() time.Duration
	Result() int
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
func (t *Test) ReelMatch(pattern string, before string, match string) *reel.Step {
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
	args := tester.Args()
	runner, err := reel.NewReel(expecter, args, errorChannel)
	if err != nil {
		return nil, err
	}
	return &Test{runner: runner, tester: tester, chain: chain}, nil
}
