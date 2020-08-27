package tnf

import (
	expect "github.com/google/goexpect"
	"github.com/redhat-nfvpe/test-network-function/internal/reel"
	"time"
)

const (
	SUCCESS = iota
	FAILURE
	ERROR
)

type Tester interface {
	Args() []string
	Timeout() time.Duration
	Result() int
}

type Test struct {
	runner *reel.Reel
	tester Tester
	chain  []reel.Handler
}

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

func (t *Test) ReelFirst() *reel.Step {
	fp := func(handler reel.Handler) *reel.Step {
		return handler.ReelFirst()
	}
	return t.dispatch(fp)
}
func (t *Test) ReelMatch(pattern string, before string, match string) *reel.Step {
	fp := func(handler reel.Handler) *reel.Step {
		return handler.ReelMatch(pattern, before, match)
	}
	return t.dispatch(fp)
}
func (t *Test) ReelTimeout() *reel.Step {
	fp := func(handler reel.Handler) *reel.Step {
		return handler.ReelTimeout()
	}
	return t.dispatch(fp)
}
func (t *Test) ReelEof() {
	for _, handler := range t.chain {
		handler.ReelEof()
	}
}

func NewTest(expecter *expect.Expecter, tester Tester, chain []reel.Handler, errorChannel <-chan error) (*Test, error) {
	var args []string
	args = tester.Args()
	runner, err := reel.NewReel(expecter, args, errorChannel)
	if err != nil {
		return nil, err
	}
	return &Test{runner: runner, tester: tester, chain: chain}, nil
}
