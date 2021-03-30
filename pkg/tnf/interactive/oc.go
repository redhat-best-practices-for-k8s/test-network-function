// Copyright (C) 2020 Red Hat, Inc.
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
	"time"

	expect "github.com/ryandgoulding/goexpect"
)

const (
	ocClientCommandSeparator = "--"
	ocCommand                = "oc"
	ocContainerArg           = "-c"
	ocDefaultShell           = "sh"
	ocExecCommand            = "exec"
	ocNamespaceArg           = "-n"
	ocInteractiveArg         = "-i"
)

// Oc provides an OpenShift Client designed to wrap the "oc" CLI.
type Oc struct {
	// name of the pod
	pod string
	// name of the container
	container string
	// namespace of the pod
	namespace string
	// timeout for commands run in expecter
	timeout time.Duration
	// options for experter, such as expect.Verbose(true)
	opts []expect.Option
	// the underlying subprocess implementation, tailored to OpenShift Client
	expecter *expect.Expecter
	// error during the spawn process
	spawnErr error
	// error channel for interactive error stream
	errorChannel <-chan error
}

// SpawnOc creates an OpenShift Client subprocess, spawning the appropriate underlying PTY.
func SpawnOc(spawner *Spawner, pod, container, namespace string, timeout time.Duration, opts ...expect.Option) (*Oc, <-chan error, error) {
	ocArgs := []string{ocExecCommand, ocNamespaceArg, namespace, ocInteractiveArg, pod, ocContainerArg, container, ocClientCommandSeparator, ocDefaultShell}
	context, err := (*spawner).Spawn(ocCommand, ocArgs, timeout, opts...)
	if err != nil {
		return nil, context.GetErrorChannel(), err
	}
	errorChannel := context.GetErrorChannel()
	return &Oc{pod: pod, container: container, namespace: namespace, timeout: timeout, opts: opts, expecter: context.GetExpecter(), spawnErr: err, errorChannel: errorChannel}, errorChannel, nil
}

// GetExpecter returns a reference to the expect.Expecter reference used to control the OpenShift client.
func (o *Oc) GetExpecter() *expect.Expecter {
	return o.expecter
}

// GetPodName returns the name of the pod.
func (o *Oc) GetPodName() string {
	return o.pod
}

// GetPodContainerName returns the name of the container.
func (o *Oc) GetPodContainerName() string {
	return o.container
}

// GetPodNamespace extracts the namespace of the pod.
func (o *Oc) GetPodNamespace() string {
	return o.namespace
}

// GetTimeout returns the timeout for the expect.Expecter.
func (o *Oc) GetTimeout() time.Duration {
	return o.timeout
}

// GetOptions returns the options, such as verbosity.
func (o *Oc) GetOptions() []expect.Option {
	return o.opts
}

// GetErrorChannel returns the error channel for interactive monitoring.
func (o *Oc) GetErrorChannel() <-chan error {
	return o.errorChannel
}
