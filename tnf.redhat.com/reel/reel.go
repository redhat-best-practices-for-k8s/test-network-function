package reel

import (
    "os/exec"
    "io"
    "bufio"
    "encoding/json"
)

const CTRL_C string = "\003" // ^C
const CTRL_D string = "\004" // ^D

type Step struct {
    Execute string      `json:"execute,omitempty"`
    Expect  []string    `json:"expect,omitempty"`
    Timeout int         `json:"timeout,omitempty"`
}

type Event struct {
    /* "event" in {"match","timeout","eof"} */
    Event   string      `json:"event"`
    /* only when "event" = "match" */
    Idx     int         `json:"idx,omitempty"`
    Pattern string      `json:"pattern,omitempty"`
    Before  string      `json:"before,omitempty"`
    Match   string      `json:"match,omitempty"`
}

type Handler interface {
    ReelFirst() (*Step)
    ReelMatch(pattern string, before string, match string) (*Step)
    ReelTimeout() (*Step)
    ReelEof() (*Step)
}

type StepFunc func(Handler) (*Step)

type Reel struct {
    subp    *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    scanner *bufio.Scanner
}

func (reel *Reel) Open() error {
    return reel.subp.Start()
}
func (reel *Reel) Step(step *Step, handler Handler) error {
    var msg []byte
    var err error

    for step != nil {
        msg, err = json.Marshal(step)
        if err != nil {
            return err
        }
        reel.stdin.Write(msg)
        reel.stdin.Write([]byte("\n"))

        if len(step.Expect) == 0 {
            return nil
        }

        var event Event
        reel.scanner.Scan()
        err = json.Unmarshal(reel.scanner.Bytes(), &event)
        if err != nil {
            return err
        }
        switch event.Event {
        case "match":
            step = handler.ReelMatch(event.Pattern, event.Before, event.Match)
        case "timeout":
            step = handler.ReelTimeout()
        case "eof":
            step = handler.ReelEof()
        }
    }
    return nil
}
func (reel *Reel) Close() {
    reel.stdin.Close()
    reel.subp.Wait()
    reel.subp = nil
}
func (reel *Reel) Run(handler Handler) error {
    err := reel.Open()
    if err == nil {
        err = reel.Step(handler.ReelFirst(), handler)
        reel.Close()
    }
    return err
}

func prependLogOption(args []string, logfile string) ([]string) {
    args = append(args, "", "")
    copy(args[2:], args)
    args[0] = "-l"
    args[1] = logfile
    return args
}

func NewReel(logfile string, args []string) (*Reel, error) {
    var err error

    if logfile != "" {
        args = prependLogOption(args, logfile)
    }

    subp := exec.Command("reel.exp", args[:]...)
    stdin, err := subp.StdinPipe()
    if err != nil {
        return nil, err
    }
    stdout, err := subp.StdoutPipe()
    if err != nil {
        stdin.Close()
        return nil, err
    }
    return &Reel{
        subp: subp,
        stdin: stdin,
        stdout: stdout,
        scanner: bufio.NewScanner(stdout),
    }, err
}
