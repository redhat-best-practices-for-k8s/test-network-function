package tnf

import (
    "bufio"
    "strings"
    "os"
    "fmt"
    "tnf.redhat.com/reel"
)

type TestFeeder struct {
    timeout int
    prompt  string
    scanner *bufio.Scanner
    tester  Tester
}

func (f *TestFeeder) ReelFirst() (*reel.Step) {
    return nil
}
func (f *TestFeeder) ReelMatch(pattern string, before string, match string) (*reel.Step) {
    if f.scanner != nil && f.scanner.Scan() {
        config, err := DecodeConfig(f.scanner.Bytes())
        if err == nil {
            // TODO: no such test => panic
            f.tester = Tests[config.Test](config)
            return &reel.Step{
                Execute: strings.Join(f.tester.Args(), " "),
                Expect: []string{f.prompt},
                Timeout: f.tester.Timeout(),
            }
            // TODO: fold in result?
        } else {
            fmt.Fprintln(os.Stderr, err)
            f.scanner = nil
        }
    }
    f.tester = nil
    return nil
}
func (f *TestFeeder) ReelTimeout() (*reel.Step) {
    if f.tester != nil {
        return &reel.Step{
            Execute: "\003", // ^C
            Expect: []string{f.prompt},
            Timeout: f.timeout,
        }
    }
    return nil
}
func (f *TestFeeder) ReelEof() (*reel.Step) {
    return nil
}

func NewTestFeeder(timeout int, prompt string, scanner *bufio.Scanner) (*TestFeeder) {
    return &TestFeeder{
        timeout: timeout,
        prompt:  prompt,
        scanner: scanner,
    }
}
