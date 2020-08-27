package interactive

import (
	expect "github.com/google/goexpect"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"time"
)

var UnitTestMode = false
var spawnFunc *SpawnFunc

func SetSpawnFunc(sFunc *SpawnFunc) {
	spawnFunc = sFunc
}

// Abstracts a wrapper interface over the required methods of the exec.Cmd API for testing purposes.
type SpawnFunc interface {
	Command(name string, arg ...string) *SpawnFunc
	Start() error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	Wait() error
}

// An implementation of SpawnFunc using exec.Cmd.
type ExecSpawnFunc struct {
	cmd *exec.Cmd
}

func (e *ExecSpawnFunc) Command(name string, arg ...string) *SpawnFunc {
	cmd := exec.Command(name, arg...)
	execSpawnFunc := &ExecSpawnFunc{cmd: cmd}
	var spawnFunc SpawnFunc = execSpawnFunc
	return &spawnFunc
}

func (e *ExecSpawnFunc) Wait() error {
	return e.cmd.Wait()
}

func (e *ExecSpawnFunc) Start() error {
	return e.cmd.Start()
}

func (e *ExecSpawnFunc) StdinPipe() (io.WriteCloser, error) {
	return e.cmd.StdinPipe()
}

func (e *ExecSpawnFunc) StdoutPipe() (io.Reader, error) {
	return e.cmd.StdoutPipe()
}

type Spawner interface {
	Spawn(command string, args []string, timeout time.Duration, opts ...expect.Option) (*Context, error)
}

// Type Context represents an interactive context.  This abstraction is meant to be overloaded, and can represent
// something as simple as a shell, to as complex as an interactive OpenShift client or SSH session.  Context follows the
// Container design pattern, and is a simple data transfer object.
type Context struct {
	expecter     *expect.Expecter
	errorChannel <-chan error
}

func (c *Context) GetExpecter() *expect.Expecter {
	return c.expecter
}

func (c *Context) GetErrorChannel() <-chan error {
	return c.errorChannel
}

func NewContext(expecter *expect.Expecter, errorChannel <-chan error) *Context {
	return &Context{expecter: expecter, errorChannel: errorChannel}
}

type GoExpectSpawner struct {
}

func NewGoExpectSpawner() *GoExpectSpawner {
	return &GoExpectSpawner{}
}

// Creates a subprocess, setting standard input and standard output appropriately.  This is the base method to create
// any interactive PTY based process.
func (g *GoExpectSpawner) Spawn(command string, args []string, timeout time.Duration, opts ...expect.Option) (*Context, error) {
	if !UnitTestMode {
		execSpawnFunc := &ExecSpawnFunc{}
		var transitionSpawnFunc SpawnFunc = execSpawnFunc
		spawnFunc = &transitionSpawnFunc
	}

	spawnFunc = (*spawnFunc).Command(command, args...)
	stdinPipe, stdoutPipe, err := g.unpackPipes(spawnFunc)
	if err != nil {
		return nil, err
	}
	err = g.startCommand(spawnFunc, command, args)
	if err != nil {
		return nil, err
	}
	return g.spawnGeneric(spawnFunc, stdinPipe, stdoutPipe, timeout, opts...)
}

// Helper method which spawns a Context.  The pseudo-terminal (PTY) as well as the underlying goroutine is set up using
// expect.SpawnGeneric(...), allowing for long-lived sessions.
func (g *GoExpectSpawner) spawnGeneric(spawnFunc *SpawnFunc, stdinPipe io.WriteCloser, stdoutPipe io.Reader, timeout time.Duration, opts ...expect.Option) (*Context, error) {
	// Spawns a generic PTY process using expect.SpawnGeneric(...).
	var gexpecter *expect.GExpect
	var errorChannel <-chan error
	var err error
	gexpecter, errorChannel, err = expect.SpawnGeneric(&expect.GenOptions{
		In:  stdinPipe,
		Out: stdoutPipe,
		Wait: func() error {
			return (*spawnFunc).Wait()
		},
		Close: func() error {
			return nil
		},
		Check: func() bool { return true },
	}, timeout, opts...)
	// coax out the typing
	var expecter expect.Expecter = gexpecter
	// Return an interactive context containing the expecter and the error channel.  The error channel should be
	// monitored by a separate goroutine for errors.
	return NewContext(&expecter, errorChannel), err
}

// Helper method to start an exec.Cmd.
func (g *GoExpectSpawner) startCommand(spawnFunc *SpawnFunc, command string, args []string) error {
	err := (*spawnFunc).Start()
	if err != nil {
		log.Errorf("Failed to invoke %s %s: %v", command, args, err)
	}
	return err
}

// Helper method to unpack stdin and stdout.
func (g *GoExpectSpawner) unpackPipes(spawnFunc *SpawnFunc) (io.WriteCloser, io.Reader, error) {
	stdinPipe, err := g.extractStdinPipe(spawnFunc)
	if err != nil {
		return nil, nil, err
	}
	stdoutPipe, err := g.extractStdoutPipe(spawnFunc)
	if err != nil {
		return nil, nil, err
	}
	return stdinPipe, stdoutPipe, err
}

// Helper method to extract stdin.
func (g *GoExpectSpawner) extractStdinPipe(spawnFunc *SpawnFunc) (io.WriteCloser, error) {
	stdin, err := (*spawnFunc).StdinPipe()
	if err != nil {
		log.Errorf("Couldn't extract stdin for the given process: %v", err)
	}
	return stdin, err
}

// Helper method to extract stdout.
func (g *GoExpectSpawner) extractStdoutPipe(spawnFunc *SpawnFunc) (io.Reader, error) {
	stdout, err := (*spawnFunc).StdoutPipe()
	if err != nil {
		log.Errorf("Couldn't extract stdout for the given process: %v", err)
	}
	return stdout, err
}
