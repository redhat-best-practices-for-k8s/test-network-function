// Run a target subprocess with programmatic control over interaction with it.
// Programmatic control uses a Read-Execute-Expect-Loop ("REEL") pattern.
// This pattern is implemented by `bin/reel.exp`, a general purpose expect script.
package reel

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
)

const CTRL_C string = "\003" // ^C
const CTRL_D string = "\004" // ^D

// A step is an instruction for a single REEL pass.
// To process a step, reel.exp first sends the `Execute` string to the target
// subprocess (if supplied); then it will block until the subprocess output to
// stdout matches one of the regular expressions in `Expect` (if any supplied).
// A positive integer `Timeout` (seconds) prevents blocking forever.
// A step is sent to reel.exp as a JSON object in a single line of text.
type Step struct {
	Execute string   `json:"execute,omitempty"`
	Expect  []string `json:"expect,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// An event is a notification related to a single REEL pass.
// reel.exp may report an `Event` which is a "match" of a step `Expect` pattern;
// a "timeout", if there is no match within the specified period; or "eof",
// indicating that the subprocess has exited.
// reel.exp sends an event as a JSON object in a single line of text.
// Value constraints are documented in comments by field.
type _Event struct {
	Event   string `json:"event"`             // only "match" or "timeout" or "eof"
	Idx     int    `json:"idx,omitempty"`     // only present when "event" is "match"
	Pattern string `json:"pattern,omitempty"` // only present when "event" is "match"
	Before  string `json:"before,omitempty"`  // only present when "event" is "match"
	Match   string `json:"match,omitempty"`   // only present when "event" is "match"
}

// A Handler implements desired programmatic control:
// `ReelFirst` returns the first step to perform;
// `ReelMatch` informs of a match event, returning the next step to perform;
// `ReelTimeout` informs of a timeout event, returning the next step to perform;
// `ReelEof` informs of the eof event.
// (If there is no step to perform, return nil.)
type Handler interface {
	ReelFirst() *Step
	ReelMatch(pattern string, before string, match string) *Step
	ReelTimeout() *Step
	ReelEof()
}

type StepFunc func(Handler) *Step

// A `Reel` instance allows interaction with a target subprocess.
type Reel struct {
	subp    *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

// Open the target subprocess for interaction.
func (reel *Reel) Open() error {
	return reel.subp.Start()
}

// Perform `step`, then, in response to events, consequent steps fed by `handler`.
// Return on first error, or when there is no next step to perform.
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

		var event _Event
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
			handler.ReelEof()
			step = nil
		}
	}
	return nil
}

// Close the target subprocess; returns when the target subprocess has exited.
func (reel *Reel) Close() {
	reel.stdin.Close()
	reel.subp.Wait()
	reel.subp = nil
}

// Run the target subprocess to completion.
// The first step to take is supplied by `handler`.
// Consequent steps are determined by `handler` in response to events.
// Return on first error, or when there is no next step to execute.
func (reel *Reel) Run(handler Handler) error {
	err := reel.Open()
	if err == nil {
		err = reel.Step(handler.ReelFirst(), handler)
		reel.Close()
	}
	return err
}

func prependLogOption(args []string, logfile string) []string {
	args = append(args, "", "")
	copy(args[2:], args)
	args[0] = "-l"
	args[1] = logfile
	return args
}

// Create a new `Reel` instance for interacting with a target subprocess.
// The command line for the target is specified in `args`.
// Optionally log dialogue with the subprocess to `logfile`.
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
		subp:    subp,
		stdin:   stdin,
		stdout:  stdout,
		scanner: bufio.NewScanner(stdout),
	}, err
}
