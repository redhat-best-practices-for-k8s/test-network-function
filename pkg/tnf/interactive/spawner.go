// Copyright (C) 2020-2021 Red Hat, Inc.
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

package interactive

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	expect "github.com/google/goexpect"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultBufferSize is the size of the input/output buffers in bytes.
	defaultBufferSize = 16384
	// defaultBufferSizeEnvironmentVariableKey is the OS environment variable name to override defaultBufferSize.
	defaultBufferSizeEnvironmentVariableKey = "TNF_DEFAULT_BUFFER_SIZE"
)

// UnitTestMode is used to determine if the context is unit test oriented v.s. an actual CNF test run, so appropriate
// mock interfaces can be injected.  This allows the spanFunc to be injected without complicating the Spawner interface.
var UnitTestMode = false
var spawnFunc *SpawnFunc

// SetSpawnFunc sets the SpawnFunc, allowing for the actual CNF tests to be run or mocked for unit test purposes.
func SetSpawnFunc(sFunc *SpawnFunc) {
	spawnFunc = sFunc
}

// SpawnFunc Abstracts a wrapper interface over the required methods of the exec.Cmd API for testing purposes.
type SpawnFunc interface {
	// Command consult exec.Cmd.Command
	Command(name string, arg ...string) *SpawnFunc

	// Start consult exec.Cmd.Start.
	Start() error

	// StdinPipe consult exec.Cmd.StdinPipe
	StdinPipe() (io.WriteCloser, error)

	// StdoutPipe consult exec.Cmd.StdoutPipe
	StdoutPipe() (io.Reader, error)

	// Wait consult exec.Cmd.Wait
	Wait() error
}

// ExecSpawnFunc is an implementation of SpawnFunc using exec.Cmd.
type ExecSpawnFunc struct {
	cmd *exec.Cmd
}

// Command wraps exec.Cmd.Command.
func (e *ExecSpawnFunc) Command(name string, arg ...string) *SpawnFunc {
	cmd := exec.Command(name, arg...)
	execSpawnFunc := &ExecSpawnFunc{cmd: cmd}
	var spawnFunc SpawnFunc = execSpawnFunc
	return &spawnFunc
}

// Wait wraps exec.Cmd.Wait.
func (e *ExecSpawnFunc) Wait() error {
	return e.cmd.Wait()
}

// Start wraps exec.Cmd.Start.
func (e *ExecSpawnFunc) Start() error {
	return e.cmd.Start()
}

// StdinPipe wraps exec.Cmd.StdinPipe
func (e *ExecSpawnFunc) StdinPipe() (io.WriteCloser, error) {
	return e.cmd.StdinPipe()
}

// StdoutPipe wraps exec.Cmd.Stdoutpipe
func (e *ExecSpawnFunc) StdoutPipe() (io.Reader, error) {
	return e.cmd.StdoutPipe()
}

// Spawner provides an interface for creating interactive sessions such as oc, ssh, or shell.
type Spawner interface {
	// Spawn creates the interactive session.
	Spawn(command string, args []string, timeout time.Duration, opts ...Option) (*Context, error)
}

// Context represents an interactive context.  This abstraction is meant to be overloaded, and can represent
// something as simple as a shell, to as complex as an interactive OpenShift client or SSH session.  Context follows the
// Container design pattern, and is a simple data transfer object.
type Context struct {
	expecter     *expect.Expecter
	errorChannel <-chan error
}

// GetExpecter returns the expect.Expecter Context.
func (c *Context) GetExpecter() *expect.Expecter {
	return c.expecter
}

// GetErrorChannel returns the error channel.
func (c *Context) GetErrorChannel() <-chan error {
	return c.errorChannel
}

// NewContext creates a Context.
func NewContext(expecter *expect.Expecter, errorChannel <-chan error) *Context {
	return &Context{expecter: expecter, errorChannel: errorChannel}
}

// GoExpectSpawner provides an implementation of a Spawner based on GoExpect.  This was abstracted for testing purposes.
// Creation through struct initialization is prohibited;  use NewGoExpectSpawner instead.
type GoExpectSpawner struct {
	// bufferSizeIsSet tracks whether the bufferSize option is set.
	bufferSizeIsSet bool
	// bufferSize is the size of the receive buffer in bytes.
	bufferSize int

	// environmentSettingsIsSet tracks whether the environmentSettings option is set.
	environmentSettingsIsSet bool
	// environmentSettings sets environment settings within the Expecter shell.
	environmentSettings []string

	// verboseIsSet tracks whether the verbose option is set.
	verboseIsSet bool
	// verbose controls verbose output.
	verbose bool

	// verboseWriterIsSet tracks whether the verboseWriter option is set.
	verboseWriterIsSet bool
	// verboseWriter is an alternate destination for verbose logs.
	verboseWriter io.Writer
}

