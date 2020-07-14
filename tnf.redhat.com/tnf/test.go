package tnf

import (
    "tnf.redhat.com/reel"
)

const (
    SUCCESS = iota
    FAILURE = iota
    ERROR = iota
)

type Tester interface {
    Args() ([]string)
    Timeout() (int)
    Result() (int)
}

type Test struct {
    runner  *reel.Reel
    tester  Tester
    chain   []reel.Handler
}

func (t *Test) Run() (int, error) {
    err := t.runner.Run(t)
    return t.tester.Result(), err
}
func (t *Test) dispatch(fp reel.StepFunc) (*reel.Step) {
    for _, handler := range t.chain {
        step := fp(handler)
        if step != nil {
            return step
        }
    }
    return nil
}
func (t *Test) ReelFirst() (*reel.Step) {
    fp := func (handler reel.Handler) (*reel.Step) {
        return handler.ReelFirst()
    }
    return t.dispatch(fp)
}
func (t *Test) ReelMatch(pattern string, before string, match string) (*reel.Step) {
    fp := func (handler reel.Handler)(*reel.Step) {
        return handler.ReelMatch(pattern, before, match)
    }
    return t.dispatch(fp)
}
func (t *Test) ReelTimeout() (*reel.Step) {
    fp := func (handler reel.Handler)(*reel.Step) {
        return handler.ReelTimeout()
    }
    return t.dispatch(fp)
}
func (t *Test) ReelEof() (*reel.Step) {
    fp := func (handler reel.Handler)(*reel.Step) {
        return handler.ReelEof()
    }
    return t.dispatch(fp)
}

func NewTest(logfile string, tester Tester, chain []reel.Handler) (*Test, error) {
    runner, err := reel.NewReel(logfile, tester.Args())
    if err != nil {
        return nil, err
    }
    return &Test{
        runner: runner,
        tester: tester,
        chain:  chain,
    }, err
}
