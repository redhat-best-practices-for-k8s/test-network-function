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
	"fmt"
	"regexp"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/common"
)

const (
	ocClientCommandSeparator = "--"
	ocCommand                = "oc"
	ocContainerArg           = "-c"
	ocDefaultShell           = "sh"
	ocExecCommand            = "exec"
	ocNamespaceArg           = "-n"
	ocInteractiveArg         = "-i"
	ocNodeArg                = "node"
)

// Oc provides an OpenShift Client designed to wrap the "oc" CLI.
type Oc struct {
	// id of the pod or the node
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
	ocArgs := []string{ocExecCommand, ocNamespaceArg, namespace, ocInteractiveArg, pod, ocContainerArg, container, ocClientCommandSeparator, ocDefaultShell}
	context, err := (*spawner).Spawn(ocCommand, ocArgs, timeout, opts...)
	if err != nil {
		return nil, context.GetErrorChannel(), err
	}
	errorChannel := context.GetErrorChannel()
	return &Oc{pod: pod, container: container, namespace: namespace, timeout: timeout, opts: opts, expecter: context.GetExpecter(), spawnErr: err, errorChannel: errorChannel, doneChannel: make(chan bool)}, errorChannel, nil
}

func SpawnNodeOc(spawner *Spawner, node, namespace string, timeout time.Duration, opts ...Option) (*Oc, <-chan error, error) {
	ocArgs := []string{common.GetDebugCommand(), ocNamespaceArg, namespace, fmt.Sprintf("%s/%s", ocNodeArg, node)}
	log.Info("spawn shell for node ", node, "using ", strings.Join(ocArgs, " "))
	context, err := (*spawner).Spawn(ocCommand, ocArgs, timeout, opts...)
	if err != nil {
		return nil, context.GetErrorChannel(), err
	}
	errorChannel := context.GetErrorChannel()
	podName := makeSimpleName(node) + "-debug"
	return &Oc{pod: podName, container: "container-00", namespace: namespace, timeout: timeout, opts: opts, expecter: context.GetExpecter(), spawnErr: err, errorChannel: errorChannel, doneChannel: make(chan bool)}, errorChannel, nil
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

// makeSimpleName is near-verbatime copy of the function used to generate debug pod
// link https://github.com/openshift/oc/blob/ecccc93cf5f51aa851d15aeffef1bda486169600/pkg/helpers/newapp/app/pipeline.go#L329
func makeSimpleName(name string) string {
	var invalidServiceChars = regexp.MustCompile("[^-a-z0-9]")
	const DNS1035LabelMaxLength int = 63
	name = strings.ToLower(name)
	name = invalidServiceChars.ReplaceAllString(name, "")
	name = strings.TrimFunc(name, func(r rune) bool { return r == '-' })
	if len(name) > DNS1035LabelMaxLength {
		name = name[:DNS1035LabelMaxLength]
	}
	return name
}