// Option is a function pointer to enable lightweight optionals for GoExpectSpawner.
type Option func(spawner *GoExpectSpawner) Option

// BufferSize sets the size of receive buffer in bytes.
func BufferSize(bufferSize int) Option {
	return func(g *GoExpectSpawner) Option {
		g.bufferSizeIsSet = true
		prev := g.bufferSize
		g.bufferSize = bufferSize
		return BufferSize(prev)
	}
}

// SetEnv sets the environmental variables of the spawned process.
func SetEnv(environmentSettings []string) Option {
	return func(g *GoExpectSpawner) Option {
		g.environmentSettingsIsSet = true
		prev := g.environmentSettings
		g.environmentSettings = environmentSettings
		return SetEnv(prev)
	}
}

// Verbose enables/disables verbose logging of matches and sends.
func Verbose(verbose bool) Option {
	return func(g *GoExpectSpawner) Option {
		g.verboseIsSet = true
		prev := g.verbose
		g.verbose = verbose
		return Verbose(prev)
	}
}

// VerboseWriter sets an alternate destination for verbose logs.
func VerboseWriter(verboseWriter io.Writer) Option {
	return func(g *GoExpectSpawner) Option {
		g.verboseWriterIsSet = true
		prev := g.verboseWriter
		g.verboseWriter = verboseWriter
		return VerboseWriter(prev)
	}
}

// getDefaultBufferSize returns the default buffer size as sourced from TNF_DEFAULT_BUFFER_SIZE.  If
// TNF_DEFAULT_BUFFER_SIZE is not set or cannot be parsed as an integer, defaultBufferSize is returned.
func getDefaultBufferSize() int {
	bufferSizeFromEnv := os.Getenv(defaultBufferSizeEnvironmentVariableKey)
	if bufferSizeFromEnv != "" {
		if bufferSize, err := strconv.Atoi(bufferSizeFromEnv); err == nil {
			log.Debugf("Utilizing buffer size as sourced from %s: %dB", defaultBufferSizeEnvironmentVariableKey, bufferSize)
			return bufferSize
		}
	}
	log.Debugf("Utilizing the default buffer size: %d", defaultBufferSize)
	return defaultBufferSize
}

// GetGoExpectOptions renders the GoExpectSpawner Option(s) as expect.Option(s).
func (g *GoExpectSpawner) GetGoExpectOptions() []expect.Option {
	opts := make([]expect.Option, 0)

	// Use BufferSize if supplied.  Otherwise, use the test-network-function default.
	if g.bufferSizeIsSet {
		opts = append(opts, expect.BufferSize(g.bufferSize))
	} else {
		opts = append(opts, expect.BufferSize(getDefaultBufferSize()))
	}

	if g.environmentSettingsIsSet {
		opts = append(opts, expect.SetEnv(g.environmentSettings))
	}

	if g.verboseIsSet {
		opts = append(opts, expect.Verbose(g.verbose))
	}

	if g.verboseWriterIsSet {
		opts = append(opts, expect.VerboseWriter(g.verboseWriter))
	}

	return opts
}

// NewGoExpectSpawner creates a new GoExpectSpawner.
func NewGoExpectSpawner() *GoExpectSpawner {
	return &GoExpectSpawner{}
}

// Spawn creates a subprocess, setting standard input and standard output appropriately.  This is the base method to
// create any interactive PTY based process.
func (g *GoExpectSpawner) Spawn(command string, args []string, timeout time.Duration, opts ...Option) (*Context, error) {
	if !UnitTestMode {
		execSpawnFunc := &ExecSpawnFunc{}
		var transitionSpawnFunc SpawnFunc = execSpawnFunc
		spawnFunc = &transitionSpawnFunc
	}

	for _, opt := range opts {
		opt(g)
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
	return g.spawnGeneric(spawnFunc, stdinPipe, stdoutPipe, timeout, g.GetGoExpectOptions()...)
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

// CreateGoExpectSpawner creates a GoExpectSpawner implementation and returns it as a *Spawner for type compatibility
// reasons.
func CreateGoExpectSpawner() *Spawner {
	goExpectSpawner := NewGoExpectSpawner()
	var spawner Spawner = goExpectSpawner
	return &spawner
}
