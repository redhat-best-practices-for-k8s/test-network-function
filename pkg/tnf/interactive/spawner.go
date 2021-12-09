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
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultBufferSize is the size of the input/output buffers in bytes.
	defaultBufferSize = 32768
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

	// Close calls the exec.Cmd.Kill to stop the process (shell).
	Close() error

	// StdinPipe consult exec.Cmd.StdinPipe
	StdinPipe() (io.WriteCloser, error)

	// StdoutPipe consult exec.Cmd.StdoutPipe
	StdoutPipe() (io.Reader, error)

	// StderrPipe consult exec.Cmd.StderrPipe
	StderrPipe() (io.Reader, error)

	// Wait consult exec.Cmd.Wait
	Wait() error

	// IsRunning returns true if the shell hasn't exited yet.
	IsRunning() bool

	// Args returns the command and arguments used to spawn the shell.
	Args() []string
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

// IsRunning returns true if e.Cmd.ProcessState is nil, false otherwise
func (e *ExecSpawnFunc) IsRunning() bool {
	return e.cmd.ProcessState == nil
}

// Args wraps e.Cmd.Args
func (e *ExecSpawnFunc) Args() []string {
	return e.cmd.Args
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

// StderrPipe wraps exec.Cmd.Stderrpipe
func (e *ExecSpawnFunc) StderrPipe() (io.Reader, error) {
	return e.cmd.StderrPipe()
}

// Close wraps exec.Cmd.Kill.
func (e *ExecSpawnFunc) Close() error {
	return e.cmd.Process.Kill()
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

	// sendTimeoutIsSet tracks whether the Send command timeout is set.
	sendTimeoutIsSet bool
	// sendTimeout is the timeout of send command
	sendTimeout time.Duration
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

// SendTimeout sets the timeout of send command
func SendTimeout(timeout time.Duration) Option {
	return func(g *GoExpectSpawner) Option {
		g.sendTimeoutIsSet = true
		prev := g.sendTimeout
		g.sendTimeout = timeout
		return SendTimeout(prev)
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

	if g.sendTimeoutIsSet {
		opts = append(opts, expect.SendTimeout(g.sendTimeout))
	}

	return opts
}

// NewGoExpectSpawner creates a new GoExpectSpawner.
func NewGoExpectSpawner() *GoExpectSpawner {
	return &GoExpectSpawner{}
}

// logCmdMirrorPipe logs specified pipe output to logger.
func logCmdMirrorPipe(cmdLine string, pipeToMirror io.Reader, name string, trace bool) io.Reader {
	originalPipe := pipeToMirror
	r, w, _ := os.Pipe()
	tr := io.TeeReader(originalPipe, w)

	log.Debugf("Creating %s mirror pipe for cmd: %s", name, cmdLine)
	go func() {
		buf := bufio.NewReader(tr)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				if err == io.EOF {
					// Graceful exit: no more to read here.
					log.Debugf("Stoping %s mirroring pipe for cmd: %s (EOF)", name, cmdLine)
				} else {
					// Some Error has happened, report it as warning.
					log.Warnf("Exiting %s log mirroring goroutine for cmd %s. Error: %s", name, cmdLine, err)
				}
				return
			}

			if trace {
				log.Trace(name + " for " + cmdLine + " : " + string(line))
			} else {
				log.Warn(name + " for " + cmdLine + " : " + string(line))
			}
		}
	}()
	return r
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
	stdinPipe, stdoutPipe, stderrPipe, err := g.unpackPipes(spawnFunc)
	if err != nil {
		return nil, err
	}

	cmdLine := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	log.Debugf("Spawning interactive shell. Cmd: %s", cmdLine)

	logCmdMirrorPipe(cmdLine, stderrPipe, "STDERR", false)
	stdoutPipe = logCmdMirrorPipe(cmdLine, stdoutPipe, "STDOUT", true)

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
			log.Debug("Killing shell cmd: " + strings.Join((*spawnFunc).Args(), " "))
			return (*spawnFunc).Close()
		},
		Check: func() bool {
			if !(*spawnFunc).IsRunning() {
				log.Error("Unable to send commands to spawned shell. Shell cmd: " + strings.Join((*spawnFunc).Args(), " "))
				return false
			}
			return true
		},
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

//nolint:gocritic // Helper method to unpack stdin and stdout.
func (g *GoExpectSpawner) unpackPipes(spawnFunc *SpawnFunc) (io.WriteCloser, io.Reader, io.Reader, error) {
	stdinPipe, err := g.extractStdinPipe(spawnFunc)
	if err != nil {
		return nil, nil, nil, err
	}
	stdoutPipe, err := g.extractStdoutPipe(spawnFunc)
	if err != nil {
		return nil, nil, nil, err
	}
	stderrPipe, err := g.extractStderrPipe(spawnFunc)
	if err != nil {
		return nil, nil, nil, err
	}
	return stdinPipe, stdoutPipe, stderrPipe, err
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

// Helper method to extract stdout.
func (g *GoExpectSpawner) extractStderrPipe(spawnFunc *SpawnFunc) (io.Reader, error) {
	stderr, err := (*spawnFunc).StderrPipe()
	if err != nil {
		log.Errorf("Couldn't extract stderr for the given process: %v", err)
	}
	return stderr, err
}

// CreateGoExpectSpawner creates a GoExpectSpawner implementation and returns it as a *Spawner for type compatibility
// reasons.
func CreateGoExpectSpawner() *Spawner {
	goExpectSpawner := NewGoExpectSpawner()
	var spawner Spawner = goExpectSpawner
	return &spawner
}
