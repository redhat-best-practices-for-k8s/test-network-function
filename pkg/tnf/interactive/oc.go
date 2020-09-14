package interactive

import (
	expect "github.com/google/goexpect"
	"time"
)

const (
	ocClientCommandSeparator = "--"
	ocCommand                = "oc"
	ocContainerArg           = "-c"
	ocDefaultShell           = "sh"
	ocExecCommand            = "exec"
	ocNamespaceArg           = "-n"
	ocInteractiveArg         = "-it"
)

// An OpenShift Client designed to wrap the "oc" CLI.
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

// Creates an OpenShift Client subprocess, spawning the appropriate underlying PTY.
func SpawnOc(spawner *Spawner, pod, container, namespace string, timeout time.Duration, opts ...expect.Option) (*Oc, <-chan error, error) {
	ocArgs := []string{ocExecCommand, ocNamespaceArg, namespace, ocInteractiveArg, pod, ocContainerArg, container, ocClientCommandSeparator, ocDefaultShell}
	context, err := (*spawner).Spawn(ocCommand, ocArgs, timeout, opts...)
	if err != nil {
		return nil, context.GetErrorChannel(), err
	}
	errorChannel := context.GetErrorChannel()
	return &Oc{pod: pod, container: container, namespace: namespace, timeout: timeout, opts: opts, expecter: context.GetExpecter(), spawnErr: err, errorChannel: errorChannel}, errorChannel, nil
}

// Extract the expect.Expecter reference used to control the OpenShift client.
func (o *Oc) GetExpecter() *expect.Expecter {
	return o.expecter
}

// Extract the name of the pod.
func (o *Oc) GetPodName() string {
	return o.pod
}

// Extract the name of the container.
func (o *Oc) GetPodContainerName() string {
	return o.container
}

// Extract the namespace of the pod.
func (o *Oc) GetPodNamespace() string {
	return o.namespace
}

// Extract the timeout for the expect.Expecter.
func (o *Oc) GetTimeout() time.Duration {
	return o.timeout
}

// Extract the options, such as verbosity.
func (o *Oc) GetOptions() []expect.Option {
	return o.opts
}

// Extract the error channel for interactive monitoring.
func (o *Oc) GetErrorChannel() <-chan error {
	return o.errorChannel
}
