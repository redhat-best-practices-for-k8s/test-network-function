# Test Development Guide

Currently, tests are all CLI driven.  That means that the commands executed in test implementations must be made
available in the target container/machine/shell's `$PATH`.  Future work will address incorporating REST-based tests.
UI-driven tests are considered out of scope for `test-network-function`.

## General Test Writing Guidelines

In general, tests should adhere to the following principles:
* Tests should be platform independent when possible, and platform-aware when not.
* Tests must be runnable in a variety of contexts (i.e., `oc`, `ssh`, and `shell`).  Internally, we have developed a
variety of `interactive.Context` implementations for each of these.  In general, so long as your command does not depend
on specific prompts, the framework handles the context transparently.
* Tests must implement the `tnf.Tester` interface.
* Tests must implement the `reel.Handler` interface.
* Tests adhere to the strict quality and style guidelines set forth in [CONTRIBUTING.md](CONTRIBUTING.md).

## Language options for writing test implementations

Currently, test implementations can only be written in Go.

## Writing a simple CLI-oriented test in Go

A `test-network-function` test must implement `tnf.Tester` and `reel.Handler` Go `interface`s.  The `tnf.Tester`
interface defines the contract required for a CLI-based test, and `reel.Handler` defines the Finite State Machine (FSM)
contract for executing the test.  A basic example is [ping.go](pkg/tnf/handlers/ping/ping.go).

We will go through implementing the required interfaces one at a time below:

### Implementing `ping.go` `tnf.Tester`

For a test to implement the `tnf.Tester` interface, it must provide definitions for `Args`, `Timeout` and `Result`.
These are the set of accessor methods used to define characteristics of the test, as well as the actual result of the
test.  Note, this does not include any expected results;  those need to be defined later.

First create a type called `Ping` which is capable of storing `result`, `timeout` and `args` variables.  Additionally,
we will restrict our test by mandating that a positive integer `count` and `destination` string must be provided during
the time of instantiation.

```go
// Ping provides a ping test implemented using command line tool `ping`.
type Ping struct {
    result      int
    timeout     time.Duration
    args        []string
    count       int
    destination string
}
```

To better enforce data encapsulation, please only export (capitalize) variables that are absolutely needed.  For
example, we use `count` not `Count` above.

If you look at the [`tnf.Tester`](pkg/tnf/test.go) interface definition, you will notice that the data types for
`result`, `timeout` and `args` match the return types for the mandated functions.

```go
type Tester interface {
	Args() []string
	Timeout() time.Duration
	Result() int
}
```

After creating the struct, define the accessors similar to the following:

```go
// Args returns the command line args for the test.
func (p *Ping) Args() []string {
	return []string{"ping", "-c", p.count, p.destination}
}

// Timeout returns the timeout in seconds for the test.
func (p *Ping) Timeout() time.Duration {
	return p.timeout
}

// Result returns the test result.
func (p *Ping) Result() int {
	return p.result
}
```

The `Args()` implementation deserves some explaining.  `Args()` is an array of the commands line argument strings.  In
this example, the command `ping -c 5 www.redhat.com` is represented as `string[]{"ping", "-c", "5", "www.redhat.com"}`.
In other words, the elements are all of the white-space separated string components of the command.

That completes our `ping.go` `tnf.Tester` implementation!  Next, implement the logic of the `reel.Handler` FSM.

### Implementing `ping.go` `reel.Handler`

The easy part is out of the way.  Implementing `reel.Handler` is slightly more involved, but should make sense after
completing this part of the tutorial.  [reel.go](internal/reel/reel.go) defines the `reel.Handler` interface:

```go
// A Handler implements desired programmatic control.
type Handler interface {
	// ReelFirst returns the first step to perform.
	ReelFirst() *Step

	// ReelMatch informs of a match event, returning the next step to perform.  ReelMatch takes three arguments:
	// `pattern` represents the regular expression pattern which was matched.
	// `before` contains all output preceding `match`.
	// `match` is the text matched by `pattern`.
	ReelMatch(pattern string, before string, match string) *Step

	// ReelTimeout informs of a timeout event, returning the next step to perform.
	ReelTimeout() *Step

	// ReelEOF informs of the eof event.
	ReelEOF()
}
```

