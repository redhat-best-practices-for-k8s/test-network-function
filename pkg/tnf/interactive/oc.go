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
	"time"

	expect "github.com/google/goexpect"
	log "github.com/sirupsen/logrus"
)

const (
	ocCommand      = "oc"
	ocContainerArg = "-c"
	ocRsh          = "rsh"
	ocNamespaceArg = "-n"
)

// Oc provides an OpenShift Client designed to wrap the "oc" CLI.
type Oc struct {
	// name of the pod or the node
	pod string
	// node set to true means the sessions is node session
	// name of the container
	container string
	// namespace of the pod
	namespace string
	// serviceAccountName of the pod
	serviceAccountName string
	// timeout for commands run in expecter
	timeout time.Duration
	// options for expecter, such as expect.Verbose(true)
	opts []Option
	// the underlying subprocess implementation, tailored to OpenShift Client
	expecter *expect.Expecter
	// error during the spawn process
	spawnErr error
	// error channel for interactive error stream
	errorChannel <-chan error
	// done channel to notify the go routine that monitors the error channel
	doneChannel chan bool
}

// SpawnOc creates an OpenShift Client subprocess, spawning the appropriate underlying PTY.
func SpawnOc(spawner *Spawner, pod, container, namespace string, timeout time.Duration, opts ...Option) (*Oc, <-chan error, error) {
	ocArgs := []string{ocRsh, ocNamespaceArg, namespace, ocContainerArg, container, pod}
	context, err := (*spawner).Spawn(ocCommand, ocArgs, timeout, opts...)
	if err != nil {
		return nil, context.GetErrorChannel(), err
	}
	errorChannel := context.GetErrorChannel()
	return &Oc{pod: pod, container: container, namespace: namespace, timeout: timeout, opts: opts, expecter: context.GetExpecter(), spawnErr: err, errorChannel: errorChannel, doneChannel: make(chan bool)}, errorChannel, nil
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

// GetServiceAccountName extracts the serviceAccountName of the pod
func (o *Oc) GetServiceAccountName() string {
	return o.serviceAccountName
}

// SetServiceAccountName sets the serviceAccountName of the pod
func (o *Oc) SetServiceAccountName(serviceAccountName string) {
	o.serviceAccountName = serviceAccountName
}

// GetTimeout returns the timeout for the expect.Expecter.
func (o *Oc) GetTimeout() time.Duration {
	return o.timeout
}

// GetOptions returns the options, such as verbosity.
func (o *Oc) GetOptions() []Option {
	return o.opts
}

// GetErrorChannel returns the error channel for interactive monitoring.
func (o *Oc) GetErrorChannel() <-chan error {
	return o.errorChannel
}

// GetDoneChannel returns the receive only done channel
func (o *Oc) GetDoneChannel() <-chan bool {
	log.Debugf("read done channel pod %s/%s", o.pod, o.container)
	return o.doneChannel
}

// Close sends the signal to the done channel
func (o *Oc) Close() {
	log.Debugf("send close to channel pod %s/%s ", o.pod, o.container)
	o.doneChannel <- true
}
