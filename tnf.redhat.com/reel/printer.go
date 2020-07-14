package reel

import (
    "fmt"
    "strings"
)

type Printer struct {
    trimr string
}

func (p *Printer) ReelFirst() (*Step) {
    return nil
}
func (p *Printer) ReelMatch(pattern string, before string, match string) (*Step) {
    if strings.TrimRight(before, p.trimr) == "" {
        fmt.Print(match)
    } else {
        fmt.Print(before, match)
    }
    return nil
}
func (p *Printer) ReelTimeout() (*Step) {
    fmt.Println("(timeout)")
    return nil
}
func (p *Printer) ReelEof() (*Step) {
    fmt.Println("(eof)")
    return nil
}

func NewPrinter(trimr string) (*Printer) {
    return &Printer{
        trimr: trimr,
    }
}