We will handle describing implementing each of these methods one by one.

#### Implementing `ping.go` `ReelEOF()`

`ReelEOF` is used to define the callback executed when EOF is encountered in the context.  Unexpected interruptions to
`ssh` or `oc` session are common reasons for EOF.

For the case of ping we can make this simple.  Since we require a `count` for `ping`, we don't need to do anything
particular for EOF.

```go
// ReelEOF does nothing;  ping requires no intervention on EOF.
func (p *Ping) ReelEOF() {
}
```

#### Implementing `ping.go` `ReelTimeout()`

`ReelTimeout` is used to define the callback executed when a test times out.

When a ping test times out, we probably ought to issue a `CTRL+C` in order to exit early and prepare the context for
future commands.

```go
// ReelTimeout returns a step which kills the ping test by sending it ^C.
func (p *Ping) ReelTimeout() *reel.Step {
	return &reel.Step{Execute: "\003"}
}
```

#### Implementing `ping.go` `ReelFirst()`

Since we supply `tnf.Test` `Args()`, we do not need to include anything for `Execute` in the returned `reel.Step`.

```go
// ReelFirst returns a step which expects the ping statistics within the test timeout.
func (p *Ping) ReelFirst() *reel.Step {
	return &reel.Step{
		Expect:  []string{`(?m)connect: Invalid argument$`, `(?m)(\d+) packets transmitted, (\d+)( packets){0,1} received, (?:\+(\d+) errors)?.*$`},
		Timeout: p.timeout,
	}
}
```

Note:  The ordering of `Expect` matters!  The framework matches `Expect` elements in index-ascending order.

### Implementing `ping.go` `ReelMatch()`

This is likely the hardest part of any test implementation.  `ReelMatch` needs to decipher what is matched, and assign
the appropriate result to the `tnf.Test`.  Let's take a look at the implementation provided for
[ping.go](pkg/tnf/handlers/ping/ping.go):

```go
// ReelMatch parses the ping statistics and set the test result on match.
// The result is success if at least one response was received and the number of
// responses received is at most one less than the number received (the "missing"
// response may be in flight).
// The result is error if ping reported a protocol error (e.g. destination host
// unreachable), no requests were sent or there was some test execution error.
// Otherwise the result is failure.
// Returns no step; the test is complete.
func (p *Ping) ReelMatch(_ string, _ string, match string) *reel.Step {
	re := regexp.MustCompile(`(?m)connect: Invalid argument$`)
	matched := re.FindStringSubmatch(match)
	if matched != nil {
		p.result = tnf.ERROR
	}
	re = regexp.MustCompile(SuccessfulOutputRegex)
	matched = re.FindStringSubmatch(match)
	if matched != nil {
		// Ignore errors in converting matches to decimal integers.
		// Regular expression `stat` is required to underwrite this assumption.
		p.transmitted, _ = strconv.Atoi(matched[1])
		p.received, _ = strconv.Atoi(matched[2])
		p.errors, _ = strconv.Atoi(matched[4])
		switch {
		case p.transmitted == 0 || p.errors > 0:
			p.result = tnf.ERROR
		case p.received > 0 && (p.transmitted-p.received) <= 1:
			p.result = tnf.SUCCESS
		default:
			p.result = tnf.FAILURE
		}
	}
	return nil
}
```

Essentially, since `ReelMatch()` always returns `nil` this function is the final state for the `reel.Step` FSM.  For
more advanced tests, `ReelMatch()` can be called an arbitrary number of times.  In this example, `ReelMatch()` is only
called once.

The logic for determining the test result is up to the test writer.  This particular implementation analyzes the match
output to determine the result.
1) If the provided `destination` results in an `Indvalid Argument`, then `tnf.ERROR` is returned.
2) If the ping summary regular expression matched, then:
* `tnf.ERROR` if there were PING transmit errors
* `tnf.SUCCESS` if a maximum of a single packet was lost
* `tnf.FAILURE` for any other case.

## Writing `ping.go` test Summary

You should now have the appropriate knowledge to write your own test implementation.  There are a variety of
implementations included out of the box in the [handlers](pkg/tnf/handlers) directory.
