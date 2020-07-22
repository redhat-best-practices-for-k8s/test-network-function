package oc

import "github.com/redhat-nfvpe/test-network-function/internal/subprocess"

const (
	containerCommandSeparator = "--"
	ocExecCommand             = "exec"
	ocExecContainerArg        = "-c"
	ocCommand                 = "oc"
	ocNamespaceArg            = "-n"
)

// Lightweight wrapper client around oc client.
func InvokeOCCommand(pod string, container string, namespace string, command []string) (string, error) {
	invokeCommandArgs := []string{ocExecCommand, pod, ocExecContainerArg, container}
	if namespace != "" {
		invokeCommandArgs = append(invokeCommandArgs, ocNamespaceArg, namespace)
	}
	invokeCommandArgs = append(invokeCommandArgs, containerCommandSeparator)
	invokeCommandArgs = append(invokeCommandArgs, command...)
	return subprocess.InvokeCommand(ocCommand, invokeCommandArgs)
}
